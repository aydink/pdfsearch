package main

import (
	"fmt"
	"os"
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
