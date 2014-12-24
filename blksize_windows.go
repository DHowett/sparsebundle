package main

import (
	"encoding/binary"
	"errors"
	"os"
	"syscall"
	"unsafe"
)

var (
	dll             = syscall.MustLoadDLL("kernel32.dll")
	DeviceIoControl = dll.MustFindProc("DeviceIoControl")
)

type DISK_GEOMETRY struct {
	Cylinders         uint64
	MediaType         uint32
	TracksPerCylinder uint32
	SectorsPerTrack   uint32
	BytesPerSector    uint32
}

type DISK_GEOMETRY_EX struct {
	Geom     DISK_GEOMETRY
	DiskSize uint64
	Data     [1]byte
}

const (
	IOCTL_DISK_GET_DRIVE_GEOMETRY_EX = 0x700A0
)

func BlockDeviceBlockSize(file *os.File) (uint32, error) {
	return 1, nil
}

func BlockDeviceBlockCount(file *os.File) (uint64, error) {
	var geometry DISK_GEOMETRY_EX
	var outBufferSize uint32
	var geomSize = binary.Size(&geometry)

	ret, _, _ := DeviceIoControl.Call(file.Fd(), IOCTL_DISK_GET_DRIVE_GEOMETRY_EX, 0, 0, uintptr(unsafe.Pointer(&geometry)), uintptr(geomSize), uintptr(unsafe.Pointer(&outBufferSize)), 0)
	if ret == 0 {
		return 0, errors.New("failed to get disk geometry")
	}
	return geometry.DiskSize, nil
}
