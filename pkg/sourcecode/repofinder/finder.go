// Author: lipixun
// Created Time : 六 12/10 23:46:19 2016
//
// File Name: finder.go
// Description:
//	The repository finder
package repofinder

import (
	"github.com/ops-openlight/openlight/pkg/workspace"
)

var (
	finders map[string]Finder = map[string]Finder{
		FinderTypeGolang: NewGolangFinder(),
		FinderTypePython: NewPythonFinder(),
	}
)

type Finder interface {
	Find(ws *workspace.Workspace, params map[string]interface{}) ([]string, error)
}

func GetFinder(t string) Finder {
	return finders[t]
}
