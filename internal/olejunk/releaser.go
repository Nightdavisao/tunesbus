package olejunk

import (
	"sync"
	"github.com/charmbracelet/log"
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
	log.Debug("OleReleaser: adding object", "object", obj)
	if obj != nil {
		obj.AddRef()
		r.objects = append(r.objects, obj)
	} // todo: warn if nil
	r.mu.Unlock()
}

func (r *OleReleaser) Release() {
	r.mu.RLock()
	defer r.mu.RUnlock()
	log.Debug("OleReleaser: releasing all objects", "length", len(r.objects))
	for _, obj := range r.objects {
		log.Debug("OleReleaser: object to be released", "object", obj)
		if obj != nil {
			refCount := obj.Release()
			log.Debug("OleReleaser: released object", "ref_count", refCount)
		}
	}
	r.objects = nil
}
