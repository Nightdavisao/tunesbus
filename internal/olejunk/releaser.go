package olejunk

import (
	"sync"

	"github.com/go-ole/go-ole"
)

type OleReleaser struct {
	mu      sync.RWMutex
	objects []*ole.IUnknown
}

func NewOleReleaser() *OleReleaser {
	return &OleReleaser{
		objects: make([]*ole.IUnknown, 0),
	}
}

func (r *OleReleaser) Add(obj *ole.IUnknown) {
	r.mu.Lock()
	if obj != nil {
		obj.AddRef()
		r.objects = append(r.objects, obj)
	} // todo: warn if nil
	r.mu.Unlock()
}

func (r *OleReleaser) Release() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, obj := range r.objects {
		if obj != nil {
			obj.Release()
		}
	}
	r.objects = nil
}
