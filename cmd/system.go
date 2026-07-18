//go:build windows

package main

import (
	"fmt"
	"os/exec"
)

func killTunes() error {
	cmd := exec.Command("taskkill", "/F", "/IM", "iTunes.exe", "/T")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("taskkill failed: %v, output: %s", err, out)
	}
	return nil
}
