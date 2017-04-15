// Author: lipixun
// Created Time : 日  3/12 15:34:05 2017
//
// File Name: model.go
// Description:
//
//	The data models

package spec

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	// PackageFinderTypeLocalPython The local python package finder type
	PackageFinderTypeLocalPython = "local-python"
	// TargetTypeGoBinary The go binary target type
	TargetTypeGoBinary = "go-binary"
	// TargetTypePythonLib The python library target type
	TargetTypePythonLib = "python-lib"
)

// Options The default options type
type Options map[string]interface{}

// NewOptions Create a new Options
func NewOptions() Options {
	return make(Options)
}

// Set Set value to options
func (o Options) Set(key string, value interface{}) error {
	if o == nil {
		return errors.New("Cannot set to nil options")
	}
	if key == "" {
		return errors.New("Require key")
	}
	if value == nil {
		return errors.New("Require value")
	}
	// Check value type
	switch value.(type) {
	case bool:
		o[key] = BoolOptionValue(value.(bool))
	case string:
		o[key] = StringOptionValue(value.(string))
	case []string:
		o[key] = StringSliceOptionValue(value.([]string))
	case int:
		o[key] = NumberOptionValue(value.(int))
	case int16:
		o[key] = NumberOptionValue(value.(int16))
	case int32:
		o[key] = NumberOptionValue(value.(int32))
	case int64:
		o[key] = NumberOptionValue(value.(int64))
	case float32:
		o[key] = NumberOptionValue(value.(float32))
	case float64:
		o[key] = NumberOptionValue(value.(float64))
	case BoolOptionValue:
		o[key] = value
	case StringOptionValue:
		o[key] = value
	case StringSliceOptionValue:
		o[key] = value
	case NumberOptionValue:
		o[key] = value
	case Options:
		o[key] = value
	default:
		return fmt.Errorf("Unsupported option value type: %s", reflect.TypeOf(value))
	}
	// Done
	return nil
}

// Merge Merge options
func (o Options) Merge(opt Options) {
	for key, value := range opt {
		o[key] = value
	}
}

// BoolOptionValue The bool value of option
type BoolOptionValue bool

// StringOptionValue The string value of option
type StringOptionValue string

// StringSliceOptionValue The string slice value of option
type StringSliceOptionValue []string

// NumberOptionValue The number value of option
type NumberOptionValue float64

// ReferenceSpec A general reference data model
type ReferenceSpec struct {
	// The package name. Empty means the package itself
	Package string `json:"package" yaml:"package"`
	// The subname in the package. Could be target, filename ... or target/filename
	SubNames []string `json:"subNames" yaml:"subNames"`
}

// ParseReferenceFromString Parse reference from string
func ParseReferenceFromString(s string) (*ReferenceSpec, error) {
	var reference ReferenceSpec
	parts := strings.Split(s, ":")
	if len(parts) > 3 {
		return nil, errors.New("Malformed reference. Too many : found")
	}
	if len(parts) == 1 {
		if strings.HasPrefix(parts[0], "//") {
			reference.Package = parts[0][2:]
		} else {
			reference.SubNames = parts
		}
	} else if len(parts) == 2 {
		if strings.HasPrefix(parts[0], "//") {
			reference.Package = parts[0][2:]
			reference.SubNames = []string{parts[1]}
		} else {
			reference.SubNames = parts
		}
	} else {
		if !strings.HasPrefix(parts[0], "//") {
			return nil, errors.New("Malformed reference. Expect // at beginning")
		}
		reference.Package = parts[0][2:]
		reference.SubNames = parts[1:]
	}
	// Done
	return &reference, nil
}

// BuildSpec The overall build data model
type BuildSpec struct {
	// The package
	Package *PackageSpec `json:"package" yaml:"references"`
	// The references
	References []*PackageReferenceSpec `json:"references" yaml:"references"`
	// The targets
	Targets []*TargetSpec `json:"targets" yaml:"targets"`
}

// GetReference Get a reference by package name
func (s *BuildSpec) GetReference(name string) *PackageReferenceSpec {
	for _, spec := range s.References {
		if spec.Name == name {
			return spec
		}
	}
	// Not found
	return nil
}

// GetTarget Get a target by target name
func (s *BuildSpec) GetTarget(name string) *TargetSpec {
	for _, spec := range s.Targets {
		if spec.Name == name {
			return spec
		}
	}
	// Not found
	return nil
}

// -*- -------------------------------------------------- Package spec -------------------------------------------------- -*-

// PackageSpec The package data model
type PackageSpec struct {
	// The package name
	Name string `json:"name" yaml:"name"`
	// The package options
	Options Options `json:"options" yaml:"options"`
	// The targets
	Targets []*TargetSpec `json:"targets" yaml:"targets"`
}

// PackageReferenceSpec the package reference data model
type PackageReferenceSpec struct {
	// The package name
	Name string `json:"name" yaml:"name"`
	// The remote uri
	Remote string `json:"remote" yaml:"remote"`
	// The finder spec
	Finders []*PackageFinderSpec `json:"finders" yaml:"finders"`
}

// PackageFinderSpec The package finder data model
type PackageFinderSpec struct {
	// The finder type
	Type string `json:"type" yaml:"type"`
	// The local python finder
	LocalPython *PackageLocalPythonFinderSpec `json:"localPython" yaml:"localPython"`
}

// PackageLocalPythonFinderSpec The local python package finder data model
type PackageLocalPythonFinderSpec struct {
	// The python module name
	Module string `json:"module" yaml:"module"`
	// The number of parent directory levels
	Options Options `json:"options" yaml:"options"`
}

// -*- -------------------------------------------------- Target spec -------------------------------------------------- -*-

// TargetSpec The target data model
type TargetSpec struct {
	// The target name
	Name string `json:"name" yaml:"name"`
	// The target type
	Type string `json:"type" yaml:"type"`
	// The go binary
	GoBinary *GoBinaryTargetSpec `json:"goBinary" yaml:"goBinary"`
	// The python library
	PythonLib *PythonLibTargetSpec `json:"pythonLib" yaml:"pythonLib"`
	// The target options
	Options Options `json:"options" yaml:"options"`
	// The target dependencies
	Dependencies []*DependencySpec `json:"dependencies" yaml:"dependencies"`
	// The specific dependencies
	GoDependencies  []*GoDependencySpec  `json:"goDependencies" yaml:"goDependencies"`
	PipDependencies []*PipDependencySpec `json:"pipDependencies" yaml:"pipDependencies"`
}

// GoBinaryTargetSpec The go binary target data model
type GoBinaryTargetSpec struct {
	// The go package to build
	Package string `json:"package" yaml:"package"`
}

// PythonLibTargetSpec The python library target data model
type PythonLibTargetSpec struct{}

// DependencySpec The dependency data model
type DependencySpec struct {
	// The reference
	Reference *ReferenceSpec `json:"reference" yaml:"reference"`
	// The options
	Options Options `json:"options" yaml:"options"`
}

// GoDependencySpec The go dependency data model
type GoDependencySpec struct {
	// The dependent go packages. This packages defined here will not be controlled by target reference
	// More specific, the packages defined here will be resolved by "go get -u <packages>"
	// A package cache will be used to speed up this progress
	Packages []string `json:"packages" yaml:"packages"`
}

// PipDependencySpec The pip python dependency data model
type PipDependencySpec struct {
	Modules []string `json:"modules" yaml:"modules"`
}
