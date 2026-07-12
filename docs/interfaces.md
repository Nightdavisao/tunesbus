[
    odl,
    uuid(4CB0915D-1E54-4727-BAF3-CE6CC9A225A1),
    helpstring("IITTrack Interface"),
    hidden,
    dual,
    oleautomation
]
interface IITTrack : IDispatch {
    [id(0x60020000), helpstring("Returns the four IDs that uniquely identify this object.")]
    HRESULT _stdcall GetITObjectIDs(
        [out] long* sourceID,
        [out] long* playlistID,
        [out] long* trackID,
        [out] long* databaseID);
    [id(0x60020001), propget, helpstring("The name of the object.")]
    HRESULT _stdcall Name([out, retval] BSTR* rhs);
    [id(0x60020001), propput, helpstring("The name of the object.")]
    HRESULT _stdcall Name([in] BSTR rhs);
    [id(0x60020003), propget, helpstring("The index of the object in internal application order (1-based).")]
    HRESULT _stdcall Index([out, retval] long* rhs);
    [id(0x60020004), propget, helpstring("The source ID of the object.")]
    HRESULT _stdcall sourceID([out, retval] long* rhs);
    [id(0x60020005), propget, helpstring("The playlist ID of the object.")]
    HRESULT _stdcall playlistID([out, retval] long* rhs);
    [id(0x60020006), propget, helpstring("The track ID of the object.")]
    HRESULT _stdcall trackID([out, retval] long* rhs);
    [id(0x60020007), propget, helpstring("The track database ID of the object.")]
    HRESULT _stdcall TrackDatabaseID([out, retval] long* rhs);
    [id(0x60030000), helpstring("Delete this track.")]
    HRESULT _stdcall Delete();
    [id(0x60030001), helpstring("Start playing this track.")]
    HRESULT _stdcall Play();
    [id(0x60030002), helpstring("Add artwork from an image file to this track.")]
    HRESULT _stdcall AddArtworkFromFile(
        [in] BSTR filePath,
        [out, retval] IITArtwork** rhs);
    [id(0x60030003), propget, helpstring("The track kind.")]
    HRESULT _stdcall Kind([out, retval] ITTrackKind* rhs);
    [id(0x60030004), propget, helpstring("The playlist that contains this track.")]
    HRESULT _stdcall Playlist([out, retval] IITPlaylist** rhs);
    [id(0x60030005), propget, helpstring("The album containing the track.")]
    HRESULT _stdcall Album([out, retval] BSTR* rhs);
    [id(0x60030005), propput, helpstring("The album containing the track.")]
    HRESULT _stdcall Album([in] BSTR rhs);
    [id(0x60030007), propget, helpstring("The artist/source of the track.")]
    HRESULT _stdcall Artist([out, retval] BSTR* rhs);
    [id(0x60030007), propput, helpstring("The artist/source of the track.")]
    HRESULT _stdcall Artist([in] BSTR rhs);
    [id(0x60030009), propget, helpstring("The bit rate of the track (in kbps).")]
    HRESULT _stdcall BitRate([out, retval] long* rhs);
    [id(0x6003000a), propget, helpstring("The tempo of the track (in beats per minute).")]
    HRESULT _stdcall BPM([out, retval] long* rhs);
    [id(0x6003000a), propput, helpstring("The tempo of the track (in beats per minute).")]
    HRESULT _stdcall BPM([in] long rhs);
    [id(0x6003000c), propget, helpstring("Freeform notes about the track.")]
    HRESULT _stdcall Comment([out, retval] BSTR* rhs);
    [id(0x6003000c), propput, helpstring("Freeform notes about the track.")]
    HRESULT _stdcall Comment([in] BSTR rhs);
    [id(0x6003000e), propget, helpstring("True if this track is from a compilation album.")]
    HRESULT _stdcall Compilation([out, retval] VARIANT_BOOL* rhs);
    [id(0x6003000e), propput, helpstring("True if this track is from a compilation album.")]
    HRESULT _stdcall Compilation([in] VARIANT_BOOL rhs);
    [id(0x60030010), propget, helpstring("The composer of the track.")]
    HRESULT _stdcall Composer([out, retval] BSTR* rhs);
    [id(0x60030010), propput, helpstring("The composer of the track.")]
    HRESULT _stdcall Composer([in] BSTR rhs);
    [id(0x60030012), propget, helpstring("The date the track was added to the playlist.")]
    HRESULT _stdcall DateAdded([out, retval] DATE* rhs);
    [id(0x60030013), propget, helpstring("The total number of discs in the source album.")]
    HRESULT _stdcall DiscCount([out, retval] long* rhs);
    [id(0x60030013), propput, helpstring("The total number of discs in the source album.")]
    HRESULT _stdcall DiscCount([in] long rhs);
    [id(0x60030015), propget, helpstring("The index of the disc containing the track on the source album.")]
    HRESULT _stdcall DiscNumber([out, retval] long* rhs);
    [id(0x60030015), propput, helpstring("The index of the disc containing the track on the source album.")]
    HRESULT _stdcall DiscNumber([in] long rhs);
    [id(0x60030017), propget, helpstring("The length of the track (in seconds).")]
    HRESULT _stdcall Duration([out, retval] long* rhs);
    [id(0x60030018), propget, helpstring("True if the track is checked for playback.")]
    HRESULT _stdcall Enabled([out, retval] VARIANT_BOOL* rhs);
    [id(0x60030018), propput, helpstring("True if the track is checked for playback.")]
    HRESULT _stdcall Enabled([in] VARIANT_BOOL rhs);
    [id(0x6003001a), propget, helpstring("The name of the EQ preset of the track.")]
    HRESULT _stdcall EQ([out, retval] BSTR* rhs);
    [id(0x6003001a), propput, helpstring("The name of the EQ preset of the track.")]
    HRESULT _stdcall EQ([in] BSTR rhs);
    [id(0x6003001c), propput, helpstring("The stop time of the track (in seconds).")]
    HRESULT _stdcall Finish([in] long rhs);
    [id(0x6003001c), propget, helpstring("The stop time of the track (in seconds).")]
    HRESULT _stdcall Finish([out, retval] long* rhs);
    [id(0x6003001e), propget, helpstring("The music/audio genre (category) of the track.")]
    HRESULT _stdcall Genre([out, retval] BSTR* rhs);
    [id(0x6003001e), propput, helpstring("The music/audio genre (category) of the track.")]
    HRESULT _stdcall Genre([in] BSTR rhs);
    [id(0x60030020), propget, helpstring("The grouping (piece) of the track.  Generally used to denote movements within classical work.")]
    HRESULT _stdcall Grouping([out, retval] BSTR* rhs);
    [id(0x60030020), propput, helpstring("The grouping (piece) of the track.  Generally used to denote movements within classical work.")]
    HRESULT _stdcall Grouping([in] BSTR rhs);
    [id(0x60030022), propget, helpstring("A text description of the track.")]
    HRESULT _stdcall KindAsString([out, retval] BSTR* rhs);
    [id(0x60030023), propget, helpstring("The modification date of the content of the track.")]
    HRESULT _stdcall ModificationDate([out, retval] DATE* rhs);
    [id(0x60030024), propget, helpstring("The number of times the track has been played.")]
    HRESULT _stdcall PlayedCount([out, retval] long* rhs);
    [id(0x60030024), propput, helpstring("The number of times the track has been played.")]
    HRESULT _stdcall PlayedCount([in] long rhs);
    [id(0x60030026), propget, helpstring("The date and time the track was last played.  A value of zero means no played date.")]
    HRESULT _stdcall PlayedDate([out, retval] DATE* rhs);
    [id(0x60030026), propput, helpstring("The date and time the track was last played.  A value of zero means no played date.")]
    HRESULT _stdcall PlayedDate([in] DATE rhs);
    [id(0x60030028), propget, helpstring("The play order index of the track in the owner playlist (1-based).")]
    HRESULT _stdcall PlayOrderIndex([out, retval] long* rhs);
    [id(0x60030029), propget, helpstring("The rating of the track (0 to 100).")]
    HRESULT _stdcall Rating([out, retval] long* rhs);
    [id(0x60030029), propput, helpstring("The rating of the track (0 to 100).")]
    HRESULT _stdcall Rating([in] long rhs);
    [id(0x6003002b), propget, helpstring("The sample rate of the track (in Hz).")]
    HRESULT _stdcall SampleRate([out, retval] long* rhs);
    [id(0x6003002c), propget, helpstring("The size of the track (in bytes).")]
    HRESULT _stdcall Size([out, retval] long* rhs);
    [id(0x6003002d), propget, helpstring("The start time of the track (in seconds).")]
    HRESULT _stdcall Start([out, retval] long* rhs);
    [id(0x6003002d), propput, helpstring("The start time of the track (in seconds).")]
    HRESULT _stdcall Start([in] long rhs);
    [id(0x6003002f), propget, helpstring("The length of the track (in MM:SS format).")]
    HRESULT _stdcall Time([out, retval] BSTR* rhs);
    [id(0x60030030), propget, helpstring("The total number of tracks on the source album.")]
    HRESULT _stdcall TrackCount([out, retval] long* rhs);
};


[
    odl,
    uuid(9DD6680B-3EDC-40DB-A771-E6FE4832E34A),
    helpstring("IiTunes Interface"),
    hidden,
    dual,
    oleautomation
]
interface IiTunes : IDispatch {
    [id(0x60020000), helpstring("Reposition to the beginning of the current track or go to the previous track if already at start of current track.")]
    HRESULT _stdcall BackTrack();
    [id(0x60020001), helpstring("Skip forward in a playing track.")]
    HRESULT _stdcall FastForward();
    [id(0x60020002), helpstring("Advance to the next track in the current playlist.")]
    HRESULT _stdcall NextTrack();
    [id(0x60020003), helpstring("Pause playback.")]
    HRESULT _stdcall Pause();
    [id(0x60020004), helpstring("Play the currently targeted track.")]
    HRESULT _stdcall Play();
    [id(0x60020005), helpstring("Play the specified file path, adding it to the library if not already present.")]
    HRESULT _stdcall PlayFile([in] BSTR filePath);
    [id(0x60020006), helpstring("Toggle the playing/paused state of the current track.")]
    HRESULT _stdcall PlayPause();
    [id(0x60020007), helpstring("Return to the previous track in the current playlist.")]
    HRESULT _stdcall PreviousTrack();
    [id(0x60020008), helpstring("Disable fast forward/rewind and resume playback, if playing.")]
    HRESULT _stdcall Resume();
    [id(0x60020009), helpstring("Skip backwards in a playing track.")]
    HRESULT _stdcall Rewind();
    [id(0x6002000a), helpstring("Stop playback.")]
    HRESULT _stdcall Stop();
    [id(0x6002000b), helpstring("Start converting the specified file path.")]
    HRESULT _stdcall ConvertFile(
        [in] BSTR filePath,
        [out, retval] IITOperationStatus** rhs);
    [id(0x6002000c), helpstring("Start converting the specified array of file paths. filePaths can be of type VT_ARRAY|VT_VARIANT, where each entry is a VT_BSTR, or VT_ARRAY|VT_BSTR.  You can also pass a JScript Array object.")]
    HRESULT _stdcall ConvertFiles(
        [in] VARIANT* filePaths,
        [out, retval] IITOperationStatus** rhs);
    [id(0x6002000d), helpstring("Start converting the specified track.  iTrackToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrack.")]
    HRESULT _stdcall ConvertTrack(
        [in] VARIANT* iTrackToConvert,
        [out, retval] IITOperationStatus** rhs);
    [id(0x6002000e), helpstring("Start converting the specified tracks.  iTracksToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrackCollection.")]
    HRESULT _stdcall ConvertTracks(
        [in] VARIANT* iTracksToConvert,
        [out, retval] IITOperationStatus** rhs);
    [id(0x6002000f), helpstring("Returns true if this version of the iTunes type library is compatible with the specified version.")]
    HRESULT _stdcall CheckVersion(
        [in] long majorVersion,
        [in] long minorVersion,
        [out, retval] VARIANT_BOOL* rhs);
    [id(0x60020010), helpstring("Returns an IITObject corresponding to the specified IDs.")]
    HRESULT _stdcall GetITObjectByID(
        [in] long sourceID,
        [in] long playlistID,
        [in] long trackID,
        [in] long databaseID,
        [out, retval] IITObject** rhs);
    [id(0x60020011), helpstring("Creates a new playlist in the main library.")]
    HRESULT _stdcall CreatePlaylist(
        [in] BSTR playlistName,
        [out, retval] IITPlaylist** rhs);
    [id(0x60020012), helpstring("Open the specified iTunes Store or streaming audio URL.")]
    HRESULT _stdcall OpenURL([in] BSTR URL);
    [id(0x60020013), helpstring("Go to the iTunes Store home page.")]
    HRESULT _stdcall GotoMusicStoreHomePage();
    [id(0x60020014), helpstring("Update the contents of the iPod.")]
    HRESULT _stdcall UpdateIPod();
    [id(0x60020015)]
    HRESULT _stdcall Authorize(
        [in] long numElems,
        [in] VARIANT* data,
        [in] BSTR* names);
    [id(0x60020016), helpstring("Exits the iTunes application.")]
    HRESULT _stdcall Quit();
    [id(0x60020017), propget, helpstring("Returns a collection of music sources (music library, CD, device, etc.).")]
    HRESULT _stdcall Sources([out, retval] IITSourceCollection** rhs);
    [id(0x60020018), propget, helpstring("Returns a collection of encoders.")]
    HRESULT _stdcall Encoders([out, retval] IITEncoderCollection** rhs);
    [id(0x60020019), propget, helpstring("Returns a collection of EQ presets.")]
    HRESULT _stdcall EQPresets([out, retval] IITEQPresetCollection** rhs);
    [id(0x6002001a), propget, helpstring("Returns a collection of visual plug-ins.")]
    HRESULT _stdcall Visuals([out, retval] IITVisualCollection** rhs);
    [id(0x6002001b), propget, helpstring("Returns a collection of windows.")]
    HRESULT _stdcall Windows([out, retval] IITWindowCollection** rhs);
    [id(0x6002001c), propget, helpstring("Returns the sound output volume (0 = minimum, 100 = maximum).")]
    HRESULT _stdcall SoundVolume([out, retval] long* rhs);
    [id(0x6002001c), propput, helpstring("Returns the sound output volume (0 = minimum, 100 = maximum).")]
    HRESULT _stdcall SoundVolume([in] long rhs);
    [id(0x6002001e), propget, helpstring("True if sound output is muted.")]
    HRESULT _stdcall Mute([out, retval] VARIANT_BOOL* rhs);
    [id(0x6002001e), propput, helpstring("True if sound output is muted.")]
    HRESULT _stdcall Mute([in] VARIANT_BOOL rhs);
    [id(0x60020020), propget, helpstring("Returns the current player state.")]
    HRESULT _stdcall PlayerState([out, retval] ITPlayerState* rhs);
    [id(0x60020021), propget, helpstring("Returns the player's position within the currently playing track in seconds.")]
    HRESULT _stdcall PlayerPosition([out, retval] long* rhs);
    [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
    HRESULT _stdcall PlayerPosition([in] long rhs);
    [id(0x60020023), propget, helpstring("Returns the currently selected encoder (AAC, MP3, AIFF, WAV, etc.).")]
    HRESULT _stdcall CurrentEncoder([out, retval] IITEncoder** rhs);
    [id(0x60020023), propput, helpstring("Returns the currently selected encoder (AAC, MP3, AIFF, WAV, etc.).")]
    HRESULT _stdcall CurrentEncoder([in] IITEncoder* rhs);
    [id(0x60020025), propget, helpstring("True if visuals are currently being displayed.")]
    HRESULT _stdcall VisualsEnabled([out, retval] VARIANT_BOOL* rhs);
    [id(0x60020025), propput, helpstring("True if visuals are currently being displayed.")]
    HRESULT _stdcall VisualsEnabled([in] VARIANT_BOOL rhs);
    [id(0x60020027), propget, helpstring("True if the visuals are displayed using the entire screen.")]
    HRESULT _stdcall FullScreenVisuals([out, retval] VARIANT_BOOL* rhs);
    [id(0x60020027), propput, helpstring("True if the visuals are displayed using the entire screen.")]
    HRESULT _stdcall FullScreenVisuals([in] VARIANT_BOOL rhs);
    [id(0x60020029), propget, helpstring("Returns the size of the displayed visual.")]
    HRESULT _stdcall VisualSize([out, retval] ITVisualSize* rhs);
    [id(0x60020029), propput, helpstring("Returns the size of the displayed visual.")]
    HRESULT _stdcall VisualSize([in] ITVisualSize rhs);
    [id(0x6002002b), propget, helpstring("Returns the currently selected visual plug-in.")]
    HRESULT _stdcall CurrentVisual([out, retval] IITVisual** rhs);
    [id(0x6002002b), propput, helpstring("Returns the currently selected visual plug-in.")]
    HRESULT _stdcall CurrentVisual([in] IITVisual* rhs);
    [id(0x6002002d), propget, helpstring("True if the equalizer is enabled.")]
    HRESULT _stdcall EQEnabled([out, retval] VARIANT_BOOL* rhs);
    [id(0x6002002d), propput, helpstring("True if the equalizer is enabled.")]
    HRESULT _stdcall EQEnabled([in] VARIANT_BOOL rhs);
    [id(0x6002002f), propget, helpstring("Returns the currently selected EQ preset.")]
    HRESULT _stdcall CurrentEQPreset([out, retval] IITEQPreset** rhs);
    [id(0x6002002f), propput, helpstring("Returns the currently selected EQ preset.")]
    HRESULT _stdcall CurrentEQPreset([in] IITEQPreset* rhs);
    [id(0x60020031), propget, helpstring("The name of the current song in the playing stream (provided by streaming server).")]
    HRESULT _stdcall CurrentStreamTitle([out, retval] BSTR* rhs);
    [id(0x60020032), propget, helpstring("The URL of the playing stream or streaming web site (provided by streaming server).")]
    HRESULT _stdcall CurrentStreamURL([out, retval] BSTR* rhs);
    [id(0x60020033), propget, helpstring("Returns the main iTunes browser window.")]
    HRESULT _stdcall BrowserWindow([out, retval] IITBrowserWindow** rhs);
    [id(0x60020034), propget, helpstring("Returns the EQ window.")]
    HRESULT _stdcall EQWindow([out, retval] IITWindow** rhs);
    [id(0x60020035), propget, helpstring("Returns the source that represents the main library.")]
    HRESULT _stdcall LibrarySource([out, retval] IITSource** rhs);
    [id(0x60020036), propget, helpstring("Returns the main library playlist in the main library source.")]
    HRESULT _stdcall LibraryPlaylist([out, retval] IITLibraryPlaylist** rhs);
    [id(0x60020037), propget, helpstring("Returns the currently targeted track.")]
    HRESULT _stdcall CurrentTrack([out, retval] IITTrack** rhs);
    [id(0x60020038), propget, helpstring("Returns the playlist containing the currently targeted track.")]
    HRESULT _stdcall CurrentPlaylist([out, retval] IITPlaylist** rhs);
    [id(0x60020039), propget, helpstring("Returns a collection containing the currently selected track or tracks.")]
    HRESULT _stdcall SelectedTracks([out, retval] IITTrackCollection** rhs);
    [id(0x6002003a), propget, helpstring("Returns the version of the iTunes application.")]
    HRESULT _stdcall Version([out, retval] BSTR* rhs);
    [id(0x6002003b)]
    HRESULT _stdcall SetOptions([in] long options);
    [id(0x6002003c), helpstring("Start converting the specified file path.")]
    HRESULT _stdcall ConvertFile2(
        [in] BSTR filePath,
        [out, retval] IITConvertOperationStatus** rhs);
    [id(0x6002003d), helpstring("Start converting the specified array of file paths. filePaths can be of type VT_ARRAY|VT_VARIANT, where each entry is a VT_BSTR, or VT_ARRAY|VT_BSTR.  You can also pass a JScript Array object.")]
    HRESULT _stdcall ConvertFiles2(
        [in] VARIANT* filePaths,
        [out, retval] IITConvertOperationStatus** rhs);
    [id(0x6002003e), helpstring("Start converting the specified track.  iTrackToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrack.")]
    HRESULT _stdcall ConvertTrack2(
        [in] VARIANT* iTrackToConvert,
        [out, retval] IITConvertOperationStatus** rhs);
    [id(0x6002003f), helpstring("Start converting the specified tracks.  iTracksToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrackCollection.")]
    HRESULT _stdcall ConvertTracks2(
        [in] VARIANT* iTracksToConvert,
        [out, retval] IITConvertOperationStatus** rhs);
    [id(0x60020040), propget, helpstring("True if iTunes will process APPCOMMAND Windows messages.")]
    HRESULT _stdcall AppCommandMessageProcessingEnabled([out, retval] VARIANT_BOOL* rhs);
    [id(0x60020040), propput, helpstring("True if iTunes will process APPCOMMAND Windows messages.")]
    HRESULT _stdcall AppCommandMessageProcessingEnabled([in] VARIANT_BOOL rhs);
    [id(0x60020042), propget, helpstring("True if iTunes will force itself to be the foreground application when it displays a dialog.")]
    HRESULT _stdcall ForceToForegroundOnDialog([out, retval] VARIANT_BOOL* rhs);
    [id(0x60020042), propput, helpstring("True if iTunes will force itself to be the foreground application when it displays a dialog.")]
    HRESULT _stdcall ForceToForegroundOnDialog([in] VARIANT_BOOL rhs);
    [id(0x60020044), helpstring("Create a new EQ preset.")]
    HRESULT _stdcall CreateEQPreset(
        [in] BSTR eqPresetName,
        [out, retval] IITEQPreset** rhs);
    [id(0x60020045), helpstring("Creates a new playlist in an existing source.")]
    HRESULT _stdcall CreatePlaylistInSource(
        [in] BSTR playlistName,
        [in] VARIANT* iSource,
        [out, retval] IITPlaylist** rhs);
    [id(0x60020046), helpstring("Retrieves the current state of the player buttons.")]
    HRESULT _stdcall GetPlayerButtonsState(
        [out] VARIANT_BOOL* previousEnabled,
        [out] ITPlayButtonState* playPauseStopState,
        [out] VARIANT_BOOL* nextEnabled);
    [id(0x60020047), helpstring("Simulate click on a player control button.")]
    HRESULT _stdcall PlayerButtonClicked(
        [in] ITPlayerButton playerButton,
        [in] long playerButtonModifierKeys);
    [id(0x60020048), propget, helpstring("True if the Shuffle property is writable for the specified playlist.")]
    HRESULT _stdcall CanSetShuffle(
        [in] VARIANT* iPlaylist,
        [out, retval] VARIANT_BOOL* rhs);
    [id(0x60020049), propget, helpstring("True if the SongRepeat property is writable for the specified playlist.")]
    HRESULT _stdcall CanSetSongRepeat(
        [in] VARIANT* iPlaylist,
        [out, retval] VARIANT_BOOL* rhs);
    [id(0x6002004a), propget, helpstring("Returns an IITConvertOperationStatus object if there is currently a conversion in progress.")]
    HRESULT _stdcall ConvertOperationStatus([out, retval] IITConvertOperationStatus** rhs);
    [id(0x6002004b), helpstring("Subscribe to the specified podcast feed URL.")]
    HRESULT _stdcall SubscribeToPodcast([in] BSTR URL);
    [id(0x6002004c), helpstring("Update all podcast feeds.")]
    HRESULT _stdcall UpdatePodcastFeeds();
    [id(0x6002004d), helpstring("Creates a new folder in the main library.")]
    HRESULT _stdcall CreateFolder(
        [in] BSTR folderName,
        [out, retval] IITPlaylist** rhs);
    [id(0x6002004e), helpstring("Creates a new folder in an existing source.")]
    HRESULT _stdcall CreateFolderInSource(
        [in] BSTR folderName,
        [in] VARIANT* iSource,
        [out, retval] IITPlaylist** rhs);
    [id(0x6002004f), propget, helpstring("True if the sound volume control is enabled.")]
    HRESULT _stdcall SoundVolumeControlEnabled([out, retval] VARIANT_BOOL* rhs);
    [id(0x60020050), propget, helpstring("The full path to the current iTunes library XML file.")]
    HRESULT _stdcall LibraryXMLPath([out, retval] BSTR* rhs);
    [id(0x60020051), propget, helpstring("Returns the high 32 bits of the persistent ID of the specified IITObject.")]
    HRESULT _stdcall ITObjectPersistentIDHigh(
        [in] VARIANT* iObject,
        [out, retval] long* rhs);
    [id(0x60020052), propget, helpstring("Returns the low 32 bits of the persistent ID of the specified IITObject.")]
    HRESULT _stdcall ITObjectPersistentIDLow(
        [in] VARIANT* iObject,
        [out, retval] long* rhs);
    [id(0x60020053), helpstring("Returns the high and low 32 bits of the persistent ID of the specified IITObject.")]
    HRESULT _stdcall GetITObjectPersistentIDs(
        [in] VARIANT* iObject,
        [out] long* highID,
        [out] long* lowID);
    [id(0x60020054), propget, helpstring("Returns the player's position within the currently playing track in milliseconds.")]
    HRESULT _stdcall PlayerPositionMS([out, retval] long* rhs);
    [id(0x60020054), propput, helpstring("Returns the player's position within the currently playing track in milliseconds.")]
    HRESULT _stdcall PlayerPositionMS([in] long rhs);
};


[
    uuid(9DD6680B-3EDC-40DB-A771-E6FE4832E34A),
    helpstring("IiTunes Interface"),
    hidden,
    dual
]
dispinterface IiTunes {
    properties:
    methods:
        [id(0x60020000), helpstring("Reposition to the beginning of the current track or go to the previous track if already at start of current track.")]
        void _stdcall BackTrack();
        [id(0x60020001), helpstring("Skip forward in a playing track.")]
        void _stdcall FastForward();
        [id(0x60020002), helpstring("Advance to the next track in the current playlist.")]
        void _stdcall NextTrack();
        [id(0x60020003), helpstring("Pause playback.")]
        void _stdcall Pause();
        [id(0x60020004), helpstring("Play the currently targeted track.")]
        void _stdcall Play();
        [id(0x60020005), helpstring("Play the specified file path, adding it to the library if not already present.")]
        void _stdcall PlayFile([in] BSTR filePath);
        [id(0x60020006), helpstring("Toggle the playing/paused state of the current track.")]
        void _stdcall PlayPause();
        [id(0x60020007), helpstring("Return to the previous track in the current playlist.")]
        void _stdcall PreviousTrack();
        [id(0x60020008), helpstring("Disable fast forward/rewind and resume playback, if playing.")]
        void _stdcall Resume();
        [id(0x60020009), helpstring("Skip backwards in a playing track.")]
        void _stdcall Rewind();
        [id(0x6002000a), helpstring("Stop playback.")]
        void _stdcall Stop();
        [id(0x6002000b), helpstring("Start converting the specified file path.")]
        IITOperationStatus* _stdcall ConvertFile([in] BSTR filePath);
        [id(0x6002000c), helpstring("Start converting the specified array of file paths. filePaths can be of type VT_ARRAY|VT_VARIANT, where each entry is a VT_BSTR, or VT_ARRAY|VT_BSTR.  You can also pass a JScript Array object.")]
        IITOperationStatus* _stdcall ConvertFiles([in] VARIANT* filePaths);
        [id(0x6002000d), helpstring("Start converting the specified track.  iTrackToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrack.")]
        IITOperationStatus* _stdcall ConvertTrack([in] VARIANT* iTrackToConvert);
        [id(0x6002000e), helpstring("Start converting the specified tracks.  iTracksToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrackCollection.")]
        IITOperationStatus* _stdcall ConvertTracks([in] VARIANT* iTracksToConvert);
        [id(0x6002000f), helpstring("Returns true if this version of the iTunes type library is compatible with the specified version.")]
        VARIANT_BOOL _stdcall CheckVersion(
            [in] long majorVersion,
            [in] long minorVersion);
        [id(0x60020010), helpstring("Returns an IITObject corresponding to the specified IDs.")]
        IITObject* _stdcall GetITObjectByID(
            [in] long sourceID,
            [in] long playlistID,
            [in] long trackID,
            [in] long databaseID);
        [id(0x60020011), helpstring("Creates a new playlist in the main library.")]
        IITPlaylist* _stdcall CreatePlaylist([in] BSTR playlistName);
        [id(0x60020012), helpstring("Open the specified iTunes Store or streaming audio URL.")]
        void _stdcall OpenURL([in] BSTR URL);
        [id(0x60020013), helpstring("Go to the iTunes Store home page.")]
        void _stdcall GotoMusicStoreHomePage();
        [id(0x60020014), helpstring("Update the contents of the iPod.")]
        void _stdcall UpdateIPod();
        [id(0x60020015)]
        void _stdcall Authorize(
            [in] long numElems,
            [in] VARIANT* data,
            [in] BSTR* names);
        [id(0x60020016), helpstring("Exits the iTunes application.")]
        void _stdcall Quit();
        [id(0x60020017), propget, helpstring("Returns a collection of music sources (music library, CD, device, etc.).")]
        IITSourceCollection* _stdcall Sources();
        [id(0x60020018), propget, helpstring("Returns a collection of encoders.")]
        IITEncoderCollection* _stdcall Encoders();
        [id(0x60020019), propget, helpstring("Returns a collection of EQ presets.")]
        IITEQPresetCollection* _stdcall EQPresets();
        [id(0x6002001a), propget, helpstring("Returns a collection of visual plug-ins.")]
        IITVisualCollection* _stdcall Visuals();
        [id(0x6002001b), propget, helpstring("Returns a collection of windows.")]
        IITWindowCollection* _stdcall Windows();
        [id(0x6002001c), propget, helpstring("Returns the sound output volume (0 = minimum, 100 = maximum).")]
        long _stdcall SoundVolume();
        [id(0x6002001c), propput, helpstring("Returns the sound output volume (0 = minimum, 100 = maximum).")]
        void _stdcall SoundVolume([in] long rhs);
        [id(0x6002001e), propget, helpstring("True if sound output is muted.")]
        VARIANT_BOOL _stdcall Mute();
        [id(0x6002001e), propput, helpstring("True if sound output is muted.")]
        void _stdcall Mute([in] VARIANT_BOOL rhs);
        [id(0x60020020), propget, helpstring("Returns the current player state.")]
        ITPlayerState _stdcall PlayerState();
        [id(0x60020021), propget, helpstring("Returns the player's position within the currently playing track in seconds.")]
        long _stdcall PlayerPosition();
        [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
        void _stdcall PlayerPosition([in] long rhs);
        [id(0x60020023), propget, helpstring("Returns the currently selected encoder (AAC, MP3, AIFF, WAV, etc.).")]
        IITEncoder* _stdcall CurrentEncoder();
        [id(0x60020023), propput, helpstring("Returns the currently selected encoder (AAC, MP3, AIFF, WAV, etc.).")]
        void _stdcall CurrentEncoder([in] IITEncoder* rhs);
        [id(0x60020025), propget, helpstring("True if visuals are currently being displayed.")]
        VARIANT_BOOL _stdcall VisualsEnabled();
        [id(0x60020025), propput, helpstring("True if visuals are currently being displayed.")]
        void _stdcall VisualsEnabled([in] VARIANT_BOOL rhs);
        [id(0x60020027), propget, helpstring("True if the visuals are displayed using the entire screen.")]
        VARIANT_BOOL _stdcall FullScreenVisuals();
        [id(0x60020027), propput, helpstring("True if the visuals are displayed using the entire screen.")]
        void _stdcall FullScreenVisuals([in] VARIANT_BOOL rhs);
        [id(0x60020029), propget, helpstring("Returns the size of the displayed visual.")]
        ITVisualSize _stdcall VisualSize();
        [id(0x60020029), propput, helpstring("Returns the size of the displayed visual.")]
        void _stdcall VisualSize([in] ITVisualSize rhs);
        [id(0x6002002b), propget, helpstring("Returns the currently selected visual plug-in.")]
        IITVisual* _stdcall CurrentVisual();
        [id(0x6002002b), propput, helpstring("Returns the currently selected visual plug-in.")]
        void _stdcall CurrentVisual([in] IITVisual* rhs);
        [id(0x6002002d), propget, helpstring("True if the equalizer is enabled.")]
        VARIANT_BOOL _stdcall EQEnabled();
        [id(0x6002002d), propput, helpstring("True if the equalizer is enabled.")]
        void _stdcall EQEnabled([in] VARIANT_BOOL rhs);
        [id(0x6002002f), propget, helpstring("Returns the currently selected EQ preset.")]
        IITEQPreset* _stdcall CurrentEQPreset();
        [id(0x6002002f), propput, helpstring("Returns the currently selected EQ preset.")]
        void _stdcall CurrentEQPreset([in] IITEQPreset* rhs);
        [id(0x60020031), propget, helpstring("The name of the current song in the playing stream (provided by streaming server).")]
        BSTR _stdcall CurrentStreamTitle();
        [id(0x60020032), propget, helpstring("The URL of the playing stream or streaming web site (provided by streaming server).")]
        BSTR _stdcall CurrentStreamURL();
        [id(0x60020033), propget, helpstring("Returns the main iTunes browser window.")]
        IITBrowserWindow* _stdcall BrowserWindow();
        [id(0x60020034), propget, helpstring("Returns the EQ window.")]
        IITWindow* _stdcall EQWindow();
        [id(0x60020035), propget, helpstring("Returns the source that represents the main library.")]
        IITSource* _stdcall LibrarySource();
        [id(0x60020036), propget, helpstring("Returns the main library playlist in the main library source.")]
        IITLibraryPlaylist* _stdcall LibraryPlaylist();
        [id(0x60020037), propget, helpstring("Returns the currently targeted track.")]
        IITTrack* _stdcall CurrentTrack();
        [id(0x60020038), propget, helpstring("Returns the playlist containing the currently targeted track.")]
        IITPlaylist* _stdcall CurrentPlaylist();
        [id(0x60020039), propget, helpstring("Returns a collection containing the currently selected track or tracks.")]
        IITTrackCollection* _stdcall SelectedTracks();
        [id(0x6002003a), propget, helpstring("Returns the version of the iTunes application.")]
        BSTR _stdcall Version();
        [id(0x6002003b)]
        void _stdcall SetOptions([in] long options);
        [id(0x6002003c), helpstring("Start converting the specified file path.")]
        IITConvertOperationStatus* _stdcall ConvertFile2([in] BSTR filePath);
        [id(0x6002003d), helpstring("Start converting the specified array of file paths. filePaths can be of type VT_ARRAY|VT_VARIANT, where each entry is a VT_BSTR, or VT_ARRAY|VT_BSTR.  You can also pass a JScript Array object.")]
        IITConvertOperationStatus* _stdcall ConvertFiles2([in] VARIANT* filePaths);
        [id(0x6002003e), helpstring("Start converting the specified track.  iTrackToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrack.")]
        IITConvertOperationStatus* _stdcall ConvertTrack2([in] VARIANT* iTrackToConvert);
        [id(0x6002003f), helpstring("Start converting the specified tracks.  iTracksToConvert is a VARIANT of type VT_DISPATCH that points to an IITTrackCollection.")]
        IITConvertOperationStatus* _stdcall ConvertTracks2([in] VARIANT* iTracksToConvert);
        [id(0x60020040), propget, helpstring("True if iTunes will process APPCOMMAND Windows messages.")]
        VARIANT_BOOL _stdcall AppCommandMessageProcessingEnabled();
        [id(0x60020040), propput, helpstring("True if iTunes will process APPCOMMAND Windows messages.")]
        void _stdcall AppCommandMessageProcessingEnabled([in] VARIANT_BOOL rhs);
        [id(0x60020042), propget, helpstring("True if iTunes will force itself to be the foreground application when it displays a dialog.")]
        VARIANT_BOOL _stdcall ForceToForegroundOnDialog();
        [id(0x60020042), propput, helpstring("True if iTunes will force itself to be the foreground application when it displays a dialog.")]
        void _stdcall ForceToForegroundOnDialog([in] VARIANT_BOOL rhs);
        [id(0x60020044), helpstring("Create a new EQ preset.")]
        IITEQPreset* _stdcall CreateEQPreset([in] BSTR eqPresetName);
        [id(0x60020045), helpstring("Creates a new playlist in an existing source.")]
        IITPlaylist* _stdcall CreatePlaylistInSource(
            [in] BSTR playlistName,
            [in] VARIANT* iSource);
        [id(0x60020046), helpstring("Retrieves the current state of the player buttons.")]
        void _stdcall GetPlayerButtonsState(
            [out] VARIANT_BOOL* previousEnabled,
            [out] ITPlayButtonState* playPauseStopState,
            [out] VARIANT_BOOL* nextEnabled);
        [id(0x60020047), helpstring("Simulate click on a player control button.")]
        void _stdcall PlayerButtonClicked(
            [in] ITPlayerButton playerButton,
            [in] long playerButtonModifierKeys);
        [id(0x60020048), propget, helpstring("True if the Shuffle property is writable for the specified playlist.")]
        VARIANT_BOOL _stdcall CanSetShuffle([in] VARIANT* iPlaylist);
        [id(0x60020049), propget, helpstring("True if the SongRepeat property is writable for the specified playlist.")]
        VARIANT_BOOL _stdcall CanSetSongRepeat([in] VARIANT* iPlaylist);
        [id(0x6002004a), propget, helpstring("Returns an IITConvertOperationStatus object if there is currently a conversion in progress.")]
        IITConvertOperationStatus* _stdcall ConvertOperationStatus();
        [id(0x6002004b), helpstring("Subscribe to the specified podcast feed URL.")]
        void _stdcall SubscribeToPodcast([in] BSTR URL);
        [id(0x6002004c), helpstring("Update all podcast feeds.")]
        void _stdcall UpdatePodcastFeeds();
        [id(0x6002004d), helpstring("Creates a new folder in the main library.")]
        IITPlaylist* _stdcall CreateFolder([in] BSTR folderName);
        [id(0x6002004e), helpstring("Creates a new folder in an existing source.")]
        IITPlaylist* _stdcall CreateFolderInSource(
            [in] BSTR folderName,
            [in] VARIANT* iSource);
        [id(0x6002004f), propget, helpstring("True if the sound volume control is enabled.")]
        VARIANT_BOOL _stdcall SoundVolumeControlEnabled();
        [id(0x60020050), propget, helpstring("The full path to the current iTunes library XML file.")]
        BSTR _stdcall LibraryXMLPath();
        [id(0x60020051), propget, helpstring("Returns the high 32 bits of the persistent ID of the specified IITObject.")]
        long _stdcall ITObjectPersistentIDHigh([in] VARIANT* iObject);
        [id(0x60020052), propget, helpstring("Returns the low 32 bits of the persistent ID of the specified IITObject.")]
        long _stdcall ITObjectPersistentIDLow([in] VARIANT* iObject);
        [id(0x60020053), helpstring("Returns the high and low 32 bits of the persistent ID of the specified IITObject.")]
        void _stdcall GetITObjectPersistentIDs(
            [in] VARIANT* iObject,
            [out] long* highID,
            [out] long* lowID);
        [id(0x60020054), propget, helpstring("Returns the player's position within the currently playing track in milliseconds.")]
        long _stdcall PlayerPositionMS();
        [id(0x60020054), propput, helpstring("Returns the player's position within the currently playing track in milliseconds.")]
        void _stdcall PlayerPositionMS([in] long rhs);
};


[
    uuid(5846EB78-317E-4B6F-B0C3-11EE8C8FEEF2),
    helpstring("_IiTunesEvents Interface")
]
dispinterface _IiTunesEvents {
    properties:
    methods:
        [id(0x00000001), helpstring("Fired when a database change occurs.")]
        HRESULT OnDatabaseChangedEvent(
            [in] VARIANT deletedObjectIDs,
            [in] VARIANT changedObjectIDs);
        [id(0x00000002), helpstring("Fired when a track has started playing.")]
        HRESULT OnPlayerPlayEvent([in] VARIANT iTrack);
        [id(0x00000003), helpstring("Fired when a track has stopped playing.")]
        HRESULT OnPlayerStopEvent([in] VARIANT iTrack);
        [id(0x00000004), helpstring("Fired when information about the currently playing track has changed.")]
        HRESULT OnPlayerPlayingTrackChangedEvent([in] VARIANT iTrack);
        [id(0x00000005), helpstring("Fired when the iTunes user interface is no longer disabled.")]
        HRESULT OnUserInterfaceEnabledEvent();
        [id(0x00000006), helpstring("Fired when calls to the iTunes COM interface will be deferred.")]
        HRESULT OnCOMCallsDisabledEvent([in] ITCOMDisabledReason reason);
        [id(0x00000007), helpstring("Fired when calls to the iTunes COM interface will no longer be deferred.")]
        HRESULT OnCOMCallsEnabledEvent();
        [id(0x00000008), helpstring("Fired when iTunes is about to quit.")]
        HRESULT OnQuittingEvent();
        [id(0x00000009), helpstring("Fired when iTunes is about to prompt the user to quit.")]
        HRESULT OnAboutToPromptUserToQuitEvent();
        [id(0x0000000a), helpstring("Fired when the sound output volume has changed.")]
        HRESULT OnSoundVolumeChangedEvent([in] long newVolume);
};
