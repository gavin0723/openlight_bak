// Author: lipixun
// Created Time : 四 10/20 16:46:05 2016
//
// File Name: logger.go
// Description:
//	The logger
package openlight

import (
	"fmt"
	"github.com/fatih/color"
	"io"
)

const (
	LogLevelAll   = 0
	LogLevelDebug = 1
	LogLevelInfo  = 2
	LogLevelWarn  = 3
	LogLevelError = 4
	LogLevelNone  = 10
)

type Logger struct {
	Level  int
	writer io.Writer
}

func NewLogger(writer io.Writer, level int) *Logger {
	return &Logger{
		Level:  level,
		writer: writer,
	}
}

func (this *Logger) NewLine() {
	this.writer.Write([]byte("\n"))
}

func (this *Logger) WriteMessage(level int, texts ...interface{}) {
	if level <= this.Level {
		this.writer.Write([]byte(fmt.Sprint(texts...)))
	}
}

func (this *Logger) WriteMessageln(level int, texts ...interface{}) {
	if level <= this.Level {
		this.writer.Write([]byte(fmt.Sprint(texts...)))
		this.NewLine()
	}
}

func (this *Logger) WriteSuccess(texts ...interface{}) {
	this.writer.Write([]byte(color.GreenString(fmt.Sprint(texts))))
}

func (this *Logger) WriteSuccessln(texts ...interface{}) {
	this.writer.Write([]byte(color.GreenString(fmt.Sprint(texts))))
	this.NewLine()
}

func (this *Logger) WriteSuccessHeader(header string, texts ...interface{}) {
	this.writer.Write([]byte(color.GreenString(fmt.Sprintf("[%-24s]", header))))
	this.writer.Write([]byte(color.GreenString(fmt.Sprint(texts...))))
}

func (this *Logger) WriteSuccessHeaderln(header string, texts ...interface{}) {
	this.writer.Write([]byte(color.GreenString(fmt.Sprintf("[%-24s]", header))))
	this.writer.Write([]byte(color.GreenString(fmt.Sprint(texts...))))
	this.NewLine()
}

func (this *Logger) WriteDebug(texts ...interface{}) {
	if this.Level <= LogLevelDebug {
		this.writer.Write([]byte(color.BlueString(fmt.Sprint(texts))))
	}
}

func (this *Logger) WriteDebugln(texts ...interface{}) {
	if this.Level <= LogLevelDebug {
		this.writer.Write([]byte(color.BlueString(fmt.Sprint(texts))))
		this.NewLine()
	}
}

func (this *Logger) WriteDebugHeader(header string, texts ...interface{}) {
	if this.Level <= LogLevelDebug {
		this.writer.Write([]byte(color.BlueString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(fmt.Sprint(texts...)))
	}
}

func (this *Logger) WriteDebugHeaderln(header string, texts ...interface{}) {
	if this.Level <= LogLevelDebug {
		this.writer.Write([]byte(color.BlueString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(fmt.Sprint(texts...)))
		this.NewLine()
	}
}

func (this *Logger) WriteInfo(texts ...interface{}) {
	if this.Level <= LogLevelInfo {
		this.writer.Write([]byte(color.BlueString(fmt.Sprint(texts))))
	}
}

func (this *Logger) WriteInfoln(texts ...interface{}) {
	if this.Level <= LogLevelInfo {
		this.writer.Write([]byte(color.BlueString(fmt.Sprint(texts))))
		this.NewLine()
	}
}

func (this *Logger) WriteInfoHeader(header string, texts ...interface{}) {
	if this.Level <= LogLevelInfo {
		this.writer.Write([]byte(color.BlueString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(fmt.Sprint(texts...)))
	}
}

func (this *Logger) WriteInfoHeaderln(header string, texts ...interface{}) {
	if this.Level <= LogLevelInfo {
		this.writer.Write([]byte(color.BlueString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(fmt.Sprint(texts...)))
		this.NewLine()
	}
}

func (this *Logger) WriteWarning(texts ...interface{}) {
	if this.Level <= LogLevelWarn {
		this.writer.Write([]byte(color.YellowString(fmt.Sprint(texts))))
	}
}

func (this *Logger) WriteWarningln(texts ...interface{}) {
	if this.Level <= LogLevelWarn {

		this.writer.Write([]byte(color.YellowString(fmt.Sprint(texts))))
		this.NewLine()
	}
}

func (this *Logger) WriteWarningHeader(header string, texts ...interface{}) {
	if this.Level <= LogLevelWarn {
		this.writer.Write([]byte(color.YellowString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(color.YellowString(fmt.Sprint(texts...))))
	}
}

func (this *Logger) WriteWarningHeaderln(header string, texts ...interface{}) {
	if this.Level <= LogLevelWarn {
		this.writer.Write([]byte(color.YellowString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(color.YellowString(fmt.Sprint(texts...))))
		this.NewLine()
	}
}

func (this *Logger) WriteError(texts ...interface{}) {
	if this.Level <= LogLevelError {
		this.writer.Write([]byte(color.RedString(fmt.Sprint(texts))))
	}
}

func (this *Logger) WriteErrorln(texts ...interface{}) {
	if this.Level <= LogLevelError {
		this.writer.Write([]byte(color.RedString(fmt.Sprint(texts))))
		this.NewLine()
	}
}
func (this *Logger) WriteErrorHeader(header string, texts ...interface{}) {
	if this.Level <= LogLevelError {
		this.writer.Write([]byte(color.RedString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(color.RedString(fmt.Sprint(texts...))))
	}
}

func (this *Logger) WriteErrorHeaderln(header string, texts ...interface{}) {
	if this.Level <= LogLevelError {
		this.writer.Write([]byte(color.RedString(fmt.Sprintf("[%-24s]", header))))
		this.writer.Write([]byte(color.RedString(fmt.Sprint(texts...))))
		this.NewLine()
	}
}
