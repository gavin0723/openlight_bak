// Author: lipixun
// Created Time : 二 10/18 21:53:23 2016
//
// File Name: workspace.go
// Description:
//	The workspace
package workspace

import (
	"errors"
	"github.com/ops-openlight/openlight"
	"github.com/ops-openlight/openlight/sourcecode/repository"
	"github.com/ops-openlight/openlight/uri"
	"os"
)

type WorkspaceOptions struct {
	Verbose    bool
	Repository struct {
		OnlyLocal bool
		LoadLocal bool
	}
	Docker struct {
		Uri string
	}
}

func NewWorkspaceOptions() *WorkspaceOptions {
	options := new(WorkspaceOptions)
	return options
}

func DefaultWorkspaceOptions() *WorkspaceOptions {
	options := NewWorkspaceOptions()
	options.Repository.OnlyLocal = false
	options.Repository.LoadLocal = false
	return options
}

func LocalWorkspaceOptions() *WorkspaceOptions {
	options := NewWorkspaceOptions()
	options.Repository.OnlyLocal = true
	options.Repository.LoadLocal = true
	return options
}

type Workspace struct {
	options    *WorkspaceOptions
	Logger     *openlight.Logger
	FileSystem WorkspaceFileSystem
	Repository *repository.RepositoryManager
}

func NewWorkspace(fs WorkspaceFileSystem, options *WorkspaceOptions, logger *openlight.Logger) (*Workspace, error) {
	// Check parameters
	if fs == nil {
		return nil, errors.New("Require file system")
	}
	if logger == nil {
		var level int
		if options.Verbose {
			level = openlight.LogLevelAll
		} else {
			level = openlight.LogLevelInfo
		}
		logger = openlight.NewLogger(os.Stderr, level)
	}
	// Create workspace
	return &Workspace{
		FileSystem: fs,
		options:    options,
		Logger:     logger,
		Repository: repository.NewRepositoryManager(options.Repository.LoadLocal, options.Repository.OnlyLocal),
	}, nil
}

func (this *Workspace) Verbose() bool {
	// Verbose or not
	return this.options.Verbose
}

func (this *Workspace) DefaultRepositoryPathFunction(reference *uri.RepositoryReference) (string, error) {
	return this.FileSystem.GetSourcePath(reference.String(), true)
}
