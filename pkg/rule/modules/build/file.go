// Author: lipixun
// File Name: file.go
// Description:

package build

import (
	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Exposed lua infos
const (
	LUANameFile     = "File"
	LUANameArtifact = "Artifact"

	LUATypeFile     = "Build-File"
	LUATypeArtifact = "Build-Artifact"
)

// FileLUAFuncs defines all lua functions for python local finder
var FileLUAFuncs = map[string]lua.LGFunction{}

// ArtifactLUAFuncs defines all lua functions for python local finder
var ArtifactLUAFuncs = map[string]lua.LGFunction{}

// FileSource represents a file source
type FileSource interface {
	LUA.Object
	// GetProto returns the protobuf object
	GetProto() (*pbSpec.FileSource, error)
}

// RegisterFileType registers package type in lua
func RegisterFileType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(LUATypeFile)
	L.SetField(mt, "new", L.NewFunction(LUAFileNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), FileLUAFuncs))
	// Add type to module
	L.SetField(mod, LUANameFile, mt)
}

// RegisterArtifactType registers package type in lua
func RegisterArtifactType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(LUATypeArtifact)
	L.SetField(mt, "new", L.NewFunction(LUAArtifactNew))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), ArtifactLUAFuncs))
	// Add type to module
	L.SetField(mod, LUANameArtifact, mt)
}

// File represents a local file
type File struct {
	LUA.Object
	Package  string
	Filename string
}

// NewFile creates a new File
func NewFile(pkg string, filename string, options *lua.LTable) *File {
	var file = File{
		Package:  pkg,
		Filename: filename,
	}
	file.Object = LUA.NewObject(LUATypeFile, options, FileSource(&file))
	// Done
	return &file
}

// GetProto returns the protobuf object
func (f *File) GetProto() (*pbSpec.FileSource, error) {
	var pbFileSource pbSpec.FileSource
	pbFileSource.Source = &pbSpec.FileSource_File{
		File: &pbSpec.File{
			Package:  f.Package,
			Filename: f.Filename,
		},
	}
	return &pbFileSource, nil
}

// Artifact represents an artifact
type Artifact struct {
	LUA.Object
	Package string
	Target  string
}

// NewArtifact creates a new Artifact
func NewArtifact(pkg string, target string, options *lua.LTable) *Artifact {
	var artifact = Artifact{
		Package: pkg,
		Target:  target,
	}
	artifact.Object = LUA.NewObject(LUATypeFile, options, FileSource(&artifact))
	// Done
	return &artifact
}

// GetProto returns the protobuf object
func (f *Artifact) GetProto() (*pbSpec.FileSource, error) {
	var pbFileSource pbSpec.FileSource
	pbFileSource.Source = &pbSpec.FileSource_Artifact{
		Artifact: &pbSpec.Artifact{
			Package: f.Package,
			Target:  f.Target,
		},
	}
	return &pbFileSource, nil
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// LUAFileNew defines package.new function in lua
func LUAFileNew(L *lua.LState) int {
	if L.GetTop() == 1 {
		// Filename
		filename := L.CheckString(1)
		if filename == "" {
			L.ArgError(1, "Require filename")
			return 0
		}
		file := NewFile("", filename, nil)
		// Return
		L.Push(file.GetLUAUserData(L))
		return 1
	} else if L.GetTop() == 2 {
		// Repository, filename
		repository := L.CheckString(1)
		if repository == "" {
			L.ArgError(1, "Require repository")
			return 0
		}
		filename := L.CheckString(2)
		if filename == "" {
			L.ArgError(2, "Require filename")
			return 0
		}
		file := NewFile(repository, filename, nil)
		// Return
		L.Push(file.GetLUAUserData(L))
		return 1
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

// LUAArtifactNew defines package.new function in lua
func LUAArtifactNew(L *lua.LState) int {
	if L.GetTop() == 1 {
		// Target
		target := L.CheckString(1)
		if target == "" {
			L.ArgError(1, "Require target")
			return 0
		}
		artifact := NewArtifact("", target, nil)
		// Return
		L.Push(artifact.GetLUAUserData(L))
		return 1
	} else if L.GetTop() == 2 {
		// Repository, target
		repository := L.CheckString(1)
		if repository == "" {
			L.ArgError(1, "Require repository")
			return 0
		}
		target := L.CheckString(2)
		if target == "" {
			L.ArgError(2, "Require target")
			return 0
		}
		artifact := NewArtifact(repository, target, nil)
		// Return
		L.Push(artifact.GetLUAUserData(L))
		return 1
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}
