package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const DefaultBandSize int64 = 8 * 1024 * 1024

var bandSize int64 = DefaultBandSize

type BandInfo struct {
	id            uint64
	read, written bool
	buffer        []byte
}

func ReadBands(diskFile *os.File, startBand, endBand int, bufferChannel chan []byte, bandChannel chan BandInfo) chan bool {
	complete := make(chan bool)
	go func() {
		bandNumber := uint64(startBand)

		diskFile.Seek(int64(startBand)*bandSize, 0)
		for {
			buffer := <-bufferChannel
			buffer = buffer[:cap(buffer)]
			nread, err := io.ReadFull(diskFile, buffer)
			fmt.Printf("RB : Read %v bytes for band %x\n", nread, bandNumber)
			if err != nil {
				fmt.Printf("RB : Error %v\n", err)
			}

			if nread > 0 {
				bandChannel <- BandInfo{
					id:      bandNumber,
					read:    true,
					written: false,
					buffer:  buffer[0:nread],
				}
			}

			bandNumber++
			if err == io.ErrUnexpectedEOF || err == io.EOF || (endBand > 0 && bandNumber > uint64(endBand)) {
				break
			}
		}
		complete <- true
	}()
	return complete
}

func ZeroCheckBands(in, out chan BandInfo, bufferChannel chan []byte) chan bool {
	complete := make(chan bool)
	go func() {
		for band := range in {
			if bytes.Count(band.buffer, []byte{0, 0, 0, 0, 0, 0, 0, 0}) == int(bandSize/8) {
				fmt.Printf("ZCB: skipping %x\n", band.id)
				bufferChannel <- band.buffer
				continue
			}

			out <- band
		}
		complete <- true
	}()
	return complete
}

func WriteBands(directory string, bandChannel chan BandInfo, bufferChannel chan []byte) chan bool {
	complete := make(chan bool)
	go func() {
		for band := range bandChannel {
			bandFilename := fmt.Sprintf("%s/%x", directory, band.id)
			fmt.Println("WB : Writing", bandFilename)
			bandFile, _ := os.Create(bandFilename)
			bandFile.Write(band.buffer)
			bandFile.Close()
			bufferChannel <- band.buffer
		}
		complete <- true
	}()
	return complete
}

func InitBuffers(bufferChannel chan []byte, nbuffers int) {
	for i := 0; i < nbuffers; i++ {
		bufferChannel <- make([]byte, bandSize, bandSize)
	}
}

func main() {
	var disk, sbname string
	var startBand, endBand int
	var nbuffers int
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sparsebundle [options] <block device>")
		flag.PrintDefaults()
	}
	flag.StringVar(&sbname, "name", "", "sparse bundle name")
	flag.IntVar(&startBand, "start", 0, "starting band")
	flag.IntVar(&endBand, "end", -1, "starting band")
	flag.Int64Var(&bandSize, "bandsize", DefaultBandSize, "the size of each sparsebundle band, in bytes")
	flag.IntVar(&nbuffers, "buffers", 5, "number of buffers to use (each one costs <bandsize> bytes).")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		disk = args[0]
	}

	if disk == "" {
		flag.Usage()
		return
	}

	if sbname == "" {
		sbname = filepath.Base(disk)
		fmt.Fprintln(os.Stderr, "No name provided, using", sbname)
	}

	diskFile, err := os.Open(disk)
	if err != nil {
		panic(err)
	}

	bundleDirectory := sbname + ".sparsebundle"

	bandDirectory := filepath.Join(bundleDirectory, "bands")
	if err := os.MkdirAll(bandDirectory, 0755); err != nil {
		panic(err)
	}

	if devSize, err := BlockDeviceSize(diskFile); err == nil {
		WriteInfoPlist(filepath.Join(bundleDirectory, "Info.plist"), bandSize, devSize)
	}

	if token, err := os.Create(filepath.Join(bundleDirectory, "token")); err == nil {
		token.Close()
	}

	bufferChannel := make(chan []byte, nbuffers)
	InitBuffers(bufferChannel, nbuffers)

	bandsToZeroCheck := make(chan BandInfo, nbuffers*2)
	bandsToWrite := make(chan BandInfo, nbuffers*2)

	go func() {
		<-ReadBands(diskFile, startBand, endBand, bufferChannel, bandsToZeroCheck)
		diskFile.Close()
		close(bandsToZeroCheck)
	}()

	go func() {
		zc1 := ZeroCheckBands(bandsToZeroCheck, bandsToWrite, bufferChannel)
		zc2 := ZeroCheckBands(bandsToZeroCheck, bandsToWrite, bufferChannel)
		<-zc1
		<-zc2
		close(bandsToWrite)
	}()

	wc1 := WriteBands(bandDirectory, bandsToWrite, bufferChannel)
	wc2 := WriteBands(bandDirectory, bandsToWrite, bufferChannel)

	<-wc1
	<-wc2
}
