package olejunk

import (
	"sync"

	"github.com/charmbracelet/log"
	"github.com/go-ole/go-ole"
)

type OleReleaser struct {
	mu      sync.RWMutex
	objects map[*ole.IUnknown]struct{}
}

func NewOleReleaser() *OleReleaser {
	return &OleReleaser{
		objects: make(map[*ole.IUnknown]struct{}),
	}
}

func (r *OleReleaser) Add(obj *ole.IUnknown) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Debug("OleReleaser: adding object", "object", obj)
	if obj == nil {
		log.Warn("OleReleaser: attempted to add nil object")
		return
	}
	if _, exists := r.objects[obj]; exists {
		log.Debug("OleReleaser: object already tracked, skipping AddRef", "object", obj)
		return
	}
	obj.AddRef()
	r.objects[obj] = struct{}{}
}

func (r *OleReleaser) Release() {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Debug("OleReleaser: releasing all objects", "length", len(r.objects))
	for obj := range r.objects {
		log.Debug("OleReleaser: object to be released", "object", obj)
		refCount := obj.Release()
		log.Debug("OleReleaser: released object", "ref_count", refCount)
	}
	r.objects = make(map[*ole.IUnknown]struct{})
}