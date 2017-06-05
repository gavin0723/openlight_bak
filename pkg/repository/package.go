// Author: lipixun
// File Name: package.go
// Description:

package repository

import (
	"errors"

	"path/filepath"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Package represents a package
type Package struct {
	path string

	// The spec
	spec *pbSpec.Package
	// The repository
	repo *LocalRepository

	// The references
	references map[string]*Reference
	// The targets
	targets map[string]*Target
}

func newPackage(path string, spec *pbSpec.Package, repo *LocalRepository) *Package {
	if spec == nil {
		panic(errors.New("Require spec"))
	}
	if repo == nil {
		panic(errors.New("Require repository"))
	}
	return &Package{path, spec, repo, nil, nil}
}

// Path returns the path
func (pkg *Package) Path() string {
	return pkg.path
}

// Spec returns the spec
func (pkg *Package) Spec() *pbSpec.Package {
	return pkg.spec
}

// GetRelativeTarget returns the target which is relatived to this package
func (pkg *Package) GetRelativeTarget(path, target string) (*Target, error) {
	if path == "" {
		return pkg.GetTarget(target), nil
	}
	// Get from repository
	relpkg, err := pkg.repo.GetPackage(filepath.Join(pkg.path, path))
	if err != nil {
		return nil, err
	}
	return relpkg.GetTarget(target), nil
}

// References returns the reference list
func (pkg *Package) References() []*Reference {
	if pkg.references == nil {
		pkg.loadReferences()
	}
	var refs []*Reference
	for _, ref := range pkg.references {
		refs = append(refs, ref)
	}
	return refs
}

// GetReference returns the reference
func (pkg *Package) GetReference(name string) *Reference {
	if pkg.references == nil {
		pkg.loadReferences()
	}
	return pkg.references[name]
}

func (pkg *Package) loadReferences() {
	pkg.references = make(map[string]*Reference)
	for name, refSpec := range pkg.spec.References {
		ref := newReference(name, refSpec)
		pkg.references[name] = ref
	}
}

// Targets returns the target list
func (pkg *Package) Targets() []*Target {
	if pkg.targets == nil {
		pkg.loadTargets()
	}
	var targets []*Target
	for _, target := range pkg.targets {
		targets = append(targets, target)
	}
	return targets
}

// GetTarget returns the target
func (pkg *Package) GetTarget(name string) *Target {
	if pkg.targets == nil {
		pkg.loadTargets()
	}
	return pkg.targets[name]
}

func (pkg *Package) loadTargets() {
	pkg.targets = make(map[string]*Target)
	for name, targetSpec := range pkg.spec.Targets {
		ref := newTarget(name, pkg.path, targetSpec, pkg)
		pkg.targets[name] = ref
	}
}
