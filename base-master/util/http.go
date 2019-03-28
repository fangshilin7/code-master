package util

import (
	"bytes"
	"git.scsv.online/go/base/logger"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

func JsonRequest(url string, body string) (code int, msg string, err error) {
	req := bytes.NewBuffer([]byte(body))

	logger.Trace("POST %s \r\n %s", url, body)

	code = http.StatusInternalServerError

	resp, err := http.Post(url, "application/json;charset=utf-8", req)

	if err == nil {
		defer resp.Body.Close()
		code = resp.StatusCode
		var tmp []byte
		tmp, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		msg = string(tmp)
		logger.Trace(msg)
	} else {
		code = http.StatusRequestTimeout
	}

	return
}

func HttpWriteFile(filename string, writer http.ResponseWriter, request *http.Request) {
	//Check if file exists and open
	Openfile, err := os.Open(filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	writer.Header().Set("Content-Type", FileContentType)
	writer.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already, so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(writer, Openfile) //'Copy' the file to the client
}

var crossdomainxml = []byte(`<?xml version="1.0" ?>
<cross-domain-policy>
	<allow-access-from domain="*" />
	<allow-http-request-headers-from domain="*" headers="*"/>
</cross-domain-policy>`)

//跨域访问
func CORS() {
	http.HandleFunc("/crossdomain.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write(crossdomainxml)
	})
}
