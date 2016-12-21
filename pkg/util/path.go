// Author: lipixun
// Created Time : 四 10/27 18:12:34 2016
//
// File Name: path.go
// Description:
//	The path helper
package util

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func GetRealPath(p string) (string, error) {
	// Extend the home directory is necessary
	if strings.HasPrefix(p, "~/") || p == "~" {
		// Replace the first ~
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		p = strings.Replace(p, "~", u.HomeDir, 1)
	}
	// Get the real absolute path of target path
	if !filepath.IsAbs(p) {
		var err error
		p, err = filepath.Abs(p)
		if err != nil {
			return "", err
		}
	}
	if info, err := os.Lstat(p); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		} else {
			return p, nil
		}
	} else if info.Mode()&os.ModeSymlink != 0 {
		// A symbol link
		return filepath.EvalSymlinks(p)
	} else {
		return p, nil
	}
}
