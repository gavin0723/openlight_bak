// Author: lipixun
// File Name: file.go
// Description:

package build

import (
	"errors"
	"fmt"

	"github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ common.Object = (*FileSource)(nil)

// Exposed lua infos
const (
	fileLUAName         = "File"
	fileLUATypeName     = "build-file-file"
	artifactLUAName     = "Artifact"
	artifactLUATypeName = "build-file-artifact"
)

// registerFileType registers package type in lua
func registerFileType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(fileLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewFileFromLUA))
	// Add type to module
	L.SetField(mod, fileLUAName, mt)
}

// registerArtifactType registers package type in lua
func registerArtifactType(L *lua.LState, mod *lua.LTable) {
	// Create meta table
	mt := L.NewTypeMetatable(artifactLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewArtifactFromLUA))
	// Add type to module
	L.SetField(mod, artifactLUAName, mt)
}

// FileSource represents a file source
type FileSource pbSpec.FileSource

// NewFileFromLUA creates a new File from LUA
func NewFileFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	reference, err := params.GetString("reference")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [reference]: %v", err)
	}
	filename, err := params.GetString("filename")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [filename]: %v", err)
	}
	// Create file source
	f := &FileSource{
		Source: &pbSpec.FileSource_File{
			File: &pbSpec.File{
				Reference: reference,
				Filename:  filename,
			},
		},
	}
	// Done
	return f.GetLUAUserData(L), nil
}

// NewArtifactFromLUA creates a new Artifact from LUA
func NewArtifactFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	reference, err := params.GetString("reference")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [reference]: %v", err)
	}
	path, err := params.GetString("path")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [path]: %v", err)
	}
	target, err := params.GetString("target")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [target]: %v", err)
	}
	if target == "" {
		return nil, errors.New("Require target")
	}
	filename, err := params.GetString("filename")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [filename]: %v", err)
	}
	// Create file source
	f := &FileSource{
		Source: &pbSpec.FileSource_Artifact{
			Artifact: &pbSpec.Artifact{
				Reference: reference,
				Path:      path,
				Target:    target,
				Filename:  filename,
			},
		},
	}
	// Done
	return f.GetLUAUserData(L), nil
}

// GetLUAUserData returns the lua user data
func (f *FileSource) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = f
	// Done
	return ud
}
