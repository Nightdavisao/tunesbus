//go:build windows

package itunes

import (
	"log"
	//"runtime"
	"syscall"
	"unsafe"

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
	pthis := (*eventReceiver)(unsafe.Pointer(this))
	pthis.ref++
	return uintptr(pthis.ref)
}

func release(this *ole.IUnknown) uintptr {
	pthis := (*eventReceiver)(unsafe.Pointer(this))
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
	dp := (*dispParams)(unsafe.Pointer(dispparams))
	log.Printf("dp: %v, dispid: %v\n", dp, dispid)

	getTrack := func() *IiTrack {
		if dp.cArgs == 0 {
			return nil
		}
		first := (*ole.VARIANT)(unsafe.Pointer(dp.rgvarg))
		//log.Printf("first argument: %s", first)
		track, _ := getCOMObjectFromVariant[IiTrack](first, IID_IiTrack)
		return track
	}

	getInteger := func() *int64 {
		if dp.cArgs == 0 {
			return nil
		}
		first := (*ole.VARIANT)(unsafe.Pointer(dp.rgvarg))
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

func connectObject(disp *ole.IDispatch, iid *ole.GUID, idisp any) (cookie uint32, err error) {
	unknown, err := disp.QueryInterface(ole.IID_IConnectionPointContainer)
	if err != nil {
		log.Fatalf("failed to query for interface while connecting to the object: %v", err)
		return
	}

	container := (*ole.IConnectionPointContainer)(unsafe.Pointer(unknown))
	log.Printf("got the connection point container")
	defer container.Release()

	var point *ole.IConnectionPoint
	err = container.FindConnectionPoint(iid, &point)
	if err != nil {
		log.Fatalf("find connection point failed: %v", err)
		return
	}
	if edisp, ok := idisp.(*ole.IUnknown); ok {
		cookie, err = point.Advise(edisp)
		if err != nil {
			log.Fatalf("advise failed: %v", err)
			return cookie, ole.NewError(ole.E_INVALIDARG)
		}
	}
	return cookie, nil
}

type COMEventSink struct {
	dispatcher      *ole.IDispatch
	callbackHandler COMEventCallback
}

func NewCOMEventSink(dispatcher *ole.IDispatch, handler TunesEventHandler) (*COMEventSink, error) {
	return &COMEventSink{
		dispatcher: dispatcher,
		callbackHandler: COMEventCallback{
			handler: handler,
		},
	}, nil
}

func NewTunesDispatch() (*ole.IDispatch, error) {
	ole.CoInitializeEx(0, 0)
	//defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject(TunesProgramID)
	if err != nil {
		log.Fatalf("failed to create object: %v", err)
		return nil, err
	}

	iTunesDispatch, err := unknown.QueryInterface(ole.IID_IDispatch)
	return iTunesDispatch, err
}

func Uninitialize() {
	ole.CoUninitialize()
}

func (c *COMEventSink) setupEventReceiver() {
	log.Printf("setting the event receiver up")
	iid, err := ole.CLSIDFromString(IID_IiTunesEvents)
	if err != nil {
		log.Fatalf("failed to query for iTunesEvents object: %v", err)
		// todo: should we?
		//panic(err)
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
		host: c.dispatcher,
	}

	log.Printf("connecting the receiver object")
	connectObject(c.dispatcher, iid, (*ole.IUnknown)(unsafe.Pointer(receiver)))

	var m ole.Msg
	for receiver.ref != 0 {
		ole.GetMessage(&m, 0, 0, 0)
		ole.DispatchMessage(&m)
	}
}

func (c *COMEventSink) StartCOMEventLoop() {
	c.setupEventReceiver()
}
