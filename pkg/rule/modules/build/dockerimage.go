// Author: lipixun
// File Name: dockerimage.go
// Description:
//	This file defines docker image target

package build

import (
	"errors"
	"fmt"

	lua "github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/rule/common"
)

// Ensure the interface is implemented
var _ Target = (*DockerImageTarget)(nil)

const (
	dockerImageTargetLUAName     = "DockerImageTarget"
	dockerImageTargetLUATypeName = "build-target-dockerimage"
)

// registerDockerImageTargetType registers DockerImageTarget type in lua
func registerDockerImageTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(dockerImageTargetLUATypeName)
	L.SetField(mt, "new", common.NewLUANewObjectFunction(L, NewDockerImageTargetFromLUA))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"description": luaFuncTargetDescription,
		"dependent":   luaFuncTargetDependent,
		"from":        luaFuncDockerImageTargetFrom,
		"label":       luaFuncDockerImageTargetLabel,
		"add":         luaFuncDockerImageTargetAdd,
		"copy":        luaFuncDockerImageTargetCopy,
		"run":         luaFuncDockerImageTargetRun,
		"entrypoint":  luaFuncDockerImageTargetEntrypoint,
		"expose":      luaFuncDockerImageTargetExpose,
		"volume":      luaFuncDockerImageTargetVolume,
		"user":        luaFuncDockerImageTargetUser,
		"workdir":     luaFuncDockerImageTargetWorkdir,
		"env":         luaFuncDockerImageTargetEnv,
	}))
	L.SetField(mod, dockerImageTargetLUAName, mt)
}

// DockerImageTarget implements the docker image target
type DockerImageTarget struct {
	_Target
}

// NewDockerImageTargetFromLUA creates a new DockerImageTarget from LUA
func NewDockerImageTargetFromLUA(L *lua.LState, params common.Parameters) (lua.LValue, error) {
	repository, err := params.GetString("repository")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [repository]: %v", err)
	}
	if repository == "" {
		return nil, errors.New("Require repository")
	}
	image, err := params.GetString("image")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [image]: %v", err)
	}
	if image == "" {
		return nil, errors.New("Require image")
	}
	push, err := params.GetBool("push")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [push]: %v", err)
	}
	setLatestTag, err := params.GetBool("setLatestTag")
	if err != nil {
		return nil, fmt.Errorf("Invalid parameter [setLatestTag]: %v", err)
	}
	// Create a new target
	target := &DockerImageTarget{
		_Target: _Target{
			Target: &pbSpec.Target_DockerImage{
				DockerImage: &pbSpec.DockerImageTarget{
					Repository:   repository,
					Image:        image,
					Push:         push,
					SetLatestTag: setLatestTag,
				},
			},
		},
	}
	// Done
	return target.GetLUAUserData(L), nil
}

// GetDockerImageTarget returns the pbSpec.DockerImageTarget
func (t *DockerImageTarget) GetDockerImageTarget() *pbSpec.DockerImageTarget {
	return (t.Target.(*pbSpec.Target_DockerImage)).DockerImage
}

// GetLUAUserData returns the lua user data
func (t *DockerImageTarget) GetLUAUserData(L *lua.LState) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(dockerImageTargetLUATypeName))
	// Done
	return ud
}

//////////////////////////////////////// LUA functions ////////////////////////////////////////

// luaDockerImageTargetSelf get lua target self for docker image
func luaDockerImageTargetSelf(L *lua.LState) *DockerImageTarget {
	ud := L.CheckUserData(1)
	if target, ok := ud.Value.(*DockerImageTarget); ok {
		return target
	}
	L.ArgError(1, "Target expected")
	return nil
}

// luaFuncDockerImageTargetFrom defines DockerImage.from function in lua
func luaFuncDockerImageTargetFrom(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require name")
		return 0
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many names")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) != 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get name
	name := L.CheckString(2)
	if name == "" {
		L.ArgError(2, "Require name")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_From_{
			From: &pbSpec.DockerImageBuildCommand_From{
				Name: name,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetLabel defines DockerImage.label function in lua
func luaFuncDockerImageTargetLabel(L *lua.LState) int {
	if L.GetTop() < 3 {
		L.ArgError(3, "Require key and value")
		return 0
	} else if L.GetTop() > 3 {
		L.ArgError(4, "Too many values")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get key
	key := L.CheckString(2)
	if key == "" {
		L.ArgError(2, "Require key")
		return 0
	}
	value := L.CheckString(3)
	if value == "" {
		L.ArgError(3, "Require value")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Label_{
			Label: &pbSpec.DockerImageBuildCommand_Label{
				Key:   key,
				Value: value,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetAdd defines DockerImage.add function in lua
func luaFuncDockerImageTargetAdd(L *lua.LState) int {
	if L.GetTop() < 3 {
		L.ArgError(3, "Require file and target")
		return 0
	} else if L.GetTop() > 3 {
		L.ArgError(4, "Too many arguments")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get file
	ud := L.CheckUserData(2)
	file, ok := ud.Value.(*FileSource)
	if !ok {
		L.ArgError(2, "Not a file")
	}
	path := L.CheckString(3)
	if path == "" {
		L.ArgError(3, "Require path")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Add_{
			Add: &pbSpec.DockerImageBuildCommand_Add{
				File: (*pbSpec.FileSource)(file),
				Path: path,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetCopy defines DockerImage.copy function in lua
func luaFuncDockerImageTargetCopy(L *lua.LState) int {
	if L.GetTop() < 3 {
		L.ArgError(3, "Require file and target")
		return 0
	} else if L.GetTop() > 3 {
		L.ArgError(4, "Too many arguments")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get file
	ud := L.CheckUserData(2)
	file, ok := ud.Value.(*FileSource)
	if !ok {
		L.ArgError(2, "Not a file")
	}
	path := L.CheckString(3)
	if path == "" {
		L.ArgError(3, "Require path")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Copy_{
			Copy: &pbSpec.DockerImageBuildCommand_Copy{
				File: (*pbSpec.FileSource)(file),
				Path: path,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetRun defines DockerImage.run function in lua
func luaFuncDockerImageTargetRun(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require command")
		return 0
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many commands")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get name
	command := L.CheckString(2)
	if command == "" {
		L.ArgError(2, "Require command")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Run_{
			Run: &pbSpec.DockerImageBuildCommand_Run{
				Command: command,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetEntrypoint defines DockerImage.entrypoint function in lua
func luaFuncDockerImageTargetEntrypoint(L *lua.LState) int {
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get arguments
	var args []string
	for i := 2; i <= L.GetTop(); i++ {
		args = append(args, L.CheckString(i))
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Entrypoint_{
			Entrypoint: &pbSpec.DockerImageBuildCommand_Entrypoint{
				Args: args,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetExpose defines DockerImage.expose function in lua
func luaFuncDockerImageTargetExpose(L *lua.LState) int {
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get ports
	var ports []int32
	for i := 2; i <= L.GetTop(); i++ {
		ports = append(ports, int32(L.CheckNumber(i)))
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Expose_{
			Expose: &pbSpec.DockerImageBuildCommand_Expose{
				Ports: ports,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetVolume defines DockerImage.volume function in lua
func luaFuncDockerImageTargetVolume(L *lua.LState) int {
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get paths
	var paths []string
	for i := 2; i <= L.GetTop(); i++ {
		paths = append(paths, L.CheckString(i))
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Volume_{
			Volume: &pbSpec.DockerImageBuildCommand_Volume{
				Paths: paths,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetUser defines DockerImage.user function in lua
func luaFuncDockerImageTargetUser(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require user")
		return 0
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many users")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get name
	user := L.CheckString(2)
	if user == "" {
		L.ArgError(2, "Require user")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_User_{
			User: &pbSpec.DockerImageBuildCommand_User{
				Name: user,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetWorkdir defines DockerImage.workdir function in lua
func luaFuncDockerImageTargetWorkdir(L *lua.LState) int {
	if L.GetTop() < 2 {
		L.ArgError(2, "Require path")
		return 0
	} else if L.GetTop() > 2 {
		L.ArgError(3, "Too many paths")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get name
	path := L.CheckString(2)
	if path == "" {
		L.ArgError(2, "Require path")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Workdir_{
			Workdir: &pbSpec.DockerImageBuildCommand_Workdir{
				Path: path,
			},
		},
	})
	// Done
	return 0
}

// luaFuncDockerImageTargetEnv defines DockerImage.env function in lua
func luaFuncDockerImageTargetEnv(L *lua.LState) int {
	if L.GetTop() < 3 {
		L.ArgError(3, "Require key and value")
		return 0
	} else if L.GetTop() > 3 {
		L.ArgError(4, "Too many values")
		return 0
	}
	// Get dockerimage
	dockerImage := luaDockerImageTargetSelf(L).GetDockerImageTarget()
	if len(dockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get key
	key := L.CheckString(2)
	if key == "" {
		L.ArgError(2, "Require key")
		return 0
	}
	value := L.CheckString(3)
	if value == "" {
		L.ArgError(3, "Require value")
		return 0
	}
	// Add command
	dockerImage.Commands = append(dockerImage.Commands, &pbSpec.DockerImageBuildCommand{
		Command: &pbSpec.DockerImageBuildCommand_Env_{
			Env: &pbSpec.DockerImageBuildCommand_Env{
				Key:   key,
				Value: value,
			},
		},
	})
	// Done
	return 0
}
