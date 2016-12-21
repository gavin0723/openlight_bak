// Author: lipixun
// Created Time : 二 12/20 15:01:29 2016
//
// File Name: spec.go
// Description:
//	The runner spec
package runner

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const (
	SpecFileName = ".op.runner.yaml"
)

type RunnerSpec struct {
	Apps map[string]*RunnerAppSpec `yaml:"apps"` // Key is app id
}

type RunnerAppSpec struct {
	Name      string   `yaml:"name"`      // The global unique name
	Command   string   `yaml:"command"`   // The command to run
	Workdir   string   `yaml:"workdir"`   // The workdir, will use the directory of the file as the "current directory"
	Args      []string `yaml:"args"`      // The command args
	Singleton bool     `yaml:"singleton"` // A singleton app or not
}

func LoadRunnerSpecFromFile(p string) (*RunnerSpec, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var spec RunnerSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return &spec, nil
}
