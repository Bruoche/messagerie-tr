package log

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

func Debug(message string, args ...any) {
	// log("DEBUG", message, args...)
}

func Info(message string, args ...any) {
	log("INFO", message, args...)
}

func Warn(message string, args ...any) {
	log("WARN", message, args...)
}

func Error(message string, args ...any) {
	log("ERROR", message, args...)
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

func log(prefix string, message string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	caller := "[unknown caller]"
	if ok {
		caller = file + ":" + strconv.Itoa(line)
	}
	fmt.Printf(
		"[%s] %s - %s : %s\n",
		prefix,
		time.Now().Format("02/01/2006 - 15:04:05"),
		caller,
		fmt.Sprintf(message, args...),
	)
}
