// Author: lipixun
// File Name: reference.go
// Description:

package repository

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Reference represents the reference
type Reference struct {
	name string
	spec *pbSpec.Reference
}

// newReference creates a new Reference
func newReference(name string, spec *pbSpec.Reference) *Reference {
	if spec == nil {
		panic(errors.New("Require spec"))
	}
	return &Reference{name, spec}
}

// Name returns the name
func (ref *Reference) Name() string {
	return ref.name
}

// Spec returns the sepc
func (ref *Reference) Spec() *pbSpec.Reference {
	return ref.spec
}

// Find reference by finders
func (ref *Reference) Find() string {
	for _, finderSpec := range ref.spec.Finders {
		finder, err := ref.getFinder(finderSpec)
		if err != nil {
			log.Errorf("Reference [%v]: Failed to create finder [%v]: %v", ref.name, finderSpec.Name, err)
			continue
		}
		log.Debugf("Reference.Find [%v]: Find by [%v]", ref.name, finderSpec.Name)
		// Find
		path, err := finder.Find()
		if err != nil {
			log.Errorf("Reference [%v]: Failed to find by [%v]: %v", ref.name, finderSpec.Name, err)
		}
		if path != "" {
			log.Infof("Reference [%v]: Find [%v] by [%v]", ref.name, path, finderSpec.Name)
			return path
		}
	}
	// Not found
	return ""
}

func (ref *Reference) getFinder(spec *pbSpec.Finder) (Finder, error) {
	// Python
	pythonSpec := spec.GetPython()
	if pythonSpec != nil {
		finder, err := NewPythonFinder(pythonSpec)
		if err != nil {
			return nil, err
		}
		return finder, nil
	}
	// Go
	goSpec := spec.GetGo()
	if goSpec != nil {
		finder, err := NewGoFinder(goSpec)
		if err != nil {
			return nil, err
		}
		return finder, nil
	}
	// Unknown
	return nil, errors.New("No finder defined")
}

// Fetch the reference from remote
// Args:
//	root		The source code root directory
func (ref *Reference) Fetch(root string) (string, error) {
	return "", errors.New("Fetch repository from remote has not been implemented yet")
}
