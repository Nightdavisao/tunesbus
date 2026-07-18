//go:build windows

package main

import (
	"fmt"
	"time"
	"tunesbus/internal/itunes"
	"tunesbus/internal/wine"

	"github.com/charmbracelet/log"

	"github.com/godbus/dbus/v5"
	"github.com/quarckster/go-mpris-server/pkg/types"
)

func secondsToMicro(seconds int64) int64 {
	duration := time.Duration(seconds) * time.Second
	return duration.Microseconds()
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
