// Author: lipixun
// Created Time : 六 12/10 13:04:17 2016
//
// File Name: options.go
// Description:
//  The workspace options
package workspace

const (
	DefaultGlobalDirPath = "/var/run/openlight"
	DefaultUserDirPath   = "~/.openlight"

	DefaultDockerServiceUri = "unix:///var/run/docker.sock"
)

type WorkspaceOptions struct {
	Dir          WorkDirOptions      // The directory of workspace options
	Verbose      bool                // Show the verbose
	EnableColor  bool                // Enable the color of the log
	ThirdService ThirdServiceOptions // The third party options
}

func NewWorkspaceOptions() *WorkspaceOptions {
	options := new(WorkspaceOptions)
	options.Dir.GlobalPath = DefaultGlobalDirPath
	options.Dir.UserPath = DefaultUserDirPath
	options.ThirdService.Docker.Uri = DefaultDockerServiceUri
	// Done
	return options
}

type WorkDirOptions struct {
	GlobalPath               string
	UserPath                 string
	ProjectPath              string
	CurrentPathAsProjectPath bool
}

type ThirdServiceOptions struct {
	Docker DockerServiceOptions
}

type DockerServiceOptions struct {
	Uri string
}
