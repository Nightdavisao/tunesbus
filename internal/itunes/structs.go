package itunes

type TunesEventHandler interface {
    OnPlayerPlayEvent(*IiTrack)
    OnPlayerStopEvent(*IiTrack)
    OnPlayerPlayingTrackChangedEvent(*IiTrack)
    OnQuittingEvent()
    OnAboutToPromptUserToQuitEvent()
    OnSoundVolumeChangedEvent(*int64)
}

type IiTrack struct {
	Name string
	Artist string
	Album string
	Duration int64
	DiscNumber int64
	TrackNumber int64
	TrackCount int64
}

// float = int64
type IiTunes struct {
	CanSetShuffle bool
	CanSetSongRepeat bool
	//CurrentTrack IiTrack
	// [id(0x60020021), propget, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// long _stdcall PlayerPosition();
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// void _stdcall PlayerPosition([in] long rhs);
	PlayerPosition int32 // can we set the position?
	PlayerPositionMS int64
	PlayerState int32
	SoundVolume int64
	Rating int64
	Time string
	trackID int64
	// BackTrack, NextTrack, Resume, Play, PlayPause
}