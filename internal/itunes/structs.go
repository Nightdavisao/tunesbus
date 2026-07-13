package itunes

import "github.com/go-ole/go-ole"

type TunesEventHandler interface {
    OnPlayerPlayEvent(*IiTrack)
    OnPlayerStopEvent(*IiTrack)
    OnPlayerPlayingTrackChangedEvent(*IiTrack)
    OnQuittingEvent()
    OnAboutToPromptUserToQuitEvent()
    OnSoundVolumeChangedEvent(*int64)
}

type IiTrack struct {
	Dispatcher *ole.IDispatch `com:"self"`
	Name string
	Artist string
	Album string
	Duration int64
	DiscNumber int64
	TrackNumber int64
	TrackCount int64
	TrackID int64 `com:"trackID"`
}

// float = int64
type IiTunes struct {
	Dispatcher *ole.IDispatch `com:"self"`
	CanSetShuffle bool
	CanSetSongRepeat bool
	//CurrentTrack IiTrack
	// [id(0x60020021), propget, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// long _stdcall PlayerPosition();
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// void _stdcall PlayerPosition([in] long rhs);
	PlayerPosition int32 // can we set the position?
	PlayerPositionMS int32
	PlayerState int32
	SoundVolume int32
	Rating int64
	Time string
	trackID int64
	// BackTrack, NextTrack, Resume, Play, PlayPause
}

type dispParams struct {
	rgvarg            uintptr
	rgdispidNamedArgs uintptr
	cArgs             uint32
	cNamedArgs        uint32
}