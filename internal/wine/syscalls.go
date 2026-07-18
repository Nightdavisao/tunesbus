//go:build windows

package wine

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

var (
	modKernel32        = syscall.NewLazyDLL("KERNEL32.dll")
	modNtdll           = syscall.NewLazyDLL("ntdll.dll")
	procGetUnix        = modKernel32.NewProc("wine_get_unix_file_name")
	procGetDos         = modKernel32.NewProc("wine_get_dos_file_name")
	procHeapFree       = modKernel32.NewProc("HeapFree")
	procGetHeap        = modKernel32.NewProc("GetProcessHeap")
	procGetWineVersion = modNtdll.NewProc("wine_get_version")
	procGetWineBuild   = modNtdll.NewProc("wine_get_build_id")
	// Windows
	user32      = syscall.NewLazyDLL("user32.dll")
	procMsgBoxA = user32.NewProc("MessageBoxA")
)

type MBButton uintptr
type MBIcon uintptr
type MBDefault uintptr
type MBModal uintptr
type MBResult uintptr

const (
	MB_OK                MBButton = 0x00000000
	MB_OKCANCEL          MBButton = 0x00000001
	MB_ABORTRETRYIGNORE  MBButton = 0x00000002
	MB_YESNOCANCEL       MBButton = 0x00000003
	MB_YESNO             MBButton = 0x00000004
	MB_RETRYCANCEL       MBButton = 0x00000005
	MB_CANCELTRYCONTINUE MBButton = 0x00000006
)

const (
	MB_ICONHAND        MBIcon = 0x00000010
	MB_ICONQUESTION    MBIcon = 0x00000020
	MB_ICONEXCLAMATION MBIcon = 0x00000030
	MB_ICONASTERISK    MBIcon = 0x00000040
	MB_ICONERROR       MBIcon = MB_ICONHAND
	MB_ICONWARNING     MBIcon = MB_ICONEXCLAMATION
	MB_ICONINFORMATION MBIcon = MB_ICONASTERISK
)

const (
	IDOK       MBResult = 1
	IDCANCEL   MBResult = 2
	IDABORT    MBResult = 3
	IDRETRY    MBResult = 4
	IDIGNORE   MBResult = 5
	IDYES      MBResult = 6
	IDNO       MBResult = 7
	IDTRYAGAIN MBResult = 10
	IDCONTINUE MBResult = 11
)

func stringToBytePtr(s string) *byte {
	b, _ := syscall.BytePtrFromString(s)
	return b
}

func MessageBox(hwnd uintptr, text, caption string, button MBButton, icon MBIcon) MBResult {
	flags := uintptr(button) | uintptr(icon)
	ret, _, _ := procMsgBoxA.Call(
		hwnd,
		uintptr(unsafe.Pointer(stringToBytePtr(text))),
		uintptr(unsafe.Pointer(stringToBytePtr(caption))),
		flags,
	)
	return MBResult(ret)
}

func getStrResultFromProc(proc *syscall.LazyProc, str string, inputWide, outputWide, freeResult bool) (string, error) {
	var ret uintptr
	if str != "" {
		var argPtr uintptr
		if inputWide {
			arg, err := syscall.UTF16PtrFromString(str)
			if err != nil {
				return "", err
			}
			argPtr = uintptr(unsafe.Pointer(arg))
		} else {
			arg, err := syscall.BytePtrFromString(str)
			if err != nil {
				return "", err
			}
			argPtr = uintptr(unsafe.Pointer(arg))
		}
		ret, _, _ = proc.Call(argPtr)
	} else {
		ret, _, _ = proc.Call()
	}

	if ret == 0 {
		return "", syscall.GetLastError()
	}

	var finalRet string
	if outputWide {
		finalRet = cStringToGoWide(ret)
	} else {
		finalRet = cStringToGoNarrow(ret)
	}

	if freeResult {
		defer func() {
			heap, _, _ := procGetHeap.Call()
			procHeapFree.Call(heap, 0, ret)
		}()
	}

	return finalRet, nil
}

// wine_get_unix_file_name(LPCWSTR dos) -> char* (narrow)
func GetUnixFilename(dosFilename string) (string, error) {
	return getStrResultFromProc(procGetUnix, dosFilename, true, false, true)
}

// wine_get_dos_file_name(LPCSTR unix) -> WCHAR* (wide)
func GetDosFilename(unixFilename string) (string, error) {
	return getStrResultFromProc(procGetDos, unixFilename, false, true, true)
}

func GetWineVersion() (string, error) {
	return getStrResultFromProc(procGetWineVersion, "", false, false, false)
}

func GetWineBuild() (string, error) {
	return getStrResultFromProc(procGetWineBuild, "", false, false, false)
}

// reads a null-terminated single-byte (ANSI/UTF-8) C string
func cStringToGoNarrow(ptr uintptr) string {
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

// reads a null-terminated UTF-16 (WCHAR*) C string
// https://stackoverflow.com/a/41272645
func cStringToGoWide(ptr uintptr) string {
	a := (*[1<<30 - 1]uint16)(unsafe.Pointer(ptr))
	size := 0
	for ; size < len(a); size++ {
		if a[size] == 0 {
			break
		}
	}
	runes := utf16.Decode(a[:size:size])
	return string(runes)
}
