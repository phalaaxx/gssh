package main

import (
	"os"
)

/* IsTerminal returns true if output device is terminal */
func IsTerminal(f *os.File) bool {
	if fileInfo, _ := f.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	}
	return false
}
