// Author: lipixun
// Created Time : 六 12/10 23:46:32 2016
//
// File Name: python.go
// Description:
//	Find repository by python path
//	Required parameters:
// 		module 			string 	The module name to find, required
// 		parent  		int 	The parent level count
//

package repofinder

import (
	"context"
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/workspace"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	FinderTypePython = "python"

	PythonFinderParamModule = "module"
	PythonFinderParamParent = "parent"

	PythonTryImportTimeoutSeconds = 10 // 10s
	PythonTryImportScript         = `import sys
import imp
file, pathName, description = imp.find_module("%s")
if pathName:
 	print pathName
	sys.exit(0)
else:
	sys.exit(0)
`
)

type PythonFinder struct{}

func NewPythonFinder() PythonFinder {
	return PythonFinder{}
}

func (this PythonFinder) Find(ws *workspace.Workspace, params map[string]interface{}) ([]string, error) {
	_module, ok := params[PythonFinderParamModule]
	if !ok {
		return nil, errors.New("Require module parameter")
	}
	module, ok := _module.(string)
	if !ok {
		return nil, errors.New("Invalid value of module parameter")
	}
	parent := 0
	_parent, ok := params[PythonFinderParamParent]
	if ok {
		parent, ok = _parent.(int)
		if !ok {
			return nil, errors.New("Invalid value of parent parameter")
		}
	}
	// Create and run the python command
	ctx, cancel := context.WithTimeout(context.Background(), PythonTryImportTimeoutSeconds*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python")
	// Write the code by stdin
	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	go func() {
		writer.Write([]byte(fmt.Sprintf(PythonTryImportScript, module)))
		writer.Close()
	}()
	// Start the command
	rtn, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	// Get the output as the path
	path := strings.Trim(string(rtn), "\t\r\n")
	if path == "" {
		// Path not found
		return nil, nil
	}
	// Get with parent
	for i := 0; i < parent; i++ {
		path = filepath.Dir(path)
	}
	// Done
	return []string{path}, nil
}
