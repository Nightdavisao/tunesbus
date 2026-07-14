//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/wine"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

const BogusTrackID = "/org/mpris/MediaPlayer2/Track/0"

type Root struct {
	dispatcher *ole.IDispatch
}

func (r Root) Raise() error {
	log.Debug("Raise is not implemented")
	return nil
}

func (r Root) Quit() error {
	_, err := oleutil.CallMethod(r.dispatcher, "Quit")
	return err
}

func (r Root) CanQuit() (bool, error) {
	return true, nil
}

func (r Root) CanRaise() (bool, error) {
	log.Debug("CanRaise is not implemented")
	return false, nil
}

func (r Root) HasTrackList() (bool, error) {
	log.Debug("HasTrackList is not implemented")
	return false, nil
}

func (r Root) Identity() (string, error) {
	return "iTunes", nil
}

func (r Root) SupportedUriSchemes() ([]string, error) {
	return []string{}, nil
}

func (r Root) SupportedMimeTypes() ([]string, error) {
	return []string{}, nil
}

type Player struct {
	state      *State
	dispatcher *ole.IDispatch
}

func (m *Player) Next() error {
	_, err := oleutil.CallMethod(m.dispatcher, "NextTrack")
	return err
}

func (m *Player) Previous() error {
	_, err := oleutil.CallMethod(m.dispatcher, "PreviousTrack")
	return err
}

func (m *Player) Pause() error {
	_, err := oleutil.CallMethod(m.dispatcher, "Pause")
	return err
}

func (m *Player) PlayPause() error {
	_, err := oleutil.CallMethod(m.dispatcher, "PlayPause")
	return err
}

func (m *Player) Stop() error {
	_, err := oleutil.CallMethod(m.dispatcher, "Stop")
	return err
}

func (m *Player) Play() error {
	_, err := oleutil.CallMethod(m.dispatcher, "Play")
	return err
}

func (m *Player) Seek(offset types.Microseconds) error {
	return nil
}

func (m *Player) SetPosition(trackId string, position types.Microseconds) error {
	seconds := (time.Duration(position) * time.Microsecond) / time.Second
	err := itunes.SetTunesPosition(m.dispatcher, int64(seconds))
	return err
}

func (m *Player) OpenUri(uri string) error {
	return nil
}

func (m *Player) PlaybackStatus() (types.PlaybackStatus, error) {
	tunes, err := itunes.GetCurrentTunes(m.dispatcher)
	mprisState := types.PlaybackStatusPaused
	if tunes != nil {
		if tunes.PlayerState == itunes.ITPlayerStatePlaying {
			mprisState = types.PlaybackStatusPlaying
		}
	}
	return mprisState, err
}

func (m *Player) Rate() (float64, error) {
	return 0, nil
}

func (m *Player) SetRate(rate float64) error {
	return nil
}

func (m *Player) Metadata() (types.Metadata, error) {
	return *m.state.currentMetadata, nil
}

func (m *Player) Volume() (float64, error) {
	return float64(m.state.currentVolume / 100), nil
}

func (m *Player) SetVolume(volume float64) error {
	_, err := oleutil.PutProperty(m.dispatcher, "SoundVolume", volume*100)
	return err
}

func (m *Player) Position() (int64, error) {
	return int64(m.state.currentPosition), nil
}

func (m *Player) MinimumRate() (float64, error) {
	return 1.0, nil
}

func (m *Player) MaximumRate() (float64, error) {
	return 1.0, nil
}

func (m *Player) CanGoNext() (bool, error) {
	_, _, nextEnabled, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return nextEnabled, err
}

func (m *Player) CanGoPrevious() (bool, error) {
	previousEnabled, _, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return previousEnabled, err
}

func (m *Player) CanPlay() (bool, error) {
	_, buttonState, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return buttonState != itunes.ITPlayButtonStatePauseDisabled &&
		buttonState != itunes.ITPlayButtonStatePlayDisabled, err
}

func (m *Player) CanPause() (bool, error) {
	_, buttonState, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return buttonState != itunes.ITPlayButtonStatePauseDisabled &&
		buttonState != itunes.ITPlayButtonStatePlayDisabled, err
}

func (m *Player) CanSeek() (bool, error) {
	return true, nil
}

func (m *Player) CanControl() (bool, error) {
	return true, nil
}

func (m *Player) Shuffle() (bool, error) {
	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.dispatcher)
	if err != nil {
		log.Error("failed to get current playlist on getting Shuffle", err)
		return false, nil
	}

	if playlistDispatcher == nil {
		log.Debug("no playlist yet")
		return false, nil
	}
	defer playlistDispatcher.Release()

	shuffleStatus, err := oleutil.GetProperty(playlistDispatcher, "Shuffle")
	if err != nil {
		log.Error("failed to get shuffle status", err)
		return false, err
	}
	defer shuffleStatus.Clear()
	return shuffleStatus.Value().(bool), nil
}

func (m *Player) SetShuffle(shuffle bool) error {
	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.dispatcher)
	if err != nil {
		log.Error("failed to get current playlist on setting Shuffle", err)
		return nil
	}

	if playlistDispatcher == nil {
		log.Debug("no playlist yet")
		return nil
	}
	defer playlistDispatcher.Release()

	_, err = oleutil.PutProperty(playlistDispatcher, "Shuffle", shuffle)
	if err != nil {
		log.Error("failed to put shuffle status", err)
		return err
	}

	return err
}

func (m *Player) LoopStatus() (types.LoopStatus, error) {
	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.dispatcher)
	if err != nil {
		log.Error("failed to get current playlist on getting Loop", err)
		return types.LoopStatusNone, err
	}

	if playlistDispatcher == nil {
		log.Debug("no playlist yet")
		return types.LoopStatusNone, nil
	}
	defer playlistDispatcher.Release()

	v, err := oleutil.GetProperty(playlistDispatcher, "SongRepeat")
	if err != nil {
		return types.LoopStatusNone, err
	}
	defer v.Clear()
	log.Debug("loop status", v)

	// ITPlayerRepeatMode: 0 = Off, 1 = One (repeat song), 2 = All (repeat playlist)
	switch v.Val {
	case 1:
		return types.LoopStatusTrack, nil
	case 2:
		return types.LoopStatusPlaylist, nil
	default:
		return types.LoopStatusNone, nil
	}
}

func (m *Player) SetLoopStatus(status types.LoopStatus) error {
	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.dispatcher)
	if err != nil {
		log.Error("failed to get current playlist on setting Loop", err)
		return err
	}
	if playlistDispatcher == nil {
		log.Debug("no playlist yet")
		return nil
	}
	defer playlistDispatcher.Release()

	var mode int32
	switch status {
	case types.LoopStatusTrack:
		mode = 1
	case types.LoopStatusPlaylist:
		mode = 2
	default:
		mode = 0
	}
	_, err = oleutil.PutProperty(playlistDispatcher, "SongRepeat", mode)
	return err
}

type State struct {
	currentMetadata  *types.Metadata
	currentVolume    int64
	currentPosition  int64
	server           *server.Server
	hasServerStarted bool
	ticker           *time.Ticker
	done             chan struct{}
}

type eventHandler struct {
	state      *State
	dispatcher *ole.IDispatch
	player     *Player
	handler    *events.EventHandler

	QuitCalled        bool
	AboutToQuitCalled bool
}

type fn func()

func setInitialMetadata(track *itunes.IiTrack, state *State, afterSetting fn) {
	if track != nil {
		*state.currentMetadata = types.Metadata{
			TrackId:     dbus.ObjectPath(fmt.Sprintf("/org/mpris/MediaPlayer2/Track/%d", track.TrackID)),
			Album:       track.Album,
			Title:       track.Name,
			Artist:      []string{track.Artist},
			Length:      types.Microseconds(secondsToMicro(track.Duration)),
			DiscNumber:  int(track.DiscNumber),
			TrackNumber: int(track.TrackNumber),
		}
		if track.Dispatcher != nil {
			go func(state *State, track *itunes.IiTrack, afterSetting fn) {
				defer track.Dispatcher.Release()
				dosFilename, err := itunes.SaveArtworkIfAvaliable(track.Dispatcher, track)
				if err != nil {
					log.Printf("failed to retrieve artwork for current track: %v", err)
					if afterSetting != nil {
						afterSetting()
					}
					return
				}

				log.Debug("dos filename for artwork: %s", dosFilename)

				unixFilename, err := wine.GetUnixFilename(dosFilename)
				if err != nil {
					log.Printf("failed to retrieve unix filename for saved artwork: %v", err)
					if afterSetting != nil {
						afterSetting()
					}
					return
				}

				log.Debug("unix filename for artwork: %s", unixFilename)

				state.currentMetadata.ArtUrl = "file://" + unixFilename
				if afterSetting != nil {
					afterSetting()
				}
			}(state, track, afterSetting)
		}

		log.Debug("successfully set new metadata: %v", state.currentMetadata)
		return
	}

	// send only the bogus trackid if we don't have anything to begin with (stops godbus/dbus from spamming the console)
	// might not be needed anymore since we are only starting the mpris server on "demand"
	state.currentMetadata = &types.Metadata{
		TrackId: dbus.ObjectPath(BogusTrackID),
	}
}

func setPosition(tunes *itunes.IiTunes, state *State) {
	if tunes != nil {
		state.currentPosition = milliToMicro(int64(tunes.PlayerPositionMS))
	}
}

func secondsToMicro(seconds int64) int64 {
	duration := time.Duration(seconds) * time.Second
	return duration.Microseconds()
}

func milliToMicro(milli int64) int64 {
	duration := time.Duration(milli) * time.Microsecond
	return duration.Microseconds()
}

func (m *eventHandler) OnPlayerPlayEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerPlayEvent %v", t)
	setInitialMetadata(t, m.state, nil)
	if !m.state.hasServerStarted {
		go startMprisServer(m.state.server)
		m.state.hasServerStarted = true
	}
}

func (m *eventHandler) OnPlayerStopEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerStopEvent", t)
	setInitialMetadata(t, m.state, func() {
		m.handler.Player.OnPlayback()
		m.handler.Player.OnPlayPause()
	})
}

func (m *eventHandler) OnPlayerPlayingTrackChangedEvent(t *itunes.IiTrack) {
	log.Printf("OnPlayerPlayingTrackChangedEvent: %v", t)
	setInitialMetadata(t, m.state, func() {
		m.handler.Player.OnPlayback()
	})
}

func (m *eventHandler) OnQuittingEvent() {
	log.Debug("received OnQuittingEvent")
	m.QuitCalled = true
	os.Exit(0)
}

func (m *eventHandler) OnAboutToPromptUserToQuitEvent() {
	log.Debug("received OnAboutToPromptUserToQuitEvent")
	m.AboutToQuitCalled = true
	m.dispatcher.Release()
	m.state.done <- struct{}{}
	// todo: 20seg~ timer to reconnect everything if that dialog happens to show up and the user clicks "Don't Quit"
}

func (m *eventHandler) OnSoundVolumeChangedEvent(val *int64) {
	log.Debug("received OnSoundVolumeChangedEvent", *val)
	m.state.currentVolume = *val
	m.handler.Player.OnVolume()
}

func startMprisServer(s *server.Server) {
	log.Info("starting MPRIS server...")
	err := s.Listen()

	if err != nil {
		log.Error("startMprisServer failed, quitting", err)
		os.Exit(1)
		return
	}
}

func main() {
	debugModePtr := flag.Bool("debug", false, "Enable debug logging")
	identityPtr := flag.String("identity", "iTunes", "Custom identity for the MPRIS server\n"+
		"Tip: Set this to \"cider\" in all lowercase (or use some other whitelisted identity) if you want to make Music Presence pick up the player.")
	flag.Parse()

	if *debugModePtr {
		log.SetLevel(log.DebugLevel)
	}

	sigtermChan := make(chan os.Signal)
	signal.Notify(sigtermChan, syscall.SIGTERM)

	state := &State{
		ticker:           time.NewTicker(50 * time.Millisecond),
		hasServerStarted: false,
		currentMetadata:  &types.Metadata{},
		done:             make(chan struct{}),
	}
	dispatcher, err := itunes.NewTunesDispatch()
	if err != nil {
		log.Error("failed to initialize dispatcher", err)
		os.Exit(67)
		return
	}

	go func() {
		<-sigtermChan
		dispatcher.Release()
		os.Exit(1)
	}()

	root := Root{
		dispatcher: dispatcher,
	}

	player := Player{
		dispatcher: dispatcher,
		state:      state,
	}

	srv := server.NewServer(*identityPtr, root, &player)
	ev := events.NewEventHandler(srv)
	state.server = srv

	handler := &eventHandler{
		state:      state,
		handler:    ev,
		dispatcher: dispatcher,
	}

	sink, err := itunes.NewCOMEventSink(dispatcher, handler)
	if err != nil {
		log.Error("something failed when setting up the event sink", err)
		os.Exit(69)
		return
	}

	curr, _ := itunes.GetCurrentTrack(dispatcher)
	log.Debug("current track", curr)
	if curr != nil {
		setInitialMetadata(curr, state, func() {
			ev.Player.OnAll()
		})
		// workaround the issue where KDE might *not* pick up our player
		if !state.hasServerStarted {
			go startMprisServer(srv)
			state.hasServerStarted = true
		}
	} else {
		setInitialMetadata(nil, state, nil)
	}

	go func() {
		for {
			select {
			case <-state.done:
				return
			case <-state.ticker.C:
				if player.dispatcher != nil {
					tunes, _ := itunes.GetCurrentTunes(player.dispatcher)
					if tunes != nil {
						if tunes.PlayerPositionMS > 0 {
							position := time.Duration(tunes.PlayerPositionMS) * time.Millisecond
							state.currentPosition = position.Microseconds()
							ev.Player.OnAll()
						}
					}
				}
			}
		}
	}()

	err = sink.ListenEvents(state.done)
	if err != nil {
		log.Debug("failed to listen for COM events", err)
	}
}
