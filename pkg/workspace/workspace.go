// Author: lipixun
// Created Time : 二 10/18 21:53:23 2016
//
// File Name: workspace.go
// Description:
//	The workspace
package workspace

import (
	"errors"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/workspace/dirdetector"
	"os"
)

const (
	WorkspaceLogHeader = "Workspace"
)

type Workspace struct {
	Verbose bool
	Logger  log.Logger
	Dir     struct {
		Global  *WorkDir
		User    *WorkDir
		Project *WorkDir
	}
	Options WorkspaceOptions
}

// Create new default worksapce
func DefaultNew() (*Workspace, error) {
	return New(nil, nil)
}

// Create new workspace
func New(options *WorkspaceOptions, logger log.Logger) (*Workspace, error) {
	ws := new(Workspace)
	if options == nil {
		options = NewWorkspaceOptions()
	}
	if logger == nil {
		if options.Verbose {
			logger = log.New(os.Stderr, log.LevelDebug, log.LevelInfo, WorkspaceLogHeader)
		} else {
			logger = log.New(os.Stderr, log.LevelInfo, log.LevelInfo, WorkspaceLogHeader)
		}
	} else {
		logger = logger.GetLoggerWithHeader(WorkspaceLogHeader)
	}
	logger.Options().EnableColor = options.EnableColor
	// Set the workspace
	ws.Verbose = options.Verbose
	ws.Logger = logger
	ws.Options = *options
	// Initialize work dir
	if err := ws.initWorkDir(&options.Dir); err != nil {
		return nil, err
	}
	// Done
	return ws, nil
}

// Init the work dir
func (this *Workspace) initWorkDir(options *WorkDirOptions) error {
	var err error
	// Create the global work dir
	if options.GlobalPath == "" {
		return errors.New("No global workdir path defined")
	}
	this.Logger.LeveledPrintf(log.LevelDebug, "Set global workdir to: %s\n", options.GlobalPath)
	this.Dir.Global, err = newWorkDir(options.GlobalPath, this, true)
	if err != nil {
		return err
	}
	// Create the user work dir
	if options.UserPath == "" {
		return errors.New("No user workdir path defined")
	}
	this.Logger.LeveledPrintf(log.LevelDebug, "Set user workdir to: %s\n", options.UserPath)
	this.Dir.User, err = newWorkDir(options.UserPath, this, false)
	if err != nil {
		return err
	}
	// Create the project workdir
	var projectPath string
	if options.ProjectPath != "" {
		projectPath = options.ProjectPath
	} else if options.CurrentPathAsProjectPath {
		projectPath, err = os.Getwd()
		if err != nil {
			return err
		}
	} else {
		p, err := os.Getwd()
		if err != nil {
			return err
		}
		// Detect project path
		for name, detector := range dirdetector.Detectors {
			projectPath, err = detector.Detect(p)
			if err != nil {
				this.Logger.LeveledPrintf(log.LevelDebug, "Project path workdir detector [%s] returns error: %s\n", name, err)
			} else {
				this.Logger.LeveledPrintf(log.LevelDebug, "Project path workdir detector [%s] returns path: %s\n", name, projectPath)
				break
			}
		}
		if projectPath == "" {
			return errors.New("No project path detected")
		}
	}
	this.Logger.LeveledPrintf(log.LevelDebug, "Set project workdir to: %s\n", projectPath)
	this.Dir.Project, err = newWorkDir(projectPath, this, false)
	if err != nil {
		return err
	}
	// Done
	return nil
}
