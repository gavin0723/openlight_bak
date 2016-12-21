// Author: lipixun
// Created Time : 四 12/15 13:32:29 2016
//
// File Name: environment.go
// Description:
//	The builder environment
//
//  GeneralEnvironment
//      /targetDir
//          /....
//      The target dir is the target key which applied the same name replace method as environment variables
//

package builder

import (
	"errors"
	"fmt"
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Environment interface {
	Path() string                             // The root path of the environment
	GetTargets() []*spec.Target               // Get all targets (The target key) in this environment
	GetTargetPath(target *spec.Target) string // Get root path of the target in this environment
	GetEnvironVars() map[string]string        // Get the environment variables
}

type GeneralEnvironment struct {
	path    string // The root paths
	targets map[string]GeneralTargetEnvironment
}

func NewGeneralEnvironment(path string) (*GeneralEnvironment, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}
	return &GeneralEnvironment{
		path:    path,
		targets: make(map[string]GeneralTargetEnvironment),
	}, nil
}

type GeneralTargetEnvironment struct {
	target *spec.Target
	path   string
}

func (this *GeneralEnvironment) Path() string {
	return this.path
}

func (this *GeneralEnvironment) GetTargets() []*spec.Target {
	var targets []*spec.Target
	for _, environ := range this.targets {
		targets = append(targets, environ.target)
	}
	return targets
}

func (this *GeneralEnvironment) GetTargetPath(target *spec.Target) string {
	environ, ok := this.targets[target.Key()]
	if ok {
		return environ.path
	}
	return ""
}

func (this *GeneralEnvironment) EnsureTargetPath(target *spec.Target) (string, error) {
	environ, ok := this.targets[target.Key()]
	if !ok {
		path := filepath.Join(this.path, GetTargetRegularKey(target))
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return "", err
		}
		environ = GeneralTargetEnvironment{target: target, path: path}
		this.targets[target.Key()] = environ
	}
	return environ.path, nil
}

func (this *GeneralEnvironment) GetEnvironVars() map[string]string {
	vars := make(map[string]string)
	for _, environ := range this.targets {
		vars[GetBuildTargetEnvironVarKey(environ.target)] = environ.path
	}
	return vars
}

func GetBuildTargetEnvironVarKey(target *spec.Target) string {
	return fmt.Sprintf("BUILD_TARGET_%s_PATH", GetTargetRegularKey(target))
}

// A general link methods
func GeneralLink(target *spec.Target, link *spec.SourceCodeLink, dest string) error {
	if link.Path == "" {
		return errors.New("Require link path")
	}
	if strings.HasSuffix(link.Path, "/") {
		return errors.New("The path of link must not end with /")
	}
	var linkTargetName string
	if link.LinkedName == "" {
		filename := filepath.Base(link.Path)
		linkTargetName = filepath.Join(dest, filename)
	} else {
		linkTargetName = filepath.Join(dest, link.LinkedName)
	}
	// Check if the target file already existed
	if _, err := os.Stat(linkTargetName); err == nil {
		return errors.New(fmt.Sprintf("Target [%s] already existed for target [%s] source [%s]", linkTargetName, target.Key(), link.Path))
	} else if !os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Failed to check link target for target [%s] source [%s] dest [%s], error: %s", target.Key(), link.Path, linkTargetName))
	}
	// Link it
	return os.Symlink(filepath.Join(target.Path(), link.Path), linkTargetName)
}

// Get the standard build metadata environment variables
func GetBuildMetadataEnvironVars(outputPath, branch, commit, tag string, t time.Time) []string {
	return []string{
		fmt.Sprintf("CI_OUTPUT=%s", outputPath),
		fmt.Sprintf("CI_BRANCH=%s", branch),
		fmt.Sprintf("CI_COMMIT=%s", commit),
		fmt.Sprintf("CI_TAG=%s", tag),
		fmt.Sprintf("CI_TIME=%s", t.Format(time.RFC3339)),
	}
}
