//go:build windows

package itunes

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"tunesbus/internal/olejunk"
	"tunesbus/internal/wine"
	"unsafe"

	"github.com/charmbracelet/log"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetCurrentTrack(disp *ole.IDispatch) (*IiTrack, error) {
	if disp != nil {
		trackProp, err := oleutil.GetProperty(disp, "CurrentTrack")
		if err != nil {
			return nil, err
		}
		defer trackProp.Clear()
	
		trackDispatcher := trackProp.ToIDispatch()
		if trackDispatcher == nil {
			return nil, nil
		}
		trackDispatcher.AddRef()
		defer trackDispatcher.Release()
		
		track, err := olejunk.GetCOMObject[IiTrack](trackDispatcher, IID_IiTrack)
		return track, err
	}
	return nil, errors.New("disp is not ready")
}

func GetCurrentTunes(disp *ole.IDispatch) (*IiTunes, error) {
	if disp != nil {
		var result IiTunes
		if err := olejunk.UnmarshalCOM(disp, &result); err != nil {
			return nil, err
		}
	
		return &result, nil
	}
	return nil, errors.New("disp is not ready")

}

var mu sync.Mutex

func SaveArtworkIfAvaliable(trackDisp *ole.IDispatch, track *IiTrack) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	artworkCollection, err := oleutil.GetProperty(trackDisp, "Artwork")
	defer artworkCollection.Clear()
	if err != nil {
		artworkCollectionDispatcher := artworkCollection.ToIDispatch()
		
		if artworkCollectionDispatcher != nil {
			artworkCollectionDispatcher.AddRef()
			defer artworkCollectionDispatcher.Release()

			count, err := olejunk.GetPropertyFromIDispatch[int32](artworkCollectionDispatcher, "Count")
			if err != nil {
				return "", err
			}

			log.Debug("artwork count", *count)

			if *count < 1 {
				log.Debug("no artwork available", count)
				return "", nil
			}

			item, err := oleutil.GetProperty(
				artworkCollectionDispatcher,
				"Item",
				1,
			)
			defer item.Clear()

			if err != nil {
				return "", err
			}

			itemDispatcher := item.ToIDispatch()
			if itemDispatcher == nil {
				return "", nil
			}
			itemDispatcher.AddRef()
			defer itemDispatcher.Release()

			artworkFormat, err := olejunk.GetPropertyFromIDispatch[ArtworkFormat](itemDispatcher, "Format")
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

			// LOL https://stackoverflow.com/questions/56803469/join-paths-with-backslash-separator-independent-of-the-underlying-os-with-the-st
			tmpDir, err := wine.UnixTmpDirAsDosPath()
			if err != nil {
				return "", err
			}
			busTmpDir := wine.WindowsPathJoin(tmpDir, "tunesbus")
			if err := os.MkdirAll(busTmpDir, 0o755); err != nil {
				return "", err
			}

			artworkPath := wine.WindowsPathJoin(busTmpDir, fmt.Sprintf("tmp-%d%s", track.TrackID, fileSuffix))
			log.Info("artwork path", artworkPath)
			
			fileInfo, err := os.Stat(artworkPath)

			if err != nil {
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

			if !fileInfo.IsDir() {
				artworkPath = strings.ReplaceAll(artworkPath, "/", "\\")
				log.Info("windows sep", artworkPath)
				return artworkPath, nil
			}
			return "", errors.New("artworkPath is a directory instead of a file...?")
		}
	}
	return "", nil
}

func SetTunesPosition(dispatcher *ole.IDispatch, seconds int64) (err error) {
	// [id(0x60020021), propput, helpstring("Returns the player's position within the currently playing track in seconds.")]
	// HRESULT _stdcall PlayerPosition([in] long rhs);
	const PlayerPositionDispId = 0x60020021
	_, err = dispatcher.Invoke(PlayerPositionDispId, ole.DISPATCH_PROPERTYPUT, float64(seconds))
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

	return prev != 0, state, next != 0, nil
}

func SafeGetCurrentPlaylist(tunesDisp *ole.IDispatch) (*ole.IDispatch, error) {
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
	
		playlistDispatcher := currentPlaylist.ToIDispatch()
	
		if playlistDispatcher != nil {
			// Increments the reference count for an interface pointer to a COM object.
			// You should call this method whenever you make a copy of an interface pointer
			// https://learn.microsoft.com/en-us/windows/win32/api/unknwn/nf-unknwn-iunknown-addref
			playlistDispatcher.AddRef()
			return playlistDispatcher, nil
		}
	}
	return nil, nil
}
