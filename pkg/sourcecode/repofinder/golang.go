// Author: lipixun
// Created Time : 六 12/10 23:46:50 2016
//
// File Name: golang.go
// Description:
//	Find repository by go path
//	Required parameters:
// 		package 		string 	The package to find
// 		parent  		int 	The parent level count
//

package repofinder

import (
	"errors"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"os"
	"path/filepath"
	"strings"
)

const (
	FinderTypeGolang = "golang"

	GolangFinderParamPackage = "package"
	GolangFinderParamParent  = "parent"
)

type GolangFinder struct{}

func NewGolangFinder() GolangFinder {
	return GolangFinder{}
}

func (this GolangFinder) Find(ws *workspace.Workspace, params map[string]interface{}) ([]string, error) {
	_package, ok := params[GolangFinderParamPackage]
	if !ok {
		return nil, errors.New("Require package parameter")
	}
	pkg, ok := _package.(string)
	if !ok {
		return nil, errors.New("Invalid value of package parameter")
	}
	parent := 0
	_parent, ok := params[GolangFinderParamParent]
	if ok {
		parent, ok = _parent.(int)
		if !ok {
			return nil, errors.New("Invalid value of parent parameter")
		}
	}
	// Get go path
	var goPath string
	for _, environ := range os.Environ() {
		if strings.HasPrefix(environ, "GOPATH=") {
			// Found the go path
			goPath = environ[7:]
			break
		}
	}
	// Find the package
	path := filepath.Join(goPath, pkg)
	if _, err := os.Stat(path); err != nil {
		// Not found
		return nil, nil
	}
	// Found it
	// Get with parent
	for i := 0; i < parent; i++ {
		path = filepath.Dir(path)
	}
	// Done
	return []string{path}, nil
}
