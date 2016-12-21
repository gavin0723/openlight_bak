// Author: lipixun
// Created Time : 五 10/21 10:46:17 2016
//
// File Name: fs.go
// Description:
//	The workspace file system
package workspace

import (
	"errors"
	"github.com/ops-openlight/openlight/pkg/util"
	"os"
	"path/filepath"
)

type WorkDir struct {
	path     string
	ws       *Workspace
	readOnly bool
}

func newWorkDir(p string, ws *Workspace, readOnly bool) (*WorkDir, error) {
	// Get the real path
	p, err := util.GetRealPath(p)
	if err != nil {
		return nil, err
	}
	// Create workdir and initialize
	wd := new(WorkDir)
	wd.path = p
	wd.ws = ws
	wd.readOnly = readOnly
	if err := wd.initialize(); err != nil {
		return nil, err
	}
	// Done
	return wd, nil
}

func (this *WorkDir) initialize() error {
	if !this.readOnly {
		if info, err := os.Stat(this.path); err != nil {
			if os.IsNotExist(err) {
				// Not exist, create new one
				if err := os.MkdirAll(this.path, os.ModePerm); err != nil {
					return err
				}
			}
		} else if !info.IsDir() {
			return errors.New("Not a directory")
		}
	}
	return nil
}

// Get the workdir root path
func (this *WorkDir) RootPath() string {
	return this.path
}

// Get the workdir is readonly
func (this *WorkDir) ReadOnly() bool {
	return this.readOnly
}

// Get (and ensure) the path, name is relative to workdir root path
func (this *WorkDir) GetPath(name string) (string, error) {
	p := filepath.Join(this.path, name)
	if !this.readOnly {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", err
		}
	}
	return p, nil
}

// Clean this workdir
func (this *WorkDir) Clean() error {
	if this.readOnly {
		return errors.New("Cannot clean readonly workdir")
	}
	// Clean
	return nil
}
