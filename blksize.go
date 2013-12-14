package main

import "os"

func BlockDeviceSize(file *os.File) (uint64, error) {
	blksize, err := BlockDeviceBlockSize(file)
	if err != nil {
		return 0, err
	}

	blkcnt, err := BlockDeviceBlockCount(file)
	if err != nil {
		return 0, err
	}

	return uint64(blksize) * blkcnt, nil
}
