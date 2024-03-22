package logger

import (
	"fmt"
	"log"
	"os"
)

var (
	stdoutLog  *log.Logger
	stoderrLog *log.Logger
)

func init() {
	stdoutLog = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds)
	stoderrLog = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds)
}

func Debug(args ...interface{}) {
	stdoutLog.SetPrefix("[DEBUG]\u0020")
	stdoutLog.Println(args...)
}
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}
func Info(args ...interface{}) {
	stdoutLog.SetPrefix("[\u0020INFO]\u0020")
	stdoutLog.Println(args...)
}
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}
func Error(args ...interface{}) {
	stoderrLog.SetPrefix("[ERROR]\u0020")
	stoderrLog.Println(args...)
}
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}
func Fatal(args ...interface{}) {
	stoderrLog.SetPrefix("[FATAL]\u0020")
	stoderrLog.Fatalln(args...)
}
func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}
