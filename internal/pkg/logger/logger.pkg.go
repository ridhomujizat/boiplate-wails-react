package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

var (
	Debug      *log.Logger
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	HTTP       *log.Logger
	errorTrace *log.Logger
)

func Setup() {
	HTTP = log.New(os.Stdout, "[HTTP]\t", log.Ldate|log.Ltime)
	Info = log.New(os.Stdout, "[INFO]\t", log.Ldate|log.Ltime)
	Warning = log.New(os.Stdout, "[WARNING]\t", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(os.Stdout, "[DEBUG]\t", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorTrace = log.New(os.Stdout, "[ERROR]\t", log.Ldate|log.Ltime)
}

func ErrorWithStack(format string, v ...interface{}) {
	stackTrace := getStackTrace()

	message := fmt.Sprintf(format, v...)
	errorTrace.Printf("%s\n%s", message, stackTrace)
}

func getStackTrace() string {
	buf := make([]byte, 1024*64)
	buf = buf[:runtime.Stack(buf, false)]

	lines := strings.Split(string(buf), "\n")
	filteredLines := []string{}

	skip := 0
	for i, line := range lines {
		if strings.Contains(line, "logger.pkg.go") ||
			strings.Contains(line, "runtime/") ||
			strings.Contains(line, "getStackTrace") ||
			strings.Contains(line, "ErrorWithStack") {
			skip = i + 1
			continue
		}

		if i <= skip {
			continue
		}

		filteredLines = append(filteredLines, line)
	}

	return strings.Join(filteredLines, "\n")
}
