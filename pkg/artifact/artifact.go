// Author: lipixun
// Created Time : 日 12/18 16:33:06 2016
//
// File Name: artifact.go
// Description:
//	The artifact
package artifact

type Artifact interface {
	GetName() string                 // Get the name
	GetType() string                 // The the type
	GetAttr(name string) interface{} // Get the attribute
	String() string                  // Get the string representation
}
