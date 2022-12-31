package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

var (
	Trace          *log.Logger
	Info           *log.Logger
	Warning        *log.Logger
	Error          *log.Logger
	TracingEnabled bool
	TraceLevel     int
	dumpPath       string
	IsDebug        bool
)

func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"Trace: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
func init() {
	InitLog()
}

func InitLog() {
	info := ioutil.Discard
	trace := ioutil.Discard
	traceRmapi := os.Getenv("RMAPI_TRACE")
	switch traceRmapi {
	case "1":
		TracingEnabled = true
		trace = os.Stdout
		fallthrough
	case "2":
		info = os.Stdout
	}

	Init(trace, info, os.Stdout, os.Stderr)

	dumpPath = os.Getenv("RMAPI_DUMPPATH")
	if dumpPath != "" {
		IsDebug = true
	}
}

// Dump the contents somewhere
func Dump(what string, content []byte) {
	timeStamp := time.Now().UTC().Format(time.RFC3339Nano)
	filename := fmt.Sprintf("%s_%s", timeStamp, what)

	fullname := path.Join(dumpPath, filename)
	os.WriteFile(fullname, content, 0600)
}
