//go:build windows

package main

import (
	"log"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/wine"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/events"
	"github.com/quarckster/go-mpris-server/pkg/server"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

const BogusTrackID = "/org/mpris/MediaPlayer2/Track/1"

type Root struct {
	dispatcher *ole.IDispatch
}

func (r Root) Raise() error {
	log.Printf("Raise is not implemented")
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
	log.Printf("CanRaise is not implemented")
	return false, nil
}

func (r Root) HasTrackList() (bool, error) {
	log.Printf("HasTrackList is not implemented")
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
	return float64(m.state.currentVolume/100), nil
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

type State struct {
	currentMetadata *types.Metadata
	currentVolume   int64
	currentPosition int64
	isPlaying       bool
	ticker          *time.Ticker
	done            chan bool
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

func setInitialMetadata(track *itunes.IiTrack, trackDispatcher *ole.IDispatch, state *State, afterSetting fn) {
	if track != nil {
		*state.currentMetadata = types.Metadata{
			TrackId:     dbus.ObjectPath("/org/mpris/MediaPlayer2/Track/1"),
			Album:       track.Album,
			Title:       track.Name,
			Artist:      []string{track.Artist},
			Length:      types.Microseconds(secondsToMicro(track.Duration)),
			DiscNumber:  int(track.DiscNumber),
			TrackNumber: int(track.TrackNumber),
		}
		if trackDispatcher != nil {
			go func(state *State, track *itunes.IiTrack) {
				dosFilename, err := itunes.SaveArtworkIfAvaliable(trackDispatcher, track); if err != nil {
					log.Printf("failed to retrieve artwork for current track: %v", err)
				}
				
				log.Printf("dos filename for artwork: %s", dosFilename)
	
				unixFilename, err := wine.GetUnixFilename(dosFilename); if err != nil {
					log.Printf("failed to retrieve unix filename for saved artwork: %v", err)
				}
	
				log.Printf("unix filename for artwork: %s", unixFilename)
	
				state.currentMetadata.ArtUrl = "file://"+unixFilename
			}(state, track)
		}
		
		log.Printf("successfully set new metadata: %v", state.currentMetadata)
		if afterSetting != nil {
			afterSetting()
		}
		return
	}

	// send only the bogus trackid if we don't have anything to begin with
	state.currentMetadata = &types.Metadata{
		TrackId: dbus.ObjectPath("/org/mpris/MediaPlayer2/Track/1"),
	}
}

func setPosition(tunes *itunes.IiTunes, state *State) {
	if tunes != nil {
		state.currentPosition = milliToMicro(tunes.PlayerPositionMS)
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
	log.Printf("OnPlayerPlayEvent %v", t)
	setInitialMetadata(t, t.Dispatcher, m.state, nil)

}

func (m *eventHandler) OnPlayerStopEvent(t *itunes.IiTrack) {
	log.Printf("OnPlayerStopEvent: %v", t)
	setInitialMetadata(t, t.Dispatcher, m.state, func() {
		m.handler.Player.OnPlayback()
		m.handler.Player.OnPlayPause()
	})
}

func (m *eventHandler) OnPlayerPlayingTrackChangedEvent(t *itunes.IiTrack) {
	log.Printf("OnPlayerPlayingTrackChangedEvent: %v", t)
	setInitialMetadata(t, t.Dispatcher, m.state, func() {
		m.handler.Player.OnPlayback()
	})
}

func (m *eventHandler) OnQuittingEvent() {
	m.QuitCalled = true
	//m.state.done<-true
	log.Printf("OnQuittingEvent")
}

func (m *eventHandler) OnAboutToPromptUserToQuitEvent() {
	m.AboutToQuitCalled = true
	//m.state.done<-true
	log.Printf("OnAboutToPromptUserToQuitEvent")
}

func (m *eventHandler) OnSoundVolumeChangedEvent(val *int64) {
	log.Printf("OnSoundVolumeChangedEvent, %d", *val)
	<-time.After(2 * time.Second) // we can't really tell if this was from mpris or itunes itself, so we'll be debouncing the emit change
	m.state.currentVolume = *val
	m.handler.Player.OnVolume()	
}

func main() {
	state := &State{
		ticker:          time.NewTicker(50 * time.Millisecond),
		currentMetadata: &types.Metadata{},
		done:            make(chan bool), // shall only be used if the program is quitting...?
	}

	log.Printf("starting up")

	dispatcher, err := itunes.NewTunesDispatch()
	if err != nil {
		log.Printf("failed to initialize dispatcher")
		panic(err)
	}

	root := Root{
		dispatcher: dispatcher,
	}

	player := Player{
		dispatcher: dispatcher,
		state:      state,
	}

	srv := server.NewServer("iTunes", root, &player)
	ev := events.NewEventHandler(srv)

	handler := &eventHandler{
		state:      state,
		handler:    ev,
		dispatcher: dispatcher,
	}

	sink, err := itunes.NewCOMEventSink(dispatcher, handler)
	if err != nil {
		log.Fatal("something failed when setting up the event sink")
		panic(err)
	}

	curr, _ := itunes.GetCurrentTrack(dispatcher)
	log.Printf("current track: %v", curr)
	if curr != nil {
		setInitialMetadata(curr, curr.Dispatcher, state, func() {
			ev.Player.OnAll()
		})
	} else {
		setInitialMetadata(nil, nil, state, nil)
	}

	go func() {
		if err := srv.Listen(); err != nil {
			log.Printf("listen failed: %v", err)
			panic(err)
		}
	}()

	go func() {
		for {
			select {
			case <-state.done:
				return
			case <-state.ticker.C:
				if player.dispatcher != nil {
					tunes, _ := itunes.GetCurrentTunes(player.dispatcher)
					if tunes != nil {
						position := time.Duration(tunes.PlayerPositionMS) * time.Millisecond
						state.currentPosition = position.Microseconds()
						ev.Player.OnAll()
					}
				}
			}
		}
	}()

	sink.StartCOMEventLoop()
}
