//go:build windows

package wine

import (
	"syscall"
	"unsafe"
)

var (
	modKernel32  = syscall.NewLazyDLL("KERNEL32.dll")
	procGetUnix  = modKernel32.NewProc("wine_get_unix_file_name")
	procHeapFree = modKernel32.NewProc("HeapFree")
	procGetHeap  = modKernel32.NewProc("GetProcessHeap")
)

func GetUnixFilename(dosFilename string) (string, error) {
    arg, err := syscall.UTF16PtrFromString(dosFilename)
    if err != nil {
        return "", err
    }

    ret, _, _ := procGetUnix.Call(uintptr(unsafe.Pointer(arg)))
    if ret == 0 {
        return "", syscall.GetLastError()
    }

    defer func() {
        heap, _, _ := procGetHeap.Call()
        procHeapFree.Call(heap, 0, ret)
    }()

    return cStringToGo(ret), nil
}

func cStringToGo(ptr uintptr) string {
    var buf []byte
    for {
        b := *(*byte)(unsafe.Pointer(ptr))
        if b == 0 {
            break
        }
        buf = append(buf, b)
        ptr++
    }
    return string(buf)
}