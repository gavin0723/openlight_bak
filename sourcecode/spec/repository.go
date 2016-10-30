// Author: lipixun
// Created Time : 四 10/20 10:26:09 2016
//
// File Name: config.go
// Description:
//	The config used by cli
package spec

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Repository struct {
	Uri     string `yaml:"uri"`
	Options struct {
		Default struct {
			Target string `yaml:"target"`
		} `yaml:"default"`
	} `yaml:"options"`
	Targets map[string]*Target `yaml:"targets"`
}

func LoadRepositorySpecFromFile(filename string) (*Repository, error) {
	// Load repository spec from file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var spec Repository
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	} else {
		return &spec, nil
	}
}
