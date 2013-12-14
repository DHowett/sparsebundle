package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	BLKGETSIZE64 = 0x80041272
)

func BlockDeviceBlockSize(file *os.File) (uint32, error) {
	return 1, nil
}

func BlockDeviceBlockCount(file *os.File) (uint64, error) {
	var blkcnt uint64

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()), uintptr(BLKGETSIZE64), uintptr(unsafe.Pointer(&blkcnt)))
	if errno != 0 {
		return 0, errno
	}

	return blkcnt, nil
}
