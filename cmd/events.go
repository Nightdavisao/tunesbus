//go:build windows

package main

import (
	"time"
	"tunesbus/internal/itunes"

	"github.com/charmbracelet/log"

	"github.com/quarckster/go-mpris-server/pkg/events"
)

type tunesEventHandler struct {
	state   *MainState
	handler *events.EventHandler
}

func (m *tunesEventHandler) OnPlayerPlayEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerPlayEvent", t)

	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}

	changes := m.state.refreshPlaybackState(true)

	m.state.ensureMprisStarted()
	if !m.state.waitForMprisReady(2 * time.Second) {
		log.Warn("MPRIS server is not ready yet, skipping play emit")
		return
	}

	m.state.emitInitialMprisState()
	m.state.emitPlaybackChanges(changes)
}

func (m *tunesEventHandler) OnPlayerStopEvent(t *itunes.IiTrack) {
	log.Debug("received OnPlayerStopEvent", t)
	err := setPlayerMetadata(t, m.state)
	if err != nil {
		log.Error("failed to set initial metadata", err)
		return
	}
	m.state.mux.Lock()
	m.state.playbackState.playerState = itunes.ITPlayerStateStopped
	m.state.playbackState.hasPlayerState = true
	m.state.mux.Unlock()
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
	m.state.emitPlaybackChanges(m.state.refreshPlaybackState(true))
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
	if val == nil {
		return
	}
	log.Debug("received OnSoundVolumeChangedEvent", *val)
	m.state.mux.Lock()
	if m.state.playbackState.currentVolume == *val {
		m.state.mux.Unlock()
		return
	}
	m.state.playbackState.currentVolume = *val
	m.state.mux.Unlock()
	m.handler.Player.OnVolume()
}
