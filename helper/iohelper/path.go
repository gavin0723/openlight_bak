// Author: lipixun
// Created Time : 四 10/27 18:12:34 2016
//
// File Name: path.go
// Description:
//	The path helper
package iohelper

import (
	"os"
	"path/filepath"
)

func GetRealPath(p string) (string, error) {
	// Get the real absolute path of target path
	if !filepath.IsAbs(p) {
		var err error
		p, err = filepath.Abs(p)
		if err != nil {
			return "", err
		}
	}
	if info, err := os.Lstat(p); err != nil {
		return "", err
	} else if info.Mode()&os.ModeSymlink != 0 {
		// A symbol link
		return filepath.EvalSymlinks(p)
	} else {
		return p, nil
	}
}
