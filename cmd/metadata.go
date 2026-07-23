//go:build windows

package main

import (
	"fmt"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/olejunk"
	"tunesbus/internal/wine"

	"github.com/charmbracelet/log"
	"github.com/go-ole/go-ole"

	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

func secondsToMicro(seconds int64) int64 {
	duration := time.Duration(seconds) * time.Second
	return duration.Microseconds()
}

func setPlayerMetadata(track *itunes.IiTrackData, state *MainState) error {
	state.mux.Lock()
	defer state.mux.Unlock()

	albumArtist := track.AlbumArtist

	if albumArtist == "" && track.Compilation {
		albumArtist = state.config.Metadata.CompilationBoolAsString
	}

	if track != nil {
		state.currentMetadata = &types.Metadata{
			Album:       track.Album,
			Title:       track.Name,
			Artist:      []string{track.Artist},
			Length:      types.Microseconds(secondsToMicro(track.Duration)),
			DiscNumber:  int(track.DiscNumber),
			TrackNumber: int(track.TrackNumber),
			TrackId:     dbus.ObjectPath(fmt.Sprintf("/org/itunes/track/%d", track.TrackID)),
		}

		if albumArtist != "" {
			state.currentMetadata.AlbumArtist = []string{albumArtist}
		}
	}
	if state.server.Conn == nil {
		log.Info("dbus server connection is not ready yet")
		return nil
	}
	return state.mprisHandler.Player.OnTitle()
}

func setCoverArt(trackId int32, dispatch *ole.IDispatch, state *MainState) error {
	releaser := olejunk.NewOleReleaser()
	defer releaser.Release()
	
	if dispatch != nil {
		log.Info("partial new metadata (not sent yet)", "metadata", state.currentMetadata, "trackId", trackId)
		val, exists := state.artworkCache.store.Get(int64(trackId))
		if exists {
			state.currentMetadata.ArtUrl = val
			log.Info("will send cached artwork from weak map", "track_id", trackId)
			if state.server.Conn == nil {
				log.Debug("dbus server connection is not ready yet")
				return nil
			}
			return state.mprisHandler.Player.OnTitle()
		}

		// if we don't have the artwork...
		log.Info("artwork for this track doesn't exist yet", "track_id", trackId)

		dosFilename, err := itunes.SaveArtworkIfAvaliable(dispatch, trackId, releaser)
		log.Info("dos filename for artwork", "dos_filename", dosFilename)
		if err != nil {
			log.Info("failed to get artwork")
			log.Error("failed to retrieve artwork for current track", err)
			if state.server == nil || state.server.Conn == nil {
				return nil
			}
			return state.mprisHandler.Player.OnTitle()
		}

		unixFilename, err := wine.GetUnixFilename(dosFilename)
		if err != nil {
			log.Error("failed to retrieve unix filename for saved artwork", err)
			if state.server == nil || state.server.Conn == nil {
				return nil
			}
			return state.mprisHandler.Player.OnTitle()
		}
		log.Debug("unix filename for artwork", unixFilename)

		artUrl := "file://" + unixFilename
		state.artworkCache.store.Set(int64(trackId), artUrl)

		state.currentMetadata.ArtUrl = artUrl
		if state.server.Conn == nil {
			log.Debug("dbus server connection is not ready yet")
			return nil
		}
		return state.mprisHandler.Player.OnTitle()
	}
	return fmt.Errorf("track.IDispatch is nil")
}