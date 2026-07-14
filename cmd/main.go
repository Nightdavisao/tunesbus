//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/wine"

	"github.com/ammario/weakmap"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

const BogusTrackID = "/org/mpris/MediaPlayer2/Track/0"

type BusRoot struct {
	state *State
}

func (r BusRoot) Raise() error {
	log.Debug("Raise is not implemented")
	return nil
}

func (r BusRoot) Quit() error {
	if r.state.tunesDispatcher != nil {
		r, err := oleutil.CallMethod(r.state.tunesDispatcher, "Quit")
		r.Clear()
		return err
	}
	r.state.QuitSafely(nil, "")
	return nil
}

func (r BusRoot) CanQuit() (bool, error) {
	return true, nil
}

func (r BusRoot) CanRaise() (bool, error) {
	log.Debug("CanRaise is not implemented")
	return false, nil
}

func (r BusRoot) HasTrackList() (bool, error) {
	log.Debug("HasTrackList is not implemented")
	return false, nil
}

func (r BusRoot) Identity() (string, error) {
	return "iTunes", nil
}

func (r BusRoot) SupportedUriSchemes() ([]string, error) {
	return []string{}, nil
}

func (r BusRoot) SupportedMimeTypes() ([]string, error) {
	return []string{}, nil
}

type BusPlayer struct {
	state      *State
	dispatcher *ole.IDispatch
}

func (m *BusPlayer) Next() error {
	r, err := oleutil.CallMethod(m.dispatcher, "NextTrack")
	r.Clear()
	return err
}

func (m *BusPlayer) Previous() error {
	r, err := oleutil.CallMethod(m.dispatcher, "PreviousTrack")
	r.Clear()
	return err
}

func (m *BusPlayer) Pause() error {
	r, err := oleutil.CallMethod(m.dispatcher, "Pause")
	r.Clear()
	return err
}

func (m *BusPlayer) PlayPause() error {
	r, err := oleutil.CallMethod(m.dispatcher, "PlayPause")
	r.Clear()
	return err
}

func (m *BusPlayer) Stop() error {
	r, err := oleutil.CallMethod(m.dispatcher, "Stop")
	r.Clear()
	return err
}

func (m *BusPlayer) Play() error {
	r, err := oleutil.CallMethod(m.dispatcher, "Play")
	r.Clear()
	return err
}

func (m *BusPlayer) Seek(offset types.Microseconds) error {
	return nil
}

func (m *BusPlayer) SetPosition(trackId string, position types.Microseconds) error {
	log.Debug("setting Position", position)

	seconds := (time.Duration(position) * time.Microsecond) / time.Second
	err := itunes.SetTunesPosition(m.dispatcher, int64(seconds))
	return err
}

func (m *BusPlayer) OpenUri(uri string) error {
	return nil
}

func (m *BusPlayer) PlaybackStatus() (types.PlaybackStatus, error) {
	log.Debug("PlaybackStatus called")
	tunes, err := itunes.GetCurrentTunes(m.dispatcher)
	mprisState := types.PlaybackStatusPaused
	if tunes != nil {
		if tunes.PlayerState == itunes.ITPlayerStatePlaying {
			mprisState = types.PlaybackStatusPlaying
		}
	}
	return mprisState, err
}

func (m *BusPlayer) Rate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) SetRate(rate float64) error {
	return nil
}

func (m *BusPlayer) Metadata() (types.Metadata, error) {
	log.Debug("Metadata called", *m.state.currentMetadata)
	return *m.state.currentMetadata, nil
}

func (m *BusPlayer) Volume() (float64, error) {
	log.Debug("Volume called")
	return float64(m.state.currentVolume / 100), nil
}

func (m *BusPlayer) SetVolume(volume float64) error {
	_, err := oleutil.PutProperty(m.dispatcher, "SoundVolume", volume*100)
	return err
}

func (m *BusPlayer) Position() (int64, error) {
	log.Debug("Position called")
	return int64(m.state.currentPosition), nil
}

func (m *BusPlayer) MinimumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) MaximumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) CanGoNext() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	log.Debug("CanGoNext called (expensive call)")
	_, _, nextEnabled, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return nextEnabled, err
}

func (m *BusPlayer) CanGoPrevious() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	log.Debug("CanGoPrevious called (expensive call)")
	previousEnabled, _, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return previousEnabled, err
}

func (m *BusPlayer) CanPlay() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	log.Debug("CanPlay called (expensive call)")
	_, buttonState, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return buttonState != itunes.ITPlayButtonStatePauseDisabled &&
		buttonState != itunes.ITPlayButtonStatePlayDisabled, err
}

func (m *BusPlayer) CanPause() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	_, buttonState, _, err := itunes.GetPlayerButtonsState(m.dispatcher)
	return buttonState != itunes.ITPlayButtonStatePauseDisabled &&
		buttonState != itunes.ITPlayButtonStatePlayDisabled, err
}

func (m *BusPlayer) CanSeek() (bool, error) {
	return true, nil // even though we don't actually support "Seek", we need to advertise that we do, clients will set "Position" anyway
}

func (m *BusPlayer) CanControl() (bool, error) {
	return true, nil
}

func (m *BusPlayer) Shuffle() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

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

func (m *BusPlayer) SetShuffle(shuffle bool) error {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

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

func (m *BusPlayer) LoopStatus() (types.LoopStatus, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

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

	property, err := oleutil.GetProperty(playlistDispatcher, "SongRepeat")
	if err != nil {
		return types.LoopStatusNone, err
	}
	songRepeatVal := property.Value().(int32)
	property.Clear()

	// ITPlayerRepeatMode: 0 = Off, 1 = One (repeat song), 2 = All (repeat playlist)
	switch songRepeatVal {
	case 1:
		return types.LoopStatusTrack, nil
	case 2:
		return types.LoopStatusPlaylist, nil
	default:
		return types.LoopStatusNone, nil
	}
}

func (m *BusPlayer) SetLoopStatus(status types.LoopStatus) error {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

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

type WeakTrackArtworkCache struct {
	store weakmap.Map[int64, string]
}

type Config struct {
	identity *string
}

type State struct {
	config          Config
	tunesDispatcher *ole.IDispatch
	mux             sync.RWMutex
	quitOnce        sync.Once
	currentMetadata *types.Metadata
	artworkCache    WeakTrackArtworkCache
	currentVolume   int64
	currentPosition int64
	server          *server.Server
	mprisHandler    *events.EventHandler
	ticker          *time.Ticker
	quit            chan struct{}
}

type tunesEventHandler struct {
	state      *State
	dispatcher *ole.IDispatch
	handler    *events.EventHandler
}

type fn func()

func ensureValidTrackID(metadata *types.Metadata) {
	if metadata.TrackId == "" {
		metadata.TrackId = dbus.ObjectPath(BogusTrackID)
	}
}

// note that this is already releasing the track's dispatcher object, don't release it yourself after using this
func setPlayerMetadata(track *itunes.IiTrack, state *State) error {
	state.mux.Lock()
	defer state.mux.Unlock()

	if track != nil {
		metadata := types.Metadata{
			TrackId:     dbus.ObjectPath(fmt.Sprintf("/org/mpris/MediaPlayer2/Track/%d", track.TrackID)),
			Album:       track.Album,
			Title:       track.Name,
			Artist:      []string{track.Artist},
			Length:      types.Microseconds(secondsToMicro(track.Duration)),
			DiscNumber:  int(track.DiscNumber),
			TrackNumber: int(track.TrackNumber),
		}
		if track.Dispatcher != nil {
			log.Debug("partial new metadata (not sent yet)", state.currentMetadata)
			defer track.Dispatcher.Release()

			log.Debug(&state.artworkCache.store)
			val, exists := state.artworkCache.store.Get(track.TrackID)
			if exists {
				metadata.ArtUrl = val
				ensureValidTrackID(&metadata)
				*state.currentMetadata = metadata
				log.Debug("will send cached artwork from weak map", track.TrackID, val)
				if state.server.Conn == nil {
					log.Debug("dbus server connection is not ready yet")
					return nil
				}
				return state.mprisHandler.Player.OnTitle()
			}

			// if we don't have the artwork...
			log.Debug("artwork for this track doesn't exist yet", track.TrackID)

			dosFilename, err := itunes.SaveArtworkIfAvaliable(track.Dispatcher, track)
			log.Debug("dos filename for artwork", dosFilename)
			if err != nil {
				log.Error("failed to retrieve artwork for current track", err)
				*state.currentMetadata = metadata
				return state.mprisHandler.Player.OnTitle()
			}

			unixFilename, err := wine.GetUnixFilename(dosFilename)
			if err != nil {
				log.Error("failed to retrieve unix filename for saved artwork", err)
				*state.currentMetadata = metadata
				return state.mprisHandler.Player.OnTitle()
			}
			log.Debug("unix filename for artwork", unixFilename)

			artUrl := "file://" + unixFilename
			state.artworkCache.store.Set(track.TrackID, artUrl)

			metadata.ArtUrl = artUrl
			ensureValidTrackID(&metadata)
			*state.currentMetadata = metadata
			if state.server.Conn == nil {
				log.Debug("dbus server connection is not ready yet")
				return nil
			}
			return state.mprisHandler.Player.OnTitle()
		}
		return fmt.Errorf("track.Dispatcher is nil")
	}

	// send only the bogus trackid if we don't have anything to begin with (stops godbus/dbus from spamming the console)
	metadata := &types.Metadata{
		TrackId: dbus.ObjectPath(BogusTrackID),
	}
	ensureValidTrackID(metadata)
	state.currentMetadata = metadata
	if state.server.Conn == nil {
		log.Debug("dbus server connection is not ready yet")
		return nil
	}
	return state.mprisHandler.Player.OnTitle()
}

func secondsToMicro(seconds int64) int64 {
	duration := time.Duration(seconds) * time.Second
	return duration.Microseconds()
}

func milliToMicro(milli int64) int64 {
	duration := time.Duration(milli) * time.Microsecond
	return duration.Microseconds()
}

func (m *tunesEventHandler) OnPlayerPlayEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerPlayEvent", t)
	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}
	m.handler.Player.OnPlayPause()
	m.handler.Player.OnPlayback()
}

func (m *tunesEventHandler) OnPlayerStopEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerStopEvent", t)
	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}
	m.handler.Player.OnPlayPause()
	m.handler.Player.OnPlayback()
	m.handler.Player.OnEnded()
}

func (m *tunesEventHandler) OnPlayerPlayingTrackChangedEvent(t *itunes.IiTrack) {
	log.Printf("OnPlayerPlayingTrackChangedEvent: %v", t)
	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}
	m.handler.Player.OnPlayPause()
	m.handler.Player.OnPlayback()
	m.handler.Player.OnTitle()
}

func (m *tunesEventHandler) OnQuittingEvent() {
	log.Debug("received OnQuittingEvent")
	m.state.QuitSafely(nil, "")
}

func (m *tunesEventHandler) OnAboutToPromptUserToQuitEvent() {
	log.Debug("received OnAboutToPromptUserToQuitEvent")
	m.state.QuitSafely(nil, "")
	// todo: 20seg~ timer to reconnect everything if that dialog happens to show up and the user clicks "Don't Quit"
}

func (m *tunesEventHandler) OnSoundVolumeChangedEvent(val *int64) {
	log.Debug("received OnSoundVolumeChangedEvent", *val)
	m.state.currentVolume = *val
	m.handler.Player.OnVolume()
}

func StartServingBus(s *server.Server) {
	log.Info("starting MPRIS server...")
	err := s.Listen()

	if err != nil {
		log.Error("startMprisServer failed, quitting", err)
		return
	}
}

func (state *State) startTicker() {
	for {
		select {
		case <-state.quit:
			return
		case <-state.ticker.C:
			if state.tunesDispatcher != nil {
				tunes, _ := itunes.GetCurrentTunes(state.tunesDispatcher)
				if tunes != nil {
					if tunes.PlayerPositionMS > 0 {
						position := time.Duration(tunes.PlayerPositionMS) * time.Millisecond
						state.currentPosition = position.Microseconds()

						state.mprisHandler.Player.OnTitle()
						state.mprisHandler.Player.OnOptions()
						state.mprisHandler.Player.OnPosition()
					}
				}
			}
		}
	}
}

func (state *State) ParseArgs() {
	debugModePtr := flag.Bool("debug", false, "Enable debug logging")
	state.config.identity = flag.String("identity", "iTunes", "Custom identity for the MPRIS server\n"+
		"Tip: Set this to \"cider\" in all lowercase (or use some other whitelisted identity) if you want to make Music Presence pick up the player.")
	flag.Parse()

	if *debugModePtr {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	state := &State{
		ticker:          time.NewTicker(500 * time.Millisecond),
		currentMetadata: &types.Metadata{},
		artworkCache: WeakTrackArtworkCache{
			store: weakmap.Map[int64, string]{},
		},
		quit: make(chan struct{}),
	}
	go func() {
		<-sigs
		state.QuitSafely(nil, "")
	}()
	state.ParseArgs()

	dispatcher, err := itunes.NewTunesDispatch()
	if err != nil {
		state.QuitSafely(err, "failed to initialize dispatcher")
		return
	}
	state.tunesDispatcher = dispatcher
	//defer state.QuitSafely(nil, "")

	busRoot := BusRoot{
		state: state,
	}

	busPlayer := BusPlayer{
		dispatcher: dispatcher,
		state:      state,
	}

	state.server = server.NewServer(*state.config.identity, busRoot, &busPlayer)
	state.mprisHandler = events.NewEventHandler(state.server)

	handler := &tunesEventHandler{
		state:      state,
		handler:    state.mprisHandler,
		dispatcher: dispatcher,
	}

	sink, err := itunes.NewCOMEventSink(dispatcher, handler)
	if err != nil {
		state.QuitSafely(err, "something failed when setting up the event sink")
	}
	setPlayerMetadata(nil, state)
	StartServingBus(state.server)
	go state.startTicker()

	curr, err := itunes.GetCurrentTrack(dispatcher)
	if state.server.Conn != nil {
		if err == nil {
			if curr != nil && curr.Dispatcher != nil {
				log.Debug("current track", curr)

				setPlayerMetadata(curr, state)
			}
		} else {
			log.Debug("failed to get current track, probably itunes doesn't have anything in the queue...")
		}
	}
	state.mprisHandler.Player.OnAll()

	err = sink.ListenEvents(state.quit)
	if err != nil {
		state.QuitSafely(err, "failed to listen for COM events")
	}

	log.Info("the end")
}

func (state *State) QuitSafely(err error, message string) {
	if state.server.Conn != nil {
		err := state.server.Stop()
		if err != nil {
			log.Warn("failed to stop the dbus server. you might want to force kill wineserver with \"wineserver -k\"...")
		}
	}
	code := 0

	if err != nil {
		code = 1
		defer os.Exit(code)

		if message != "" {
			log.Debug(message, err)
		} else {
			log.Debug("quitting because of critical error", err)
		}
		return
	} else {
		log.Info("now quitting...")
		defer os.Exit(code)
	}

	if state.tunesDispatcher != nil {
		state.tunesDispatcher.Release()
	}
	state.quitOnce.Do(func() {
		close(state.quit)
	})
}
