// Author: lipixun
// Created Time : 六 12/10 13:03:33 2016
//
// File Name: types.go
// Description:
//
package dirdetector

var (
	Detectors map[string]DirDetector = map[string]DirDetector{
		"git": newGitDirDetector(),
	}
)

type DirDetector interface {
	Detect(p string) (string, error)
}
