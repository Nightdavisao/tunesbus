package itunes

const TunesProgramID = "iTunes.Application"

const IID_IiTunesEvents = "{5846EB78-317E-4B6F-B0C3-11EE8C8FEEF2}"
const IID_IiTrack = "{4CB0915D-1E54-4727-BAF3-CE6CC9A225A1}"
const IID_IiTunes = "(9DD6680B-3EDC-40DB-A771-E6FE4832E34A)"

const (
	OnPlayerPlayEventNum                = 0x00000002
	OnPlayerStopEventNum                = 0x00000003
	OnPlayerPlayingTrackChangedEventNum = 0x00000004
	OnQuittingEventNum                  = 0x00000008
	OnAboutToPromptUserToQuitEventNum   = 0x00000009
	OnSoundVolumeChangedEventNum        = 0x0000000a
)

const (
	ITPlayButtonStatePlayDisabled  = 0
	ITPlayButtonStatePlayEnabled   = 1
	ITPlayButtonStatePauseEnabled  = 2
	ITPlayButtonStatePauseDisabled = 3
	ITPlayButtonStateStopEnabled   = 4
	ITPlayButtonStateStopDisabled  = 5
)
