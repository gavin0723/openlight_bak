// Author: lipixun
// File Name: finders.go
// Description:

package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Finder is used to find local repository
type Finder interface {
	// Find local repository
	Find() (string, error)
}

// Ensure the interface is implements
var _ Finder = (*PythonFinder)(nil)
var _ Finder = (*GoFinder)(nil)

//
//	The python finder
//	Find local repository by searching python modules
//

const (
	_PythonTryImportTimeoutSeconds = 30 // 30s
	_PythonTryImportScript         = `import sys
import imp
file, pathName, description = imp.find_module("%s")
if pathName:
 	print pathName
	sys.exit(0)
else:
	sys.exit(0)
`
)

// PythonFinder implements Finder interface for python module
type PythonFinder struct {
	spec *pbSpec.PythonFinder
}

// NewPythonFinder creates a new PythonFinder
func NewPythonFinder(spec *pbSpec.PythonFinder) (*PythonFinder, error) {
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &PythonFinder{spec}, nil
}

// Spec returns the spec
func (finder *PythonFinder) Spec() *pbSpec.PythonFinder {
	return finder.spec
}

// Find local repository
func (finder *PythonFinder) Find() (string, error) {
	if finder.spec.Module == "" {
		return "", errors.New("Require module")
	}
	// Create and run the python command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), _PythonTryImportTimeoutSeconds*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python")
	// Write the code by stdin
	writer, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("Failed to run python import script: %v", err)
	}
	go func() {
		writer.Write([]byte(fmt.Sprintf(_PythonTryImportScript, finder.spec.Module)))
		writer.Close()
	}()
	// Start the command
	rtn, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run python import script: %v", err)
	}
	// Get the output as the path
	path := strings.Trim(string(rtn), "\t\r\n")
	if path == "" {
		return "", nil
	}
	// Get with parent
	parent := int(finder.spec.GetParent())
	for i := 0; i < parent; i++ {
		path = filepath.Dir(path)
	}
	// Done
	return path, nil
}

//
//	The go finder
//	Find local repository by go package
//

// GoFinder implements Finder interface for go package
type GoFinder struct {
	spec *pbSpec.GoFinder
}

// NewGoFinder creates a new GoFinder
func NewGoFinder(spec *pbSpec.GoFinder) (*GoFinder, error) {
	if spec == nil {
		return nil, errors.New("Require spec")
	}
	return &GoFinder{spec}, nil
}

// Spec returns the spec
func (finder *GoFinder) Spec() *pbSpec.GoFinder {
	return finder.spec
}

// Find local repository
func (finder *GoFinder) Find() (string, error) {
	if finder.spec.Package == "" {
		return "", errors.New("Require package parameter")
	}
	// Get go path
	var goPath string
	if finder.spec.Root != "" {
		goPath = finder.spec.Root
	} else {
		for _, environ := range os.Environ() {
			if strings.HasPrefix(environ, "GOPATH=") {
				// Found the go path
				goPath = environ[7:]
				break
			}
		}
		if goPath == "" {
			return "", errors.New("Environment variable [GOPATH] not found")
		}
	}
	// Find the package
	path := filepath.Join(goPath, finder.spec.Package)
	if _, err := os.Stat(path); err != nil {
		return "", nil
	}
	// Found it
	parent := int(finder.spec.Parent)
	for i := 0; i < parent; i++ {
		path = filepath.Dir(path)
	}
	// Done
	return path, nil
}
