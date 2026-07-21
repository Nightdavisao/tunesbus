//go:build windows

package main

import (
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/olejunk"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

type PlaybackState struct {
	currentVolume   int64
	currentPosition int64
	hasPosition     bool

	playerState    itunes.ITPlayerState
	hasPlayerState bool

	playButtonState int32
	hasButtonState  bool
	canGoNext       bool
	canGoPrevious   bool

	shuffle    bool
	hasShuffle bool

	loopStatus    types.LoopStatus
	hasLoopStatus bool
}

func (state PlaybackState) canPlayPause() bool {
	if !state.hasButtonState {
		return true
	}
	return state.playButtonState != itunes.ITPlayButtonStatePauseDisabled &&
		state.playButtonState != itunes.ITPlayButtonStatePlayDisabled
}

type playbackChanges struct {
	playbackStatus bool
	volume         bool
	options        bool
}

func repeatModeToLoopStatus(mode itunes.ITPlayerRepeatMode) types.LoopStatus {
	switch mode {
	case itunes.ITPlayerRepeatModeOne:
		return types.LoopStatusTrack
	case itunes.ITPlayerRepeatModeAll:
		return types.LoopStatusPlaylist
	default:
		return types.LoopStatusNone
	}
}

func getPlaylistOptions(tunesDisp *ole.IDispatch) (bool, types.LoopStatus, bool, error) {
	releaser := olejunk.NewOleReleaser()
	defer releaser.Release()
	
	playlistDisp, err := itunes.SafeGetCurrentPlaylist(tunesDisp, releaser)
	if err != nil {
		return false, types.LoopStatusNone, false, err
	}
	if playlistDisp == nil {
		return false, types.LoopStatusNone, true, nil
	}

	shuffleStatus, err := oleutil.GetProperty(playlistDisp, "Shuffle")
	if err != nil {
		return false, types.LoopStatusNone, false, err
	}
	defer shuffleStatus.Clear()

	shuffle, err := olejunk.GetVariantValue[bool](shuffleStatus)
	if err != nil {
		return false, types.LoopStatusNone, false, err
	}

	songRepeat, err := olejunk.GetPropertyFromIDispatch[itunes.ITPlayerRepeatMode](playlistDisp, "SongRepeat")
	if err != nil {
		return false, types.LoopStatusNone, false, err
	}
	if songRepeat == nil {
		return *shuffle, types.LoopStatusNone, true, nil
	}

	return *shuffle, repeatModeToLoopStatus(*songRepeat), true, nil
}

func (state *MainState) refreshPlaybackState(includeOptions bool) playbackChanges {
	if state.tunesDisp == nil {
		return playbackChanges{}
	}

	tunes, err := itunes.GetCurrentTunes(state.tunesDisp)
	if err != nil {
		log.Debug("failed to get current iTunes state", "error", err)
		return playbackChanges{}
	}
	if tunes == nil {
		return playbackChanges{}
	}

	var (
		prevEnabled bool
		buttonState int32
		nextEnabled bool
		buttonsOK   bool
		shuffle     bool
		loopStatus  types.LoopStatus
		playlistOK  bool
	)

	if includeOptions {
		prevEnabled, buttonState, nextEnabled, err = itunes.GetPlayerButtonsState(state.tunesDisp)
		if err != nil {
			log.Debug("failed to get player buttons state", "error", err)
		} else {
			buttonsOK = true
		}

		shuffle, loopStatus, playlistOK, err = getPlaylistOptions(state.tunesDisp)
		if err != nil {
			log.Debug("failed to get playlist options", "error", err)
		}
	}

	changes := playbackChanges{}
	position := time.Duration(tunes.PlayerPositionMS) * time.Millisecond

	state.mux.Lock()
	defer state.mux.Unlock()

	if !state.playbackState.hasPlayerState || tunes.PlayerState != state.playbackState.playerState {
		state.playbackState.playerState = tunes.PlayerState
		state.playbackState.hasPlayerState = true
		changes.playbackStatus = true
	}

	if int64(tunes.SoundVolume) != state.playbackState.currentVolume {
		state.playbackState.currentVolume = int64(tunes.SoundVolume)
		changes.volume = true
	}

	if tunes.PlayerPositionMS >= 0 {
		state.playbackState.currentPosition = position.Microseconds()
		state.playbackState.hasPosition = true
	}

	if buttonsOK {
		if !state.playbackState.hasButtonState ||
			state.playbackState.playButtonState != buttonState ||
			state.playbackState.canGoPrevious != prevEnabled ||
			state.playbackState.canGoNext != nextEnabled {
			state.playbackState.playButtonState = buttonState
			state.playbackState.hasButtonState = true
			state.playbackState.canGoPrevious = prevEnabled
			state.playbackState.canGoNext = nextEnabled
			changes.options = true
		}
	}

	if playlistOK {
		if !state.playbackState.hasShuffle || state.playbackState.shuffle != shuffle {
			state.playbackState.shuffle = shuffle
			state.playbackState.hasShuffle = true
			changes.options = true
		}
		if !state.playbackState.hasLoopStatus || state.playbackState.loopStatus != loopStatus {
			state.playbackState.loopStatus = loopStatus
			state.playbackState.hasLoopStatus = true
			changes.options = true
		}
	}

	return changes
}

func (state *MainState) emitPlaybackChanges(changes playbackChanges) {
	if changes.playbackStatus {
		state.mprisHandler.Player.OnPlayPause()
	}
	if changes.volume {
		state.mprisHandler.Player.OnVolume()
	}
	if changes.options {
		state.mprisHandler.Player.OnOptions()
	}
}
