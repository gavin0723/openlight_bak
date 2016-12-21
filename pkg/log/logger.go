// Author: lipixun
// Created Time : 四 10/20 16:46:05 2016
//
// File Name: logger.go
// Description:
//	The logger
package log

import (
	"fmt"
	"github.com/fatih/color"
	"io"
)

const (
	LevelAll     = 0
	LevelDebug   = 1
	LevelInfo    = 2
	LevelWarn    = 3
	LevelSuccess = 4
	LevelFail    = 5
	LevelError   = 10
	LevelNo      = 100
)

var (
	NoColor = color.New()

	DefaultOptions = Options{
		HeaderLength: 24,
		EnableColor:  false,
		ColorMapping: map[int]ColorSchema{
			LevelDebug:   NewColorSchema(color.New(color.FgCyan), color.New(color.FgCyan)),
			LevelInfo:    NewColorSchema(color.New(color.FgBlue), NoColor),
			LevelWarn:    NewColorSchema(color.New(color.FgYellow), color.New(color.FgYellow)),
			LevelSuccess: NewColorSchema(color.New(color.FgGreen), color.New(color.FgGreen)),
			LevelFail:    NewColorSchema(color.New(color.FgRed), color.New(color.FgRed)),
			LevelError:   NewColorSchema(color.New(color.FgRed), color.New(color.FgRed)),
		},
	}
)

type Logger interface {
	// Get / Set level
	GetLevel() int
	SetLevel(level int)
	GetDefaultLevel() int
	SetDefaultLevel(level int)
	// Get / Set Header
	GetDefaultHeader() string
	SetDefaultHeader(header string)
	// The options
	Options() *Options
	// Log
	Print(text ...interface{})
	Printf(format string, text ...interface{})
	Println(text ...interface{})
	LeveledPrint(level int, text ...interface{})
	LeveledPrintf(level int, format string, text ...interface{})
	LeveledPrintln(level int, text ...interface{})
	HeadedPrint(header string, text ...interface{})
	HeadedPrintf(header string, format string, text ...interface{})
	HeadedPrintln(header string, level int, text ...interface{})
	LeveledHeadedPrint(header string, level int, text ...interface{})
	LeveledHeadedPrintf(header string, level int, format string, text ...interface{})
	LeveledHeadedPrintln(header string, level int, text ...interface{})
	// Sub loggers
	GetLogger(level int, defaultLevel int, defaultHeader string) Logger
	GetLoggerWithHeader(defaultHeader string) Logger
}

type Options struct {
	HeaderLength int
	EnableColor  bool
	ColorMapping map[int]ColorSchema
}

type ColorSchema struct {
	HeaderColor  *color.Color
	MessageColor *color.Color
}

func NewColorSchema(header, message *color.Color) ColorSchema {
	if header == nil {
		header = NoColor
	}
	if message == nil {
		message = NoColor
	}
	return ColorSchema{HeaderColor: header, MessageColor: message}
}

func (this *Options) Copy() *Options {
	newOptions := Options{
		HeaderLength: this.HeaderLength,
		EnableColor:  this.EnableColor,
		ColorMapping: make(map[int]ColorSchema),
	}
	for l, c := range this.ColorMapping {
		newOptions.ColorMapping[l] = c
	}
	return &newOptions
}

type stdlogger struct {
	level         int
	defaultLevel  int
	defaultHeader string
	options       *Options
	writer        io.Writer
}

func New(writer io.Writer, level int, defaultLevel int, defaultHeader string) Logger {
	return &stdlogger{
		writer:        writer,
		level:         level,
		defaultLevel:  defaultLevel,
		defaultHeader: defaultHeader,
		options:       DefaultOptions.Copy(),
	}
}

func (this *stdlogger) GetLevel() int {
	return this.level
}

func (this *stdlogger) SetLevel(level int) {
	this.level = level
}

func (this *stdlogger) GetDefaultLevel() int {
	return this.defaultLevel
}

func (this *stdlogger) SetDefaultLevel(level int) {
	this.defaultLevel = level
}

func (this *stdlogger) GetDefaultHeader() string {
	return this.defaultHeader
}

func (this *stdlogger) SetDefaultHeader(header string) {
	this.defaultHeader = header
}

func (this *stdlogger) Options() *Options {
	return this.options
}

func (this *stdlogger) Print(text ...interface{}) {
	this.LeveledHeadedPrint(this.defaultHeader, this.defaultLevel, text...)
}

func (this *stdlogger) Printf(format string, text ...interface{}) {
	this.LeveledHeadedPrintf(this.defaultHeader, this.defaultLevel, format, text...)
}

func (this *stdlogger) Println(text ...interface{}) {
	this.LeveledHeadedPrintln(this.defaultHeader, this.defaultLevel, text...)
}

func (this *stdlogger) LeveledPrint(level int, text ...interface{}) {
	this.LeveledHeadedPrint(this.defaultHeader, level, text...)
}

func (this *stdlogger) LeveledPrintf(level int, format string, text ...interface{}) {
	this.LeveledHeadedPrintf(this.defaultHeader, level, format, text...)
}

func (this *stdlogger) LeveledPrintln(level int, text ...interface{}) {
	this.LeveledHeadedPrintln(this.defaultHeader, level, text...)
}

func (this *stdlogger) HeadedPrint(header string, text ...interface{}) {
	this.LeveledHeadedPrint(header, this.defaultLevel, text...)
}

func (this *stdlogger) HeadedPrintf(header string, format string, text ...interface{}) {
	this.LeveledHeadedPrintf(header, this.defaultLevel, format, text...)
}

func (this *stdlogger) HeadedPrintln(header string, level int, text ...interface{}) {
	this.LeveledHeadedPrintln(header, this.defaultLevel, text...)
}

func (this *stdlogger) LeveledHeadedPrint(header string, level int, text ...interface{}) {
	// Check level
	if level < this.level {
		return
	}
	// Check color
	if this.options.EnableColor {
		headerColor, messageColor := NoColor, NoColor
		c, ok := this.options.ColorMapping[level]
		if ok {
			headerColor, messageColor = c.HeaderColor, c.MessageColor
		}
		// Check header
		if header == "" {
			fmt.Fprint(this.writer, messageColor.SprintFunc()(text...))
		} else {
			header = fmt.Sprintf(fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprint(this.writer, headerColor.SprintFunc()(header))
			fmt.Fprint(this.writer, messageColor.SprintFunc()(text...))
		}
	} else {
		// Check header
		if header == "" {
			fmt.Fprint(this.writer, text...)
		} else {
			fmt.Fprintf(this.writer, fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprint(this.writer, text...)
		}
	}
}

func (this *stdlogger) LeveledHeadedPrintf(header string, level int, format string, text ...interface{}) {
	// Check level
	if level < this.level {
		return
	}
	// Check color
	if this.options.EnableColor {
		headerColor, messageColor := NoColor, NoColor
		c, ok := this.options.ColorMapping[level]
		if ok {
			headerColor, messageColor = c.HeaderColor, c.MessageColor
		}
		// Check header
		if header == "" {
			fmt.Fprint(this.writer, messageColor.SprintfFunc()(format, text...))
		} else {
			header = fmt.Sprintf(fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprint(this.writer, headerColor.SprintFunc()(header))
			fmt.Fprint(this.writer, messageColor.SprintfFunc()(format, text...))
		}
	} else {
		// Check header
		if header == "" {
			fmt.Fprintf(this.writer, format, text...)
		} else {
			fmt.Fprintf(this.writer, fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprintf(this.writer, format, text...)
		}
	}
}

func (this *stdlogger) LeveledHeadedPrintln(header string, level int, text ...interface{}) {
	// Check level
	if level < this.level {
		return
	}
	// Check color
	if this.options.EnableColor {
		headerColor, messageColor := NoColor, NoColor
		c, ok := this.options.ColorMapping[level]
		if ok {
			headerColor, messageColor = c.HeaderColor, c.MessageColor
		}
		// Check header
		if header == "" {
			fmt.Fprintln(this.writer, messageColor.SprintFunc()(text...))
		} else {
			header = fmt.Sprintf(fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprint(this.writer, headerColor.SprintFunc()(header))
			fmt.Fprint(this.writer, messageColor.SprintlnFunc()(text...))
		}
	} else {
		// Check header
		if header == "" {
			fmt.Fprintln(this.writer, text...)
		} else {
			fmt.Fprintf(this.writer, fmt.Sprintf("[%%-%ds] ", this.options.HeaderLength), header)
			fmt.Fprintln(this.writer, text...)
		}
	}
}

func (this *stdlogger) GetLogger(level int, defaultLevel int, defaultHeader string) Logger {
	return &stdlogger{
		level:         level,
		defaultLevel:  level,
		defaultHeader: defaultHeader,
		options:       this.options.Copy(),
		writer:        this.writer,
	}
}

func (this *stdlogger) GetLoggerWithHeader(defaultHeader string) Logger {
	return &stdlogger{
		level:         this.level,
		defaultLevel:  this.defaultLevel,
		defaultHeader: defaultHeader,
		options:       this.options.Copy(),
		writer:        this.writer,
	}
}
