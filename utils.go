package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func printMemUsage() {

	for tick := range time.Tick(3 * time.Second) {

		// Prints UTC time and date
		fmt.Println(tick, UTCtime())

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		//   https://golang.org/pkg/runtime/#MemStats
		fmt.Printf("Alloc = %0.2f MiB", bToMbyte(m.Alloc))
		fmt.Printf("\tTotalAlloc = %0.2f MiB", bToMbyte(m.TotalAlloc))
		fmt.Printf("\tSys = %0.2f MiB", bToMbyte(m.Sys))
		fmt.Printf("\tNumGC = %v\n", m.NumGC)
	}
}

// Defining UTCtime
func UTCtime() string {
	return ""
}

func bToMbyte(b uint64) float64 {
	return float64(b) / float64(1024) / float64(1024)
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// exists returns whether the given file or directory exists or not
func isDirectory(path string) (bool, error) {
	if fileInfo, err := os.Stat(path); !os.IsNotExist(err) {
		if fileInfo.IsDir() {
			return true, nil
		} else {
			return false, errors.New("'" + path + "' is not a valid path")
		}
	} else {
		return false, err
	}

}

type byValue []uint32

func (f byValue) Len() int {
	return len(f)
}

func (f byValue) Less(i, j int) bool {
	return f[i] < f[j]
}

func (f byValue) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type byBookTitle []Book

func (f byBookTitle) Len() int {
	return len(f)
}

func (f byBookTitle) Less(i, j int) bool {
	return f[i].Title < f[j].Title
}

func (f byBookTitle) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func uint32ToBytes(x uint32) []byte {
	var buf [4]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	return buf[:]
}

func bytesToUint32le(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func openBrowser(url string) {

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll.FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Unsupported platform")
	}

	if err != nil {
		log.Println(err)
	}
}
