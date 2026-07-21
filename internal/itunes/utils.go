//go:build windows

package itunes

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"tunesbus/internal/olejunk"
	"tunesbus/internal/wine"
	"unsafe"

	"github.com/charmbracelet/log"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetCurrentTrack(disp *ole.IDispatch) (*IiTrackData, error) {
	releaser := olejunk.NewOleReleaser()
	defer releaser.Release()
	
	if disp != nil {
		trackProp, err := oleutil.GetProperty(disp, "CurrentTrack")
		if err != nil {
			return nil, err
		}

		trackDispatch := trackProp.ToIDispatch()
		if trackDispatch == nil {
			return nil, nil
		}
		releaser.Add(&trackDispatch.IUnknown)

		track, err := olejunk.GetCOMObject[IiTrackData](trackDispatch, IID_IiTrack, releaser)
		return track, err
	}
	return nil, errors.New("disp is not ready")
}

// only used for getting properties such as volume and player position
func GetCurrentTunes(disp *ole.IDispatch) (*IiTunes, error) {
	if disp == nil {
		return nil, errors.New("disp is not ready")
	}

	properties, err := olejunk.GetPropertiesFromIDispatch[int32](disp, []string{
		"SoundVolume",
		"PlayerPosition",
		"PlayerPositionMS",
		"PlayerState",
	})
	if err != nil {
		return nil, err
	}

	return &IiTunes{
		SoundVolume:      *properties["SoundVolume"],
		PlayerPosition:   *properties["PlayerPosition"],
		PlayerPositionMS: *properties["PlayerPositionMS"],
		PlayerState:      ITPlayerState(*properties["PlayerState"]),
	}, nil
}


func SaveArtworkIfAvaliable(trackDisp *ole.IDispatch, trackId int32, releaser *olejunk.OleReleaser) (string, error) {
	artworkCollection, err := oleutil.GetProperty(trackDisp, "Artwork")
	if err != nil {
		return "", err
	}

	artworkCollectionDisp := artworkCollection.ToIDispatch()
	if artworkCollectionDisp == nil {
		return "", nil
	}
	releaser.Add(&artworkCollectionDisp.IUnknown)

	count, err := olejunk.GetPropertyFromIDispatch[int32](artworkCollectionDisp, "Count")
	if err != nil {
		return "", err
	}
	if *count < 1 {
		return "", nil
	}

	item, err := oleutil.GetProperty(artworkCollectionDisp, "Item", 1)
	if err != nil {
		return "", err
	}
	defer item.Clear()

	itemDisp := item.ToIDispatch()
	if itemDisp == nil {
		return "", nil
	}
	releaser.Add(&itemDisp.IUnknown)

	artworkFormat, err := olejunk.GetPropertyFromIDispatch[ArtworkFormat](itemDisp, "Format")
	if err != nil {
		return "", err
	}
	fileSuffix := ""
	switch *artworkFormat {
	case JPEG:
		fileSuffix = ".jpg"
	case PNG:
		fileSuffix = ".png"
	case BMP:
		fileSuffix = ".bmp"
	}

	tmpDir, err := wine.UnixTmpDirAsDosPath()
	if err != nil {
		return "", err
	}
	busTmpDir := wine.WindowsPathJoin(tmpDir, "tunesbus")
	if err := os.MkdirAll(busTmpDir, 0o755); err != nil {
		return "", err
	}
	artworkPath := wine.WindowsPathJoin(busTmpDir, fmt.Sprintf("tmp-%d%s", trackId, fileSuffix))

	fileInfo, statErr := os.Stat(artworkPath)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return "", statErr
		}
		r, err := oleutil.CallMethod(itemDisp, "SaveArtworkToFile", artworkPath)
		if r != nil {
			defer r.Clear()
		}
		if err != nil {
			return "", err
		}
		log.Info("saved artwork", "path", artworkPath)
		return artworkPath, nil
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("artwork path is a directory: %s", artworkPath)
	}
	return artworkPath, nil
}

func SetTunesPosition(disp *ole.IDispatch, seconds int64) (err error) {
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// HRESULT _stdcall PlayerPosition([in] long rhs);
	const PlayerPositionDispId = 0x60020021
	_, err = disp.Invoke(PlayerPositionDispId, ole.DISPATCH_PROPERTYPUT, float64(seconds))
	return err
}

func GetPlayerButtonsState(disp *ole.IDispatch) (prevEnabled bool, state int32, nextEnabled bool, err error) {
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

	vtbl := disp.VTable()

	hr, _, _ := syscall.SyscallN(
		vtbl.Invoke,
		uintptr(unsafe.Pointer(disp)),          // this
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

	return prev != 0, state, next != 0, nil
}

func SafeGetCurrentPlaylist(tunesDisp *ole.IDispatch, releaser *olejunk.OleReleaser) (*ole.IDispatch, error) {
	// safety check
	track, err := GetCurrentTrack(tunesDisp)
	if err != nil {
		return nil, err
	}
	if track != nil {
		currentPlaylist, err := oleutil.GetProperty(tunesDisp, "CurrentPlaylist")
		if err != nil {
			log.Error("failed to get current playlist", "currentPlaylist", currentPlaylist)
			return nil, err
		}
		defer currentPlaylist.Clear()

		playlistDisp := currentPlaylist.ToIDispatch()

		if playlistDisp != nil {
			releaser.Add(&playlistDisp.IUnknown)
			return playlistDisp, nil
		}
	}
	return nil, nil
}
