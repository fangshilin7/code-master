package util

import (
	"fmt"
	"git.scsv.online/go/base/logger"
	"net/http"
	np "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var basePath_ string

func DumpBytes(buf []byte) {
	var s string
	for i, tmp := range buf {
		if i%16 == 0 {
			s += "\r\n"
		}
		s += fmt.Sprintf("%02x ", tmp)
	}
	fmt.Println(s)
}

//Debug Handler
//basePath: 路径前缀，如 "/ms", "/cms"
func HandleDebug(basePath string) {
	basePath_ = basePath
	//采用官方实现
	//http.HandleFunc(basePath + "/debug/pprof/", np.Index)
	http.HandleFunc(basePath+"/debug/pprof/", pprofHandler)

	http.HandleFunc(basePath+"/debug/pprof/cmdline", np.Cmdline)
	http.HandleFunc(basePath+"/debug/pprof/profile", np.Profile)
	http.HandleFunc(basePath+"/debug/pprof/symbol", np.Symbol)
	http.HandleFunc(basePath+"/debug/pprof/trace", np.Trace)

	//自定义heap handler，dump堆信息
	http.HandleFunc(basePath+"/debug/pprof/heapfile", HandlerHeap)

	http.HandleFunc(basePath+"/debug", CommonHandler)
}

func pprofHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, basePath_+"/debug/pprof/") {
		name := strings.TrimPrefix(r.URL.Path, basePath_+"/debug/pprof/")
		if name != "" {
			np.Handler(name).ServeHTTP(w, r)
			return
		}
	}
	np.Index(w, r)
}

//通用Debug，用于设置日志级别等
//http://ip:port/{prefix}/debug?loglevel=1
func CommonHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	level := r.FormValue("loglevel")
	if level != "" {
		n, _ := strconv.Atoi(level)
		if n < logger.LOG_OFF {
			n = logger.LOG_INFO
		} else if n > logger.LOG_ALL {
			n = logger.LOG_ALL
		}
		logger.Level = n
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("Set OK, loglevel = %d\n", logger.Level)))
		return
	}

	heap := r.FormValue("heap")
	if heap == "1" {
		runtime.GC()

		f, err := os.Create(fmt.Sprintf("heap_%d.prof", time.Now().Unix()))
		if err != nil {
			return
		}
		defer f.Close()
		pprof.WriteHeapProfile(f)

		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf("Heap dump OK")))
		return
	}

	w.WriteHeader(500)
	w.Write([]byte("Failed"))
}

func HandlerHeap(w http.ResponseWriter, r *http.Request) {
	runtime.GC()

	filename := fmt.Sprintf("heap_%d.prof", time.Now().Unix())
	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()
	pprof.WriteHeapProfile(f)

	HttpWriteFile(filename, w, r)
}
