// Author: lipixun
// Created Time : 四 12/15 19:28:36 2016
//
// File Name: options.go
// Description:
//	The builder options
package builder

import (
	"time"
)

// The build option
type BuilderOptions struct {
	Tag        string            // The build tag
	Time       time.Time         // The build time
	OutputPath string            // The find build artifacts will be copied to this path
	ThirdParty ThirdPartyOptions // The third party options
}

// Create a new BuildOption
func NewBuilderOptions(tag string, outputPath string) BuilderOptions {
	return BuilderOptions{
		Tag:        tag,
		Time:       time.Now(),
		OutputPath: outputPath,
		ThirdParty: ThirdPartyOptions{
			Docker: DockerOptions{
				Push: true,
			},
		},
	}
}

type ThirdPartyOptions struct {
	Docker DockerOptions
}

type DockerOptions struct {
	Push bool
}
