// Author: lipixun
// File Name: target.go
// Description:

package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Target represents a target
type Target struct {
	key  string
	name string
	path string
	spec *pbSpec.Target
	pkg  *Package
}

func newTarget(name, path string, spec *pbSpec.Target, pkg *Package) *Target {
	if spec == nil {
		panic(errors.New("Require spec"))
	}
	if pkg == nil {
		panic(errors.New("Require package"))
	}
	// Get key
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			panic(fmt.Errorf("Failed to get absoluate path of path [%v]: %v", path, err))
		}
	}
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		panic(fmt.Errorf("Failed to get inode of path [%v]: %v", path, err))
	}
	key := fmt.Sprintf("%v:%v", stat.Ino, name)
	// Create target
	return &Target{key, name, path, spec, pkg}
}

// Key returns the target key
func (target *Target) Key() string {
	return target.key
}

// Name returns the name
func (target *Target) Name() string {
	return target.name
}

// Path returns the path
func (target *Target) Path() string {
	return target.path
}

// Spec returns the spec
func (target *Target) Spec() *pbSpec.Target {
	return target.spec
}

// Package returns the package this target belongs to
func (target *Target) Package() *Package {
	return target.pkg
}

// InitOutputDir inits output dir for this target
func (target *Target) InitOutputDir(name string) (string, error) {
	dirname := target.GetOutputDir(name)
	if info, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			// Create a new output dir
			if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
				return "", fmt.Errorf("Failed to create directory [%v]: %v", dirname, err)
			}
		} else {
			return "", fmt.Errorf("Failed to check path [%v]: %v", dirname, err)
		}
	} else if !info.IsDir() {
		return "", fmt.Errorf("Target output path [%v] is not a directory", dirname)
	}
	// Done
	return dirname, nil
}

// GetOutputDir returns the output dir name
func (target *Target) GetOutputDir(name string) string {
	return filepath.Join(target.path, "op-out", name)
}
