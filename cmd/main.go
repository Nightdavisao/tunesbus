//go:build windows

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/olejunk"
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

type BusRoot struct {
	state *MainState
}

func (r BusRoot) Raise() error {
	log.Debug("Raise is not implemented")
	return nil
}

func (r BusRoot) Quit() error {
	if r.state.tunesDisp != nil {
		r, err := oleutil.CallMethod(r.state.tunesDisp, "Quit")
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
	state *MainState
}

func (m *BusPlayer) Next() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "NextTrack")
	r.Clear()
	return err
}

func (m *BusPlayer) Previous() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "PreviousTrack")
	r.Clear()
	return err
}

func (m *BusPlayer) Pause() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "Pause")
	r.Clear()
	return err
}

func (m *BusPlayer) PlayPause() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "PlayPause")
	r.Clear()
	return err
}

func (m *BusPlayer) Stop() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "Stop")
	r.Clear()
	return err
}

func (m *BusPlayer) Play() error {
	r, err := oleutil.CallMethod(m.state.tunesDisp, "Play")
	r.Clear()
	return err
}

func (m *BusPlayer) Seek(offset types.Microseconds) error {
	return nil
}

func (m *BusPlayer) SetPosition(trackId dbus.ObjectPath, position types.Microseconds) error {
	log.Debug("setting Position", position)

	seconds := (time.Duration(position) * time.Microsecond) / time.Second
	err := itunes.SetTunesPosition(m.state.tunesDisp, int64(seconds))
	return err
}

func (m *BusPlayer) OpenUri(uri string) error {
	return nil
}

func (m *BusPlayer) PlaybackStatus() (types.PlaybackStatus, error) {
	log.Debug("PlaybackStatus called")
	tunes, err := itunes.GetCurrentTunes(m.state.tunesDisp)
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
	if m.state.currentMetadata != nil {
		log.Debug("Metadata called", *m.state.currentMetadata)
		if m.state.currentMetadata.TrackId.IsValid() {
			return *m.state.currentMetadata, nil
		}
	}
	if m.state.currentMetadata == nil {
		log.Info("Metadata called", "metadata is nil, using fallback")
	} else {
		log.Info("Metadata called", "metadata has invalid track id, using fallback", "track_id", m.state.currentMetadata.TrackId)
	}

	return types.Metadata{}, nil
}

func (m *BusPlayer) Volume() (float64, error) {
	log.Debug("Volume called")
	return float64(m.state.playbackState.currentVolume / 100), nil
}

func (m *BusPlayer) SetVolume(volume float64) error {
	r, err := oleutil.PutProperty(m.state.tunesDisp, "SoundVolume", volume*100)
	r.Clear()
	return err
}

func (m *BusPlayer) Position() (int64, error) {
	log.Debug("Position called")
	return int64(m.state.playbackState.currentPosition), nil
}

func (m *BusPlayer) MinimumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) MaximumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) CanGoNext() (bool, error) {
	return m.state.playbackState.canGoNext, nil
}

func (m *BusPlayer) CanGoPrevious() (bool, error) {
	return m.state.playbackState.canGoPrevious, nil
}

func (m *BusPlayer) CanPlay() (bool, error) {
	log.Debug("CanPlay called", "buttonState", m.state.playbackState.lastPlayerState)
	return m.state.playbackState.lastPlayerState != itunes.ITPlayButtonStatePauseDisabled &&
		m.state.playbackState.lastPlayerState != itunes.ITPlayButtonStatePlayDisabled, nil
}

func (m *BusPlayer) CanPause() (bool, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	return m.state.playbackState.lastPlayerState != itunes.ITPlayButtonStatePauseDisabled &&
		m.state.playbackState.lastPlayerState != itunes.ITPlayButtonStatePlayDisabled, nil
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

	playlistDispatch, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp)
	if err != nil {
		log.Error("failed to get current playlist on getting Shuffle", err)
		return false, nil
	}

	if playlistDispatch != nil {
		defer playlistDispatch.Release()

		shuffleStatus, err := oleutil.GetProperty(playlistDispatch, "Shuffle")
		if err != nil {
			log.Error("failed to get shuffle status", err)
			return false, err
		}
		r, err := olejunk.GetVariantValue[bool](shuffleStatus)
		if err != nil {
			return false, err
		}
		return *r, err
	}
	return false, nil
}

func (m *BusPlayer) SetShuffle(shuffle bool) error {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp)
	if err != nil {
		log.Error("failed to get current playlist on setting Shuffle", err)
		return nil
	}

	if playlistDispatcher != nil {
		defer playlistDispatcher.Release()

		result, err := oleutil.PutProperty(playlistDispatcher, "Shuffle", shuffle)
		if err != nil {
			log.Error("failed to put shuffle status", "error", err)
			return err
		}
		result.Clear()
	}
	return nil
}

func (m *BusPlayer) LoopStatus() (types.LoopStatus, error) {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	playlistDisp, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp)
	if err != nil {
		log.Error("failed to get current playlist on getting Loop", err)
		return types.LoopStatusNone, err
	}
	if playlistDisp == nil {
		return types.LoopStatusNone, nil
	}
	defer playlistDisp.Release()

	songRepeat, err := olejunk.GetPropertyFromIDispatch[itunes.ITPlayerRepeatMode](playlistDisp, "SongRepeat")
	if err != nil {
		return types.LoopStatusNone, err
	}
	if songRepeat == nil {
		return types.LoopStatusNone, nil
	}

	switch *songRepeat {
	case itunes.ITPlayerRepeatModeOne:
		return types.LoopStatusTrack, nil
	case itunes.ITPlayerRepeatModeAll:
		return types.LoopStatusPlaylist, nil
	default:
		return types.LoopStatusNone, nil
	}
}

func (m *BusPlayer) SetLoopStatus(status types.LoopStatus) error {
	m.state.mux.Lock()
	defer m.state.mux.Unlock()

	playlistDispatch, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp)
	if err != nil {
		log.Error("failed to get current playlist on setting Loop", err)
		return err
	}
	if playlistDispatch == nil {
		log.Debug("no playlist yet")
		return nil
	}
	defer playlistDispatch.Release()

	var mode int32
	switch status {
	case types.LoopStatusTrack:
		mode = 1
	case types.LoopStatusPlaylist:
		mode = 2
	default:
		mode = 0
	}
	_, err = oleutil.PutProperty(playlistDispatch, "SongRepeat", mode)
	return err
}

type WeakTrackArtworkCache struct {
	store weakmap.Map[int64, string]
}

type Config struct {
	identity        *string
	diagnosticsOnly *bool
}

type OnceGroup struct {
	quitOnce        sync.Once
	mprisStartOnce  sync.Once
	initialSyncOnce sync.Once
}

type PlaybackState struct {
	currentVolume   int64
	currentPosition int64
	lastPlayerState itunes.ITPlayerState
	hasPlayerState  bool
	canGoNext       bool
	canGoPrevious   bool
}

type MainState struct {
	config    Config
	tunesDisp *ole.IDispatch

	mux  sync.RWMutex
	sync OnceGroup

	currentMetadata *types.Metadata
	playbackState   PlaybackState

	server       *server.Server
	mprisHandler *events.EventHandler

	comSink *itunes.COMEventSink

	ticker *time.Ticker
	quit   chan struct{}

	artworkCache WeakTrackArtworkCache
}

type tunesEventHandler struct {
	state         *MainState
	tunesDispatch *ole.IDispatch
	handler       *events.EventHandler
}

func (state *MainState) ensureMprisStarted() {
	state.sync.mprisStartOnce.Do(func() {
		go state.startServingBus(state.server)
	})
}

func (state *MainState) waitForMprisReady(timeout time.Duration) bool {
	if state.server == nil {
		return false
	}
	if state.server.Conn != nil {
		return true
	}
	if timeout <= 0 {
		select {
		case <-state.server.Ready():
			return true
		default:
			return false
		}
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-state.server.Ready():
		return true
	case <-state.quit:
		return false
	case <-timer.C:
		return state.server.Conn != nil
	}
}

func (state *MainState) emitInitialMprisState() {
	state.sync.initialSyncOnce.Do(func() {
		state.mprisHandler.Player.OnAll()
	})
}

// note that this is already releasing the track's IDispatch object, don't release it yourself after using this
func setPlayerMetadata(track *itunes.IiTrack, state *MainState) error {
	state.mux.Lock()
	defer state.mux.Unlock()

	if track != nil {
		metadata := types.Metadata{
			Album:       track.Album,
			Title:       track.Name,
			Artist:      []string{track.Artist},
			Length:      types.Microseconds(secondsToMicro(track.Duration)),
			DiscNumber:  int(track.DiscNumber),
			TrackNumber: int(track.TrackNumber),
			TrackId:     dbus.ObjectPath(fmt.Sprintf("/org/itunes/track/%d", track.TrackID)),
		}

		if track.IDispatch != nil {
			log.Info("partial new metadata (not sent yet)", "metadata", state.currentMetadata, "track", track, "dispatch", track.IDispatch)
			defer track.IDispatch.Release()

			val, exists := state.artworkCache.store.Get(track.TrackID)
			if exists {
				metadata.ArtUrl = val
				*state.currentMetadata = metadata
				log.Info("will send cached artwork from weak map", "track_id", track.TrackID, "value", val)
				if state.server.Conn == nil {
					log.Debug("dbus server connection is not ready yet")
					return nil
				}
				return state.mprisHandler.Player.OnTitle()
			}

			// if we don't have the artwork...
			log.Info("artwork for this track doesn't exist yet", "track_id", track.TrackID)

			dosFilename, err := itunes.SaveArtworkIfAvaliable(track.IDispatch, track)
			log.Info("dos filename for artwork", "dos_filename", dosFilename)
			if err != nil {
				log.Info("failed to get artwork")
				log.Error("failed to retrieve artwork for current track", err)
				*state.currentMetadata = metadata
				if state.server == nil || state.server.Conn == nil {
					return nil
				}
				return state.mprisHandler.Player.OnTitle()
			}

			unixFilename, err := wine.GetUnixFilename(dosFilename)
			if err != nil {
				log.Error("failed to retrieve unix filename for saved artwork", err)
				*state.currentMetadata = metadata
				if state.server == nil || state.server.Conn == nil {
					return nil
				}
				return state.mprisHandler.Player.OnTitle()
			}
			log.Debug("unix filename for artwork", unixFilename)

			artUrl := "file://" + unixFilename
			state.artworkCache.store.Set(track.TrackID, artUrl)

			metadata.ArtUrl = artUrl
			*state.currentMetadata = metadata
			if state.server.Conn == nil {
				log.Debug("dbus server connection is not ready yet")
				return nil
			}
			return state.mprisHandler.Player.OnTitle()
		}
		return fmt.Errorf("track.IDispatch is nil")
	}

	if state.server.Conn == nil {
		log.Info("dbus server connection is not ready yet")
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

	prevEnabled, _, nextEnabled, err := itunes.GetPlayerButtonsState(m.state.tunesDisp)
	if err != nil {
		log.Debug("failed to get buttons state, setting true for both previous and next anyway", "error", err)
		m.state.playbackState.canGoNext = true
		m.state.playbackState.canGoPrevious = true
	} else {
		log.Debug("button state", "prevEnabled", prevEnabled, "nextEnabled", nextEnabled)
		m.state.playbackState.canGoPrevious = prevEnabled
		m.state.playbackState.canGoNext = nextEnabled
	}


	m.state.ensureMprisStarted()
	if !m.state.waitForMprisReady(2 * time.Second) {
		log.Warn("MPRIS server is not ready yet, skipping play emit")
		return
	}

	m.state.emitInitialMprisState()
	m.handler.Player.OnPlayPause()
}

func (m *tunesEventHandler) OnPlayerStopEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerStopEvent", t)
	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}
	m.handler.Player.OnTitle()
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
	m.handler.Player.OnTitle()
}

func (m *tunesEventHandler) OnQuittingEvent() {
	log.Debug("received OnQuittingEvent")
	m.state.QuitSafely(nil, "")
}

func (m *tunesEventHandler) OnAboutToPromptUserToQuitEvent() {
	log.Debug("received OnAboutToPromptUserToQuitEvent")
	m.state.QuitSafely(nil, "")
	// be evil
	err := killTunes()
	if err != nil {
		log.Error("failed to kill iTunes.exe via taskkill", "err", err)
	}
}

func (m *tunesEventHandler) OnSoundVolumeChangedEvent(val *int64) {
	log.Debug("received OnSoundVolumeChangedEvent", *val)
	if m.state.playbackState.currentVolume == *val {
		return
	}
	m.state.playbackState.currentVolume = *val
	m.handler.Player.OnVolume()
}

func (state *MainState) startServingBus(s *server.Server) {
	log.Info("starting MPRIS server...")
	err := s.Listen()

	if err != nil {
		state.QuitSafely(err, "startMprisServer failed, quitting")
	}
}

func (state *MainState) startTicker() {
	for {
		select {
		case <-state.quit:
			return
		case <-state.ticker.C:
			if !state.waitForMprisReady(0) {
				continue
			}
			if state.tunesDisp != nil {
				tunes, _ := itunes.GetCurrentTunes(state.tunesDisp)
				if tunes != nil {
					if !state.playbackState.hasPlayerState || tunes.PlayerState != state.playbackState.lastPlayerState {
						state.playbackState.lastPlayerState = tunes.PlayerState
						state.playbackState.hasPlayerState = true
						state.mprisHandler.Player.OnPlayPause()
					}

					if int64(tunes.SoundVolume) != state.playbackState.currentVolume {
						state.playbackState.currentVolume = int64(tunes.SoundVolume)
						state.mprisHandler.Player.OnVolume()
					}

					if tunes.PlayerPositionMS > 0 {
						position := time.Duration(tunes.PlayerPositionMS) * time.Millisecond
						state.playbackState.currentPosition = position.Microseconds()

						if tunes.PlayerState == itunes.ITPlayerStatePlaying {
							state.mprisHandler.Player.OnPosition()
						}
					}

					state.mprisHandler.Player.OnOptions()
				}
			}
		}
	}
}

func (state *MainState) ParseArgs() {
	debugModePtr := flag.Bool("debug", false, "Enable debug logging")

	state.config.identity = flag.String("identity", "iTunes", "Custom identity for the MPRIS server\n"+
		"Tip: Set this to \"cider\" in all lowercase (or use some other whitelisted identity) if you want to make Music Presence pick up the player.")
	state.config.diagnosticsOnly = flag.Bool("diagnostics", false, "Show diagnostics and quit")
	flag.Parse()

	if *debugModePtr {
		log.SetLevel(log.DebugLevel)
		return
	}
}

func main() {
	runtime.LockOSThread()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	state := &MainState{
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

	const title = "tunesbus"

	if *state.config.diagnosticsOnly {
		text := ""

		wineVersion, err := wine.GetWineVersion()
		if err != nil {
			wine.ErrorMessageBox(
				title,
				fmt.Sprintf("Error on getting Wine version: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Wine version: %s\n", wineVersion)

		wineBuild, err := wine.GetWineBuild()
		if err != nil {
			wine.ErrorMessageBox(
				title,
				fmt.Sprintf("Error on getting Wine build ID: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Wine build: %s\n", wineBuild)

		tmpDir, err := wine.UnixTmpDirAsDosPath()
		if err != nil {
			wine.ErrorMessageBox(
				title,
				fmt.Sprintf("Error on getting temporary Unix directory: %v\n", err),
			)
			return
		}
		text = text + fmt.Sprintf("Temporary Unix directory (as DOS): %s\n", tmpDir)
		text = text + fmt.Sprintf("WINEPREFIX env: %s\n", wine.GetWinePrefix())

		wine.InfoMessageBox(title, text)
		return
	}

	var err error

	state.tunesDisp, err = itunes.NewTunesDispatch()
	if err != nil {
		state.QuitSafely(err, "failed to initialize dispatcher")
		return
	}

	busRoot := BusRoot{
		state: state,
	}

	busPlayer := BusPlayer{
		state: state,
	}

	state.server = server.NewServer(*state.config.identity, busRoot, &busPlayer)
	state.mprisHandler = events.NewEventHandler(state.server)

	handler := &tunesEventHandler{
		state:   state,
		handler: state.mprisHandler,
	}

	state.comSink, err = itunes.NewCOMEventSink(state.tunesDisp, handler)
	if err != nil {
		state.QuitSafely(err, "something failed when setting up the event sink")
	}

	go state.startTicker()

	err = state.comSink.ListenEvents()
	if err != nil {
		state.QuitSafely(err, "failed to listen for COM events")
	}

	log.Info("the end")
}

func (state *MainState) QuitSafely(err error, message string) {
	if state.server.Conn != nil {
		err := state.server.Stop()
		if err != nil {
			log.Warn("failed to stop the dbus server. you might want to force kill wineserver with \"wineserver -k\"...")
			wine.WarningMessageBox(
				"tunesbus",
				"Failed to stop the dbus server. You might want to force kill wineserver with the command below.\n\n"+
					fmt.Sprintf("WINEPREFIX=%s wineserver -k", wine.GetWinePrefix()),
			)
		}
	}
	if state.comSink != nil {
		state.comSink.DisconnectObject()
	}

	code := 0

	if err != nil {
		code = 1
		defer os.Exit(code)

		if message != "" {
			log.Error(message, "error", err)
			wine.ErrorMessageBox("tunesbus", message)
		} else {
			log.Error("quitting because of critical error", "error", err)
		}
		return
	} else {
		log.Info("now quitting...")
		defer os.Exit(code)
	}

	if state.tunesDisp != nil {
		state.tunesDisp.Release()
	}
	state.sync.quitOnce.Do(func() {
		close(state.quit)
	})
}

func killTunes() error {
	cmd := exec.Command("taskkill", "/F", "/IM", "iTunes.exe", "/T")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill failed: %v, output: %s", err, out)
	}
	return nil
}