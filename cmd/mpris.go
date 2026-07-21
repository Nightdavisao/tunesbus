//go:build windows

package main

import (
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/olejunk"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole/oleutil"
	"github.com/godbus/dbus/v5"
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
	return r.state.config.MPRIS.Identity, nil
}

func (r BusRoot) DesktopEntry() (string, error) {
	return r.state.config.MPRIS.DesktopEntry, nil
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
	if err == nil {
		m.state.mux.Lock()
		m.state.playbackState.currentPosition = int64(position)
		m.state.playbackState.hasPosition = true
		m.state.mux.Unlock()
	}
	return err
}

func (m *BusPlayer) OpenUri(uri string) error {
	return nil
}

func (m *BusPlayer) PlaybackStatus() (types.PlaybackStatus, error) {
	log.Debug("PlaybackStatus called")
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	if !m.state.playbackState.hasPlayerState {
		return types.PlaybackStatusPaused, nil
	}

	switch m.state.playbackState.playerState {
	case itunes.ITPlayerStatePlaying:
		return types.PlaybackStatusPlaying, nil
	case itunes.ITPlayerStateStopped:
		return types.PlaybackStatusStopped, nil
	default:
		return types.PlaybackStatusPaused, nil
	}
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
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return float64(m.state.playbackState.currentVolume) / 100, nil
}

func (m *BusPlayer) SetVolume(volume float64) error {
	r, err := oleutil.PutProperty(m.state.tunesDisp, "SoundVolume", volume*100)
	if r != nil {
		r.Clear()
	}
	if err == nil {
		m.state.mux.Lock()
		m.state.playbackState.currentVolume = int64(volume * 100)
		m.state.mux.Unlock()
	}
	return err
}

func (m *BusPlayer) Position() (int64, error) {
	log.Debug("Position called")
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return int64(m.state.playbackState.currentPosition), nil
}

func (m *BusPlayer) MinimumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) MaximumRate() (float64, error) {
	return 1.0, nil
}

func (m *BusPlayer) CanGoNext() (bool, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return m.state.playbackState.canGoNext, nil
}

func (m *BusPlayer) CanGoPrevious() (bool, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return m.state.playbackState.canGoPrevious, nil
}

func (m *BusPlayer) CanPlay() (bool, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	log.Debug("CanPlay called", "buttonState", m.state.playbackState.playButtonState)
	return m.state.playbackState.canPlayPause(), nil
}

func (m *BusPlayer) CanPause() (bool, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return m.state.playbackState.canPlayPause(), nil
}

func (m *BusPlayer) CanSeek() (bool, error) {
	return true, nil // even though we don't actually support "Seek", we need to advertise that we do, clients will set "Position" anyway
}

func (m *BusPlayer) CanControl() (bool, error) {
	return true, nil
}

func (m *BusPlayer) Shuffle() (bool, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	return m.state.playbackState.shuffle, nil
}

func (m *BusPlayer) SetShuffle(shuffle bool) error {
	releaser := olejunk.NewOleReleaser()
	defer releaser.Release()
	
	playlistDispatcher, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp, releaser)
	if err != nil {
		log.Error("failed to get current playlist on setting Shuffle", err)
		return nil
	}

	if playlistDispatcher != nil {
		result, err := oleutil.PutProperty(playlistDispatcher, "Shuffle", shuffle)
		if err != nil {
			log.Error("failed to put shuffle status", "error", err)
			return err
		}
		result.Clear()
	} else {
		log.Debug("no playlist yet")
		return nil
	}
	m.state.mux.Lock()
	m.state.playbackState.shuffle = shuffle
	m.state.playbackState.hasShuffle = true
	m.state.mux.Unlock()
	return nil
}

func (m *BusPlayer) LoopStatus() (types.LoopStatus, error) {
	m.state.mux.RLock()
	defer m.state.mux.RUnlock()

	if !m.state.playbackState.hasLoopStatus {
		return types.LoopStatusNone, nil
	}
	return m.state.playbackState.loopStatus, nil
}

func (m *BusPlayer) SetLoopStatus(status types.LoopStatus) error {
	releaser := olejunk.NewOleReleaser()
	defer releaser.Release()
	
	playlistDispatch, err := itunes.SafeGetCurrentPlaylist(m.state.tunesDisp, releaser)
	if err != nil {
		log.Error("failed to get current playlist on setting Loop", err)
		return err
	}
	if playlistDispatch == nil {
		log.Debug("no playlist yet")
		return nil
	}

	var mode int32
	switch status {
	case types.LoopStatusTrack:
		mode = 1
	case types.LoopStatusPlaylist:
		mode = 2
	default:
		mode = 0
	}
	r, err := oleutil.PutProperty(playlistDispatch, "SongRepeat", mode)
	if err != nil {
		return err
	}
	m.state.mux.Lock()
	r.Clear()
	m.state.playbackState.loopStatus = status
	m.state.playbackState.hasLoopStatus = true
	m.state.mux.Unlock()
	return err
}
