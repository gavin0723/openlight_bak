// Author: lipixun
// Created Time : 二 11/22 14:30:08 2016
//
// File Name: workdir.go
// Description:
//	The cli work directory
package workdir

import (
	"github.com/ops-openlight/openlight/helper/iohelper"
	"os"
	"path/filepath"
)

const (
	ConfigDirName = "config"
	RunnerDirName = "run"
)

type WorkDir struct {
	path string
}

func NewWorkDir(p string) (*WorkDir, error) {
	p, err := iohelper.GetRealPath(p)
	if err != nil {
		return nil, err
	}
	return &WorkDir{path: p}, nil
}

// Ensure this work dir
func (this *WorkDir) Ensure() error {
	if err := os.MkdirAll(this.path, os.ModePerm); err != nil {
		return err
	}
	// Done
	return nil
}

func (this *WorkDir) GetConfigPath() string {
	return filepath.Join(this.path, ConfigDirName)
}

func (this *WorkDir) GetRunnerPath() string {
	return filepath.Join(this.path, RunnerDirName)
}
