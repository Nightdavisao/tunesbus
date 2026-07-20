//go:build windows

package itunes

import (
	"errors"
	"runtime"
	olejunk "tunesbus/internal/olejunk"

	"syscall"
	"unsafe"

	"github.com/charmbracelet/log"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type eventReceiver struct {
	lpVtbl *eventReceiverVtbl
	ref    int32
	host   *ole.IDispatch
}

type eventReceiverVtbl struct {
	ole.IUnknownVtbl
	GetTypeInfoCount uintptr
	GetTypeInfo      uintptr
	GetIDsOfNames    uintptr
	Invoke           uintptr
}

type COMEventCallback struct {
	handler TunesEventHandler
}

func queryInterface(this *ole.IUnknown, iid *ole.GUID, punk **ole.IUnknown) uintptr {
	s, _ := ole.StringFromCLSID(iid)

	*punk = nil
	if ole.IsEqualGUID(iid, ole.IID_IUnknown) ||
		ole.IsEqualGUID(iid, ole.IID_IDispatch) {
		addRef(this)
		*punk = this
		return ole.S_OK
	}
	if s == IID_IiTunesEvents {
		addRef(this)
		*punk = this
		return ole.S_OK
	}
	return ole.E_NOINTERFACE
}

func addRef(this *ole.IUnknown) uintptr {
	ptr := unsafe.Pointer(this)
	olejunk.PtrCache.Add(ptr)
	
	pthis := (*eventReceiver)(ptr)
	pthis.ref++
	return uintptr(pthis.ref)
}

func release(this *ole.IUnknown) uintptr {
	ptr := unsafe.Pointer(this)
	olejunk.PtrCache.Add(ptr)
	
	pthis := (*eventReceiver)(ptr)
	pthis.ref--
	return uintptr(pthis.ref)
}

func getIDsOfNames(this *ole.IUnknown, iid *ole.GUID, wnames **uint16, namelen int, lcid int, pdisp *int32) uintptr {
	return ole.E_NOTIMPL
}

func getTypeInfoCount(pcount *int) uintptr {
	if pcount != nil {
		*pcount = 0
	}
	return ole.S_OK
}

func getTypeInfo(ptypeif *uintptr) uintptr {
	*ptypeif = uintptr(0)
	return ole.E_NOTIMPL
}

func (ev *COMEventCallback) invoke(this *ole.IDispatch, dispid int, riid *ole.GUID, lcid int, flags int16, dispparams *ole.DISPPARAMS, result *ole.VARIANT, pexcepinfo *ole.EXCEPINFO, nerr *uint) uintptr {
	ptr := unsafe.Pointer(dispparams)
	olejunk.PtrCache.Add(ptr)
	dp := (*dispParams)(unsafe.Pointer(dispparams))
	log.Debug("disp", dp, dispid)

	getTrack := func() *IiTrack {
		if dp.cArgs == 0 {
			return nil
		}
		ptr := unsafe.Pointer(&dp.rgvarg)
		olejunk.PtrCache.Add(ptr)
		
		first := (*ole.VARIANT)(*(*unsafe.Pointer)(ptr))
		track, err := olejunk.GetCOMObjectFromVariant[IiTrack](first, IID_IiTrack)
		if err != nil {
			return nil
		}
		return track
	}

	getInteger := func() *int64 {
		if dp.cArgs == 0 {
			return nil
		}
		ptr := unsafe.Pointer(&dp.rgvarg)
		olejunk.PtrCache.Add(ptr)
		
		first := (*ole.VARIANT)(*(*unsafe.Pointer)(ptr))
		switch first.VT {
		case ole.VT_I4, ole.VT_I8, ole.VT_INT:
			return &first.Val
		default:
			return nil
		}
	}

	switch dispid {
	case OnPlayerPlayEventNum:
		ev.handler.OnPlayerPlayEvent(getTrack())
	case OnPlayerStopEventNum:
		ev.handler.OnPlayerStopEvent(getTrack())
	case OnPlayerPlayingTrackChangedEventNum:
		ev.handler.OnPlayerPlayingTrackChangedEvent(getTrack())
	case OnQuittingEventNum:
		ev.handler.OnQuittingEvent()
	case OnAboutToPromptUserToQuitEventNum:
		ev.handler.OnAboutToPromptUserToQuitEvent()
	case OnSoundVolumeChangedEventNum:
		ev.handler.OnSoundVolumeChangedEvent(getInteger())
	}

	return ole.S_OK
}

func connectObject(disp *ole.IDispatch, iid *ole.GUID, idisp any) (point *ole.IConnectionPoint, cookie uint32, err error) {
	unknown, err := disp.QueryInterface(ole.IID_IConnectionPointContainer)
	unknown.AddRef()
	if err != nil {
		log.Fatalf("failed to query for interface while connecting to the object: %v", err)
		return
	}

	container := (*ole.IConnectionPointContainer)(unsafe.Pointer(unknown))
	container.AddRef()
	log.Debug("got the connection point container")
	defer container.Release()

	point = nil
	err = container.FindConnectionPoint(iid, &point)
	if err != nil {
		log.Fatalf("find connection point failed: %v", err)
		return
	}
	point.AddRef()
	
	if edisp, ok := idisp.(*ole.IUnknown); ok {
		cookie, err = point.Advise(edisp)
		if err != nil {
			log.Fatalf("advise failed: %v", err)
			point.Release()
			return nil, cookie, ole.NewError(ole.E_INVALIDARG)
		}
	}
	return point, cookie, nil
}

type COMEventSink struct {
	disp      *ole.IDispatch
	callbackHandler COMEventCallback
	connectionPoint *ole.IConnectionPoint
	cookie          uint32
}

func NewCOMEventSink(disp *ole.IDispatch, handler TunesEventHandler) (*COMEventSink, error) {
	return &COMEventSink{
		disp: disp,
		callbackHandler: COMEventCallback{
			handler: handler,
		},
	}, nil
}

func NewTunesDispatch() (*ole.IDispatch, error) {
	// https://learn.microsoft.com/en-us/windows/win32/api/objbase/ne-objbase-coinit
	ole.CoInitializeEx(uintptr(0), ole.COINIT_MULTITHREADED)

	unknown, err := oleutil.CreateObject(TunesProgramID)
	if err != nil {
		log.Fatalf("failed to create object: %v", err)
		return nil, err
	}
	defer unknown.Release()

	iTunesDispatch, err := unknown.QueryInterface(ole.IID_IDispatch)
	return iTunesDispatch, err
}

func (c *COMEventSink) DisconnectObject() {
	if c.connectionPoint == nil {
		return
	}
	if c.cookie != 0 {
		if err := c.connectionPoint.Unadvise(c.cookie); err != nil {
			log.Error("failed to unadvise event sink", err)
		}
		c.cookie = 0
	}
	c.connectionPoint.Release()
	c.connectionPoint = nil
}

func (c *COMEventSink) ListenEvents() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	log.Info("setting the event receiver up")
	defer ole.CoUninitialize()
	iid, err := ole.CLSIDFromString(IID_IiTunesEvents)
	if err != nil {
		log.Error("failed to query for iTunesEvents object", err)
		return err
	}

	receiver := &eventReceiver{
		lpVtbl: &eventReceiverVtbl{
			IUnknownVtbl: ole.IUnknownVtbl{
				QueryInterface: syscall.NewCallback(queryInterface),
				AddRef:         syscall.NewCallback(addRef),
				Release:        syscall.NewCallback(release),
			},
			GetTypeInfoCount: syscall.NewCallback(getTypeInfoCount),
			GetTypeInfo:      syscall.NewCallback(getTypeInfo),
			GetIDsOfNames:    syscall.NewCallback(getIDsOfNames),
			Invoke:           syscall.NewCallback(c.callbackHandler.invoke),
		},
		host: c.disp,
	}
	ptr := unsafe.Pointer(receiver)
	olejunk.PtrCache.Add(ptr)

	c.connectionPoint, c.cookie, err = connectObject(c.disp, iid, (*ole.IUnknown)(ptr))
	if err != nil {
		log.Error("failed to connect the eventReceiver object", err)
		return err
	}

	var msg ole.Msg
	for {
		if receiver.ref != 0 {
			ret, err := ole.GetMessage(&msg, 0, 0, 0)
			if err != nil {
				return err
			}
			if ret == 0 {
				break
			}
			ole.DispatchMessage(&msg)
		}

		if receiver.ref == -1 {
			log.Warn("receiver.ref is -1...? we should probably quit.")
			return errors.New("receiver.ref is -1")
		}
	}
	return nil
}
