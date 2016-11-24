// Author: lipixun
// Created Time : 三 11/23 18:47:12 2016
//
// File Name: spec.go
// Description:
//	The runner spec
package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type RunnerConfig struct {
	Apps map[string]RunnerApplicationConfig `yaml:"apps"`
}

type RunnerApplicationConfig struct {
	Name      string   `yaml:"name"`
	Command   string   `yaml:"command"`
	Workdir   string   `yaml:"workdir"`
	Args      []string `yaml:"args"`
	Singleton bool     `yaml:"singleton"`
}

func LoadRunnerConfigFromFile(p string) (*RunnerConfig, error) {
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var config RunnerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
