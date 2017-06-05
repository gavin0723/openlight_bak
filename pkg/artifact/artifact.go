// Author: lipixun
// Created Time : 日 12/18 16:33:06 2016
//
// File Name: artifact.go
// Description:
//

package artifact

// Artifact defines the artifact
type Artifact interface {
	// GetType returns the artifact type
	GetType() string
	// GetPath returns the (original) path of this artifact
	GetPath() string
	// String returns the string
	String() string
}
