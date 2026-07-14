//go:build windows

package itunes

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/log"

	//"os"
	"path"
	"syscall"
	"unsafe"

	//"log"

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

func getInt32Property(dispatcher *ole.IDispatch, property string) (int32, error) {
	variant, err := oleutil.GetProperty(dispatcher, property)
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

func SaveArtworkIfAvaliable(trackDispatcher *ole.IDispatch, track *IiTrack) (dosFilePath string, err error) {
	artworkPath := ""

	artworkCollection, err := oleutil.GetProperty(trackDispatcher, "Artwork")
	if err != nil {
		return artworkPath, err
	}
	defer artworkCollection.Clear()

	artworkCollectionDispatcher := artworkCollection.ToIDispatch()
	if artworkCollectionDispatcher == nil {
		return artworkPath, nil
	}

	count, err := oleutil.GetProperty(artworkCollectionDispatcher, "Count")
	if err != nil {
		return artworkPath, err
	}
	defer count.Clear()

	log.Debug("artwork count: %d", count.Value())

	if count.Value().(int32) > 0 {
		item, err := oleutil.GetProperty(
			artworkCollectionDispatcher,
			"Item",
			1,
		)
		if err != nil {
			return artworkPath, err
		}
		defer item.Clear()

		itemDispatcher := item.ToIDispatch()
		if itemDispatcher == nil {
			return artworkPath, nil
		}

		artworkFormat, err := oleutil.GetProperty(itemDispatcher, "Format")
		if err != nil {
			return artworkPath, err
		}
		defer artworkFormat.Clear()

		// 0 = Unknown, 1 = JPEG, 2 = PNG, 3 = BMP
		fileSuffix := ""

		switch artworkFormat.Value().(int32) {
		case 1:
			fileSuffix = ".jpg"
		case 2:
			fileSuffix = ".png"
		case 3:
			fileSuffix = ".bmp"
		}

		//tmpDir, err := os.MkdirTemp("", "tunesbus")
		//log.Printf("dir temp is %s", tmpDir)

		if err == nil {
			// TODO: check if the file already exists
			//os.Mkdir("C:\\Temp", 0755)
			// :)
			artworkPath = path.Join("C:\\", fmt.Sprintf("%d%s", track.TrackID, fileSuffix))
			log.Debug("successfully saved artwork", artworkPath)

			r, err := oleutil.CallMethod(
				itemDispatcher,
				"SaveArtworkToFile",
				artworkPath,
			)
			if err != nil {
				return "", err
			}
			defer r.Clear()
			log.Debug("result for artwork", r)
		}
	}
	return artworkPath, err
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
