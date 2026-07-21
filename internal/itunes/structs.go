package itunes

import "github.com/go-ole/go-ole"

type TunesEventHandler interface {
	OnPlayerPlayEvent(*IiTrackData, *ole.IDispatch)
	OnPlayerStopEvent(*IiTrackData, *ole.IDispatch)
	OnPlayerPlayingTrackChangedEvent(*IiTrackData, *ole.IDispatch)
	OnQuittingEvent()
	OnAboutToPromptUserToQuitEvent()
	OnSoundVolumeChangedEvent(*int64)
}

type IiTrackData struct {
	Name        string
	Artist      string
	Album       string
	Duration    int64
	DiscNumber  int64
	TrackNumber int64
	TrackCount  int64
	TrackID     int64 `com:"trackID"`
}

type ArtworkFormat int32

const (
	Unknown ArtworkFormat = iota
	JPEG
	PNG
	BMP
)

type ITPlayerState int32

const (
	ITPlayerStateStopped ITPlayerState = iota
	ITPlayerStatePlaying
	ITPlayerStateFastForward
	ITPlayerStateRewind
)

type ITPlayerRepeatMode int32

const (
	ITPlayerRepeatModeNone ITPlayerRepeatMode = iota
	ITPlayerRepeatModeOne
	ITPlayerRepeatModeAll
)

// float = int64
type IiTunes struct {
	CanSetShuffle    bool
	CanSetSongRepeat bool
	//CurrentTrack IiTrack
	// [id(0x60020021), propget, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// long _stdcall PlayerPosition();
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// void _stdcall PlayerPosition([in] long rhs);
	// Player position in seconds
	PlayerPosition int32
	// Player position in milliseconds
	PlayerPositionMS int32
	PlayerState      ITPlayerState
	SoundVolume      int32
	Rating           int64
	Time             string
	TrackID          int64 `com:"trackID"`
	// BackTrack, NextTrack, Resume, Play, PlayPause
}

type dispParams struct {
	rgvarg            uintptr
	rgdispidNamedArgs uintptr
	cArgs             uint32
	cNamedArgs        uint32
}
