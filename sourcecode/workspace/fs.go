// Author: lipixun
// Created Time : 五 10/21 10:46:17 2016
//
// File Name: fs.go
// Description:
//	The workspace file system
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

const (
	UserWorkspaceSymbolName  = "op-workspace"
	UserWorkspaceProfileName = ".op.workspace.json"

	UserWorksapceDataDirPrefix = "op-ws-"
	UserWorkspaceSourceCodeDir = "source"
	UserWorkspaceGenerateDir   = "generate"
	UserWorkspaceOutputDir     = "output"
)

type WorkspaceFileSystem interface {
	Path() string
	Initialize() error
	Clean() error
	GetSourcePath(dest string, ensure bool) (string, error)
	GetGeneratePath(dest string, ensure bool) (string, error)
	GetOutputPath(ensure bool) (string, error)
}

type UserWorkspaceFileSystem struct {
	userPath string
	dataPath string
}

func NewUserWorkspaceFileSystem(p string) *UserWorkspaceFileSystem {
	return &UserWorkspaceFileSystem{userPath: p}
}

type UserWorkspaceProfile struct {
	Path string `json:"path"`
}

func (this *UserWorkspaceFileSystem) Path() string {
	return this.userPath
}

func (this *UserWorkspaceFileSystem) Initialize() error {
	// Initialize the workspace file system
	// Try to load profile
	if err := this.tryLoadPreviousFileSystem(); err != nil {
		// Create new file system
		return this.newFileSystem()
	}
	return nil
}

func (this *UserWorkspaceFileSystem) tryLoadPreviousFileSystem() error {
	data, err := ioutil.ReadFile(path.Join(this.userPath, UserWorkspaceProfileName))
	if err != nil {
		return err
	}
	var profile UserWorkspaceProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return err
	}
	if info, err := os.Stat(profile.Path); err != nil {
		return err
	} else if !info.IsDir() {
		return errors.New(fmt.Sprint("Workspace path is not a directory: ", profile.Path))
	}
	this.dataPath = profile.Path
	// Ensure
	if err := this.ensure(); err != nil {
		return err
	}
	// Done
	return nil
}

func (this *UserWorkspaceFileSystem) newFileSystem() error {
	// Create a new temp dir
	dirName, err := ioutil.TempDir(os.TempDir(), UserWorksapceDataDirPrefix)
	if err != nil {
		return err
	}
	this.dataPath = dirName
	// Create & save the profile
	profile := UserWorkspaceProfile{Path: dirName}
	data, err := json.Marshal(&profile)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(this.userPath, UserWorkspaceProfileName), data, 0666); err != nil {
		return err
	}
	// Ensure
	if err := this.ensure(); err != nil {
		return err
	}
	// Done
	return nil
}

func (this *UserWorkspaceFileSystem) ensure() error {
	// Ensure the workspace directory
	if err := os.MkdirAll(path.Join(this.dataPath, UserWorkspaceSourceCodeDir), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(this.dataPath, UserWorkspaceGenerateDir), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(this.dataPath, UserWorkspaceOutputDir), os.ModePerm); err != nil {
		return err
	}
	// Ensure the workspace symbol (from user path to data path)
	symlinkPath := path.Join(this.userPath, UserWorkspaceSymbolName)
	if info, err := os.Lstat(symlinkPath); err != nil {
		if os.IsNotExist(err) {
			// Create new symbol link
			return os.Symlink(this.dataPath, symlinkPath)
		} else {
			return err
		}
	} else if info.Mode()&os.ModeSymlink == 0 {
		// Not a symbol link
		return errors.New(fmt.Sprintf("Directory [%s] exists and cannot be managed by openlight cli", symlinkPath))
	} else {
		// Check the link
		oldLinkTarget, err := os.Readlink(symlinkPath)
		if err != nil {
			return err
		}
		if oldLinkTarget != this.dataPath {
			// Mismatch the link, remote the symbol link and create a new one
			if err := os.Remove(symlinkPath); err != nil {
				return err
			}
			return os.Symlink(this.dataPath, symlinkPath)
		}
	}
	// Done
	return nil
}

func (this *UserWorkspaceFileSystem) Clean() error {
	// Clean this workspace filesystem
	// Remove the workspace directory
	os.RemoveAll(this.dataPath)
	// Remove the profile
	os.Remove(path.Join(this.userPath, UserWorkspaceProfileName))
	// Remove the symbol
	os.Remove(path.Join(this.userPath, UserWorkspaceSymbolName))
	// Done
	return nil
}

func (this *UserWorkspaceFileSystem) GetSourcePath(dest string, ensure bool) (string, error) {
	p := path.Join(this.dataPath, UserWorkspaceSourceCodeDir, dest)
	if ensure {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", err
		}
	}
	// Done
	return p, nil
}

func (this *UserWorkspaceFileSystem) GetGeneratePath(dest string, ensure bool) (string, error) {
	p := path.Join(this.dataPath, UserWorkspaceGenerateDir, dest)
	if ensure {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", err
		}
	}
	// Done
	return p, nil
}

func (this *UserWorkspaceFileSystem) GetOutputPath(ensure bool) (string, error) {
	p := path.Join(this.dataPath, UserWorkspaceOutputDir)
	if ensure {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", err
		}
	}
	// Done
	return p, nil
}
