//go:build windows

package itunes

import (
	"errors"
	"fmt"
	"log"
	"time"

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
	//log.Printf("called GetCurrentTrack")
	var err error = nil
	trackProp, err := oleutil.GetProperty(dispatcher, "CurrentTrack")
	track, err := getCOMObjectFromVariant[IiTrack](trackProp, IID_IiTrack)
	return track, err
}

func GetCurrentTunes(dispatcher *ole.IDispatch) (*IiTunes, error) {
	if dispatcher == nil {
		return nil, errors.New("dispatcher is not ready")
	}

	var err error = nil

	soundVolumeVar, err := oleutil.GetProperty(dispatcher, "SoundVolume")
	soundVolume := int(soundVolumeVar.Val)

	playerPositionVar, err := oleutil.GetProperty(dispatcher, "PlayerPosition")
	playerPosition := int(playerPositionVar.Val)

	playerPositionMSVar, err := oleutil.GetProperty(dispatcher, "PlayerPositionMS")
	playerPositionMS := int(playerPositionMSVar.Val)

	playerStateVar, err := oleutil.GetProperty(dispatcher, "PlayerState")
	playerState := int(playerStateVar.Val)

	tunes := &IiTunes{
		SoundVolume:      int64(soundVolume),
		PlayerPosition:   int32(playerPosition),
		PlayerPositionMS: int64(playerPositionMS),
		PlayerState:      int32(playerState),
	}

	return tunes, err
}

func SaveArtworkIfAvaliable(trackDispatcher *ole.IDispatch, track *IiTrack) (dosFilePath string, err error) {
	artworkPath := ""
	
	artworkCollection, err := oleutil.GetProperty(trackDispatcher, "Artwork")
	if err != nil {
		return artworkPath, err
	}
	log.Printf("VT = %#x", artworkCollection.VT)
	count, err := oleutil.GetProperty(artworkCollection.ToIDispatch(), "Count")
	if err != nil {
		return artworkPath, err
	}
	log.Printf("artwork count: %d", count.Value())

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
			<-time.After(2 * time.Second)
			//os.Mkdir("C:\\Temp", 0755)
			// :)
			artworkPath = path.Join("C:\\", fmt.Sprintf("%d%s", track.TrackID, fileSuffix))
			log.Printf("saved artwork at %s", artworkPath)
			
			artwork := item.ToIDispatch()
			r, err := oleutil.CallMethod(
				artwork,
				"SaveArtworkToFile",
				artworkPath,
			); if err != nil {
				return "", err
			}
			log.Printf("result for artwork: %v", r)
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
