package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

const WRITE_BYTES = 1024 * 1024
const DEFAULT_GB = 2

var baseDir string

func usage() {
	fmt.Println(`usage: db path channels [GB]`)
}

func writeData(num int, wc int64, ch chan int64) {
	log.Printf("Thread[%d] Writing ...", num)

	path := fmt.Sprintf("%s%sdbtmp%d", baseDir, string(os.PathSeparator), num)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	var sum int64
	defer func() {
		f.Close()
		os.Remove(path)
		ch <- sum
	}()

	b := make([]byte, WRITE_BYTES)
	for {
		n, err := f.Write(b)
		if err != nil {
			log.Println(err)
			break
		}
		sum += int64(n)
		if sum >= wc {
			break
		}
	}

	f.Sync()

	log.Printf("Thread[%d] Write OK", num)
}

func main() {
	if len(os.Args) < 3 {
		usage()
		return
	}
	baseDir = os.Args[1]
	threads, _ := strconv.Atoi(os.Args[2])
	var total = DEFAULT_GB
	if len(os.Args) > 3 {
		total, _ = strconv.Atoi(os.Args[3])
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	ch := make(chan int64, threads)
	wc := (int64(total) << 30) / int64(threads)
	log.Print("Disk Write Benchmark Start...")
	start := time.Now().UnixNano()
	for i := 0; i < threads; i++ {
		go writeData(i, wc, ch)
	}

	var sum int64
	for i := 0; i < threads; i++ {
		sum += <-ch
	}
	use := time.Now().UnixNano() - start

	log.Printf("Disk Write: %d GB, Use %.1fs, Speed: %d MB/s", sum>>30, float64(use)/1000000000,
		int32(float64(sum)*1000000000/(float64(use)*1024*1024)))
}
