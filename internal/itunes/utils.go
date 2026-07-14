//go:build windows

package itunes

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"github.com/charmbracelet/log"
	"path"
	"syscall"
	"unsafe"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetCurrentTrack(dispatcher *ole.IDispatch) (*IiTrack, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}

	trackProp, err := oleutil.GetProperty(dispatcher, "CurrentTrack")
	if err != nil {
		return nil, err
	}
	defer trackProp.Clear()

	trackDispatcher := trackProp.ToIDispatch()
	if trackDispatcher == nil {
		return nil, nil
	}

	track, err := getCOMObject[IiTrack](trackDispatcher, IID_IiTrack)
	return track, err
}

func getInt32Property(dispatcher *ole.IDispatch, property string, params ...any) (int32, error) {
	variant, err := oleutil.GetProperty(dispatcher, property, params...)
	if err != nil {
		return 0, err
	}
	defer variant.Clear()

	v, ok := variant.Value().(int32)
	if !ok {
		return 0, fmt.Errorf("property %q did not return int32", property)
	}
	return v, nil
}

func GetCurrentTunes(dispatcher *ole.IDispatch) (*IiTunes, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}

	soundVolume, err := getInt32Property(dispatcher, "SoundVolume")
	if err != nil {
		return nil, err
	}

	playerPosition, err := getInt32Property(dispatcher, "PlayerPosition")
	if err != nil {
		return nil, err
	}

	playerPositionMS, err := getInt32Property(dispatcher, "PlayerPositionMS")
	if err != nil {
		return nil, err
	}

	playerState, err := getInt32Property(dispatcher, "PlayerState")
	if err != nil {
		return nil, err
	}

	tunes := &IiTunes{
		SoundVolume:      soundVolume,
		PlayerPosition:   playerPosition,
		PlayerPositionMS: playerPositionMS,
		PlayerState:      playerState,
	}

	return tunes, nil
}

var mu sync.Mutex

func SaveArtworkIfAvaliable(trackDispatcher *ole.IDispatch, track *IiTrack) (dosFilePath string, err error) {
	mu.Lock()
	defer mu.Unlock()
	
	artworkCollection, err := oleutil.GetProperty(trackDispatcher, "Artwork")
	if err != nil {
		return "", err
	}

	artworkCollectionDispatcher := artworkCollection.ToIDispatch()
	if artworkCollectionDispatcher == nil {
		return "", nil
	}

	count, err := getInt32Property(artworkCollectionDispatcher, "Count")
	if err != nil {
		return "", err
	}

	log.Debug("artwork count", count)

	if count < 1 {
		log.Debug("no artwork available", count)
		return "", nil
	}
	
	item, err := oleutil.GetProperty(
		artworkCollectionDispatcher,
		"Item",
		1,
	)
	if err != nil {
		return "", err
	}

	itemDispatcher := item.ToIDispatch()
	if itemDispatcher == nil {
		return "", nil
	}

	artworkFormat, err := getInt32Property(itemDispatcher, "Format")
	if err != nil {
		return "", err
	}

	// 0 = Unknown, 1 = JPEG, 2 = PNG, 3 = BMP
	fileSuffix := ""

	switch artworkFormat {
	case 1:
		fileSuffix = ".jpg"
	case 2:
		fileSuffix = ".png"
	case 3:
		fileSuffix = ".bmp"
	}

	tmpDir := path.Join("Z:", "tmp", "tunesbus")
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return "", err
	}

	artworkPath := path.Join(tmpDir, fmt.Sprintf("tmp-%d%s", track.TrackID, fileSuffix))
	log.Info("artwork path", artworkPath)
	fileInfo, err := os.Stat(artworkPath)
	
	// LOL https://stackoverflow.com/questions/56803469/join-paths-with-backslash-separator-independent-of-the-underlying-os-with-the-st
	if err != nil {
		artworkPath = strings.ReplaceAll(artworkPath, "/", "\\")
		log.Info("windows sep", artworkPath)
		r, err := oleutil.CallMethod(
			itemDispatcher,
			"SaveArtworkToFile",
			artworkPath,
		)
		defer r.Clear()
		log.Info("successfully saved artwork", artworkPath)

		if err != nil {
			return "", err
		}
		return artworkPath, nil
	}
	// apparently you can't release/clear everything here....????
	defer artworkCollectionDispatcher.Release()
	defer artworkCollection.Clear()
	// defer item.Clear()
	// defer itemDispatcher.Release()

	if !fileInfo.IsDir() {
		artworkPath = strings.ReplaceAll(artworkPath, "/", "\\")
		log.Info("windows sep", artworkPath)
		return artworkPath, nil
	}
	return "", errors.New("artworkPath is a directory instead of a file...?")
}

func SetTunesPosition(dispatcher *ole.IDispatch, seconds int64) (err error) {
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// HRESULT _stdcall PlayerPosition([in] long rhs);
	const PlayerPositionDispId = 0x60020021
	const DISPATCH_PROPERTYPUT = 0x4

	_, err = dispatcher.Invoke(PlayerPositionDispId, DISPATCH_PROPERTYPUT, float64(seconds))
	return err
}

func GetPlayerButtonsState(dispatcher *ole.IDispatch) (prevEnabled bool, state int32, nextEnabled bool, err error) {
	var next int16
	var prev int16

	args := []ole.VARIANT{
		{
			VT:  ole.VT_BYREF | ole.VT_BOOL,
			Val: int64(uintptr(unsafe.Pointer(&next))),
		},
		{
			VT:  ole.VT_BYREF | ole.VT_I4,
			Val: int64(uintptr(unsafe.Pointer(&state))),
		},
		{
			VT:  ole.VT_BYREF | ole.VT_BOOL,
			Val: int64(uintptr(unsafe.Pointer(&prev))),
		},
	}

	params := dispParams{
		cArgs:  uint32(len(args)),
		rgvarg: uintptr(unsafe.Pointer(&args[0])),
	}

	var result ole.VARIANT
	defer result.Clear()
	var excep ole.EXCEPINFO
	var argErr uint32

	vtbl := dispatcher.VTable()

	hr, _, _ := syscall.SyscallN(
		vtbl.Invoke,
		uintptr(unsafe.Pointer(dispatcher)),    // this
		uintptr(0x60020046),                    // DISPID GetPlayerButtonsState
		uintptr(unsafe.Pointer(&ole.IID_NULL)), // riid
		uintptr(0),                             // lcid
		uintptr(ole.DISPATCH_METHOD),           // flags
		uintptr(unsafe.Pointer(&params)),       // DISPPARAMS
		uintptr(unsafe.Pointer(&result)),       // retval (unused)
		uintptr(unsafe.Pointer(&excep)),        // EXCEPINFO
		uintptr(unsafe.Pointer(&argErr)),       // argErr
	)

	if hr != 0 {
		return false, 0, false, nil
	}

	//log.Printf("prev is %d, state is %d, next is %d", prev, state, next)

	return prev != 0, state, next != 0, nil
}

func SafeGetCurrentPlaylist(tunesDispatcher *ole.IDispatch) (*ole.IDispatch, error) {
	// safety check
	track, err := GetCurrentTrack(tunesDispatcher)
	if err != nil {
		return nil, nil
	}
	if track == nil {
		return nil, nil
	}
	if track.Dispatcher != nil {
		track.Dispatcher.Release()
	}

	currentPlaylist, err := oleutil.GetProperty(tunesDispatcher, "CurrentPlaylist")
	if err != nil {
		log.Error("failed to get current playlist", currentPlaylist)
		return nil, err
	}
	defer currentPlaylist.Clear()

	playlistDispatcher := currentPlaylist.ToIDispatch()

	if playlistDispatcher != nil {
		playlistDispatcher.AddRef()
		return playlistDispatcher, nil
	}
	return nil, nil
}
