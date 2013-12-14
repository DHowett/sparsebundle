package main

import (
	"howett.net/plist"
	"os"
)

type InfoDictionary struct {
	InfoVersion    string `plist:"CFBundleInfoDictionaryVersion"`
	Version        int    `plist:"CFBundleVersion,omitempty"`
	DisplayVersion string `plist:"CFBundleDisplayVersion,omitempty"`
}

type SparseBundleHeader struct {
	InfoDictionary
	BandSize            uint64 `plist:"band-size"`
	BackingStoreVersion int    `plist:"bundle-backingstore-version"`
	DiskImageBundleType string `plist:"diskimage-bundle-type"`
	Size                uint64 `plist:"size"`
}

func WriteInfoPlist(path string, bandSize int64, deviceSize uint64) error {
	sbheader := &SparseBundleHeader{
		InfoDictionary:      InfoDictionary{InfoVersion: "6.0"},
		BandSize:            uint64(bandSize),
		Size:                deviceSize,
		DiskImageBundleType: "com.apple.diskimage.sparsebundle",
		BackingStoreVersion: 1,
	}

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}

	encoder := plist.NewEncoderForFormat(file, plist.XMLFormat)
	encoder.Indent("\t")
	err = encoder.Encode(sbheader)
	if err != nil {
		return err
	}

	return nil
}
