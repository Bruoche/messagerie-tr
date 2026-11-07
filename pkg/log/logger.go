package log

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var ShowDebug = false

func Debug(message string, args ...any) {
	if ShowDebug {
		log("DEBUG", message, args...)
	}
}

func ServCall(request string) {
	fmt.Printf(
		"[SERVER] %s - Request %s called\n",
		time.Now().Format("02/01/2006 - 15:04:05"),
		request,
	)
}

func Info(message string, args ...any) {
	log("INFO", message, args...)
}

func Warn(message string, args ...any) {
	log("WARN", message, args...)
}

func Error(message string, args ...any) {
	log("ERROR", message, args...)
	if ShowDebug {
		stackDepth := 2
		ok := true
		var file string
		var line int
		for ok {
			_, file, line, ok = runtime.Caller(stackDepth)
			fmt.Printf("    %s:%d\n", file, line)
			stackDepth++
		}
	}
}

func log(prefix string, message string, args ...any) {
	location := ""
	if ShowDebug {
		_, file, line, ok := runtime.Caller(2)
		caller := "[unknown caller]"
		if ok {
			caller = file + ":" + strconv.Itoa(line)
		}
		location = " - " + caller
	}
	fmt.Printf(
		"[%s] %s%s : %s\n",
		prefix,
		time.Now().Format("02/01/2006 - 15:04:05"),
		location,
		fmt.Sprintf(message, args...),
	)
}
