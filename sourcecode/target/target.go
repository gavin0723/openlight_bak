// Author: lipixun
// Created Time : 二 10/18 21:59:20 2016
//
// File Name: target.go
// Description:
//	The build target
package target

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/sourcecode/repository"
	"github.com/ops-openlight/openlight/sourcecode/spec"
)

var (
	// The builder creators
	BuilderCreators = map[string]func(s *spec.Target) (Builder, error){
		BuilderTypePython:   PythonBuilderCreator,
		BuilderTypeGolang:   GolangBuilderCreator,
		BuilderTypeMakefile: MakefileBuilderCreator,
		BuilderTypeShell:    ShellBuilderCreator,
		BuilderTypeDocker:   DockerBuilderCreator,
	}
	// The source code creators
	SourceCodeCreators = map[string]func(s *spec.Target) (SourceCode, error){
		SourceCodeTypeGolang: GolangSourceCodeCreator,
		SourceCodeTypePython: PythonSourceCodeCreator,
	}
)

type Target struct {
	Name       string
	Spec       *spec.Target
	Deps       []*TargetDependency
	Repository *repository.Repository
	Builder    Builder
	SourceCode SourceCode
}

type TargetDependency struct {
	Spec   spec.TargetDependency // The dependency spec
	Target *Target               // The resolved target object
}

func NewTarget(name string, s *spec.Target, r *repository.Repository) (*Target, error) {
	if r == nil {
		return nil, errors.New("Require repository")
	}
	// Create target
	target := Target{
		Name:       name,
		Spec:       s,
		Repository: r,
	}
	// Check builder
	if s.Builder != nil {
		// Create builder
		creator, ok := BuilderCreators[s.Builder.Type]
		if ok {
			builder, err := creator(s)
			if err != nil {
				return nil, err
			}
			// Set the builder
			target.Builder = builder
		} else {
			return nil, errors.New(fmt.Sprintf("Unknown builder type [%s]", s.Builder.Type))

		}
	}
	if s.SourceCode != nil {
		// Create source code
		creator, ok := SourceCodeCreators[s.SourceCode.Type]
		if ok {
			sourcecode, err := creator(s)
			if err != nil {
				return nil, err
			}
			// Set the sourcecode
			target.SourceCode = sourcecode
		} else {
			return nil, errors.New(fmt.Sprintf("Unknown sourcecode type [%s]", s.Builder.Type))

		}
	}
	// Done
	return &target, nil
}

func (this *Target) Key() string {
	// Get the unique key of target
	return fmt.Sprintf("%s:%s", this.Repository.Uri(), this.Name)
}

func GetTargetKey(target string, repo *repository.Repository) string {
	return fmt.Sprintf("%s:%s", repo.Uri(), target)
}
