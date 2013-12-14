package main

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	DKIOCGETBLOCKSIZE  = 0x40046418 // takes a *uint32
	DKIOCGETBLOCKCOUNT = 0x40086419 // takes a *uint64
)

func BlockDeviceBlockSize(file *os.File) (uint32, error) {
	var blksize uint32

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()), uintptr(DKIOCGETBLOCKSIZE), uintptr(unsafe.Pointer(&blksize)))
	if errno != 0 {
		return 0, errno
	}

	return blksize, nil
}

func BlockDeviceBlockCount(file *os.File) (uint64, error) {
	var blkcnt uint64

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()), uintptr(DKIOCGETBLOCKCOUNT), uintptr(unsafe.Pointer(&blkcnt)))
	if errno != 0 {
		return 0, errno
	}

	return blkcnt, nil
}
