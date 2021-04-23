package utils

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/megamon/core/config"
)

var (
	//WarningLogger : log warnings
	WarningLogger *log.Logger

	//InfoLogger : log info messages
	InfoLogger *log.Logger

	//ErrorLogger : Log errors
	ErrorLogger *log.Logger
)

//TrimSpecials : reduce sequence of special characters \n,\t to single one
func TrimSpecials(text string) (trimmed string) {
	last := text
	trimmed = strings.Replace(text, "\n\n", "\n", -1)
	trimmed = strings.Replace(trimmed, "\t\t", "\t", -1)

	for last != trimmed {
		last = trimmed
		trimmed = strings.Replace(trimmed, "\n\n", "\n", -1)
		trimmed = strings.Replace(trimmed, "\t\t", "\t", -1)
	}
	return
}

//ReadFile : read file content
func ReadFile(filename string) (fileData []byte, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	data := make([]byte, 64)

	for {
		n, err := file.Read(data)
		if err == io.EOF {
			break
		}

		fileData = append(fileData, data[:n]...)

	}
	return
}

//DoRequest : do http request
func DoRequest(req *http.Request) (resp *http.Response, err error) {
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	resp, err = client.Do(req)
	return resp, err
}

//GetBodyReader : body reader for response with gzip/text encoding
func GetBodyReader(resp *http.Response) (bodyReader io.ReadCloser, err error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(resp.Body)

	default:
		bodyReader = resp.Body
	}

	if err != nil {
		resp.Body.Close()
	}

	return bodyReader, err
}

//InitConfig : Read file with settings and parse it
func InitConfig(confFile string) (err error) {
	confData, err := ReadFile(confFile)

	if err != nil {
		return
	}

	err = json.Unmarshal(confData, &config.Settings)
	return
}

//InitLoggers : init loggers
func InitLoggers(logFile string) (err error) {
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}
