// Author: lipixun
// File Name: local.go
// Description:

package repository

import (
	"errors"
	"fmt"
	"os"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"path/filepath"

	"github.com/ops-openlight/openlight/pkg/rule"
	"github.com/ops-openlight/openlight/pkg/rule/engine"
)

const (
	// DefaultSpecFilename defaults the default filename of spec file
	DefaultSpecFilename = "opspec"
)

// LocalRepository represents a local repository which has a spec file defined at root path
type LocalRepository struct {
	path string

	//
	//	Repository can determine which engine to use. This feature makes rule file forward-compitable available
	//
	engine *engine.Engine

	// The rule files
	ruleFiles map[string]*pbSpec.RuleFiles

	// The packages
	packages map[string]*Package
}

// NewLocalRepository creates a new LocalRepository
func NewLocalRepository(path string) (*LocalRepository, error) {
	return &LocalRepository{
		path:      path,
		engine:    rule.NewEngine(),
		ruleFiles: make(map[string]*pbSpec.RuleFiles),
		packages:  make(map[string]*Package),
	}, nil
}

// Path returns the local root path
func (repo *LocalRepository) Path() string {
	return repo.path
}

// Engine returns the engine used to parse all spec files in this repository
func (repo *LocalRepository) Engine() *engine.Engine {
	return repo.engine
}

// RootRuleFiles returns the root rule files
func (repo *LocalRepository) RootRuleFiles() (*pbSpec.RuleFiles, error) {
	return repo.GetRuleFiles("")
}

// GetRuleFiles returns the rule rules for a given path in the repository
func (repo *LocalRepository) GetRuleFiles(path string) (*pbSpec.RuleFiles, error) {
	if ruleFiles := repo.ruleFiles[path]; ruleFiles != nil {
		return ruleFiles, nil
	}
	ruleFiles, err := repo.loadRuleFiles(path)
	if err != nil {
		return nil, err
	}
	repo.ruleFiles[path] = ruleFiles
	return ruleFiles, nil
}

// RootPackage returns the root package
func (repo *LocalRepository) RootPackage() (*Package, error) {
	return repo.GetPackage("")
}

// GetPackage returns the package for a given path in the repository
func (repo *LocalRepository) GetPackage(path string) (*Package, error) {
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("Invalid package path [%v]. Path must not be a absolute path", path)
	}
	if pkg := repo.packages[path]; pkg != nil {
		return pkg, nil
	}
	pkg, err := repo.loadPackage(path)
	if err != nil {
		return nil, err
	}
	repo.packages[path] = pkg
	return pkg, nil
}

// GetTarget returns the target for a given path in the repository
func (repo *LocalRepository) GetTarget(path, target string) (*Target, error) {
	pkg, err := repo.GetPackage(path)
	if err != nil {
		return nil, err
	}
	return pkg.GetTarget(target), nil
}

// InitOutputDir initializes an output dir of a specific name in the given path of this repository
func (repo *LocalRepository) InitOutputDir(path, name string) (string, error) {
	dirname := filepath.Join(repo.path, path, "op-out", name)
	if info, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			// Create a new output dir
			if err := os.MkdirAll(dirname, os.ModePerm); err != nil {
				return "", fmt.Errorf("Failed to create directory [%v]: %v", dirname, err)
			}
		} else {
			return "", fmt.Errorf("Failed to check path [%v]: %v", dirname, err)
		}
	} else if !info.IsDir() {
		return "", fmt.Errorf("Target output path [%v] is not a directory", dirname)
	}
	// Done
	return dirname, nil
}

func (repo *LocalRepository) loadRuleFiles(path string) (*pbSpec.RuleFiles, error) {
	filename := filepath.Join(repo.path, path, DefaultSpecFilename)
	if info, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("Failed to check file [%v]: %v", filename, err)
	} else if info.IsDir() {
		// A directory
		filename = filepath.Join(filename, "main")
		if info, err := os.Stat(filename); err != nil {
			return nil, fmt.Errorf("Failed to check file [%v] in directory: %v", filename, err)
		} else if info.IsDir() {
			return nil, fmt.Errorf("Rule directory in directory is not supported: %v", filename)
		}
	}
	// Load the spec file
	ctx, err := repo.engine.ParseFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load rule file [%v]: %v", filename, err)
	}
	// Get rule and return
	ruleFiles, err := ctx.GetRule()
	if err != nil {
		return nil, fmt.Errorf("Failed to get rule of file [%v]: %v", filename, err)
	}
	return ruleFiles, nil
}

func (repo *LocalRepository) loadPackage(path string) (*Package, error) {
	ruleFiles, err := repo.GetRuleFiles(path)
	if err != nil {
		return nil, err
	}
	pkgSpec := ruleFiles.GetBuild().GetPackage()
	if pkgSpec == nil {
		return nil, errors.New("No package defined")
	}
	return newPackage(filepath.Join(repo.path, path), pkgSpec, repo), nil
}
