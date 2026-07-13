//go:build windows

package itunes

import (
	"errors"
	"fmt"
	"time"

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
	track, err := getCOMObjectFromVariant[IiTrack](trackProp, IID_IiTrack)
	return track, err
}

func GetCurrentTunes(dispatcher *ole.IDispatch) (*IiTunes, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}

	soundVolumeVar, err := oleutil.GetProperty(dispatcher, "SoundVolume")
	if err != nil {
		return nil, err
	}
	soundVolume := soundVolumeVar.Value().(int32)

	playerPositionVar, err := oleutil.GetProperty(dispatcher, "PlayerPosition")
	if err != nil {
		return nil, err
	}
	playerPosition := playerPositionVar.Value().(int32)

	playerPositionMSVar, err := oleutil.GetProperty(dispatcher, "PlayerPositionMS")
	if err != nil {
		return nil, err
	}
	playerPositionMS := playerPositionMSVar.Value().(int32)

	playerStateVar, err := oleutil.GetProperty(dispatcher, "PlayerState")
	if err != nil {
		return nil, err
	}
	playerState := playerStateVar.Value().(int32)

	tunes := &IiTunes{
		SoundVolume:      soundVolume,
		PlayerPosition:   playerPosition,
		PlayerPositionMS: playerPositionMS,
		PlayerState:      playerState,
	}

	return tunes, err
}

func SaveArtworkIfAvaliable(trackDispatcher *ole.IDispatch, track *IiTrack) (dosFilePath string, err error) {
	artworkPath := ""

	artworkCollection, err := oleutil.GetProperty(trackDispatcher, "Artwork")
	if err != nil {
		return artworkPath, err
	}
	count, err := oleutil.GetProperty(artworkCollection.ToIDispatch(), "Count")
	if err != nil {
		return artworkPath, err
	}
	log.Debug("artwork count: %d", count.Value())

	if count.Value().(int32) > 0 {
		item, err := oleutil.GetProperty(
			artworkCollection.ToIDispatch(),
			"Item",
			1,
		)
		if err != nil {
			return artworkPath, err
		}

		artworkFormat, err := oleutil.GetProperty(item.ToIDispatch(), "Format")
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
			<-time.After(200 * time.Millisecond)
			//os.Mkdir("C:\\Temp", 0755)
			// :)
			artworkPath = path.Join("C:\\", fmt.Sprintf("%d%s", track.TrackID, fileSuffix))
			log.Debug("successfully saved artwork", artworkPath)

			artwork := item.ToIDispatch()
			r, err := oleutil.CallMethod(
				artwork,
				"SaveArtworkToFile",
				artworkPath,
			)
			if err != nil {
				return "", err
			}
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
	const DISPID_PROPERTYPUT = -3

	value := ole.VARIANT{}
	value.VT = ole.VT_R8
	value.Val = seconds

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
	_, err := GetCurrentTrack(tunesDispatcher)
	if err != nil {
		return nil, nil
	}
	
	currentPlaylist, err := oleutil.GetProperty(tunesDispatcher, "CurrentPlaylist")
	if err != nil {
		log.Error("failed to get current playlist", currentPlaylist)
		return nil, err
	}
	playlistDispatcher := currentPlaylist.ToIDispatch()

	if playlistDispatcher != nil {
		return playlistDispatcher, nil
	}
	return nil, nil
}
