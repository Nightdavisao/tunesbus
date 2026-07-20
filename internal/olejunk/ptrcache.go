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

func (me *_PtrCache) Add(ptr unsafe.Pointer) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if me.cache == nil {
		me.cache = make(map[unsafe.Pointer]struct{})
	}
	me.cache[ptr] = struct{}{}
}

func (me *_PtrCache) Delete(ptr unsafe.Pointer) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	delete(me.cache, ptr)
}