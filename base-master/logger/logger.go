package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

const (
	LOG_OFF   = 0 //iota
	LOG_FATAL = 1
	LOG_ERROR = 2
	LOG_WARN  = 3
	LOG_INFO  = 4
	LOG_DEBUG = 5
	LOG_TRACE = 6
	LOG_ALL   = 7
)

var Level int = LOG_DEBUG
var LEVELS []string = []string{"OFF", "FATAL", "ERROR", "WARN", "INFO", "DEBUG", "TRACE", "ALL"}
var Prefix = ""
var MaxLogDays = 7
var LogDirectory = ""
var LogToFile = true

var logFile *os.File = nil
var fileTime time.Time
var fileLock sync.Mutex

func Output(level int, format string, a ...interface{}) {
	if level <= Level {
		s := fmt.Sprintf(format, a...)
		tm := time.Now()
		s = fmt.Sprintf("[%04d-%02d-%02d %02d:%02d:%02d.%03d] [%s] %s ",
			tm.Year(), tm.Month(), tm.Day(), tm.Hour(),
			tm.Minute(), tm.Second(), tm.Nanosecond()/1000000,
			LEVELS[level], Prefix) + s + "\r\n"
		fmt.Fprint(os.Stdout, s)

		if level == LOG_FATAL {
			outputFatal(&tm, s)
		} else if LogToFile {
			outputFile(&tm, s)
		}
	}
}

func Fatal(format string, a ...interface{}) {
	Output(LOG_FATAL, format, a...)
}

func Error(format string, a ...interface{}) {
	Output(LOG_ERROR, format, a...)
}

func Warn(format string, a ...interface{}) {
	Output(LOG_WARN, format, a...)
}

func Info(format string, a ...interface{}) {
	Output(LOG_INFO, format, a...)
}

func Debug(format string, a ...interface{}) {
	Output(LOG_DEBUG, format, a...)
}

func Trace(format string, a ...interface{}) {
	Output(LOG_TRACE, format, a...)
}

func All(format string, a ...interface{}) {
	Output(LOG_ALL, format, a...)
}

func outputFile(tm *time.Time, s string) {
	fileLock.Lock()
	defer fileLock.Unlock()
	var err error
	if logFile == nil {
		os.MkdirAll(LogDirectory, os.ModePerm)
		path := LogDirectory + "server.log"

		// 取文件属性
		if info, err := os.Stat(path); err == nil {
			t := info.ModTime().Format("20060102")
			now := time.Now().Format("20060102")
			if t != now {
				// 重命名已有日志文件
				err = os.Rename(path, LogDirectory+t+".log")
			}
		}

		logFile, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			//fmt.Fprint(os.Stderr, err)
			return
		}
		fileTime = time.Now()
		go removeExpiredFile()
	} else {
		checkFile(tm)
	}

	logFile.WriteString(s)
}

func outputFatal(tm *time.Time, s string) {
	os.MkdirAll(LogDirectory, os.ModePerm)
	path := LogDirectory + "server.dumplog"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		//fmt.Fprint(os.Stderr, err)
		return
	}
	file.WriteString(s)
}

func checkFile(tm *time.Time) {
	days := tm.Day() - fileTime.Day()
	if days != 0 {
		//日期不一致，备份日志文件
		dst := fmt.Sprintf("%s%04d%02d%02d.log", LogDirectory,
			fileTime.Year(), fileTime.Month(), fileTime.Day())

		logPath := LogDirectory + "server.log"
		copyFile(logPath, dst)
		fileTime = *tm

		//windows下必须关闭文件才能truncate
		var err error
		logFile.Close()
		err = os.Truncate(logPath, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		logFile, err = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		go removeExpiredFile()
	}
}

//删除超过限定天数的日志
func removeExpiredFile() {
	files, err := ioutil.ReadDir(LogDirectory + ".")
	if err != nil {
		//fmt.Fprint(os.Stderr, err)
		return
	}
	for _, file := range files {
		diff := fileTime.Unix() - file.ModTime().Unix()
		s := file.Name()
		if len(s) < 5 {
			continue
		}
		s = s[len(s)-4:]
		if s != ".log" {
			continue
		}

		if diff > int64(MaxLogDays*24*3600) {
			os.Remove(LogDirectory + file.Name())
			fmt.Println("remove: ", file.Name())
			break
		}
	}
}

func copyFile(srcFile, destFile string) error {
	file, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer file.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	io.Copy(dest, file)

	return nil
}
