// Author: lipixun
// Created Time : 二  3/21 23:28:03 2017
//
// File Name: pkg.go
// Description:

package btypes

import (
	"github.com/ops-openlight/openlight/pkg/build/errors"
)

// Package The package
type Package interface {
	// Name Get the package name
	Name() string
	// Path Get the path of the repository
	Path() string
	// Repository Get the repository of this package
	Repository() Repository
	// AddTarget Add a target
	AddTarget(target Target) error
}

// NewPackage Create a new package
func NewPackage(name string, path string, repository Repository) Package {
	return &_Package{
		name:       name,
		path:       path,
		repository: &repository,
		targets:    make(map[string]Target),
	}
}

type _Package struct {
	name       string
	path       string
	repository *Repository
	targets    map[string]Target
}

func (pkg *_Package) Name() string {
	return pkg.Name()
}

func (pkg *_Package) Path() string {
	return pkg.path
}

func (pkg *_Package) Repository() Repository {
	return *pkg.repository
}

func (pkg *_Package) AddTarget(target Target) error {
	if t := pkg.targets[target.Fullname()]; t != nil {
		return errors.ErrDuplicatedTarget
	}
	pkg.targets[target.Fullname()] = target
	return nil
}
