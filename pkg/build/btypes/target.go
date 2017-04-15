// Author: lipixun
// Created Time : 一  3/13 17:20:45 2017
//
// File Name: target.go
// Description:

package btypes

// Target the target
type Target interface {
	// Name Get target name
	Name() string
	// Fullname Get target full name (in the repository)
	Fullname() string
	// Path Get target local path
	Path() string
	// Package Get the package this target belongs to
	Package() Package
}

// TargetImpl The target implementation
type TargetImpl interface {
}

type _Target struct {
	name     string
	fullname string
	path     string
	pkg      Package
	impl     TargetImpl
}
