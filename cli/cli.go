// Author: lipixun
// Created Time : 二 11/22 14:37:16 2016
//
// File Name: cli.go
// Description:
//	The cli object
package cli

import (
	"github.com/ops-openlight/openlight/cli/workdir"
	"github.com/ops-openlight/openlight/log"
	"gopkg.in/urfave/cli.v1"
	"io"
	"os"
)

const (
	LogHeader      = "OpenlightCLI"
	DefaultWorkDir = "~/.op"
)

type OpenlightCLI struct {
	// I/O
	inputStream  io.Reader
	outputStream io.Writer
	errorStream  io.Writer
	// The logger
	logger log.Logger
	// The workdir
	workdir *workdir.WorkDir
	// Parameters
	debug bool
}

func New(input io.Reader, output, errStream io.Writer, logger log.Logger, workdir *workdir.WorkDir, debug bool) (*OpenlightCLI, error) {
	if err := workdir.Ensure(); err != nil {
		return nil, err
	}
	return &OpenlightCLI{
		inputStream:  input,
		outputStream: output,
		errorStream:  errStream,
		logger:       logger,
		workdir:      workdir,
		debug:        debug,
	}, nil
}

func NewStandard(workdirPath string, debug bool) (*OpenlightCLI, error) {
	w, err := workdir.NewWorkDir(workdirPath)
	if err != nil {
		return nil, err
	}
	logLevel := log.LevelInfo
	if debug {
		logLevel = log.LevelDebug
	}
	logger := log.NewLogger(os.Stderr, logLevel, log.LevelInfo, LogHeader)
	options := logger.Options()
	options.EnableColor = true
	// Create the cli
	return New(os.Stdin, os.Stdout, os.Stderr, logger, w, debug)
}

func NewFromCLI(c *cli.Context) (*OpenlightCLI, error) {
	verbose := c.GlobalBool("verbose")
	workdir := c.GlobalString("workdir")
	if workdir == "" {
		workdir = DefaultWorkDir
	}
	return NewStandard(workdir, verbose)
}

func (this *OpenlightCLI) InputStream() io.Reader {
	return this.inputStream
}

func (this *OpenlightCLI) OutputStream() io.Writer {
	return this.outputStream
}

func (this *OpenlightCLI) ErrorStream() io.Writer {
	return this.errorStream
}

// Return the logger, this method ensures the return value is not nil
func (this *OpenlightCLI) Logger() log.Logger {
	return this.logger
}

// Return the work dir
func (this *OpenlightCLI) WorkDir() *workdir.WorkDir {
	return this.workdir
}

// Is debug
func (this *OpenlightCLI) Debug() bool {
	return this.debug
}
