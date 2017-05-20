package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

// LogFlag 控制日志的前缀
const logFlag = log.LstdFlags | log.Lmicroseconds | log.Lshortfile

var (
	errLogger  *log.Logger
	infoLogger *log.Logger
)

func init() {
	errLogger = log.New(os.Stderr, "", logFlag)
	infoLogger = log.New(os.Stdout, "", logFlag)
}

// Fatalf 打印错误日志并退出
func Fatalf(format string, v ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString("FATAL ")
	buf.WriteString(format)

	errLogger.Output(2, fmt.Sprintf(buf.String(), v...))

	os.Exit(1)
}

// Errorf 打印错误日志
func Errorf(format string, v ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString("ERROR ")
	buf.WriteString(format)

	errLogger.Output(2, fmt.Sprintf(buf.String(), v...))
}

// Infof 打印信息日志
func Infof(format string, v ...interface{}) {
	var buf bytes.Buffer
	buf.WriteString("INFO ")
	buf.WriteString(format)

	infoLogger.Output(2, fmt.Sprintf(buf.String(), v...))
}
