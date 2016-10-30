// Author: lipixun
// Created Time : 五 10/21 13:35:12 2016
//
// File Name: uri.go
// Description:
// 	The repository uri format:
// 		<uri>(///(@<branch>)|(=<commit>))?
//	The target uri, format:
// 		(<uri>(///(@<branch>)|(=<commit>))?::)?<target>
package uri

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	UriRegex    = "(?P<uri>(([^/|:]*((/|//|:)[^/:]+)?)+))"
	BranchRegex = "(@(?P<branch>[a-zA-Z0-9-_]+))"
	CommitRegex = "(=(?P<commit>[a-zA-Z0-9-_]+))"
	TargetRegex = "(?P<target>[a-zA-Z0-9_-]+)"
)

var (
	RepositoryUriRegex      = regexp.MustCompile(fmt.Sprintf("^%s(///(%s|%s))?$", UriRegex, BranchRegex, CommitRegex))
	RepositoryUriRegexNames = RepositoryUriRegex.SubexpNames()
	TargetUriRegex          = regexp.MustCompile(fmt.Sprintf("^(%s|(%s(///(%s|%s))?::%s)|(%s(///(%s|%s))?))$", TargetRegex, UriRegex, BranchRegex, CommitRegex, TargetRegex, UriRegex, BranchRegex, CommitRegex))
	TargetUriRegexNames     = TargetUriRegex.SubexpNames()
)

type RepositoryUri struct {
	Uri    string `json:"uri" yaml:"uri"`
	Branch string `json:"branch" yaml:"branch"`
	Commit string `json:"commit" yaml:"commit"`
}

func ParseRepositoryUri(u string) *RepositoryUri {
	// Parse repository uri
	matches := RepositoryUriRegex.FindStringSubmatch(u)
	if matches == nil {
		return nil
	}
	var uri RepositoryUri
	for i, match := range matches {
		if match != "" {
			switch RepositoryUriRegexNames[i] {
			case "uri":
				uri.Uri = match
			case "branch":
				uri.Branch = match
			case "commit":
				uri.Commit = match
			}
		}
	}
	// Done
	return &uri
}

func (this *RepositoryUri) Validate() error {
	if this.Uri == "" {
		return errors.New("Require uri")
	}
	if this.Branch != "" && this.Commit != "" {
		return errors.New("Cannot specify both branch and commit")
	}
	// Done
	return nil
}

func (this *RepositoryUri) Equal(uri *RepositoryUri) bool {
	return this.Uri == uri.Uri && this.Branch == uri.Branch && this.Commit == uri.Commit
}

func (this *RepositoryUri) String() string {
	if this.Branch == "" && this.Commit == "" {
		return this.Uri
	} else if this.Branch != "" {
		return fmt.Sprintf("%s///@%s", this.Uri, this.Branch)
	} else if this.Commit != "" {
		return fmt.Sprintf("%s///=%s", this.Uri, this.Commit)
	} else {
		return fmt.Sprintf("<!RepositoryUri Invalid>(Uri=%s,Branch=%s,Commit=%s)", this.Uri, this.Branch, this.Commit)
	}
}

type RepositoryReference struct {
	Local  string         `json:"local" yaml:"local"`
	Remote *RepositoryUri `json:"remote" yaml:"remote"`
}

func (this *RepositoryReference) Equal(reference *RepositoryReference) bool {
	if this.Local != reference.Local {
		return false
	}
	if this.Remote == nil && reference.Remote == nil {
		return true
	} else if this.Remote == nil && reference.Remote != nil || this.Remote != nil && reference.Remote == nil {
		return false
	} else {
		return this.Remote.Equal(reference.Remote)
	}
}

func (this *RepositoryReference) String() string {
	// Convert to string
	if this.Local == "" && this.Remote == nil {
		return ""
	} else if this.Local != "" && this.Remote == nil {
		return this.Local
	} else if this.Local == "" && this.Remote != nil {
		return this.Remote.String()
	} else {
		return fmt.Sprintf("<!RepositoryReference Invalid>(Local=%s,Remote=%s)", this.Local, this.Remote.String())
	}
}

type TargetUri struct {
	Repository RepositoryReference `json:"repository" yaml:"repository"`
	Target     string              `json:"target" yaml:"target"`
}

func ParseTargetUri(u string) *TargetUri {
	// Parse repository uri
	// NOTE: We don't support specified a local repository reference in target uri string
	matches := TargetUriRegex.FindStringSubmatch(u)
	if matches == nil {
		return nil
	}
	var uri TargetUri
	for i, match := range matches {
		if match != "" {
			switch TargetUriRegexNames[i] {
			case "uri":
				if uri.Repository.Remote == nil {
					uri.Repository.Remote = new(RepositoryUri)
				}
				uri.Repository.Remote.Uri = match
			case "branch":
				if uri.Repository.Remote == nil {
					uri.Repository.Remote = new(RepositoryUri)
				}
				uri.Repository.Remote.Branch = match
			case "commit":
				if uri.Repository.Remote == nil {
					uri.Repository.Remote = new(RepositoryUri)
				}
				uri.Repository.Remote.Commit = match
			case "target":
				uri.Target = match
			}
		}
	}
	// Done
	return &uri
}

func (this *TargetUri) Equal(uri *TargetUri) bool {
	return this.Repository.Equal(&uri.Repository) && this.Target == uri.Target
}

func (this *TargetUri) String() string {
	v := this.Repository.String()
	if v == "" {
		return this.Target
	} else if this.Target == "" {
		return v
	} else {
		return fmt.Sprintf("%s::%s", v, this.Target)
	}
}
