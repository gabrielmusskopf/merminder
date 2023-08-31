package logger

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

func filename(f string) string {
    nameParts := strings.Split(f, "/")
    return nameParts[len(nameParts) - 1]

}

func Error(err error) {
    if err != nil {
        _, fn, line, _ := runtime.Caller(1)
        log.Printf("[error] in [%s:%d] %v", filename(fn), line, err)
    }
}

func Fatal(err error) {
    if err != nil {
        _, fn, line, _ := runtime.Caller(1)
        log.Fatalf("[fatal] in [%s:%d] %v", filename(fn), line, err)
    }
}

func Fatals(err string) {
    _, fn, line, _ := runtime.Caller(1)
    log.Fatalf("[fatal] in [%s:%d] %v", filename(fn), line, err)
}

func Warning(f string, v ...any) {
    _, fn, line, _ := runtime.Caller(1)
    s := fmt.Sprintf(f, v...)

    log.Printf("[warning] in [%s:%d] %v", filename(fn), line, s)
}

func Info(f string, v ...any) {
    _, fn, line, _ := runtime.Caller(1)
    s := fmt.Sprintf(f, v...)

    log.Printf("[info] in [%s:%d] %v", filename(fn), line, s)
}

