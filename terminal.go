package main

import (
	"syscall"
	"unsafe"
)

// determine if output device is terminal
func IsTerminal(fd uintptr) bool {
	var term syscall.Termios
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&term)))
	return err == 0
}
