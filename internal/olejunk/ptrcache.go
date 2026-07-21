//go:build windows

package olejunk

import (
	"sync"
	"unsafe"
)

type _PtrCache struct {
	mutex sync.Mutex
	cache map[unsafe.Pointer]struct{}
}

var PtrCache _PtrCache

func (p *_PtrCache) Add(ptr unsafe.Pointer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cache == nil {
		p.cache = make(map[unsafe.Pointer]struct{})
	}
	p.cache[ptr] = struct{}{}
}

func (p *_PtrCache) Delete(ptr unsafe.Pointer) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	delete(p.cache, ptr)
}
