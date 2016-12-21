// Author: lipixun
// Created Time : 日 12/11 18:40:31 2016
//
// File Name: util.go
// Description:
//	The utility
package repoloader

import (
	"github.com/ops-openlight/openlight/pkg/sourcecode/spec"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func LoadRepositorySpecFromFile(filename string) (*spec.RepositorySpec, error) {
	// Load repository spec from file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var repoSpec spec.RepositorySpec
	if err := yaml.Unmarshal(data, &repoSpec); err != nil {
		return nil, err
	} else {
		return &repoSpec, nil
	}
}
