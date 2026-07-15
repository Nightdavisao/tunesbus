//go:build windows

package wine

import (
	"os"
	"strings"
)

func WindowsPathJoin(elem ...string) string {
	return strings.Join(elem, "\\")
}

func UnixPathJoin(elem ...string) string {
	return strings.Join(elem, "/")
}

func UnixTmpDirAsDosPath() (string, error) {
	dir := UnixPathJoin("/tmp")
	dos, err := GetDosFilename(dir)
	if err != nil {
		return "", err
	}
	return WindowsPathJoin(dos), nil
}

// can be empty
func GetWinePrefix() string {
	return os.Getenv("WINEPREFIX")
}

func InfoMessageBox(caption string, text string) MBResult {
	return MessageBox(0, text, caption, MB_OK, MB_ICONINFORMATION)
}

func ErrorMessageBox(caption string, text string) MBResult {
	return MessageBox(0, text, caption, MB_OK, MB_ICONERROR)
}

func WarningMessageBox(caption string, text string) MBResult {
	return MessageBox(0, text, caption, MB_OK, MB_ICONWARNING)
}