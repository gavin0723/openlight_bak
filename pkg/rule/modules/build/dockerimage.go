// Author: lipixun
// File Name: dockerimage.go
// Description:
//	This file defines docker image target

package build

import (
	lua "github.com/yuin/gopher-lua"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// DockerImageTargetLUAFuncs defines all lua functions for docker image target
var DockerImageTargetLUAFuncs = map[string]lua.LGFunction{
	"repository":  LUAFuncDockerImageTargetRepository,
	"dockerImage": LUAFuncDockerImageTargetDockerImage,
	"from":        LUAFuncDockerImageTargetFrom,
	"label":       LUAFuncDockerImageTargetLabel,
	"add":         LUAFuncDockerImageTargetAdd,
	"copy":        LUAFuncDockerImageTargetCopy,
	"run":         LUAFuncDockerImageTargetRun,
	"entrypoint":  LUAFuncDockerImageTargetEntrypoint,
	"expose":      LUAFuncDockerImageTargetExpose,
	"volume":      LUAFuncDockerImageTargetVolume,
	"user":        LUAFuncDockerImageTargetUser,
	"workdir":     LUAFuncDockerImageTargetWorkdir,
	"env":         LUAFuncDockerImageTargetEnv,
}

// DockerImageTarget represents the docker target of build package
type DockerImageTarget struct {
	Repository string
	ImageName  string
	Commands   []DockerImageCommand
}

// RegisterDockerImageTargetType registers DockerImageTarget type in lua
func RegisterDockerImageTargetType(L *lua.LState, mod *lua.LTable) {
	mt := L.NewTypeMetatable(LUATypeDockerImageTarget)
	var funcs = make(map[string]lua.LGFunction)
	for name, function := range TargetLUAFuncs {
		funcs[name] = function
	}
	for name, function := range DockerImageTargetLUAFuncs {
		funcs[name] = function
	}
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), funcs))
}

// NewDockerImageTarget creates a new DockerImageTarget
func NewDockerImageTarget(name string, repository string, imageName string, options *lua.LTable) *Target {
	target := NewTarget(LUATypeDockerImageTarget, name, options)
	target.DockerImage = &DockerImageTarget{
		Repository: repository,
		ImageName:  imageName,
	}
	// Done
	return target
}

// GetProto returns the protobuf object
func (target *DockerImageTarget) GetProto(options *lua.LTable) (*pbSpec.DockerImageTarget, error) {
	var pbTarget pbSpec.DockerImageTarget
	pbTarget.Repository = target.Repository
	pbTarget.ImageName = target.ImageName
	// Add commands
	for _, cmd := range target.Commands {
		pbCmd, err := cmd.GetProto()
		if err != nil {
			return nil, err
		}
		pbTarget.Commands = append(pbTarget.Commands, pbCmd)
	}
	// Done
	return &pbTarget, nil
}

// DockerImageCommand represents the docker build command
type DockerImageCommand interface {
	// GetProto returns the protobuf object
	GetProto() (*pbSpec.DockerImageBuildCommand, error)
}

// DockerImageCommandFrom define docker command: FROM
type DockerImageCommandFrom struct {
	Name string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandFrom) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_From_{
		From: &pbSpec.DockerImageBuildCommand_From{
			Name: cmd.Name,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandLabel define docker command: LABEL
type DockerImageCommandLabel struct {
	Key   string
	Value string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandLabel) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Label_{
		Label: &pbSpec.DockerImageBuildCommand_Label{
			Key:   cmd.Key,
			Value: cmd.Value,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandAdd define docker command: ADD
type DockerImageCommandAdd struct {
	File FileSource
	Path string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandAdd) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	var file *pbSpec.FileSource
	if cmd.File != nil {
		var err error
		file, err = cmd.File.GetProto()
		if err != nil {
			return nil, err
		}
	}
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Add_{
		Add: &pbSpec.DockerImageBuildCommand_Add{
			File: file,
			Path: cmd.Path,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandCopy define docker command: COPY
type DockerImageCommandCopy struct {
	File FileSource
	Path string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandCopy) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	var file *pbSpec.FileSource
	if cmd.File != nil {
		var err error
		file, err = cmd.File.GetProto()
		if err != nil {
			return nil, err
		}
	}
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Copy_{
		Copy: &pbSpec.DockerImageBuildCommand_Copy{
			File: file,
			Path: cmd.Path,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandRun define docker command: RUN
type DockerImageCommandRun struct {
	Command string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandRun) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Run_{
		Run: &pbSpec.DockerImageBuildCommand_Run{
			Command: cmd.Command,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandEntrypoint define docker command: ENTRYPOINT
type DockerImageCommandEntrypoint struct {
	Args []string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandEntrypoint) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Entrypoint_{
		Entrypoint: &pbSpec.DockerImageBuildCommand_Entrypoint{
			Args: cmd.Args,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandExpose define docker command: EXPOSE
type DockerImageCommandExpose struct {
	Ports []int
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandExpose) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	var ports []int32
	for _, port := range cmd.Ports {
		ports = append(ports, int32(port))
	}
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Expose_{
		Expose: &pbSpec.DockerImageBuildCommand_Expose{
			Ports: ports,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandVolume define docker command: VOLUME
type DockerImageCommandVolume struct {
	Paths []string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandVolume) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Volume_{
		Volume: &pbSpec.DockerImageBuildCommand_Volume{
			Paths: cmd.Paths,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandUser define docker command: USER
type DockerImageCommandUser struct {
	Name string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandUser) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_User_{
		User: &pbSpec.DockerImageBuildCommand_User{
			Name: cmd.Name,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandWorkdir define docker command: WORKDIR
type DockerImageCommandWorkdir struct {
	Path string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandWorkdir) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Workdir_{
		Workdir: &pbSpec.DockerImageBuildCommand_Workdir{
			Path: cmd.Path,
		},
	}
	return &pbCommand, nil
}

// DockerImageCommandEnv define docker command: ENV
type DockerImageCommandEnv struct {
	Key   string
	Value string
}

// GetProto returns the protobuf object
func (cmd *DockerImageCommandEnv) GetProto() (*pbSpec.DockerImageBuildCommand, error) {
	var pbCommand pbSpec.DockerImageBuildCommand
	pbCommand.Command = &pbSpec.DockerImageBuildCommand_Env_{
		Env: &pbSpec.DockerImageBuildCommand_Env{
			Key:   cmd.Key,
			Value: cmd.Value,
		},
	}
	return &pbCommand, nil
}

// LUADockerImageTargetSelf get lua target self for docker image
func LUADockerImageTargetSelf(L *lua.LState) *Target {
	target := LUATargetSelf(L)
	if target == nil {
		return nil
	}
	if target.DockerImage == nil {
		L.ArgError(0, "Not a docker image")
		return nil
	}
	return target
}

// LUAFuncDockerImageTargetRepository defines target.repository function in lua
func LUAFuncDockerImageTargetRepository(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if L.GetTop() == 1 {
		// Get
		L.Push(lua.LString(target.DockerImage.Repository))
		return 1
	} else if L.GetTop() == 2 {
		// Set
		repository := L.CheckString(2)
		if repository == "" {
			L.ArgError(2, "Require value")
			return 0
		}
		target.DockerImage.Repository = repository
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

// LUAFuncDockerImageTargetDockerImage defines target.dockerImage function in lua
func LUAFuncDockerImageTargetDockerImage(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if target.DockerImage == nil {
		L.ArgError(0, "Not a DockerImage target")
		return 0
	}
	if L.GetTop() == 1 {
		// Get
		L.Push(lua.LString(target.DockerImage.ImageName))
		return 1
	} else if L.GetTop() == 2 {
		// Set
		imageName := L.CheckString(2)
		if imageName == "" {
			L.ArgError(2, "Require value")
			return 0
		}
		target.DockerImage.ImageName = imageName
		return 0
	}
	// Invalid arguments
	L.ArgError(0, "Invalid arguments")
	return 0
}

// LUAFuncDockerImageTargetFrom defines DockerImage.from function in lua
func LUAFuncDockerImageTargetFrom(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) != 0 {
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
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandFrom{Name: name})
	// Done
	return 0
}

// LUAFuncDockerImageTargetLabel defines DockerImage.label function in lua
func LUAFuncDockerImageTargetLabel(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get key
	key := L.CheckString(2)
	if key == "" {
		L.ArgError(2, "Require key")
		return 0
	}
	// Get value
	value := L.CheckString(3)
	if value == "" {
		L.ArgError(3, "Require value")
		return 0
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandLabel{Key: key, Value: value})
	// Done
	return 0
}

// LUAFuncDockerImageTargetAdd defines DockerImage.add function in lua
func LUAFuncDockerImageTargetAdd(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get file source
	ud := L.CheckUserData(2)
	if ud == nil {
		return 0
	}
	fileSource, ok := ud.Value.(FileSource)
	if !ok {
		L.ArgError(2, "Not a file source")
		return 0
	}
	// Get path
	path := L.CheckString(3)
	if path == "" {
		L.ArgError(3, "Require path")
		return 0
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandAdd{File: fileSource, Path: path})
	// Done
	return 0
}

// LUAFuncDockerImageTargetCopy defines DockerImage.copy function in lua
func LUAFuncDockerImageTargetCopy(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get file source
	ud := L.CheckUserData(2)
	if ud == nil {
		return 0
	}
	fileSource, ok := ud.Value.(FileSource)
	if !ok {
		L.ArgError(2, "Not a file source")
		return 0
	}
	// Get path
	path := L.CheckString(3)
	if path == "" {
		L.ArgError(3, "Require path")
		return 0
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandCopy{File: fileSource, Path: path})
	// Done
	return 0
}

// LUAFuncDockerImageTargetRun defines DockerImage.run function in lua
func LUAFuncDockerImageTargetRun(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get command
	command := L.CheckString(2)
	if command == "" {
		L.ArgError(2, "Require command")
		return 0
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandRun{Command: command})
	// Done
	return 0
}

// LUAFuncDockerImageTargetEntrypoint defines DockerImage.entrypoint function in lua
func LUAFuncDockerImageTargetEntrypoint(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	if L.GetTop() < 2 {
		L.ArgError(2, "Require command and arguments")
		return 0
	}
	var args []string
	for i := 2; i <= L.GetTop(); i++ {
		args = append(args, L.CheckString(i))
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandEntrypoint{Args: args})
	// Done
	return 0
}

// LUAFuncDockerImageTargetExpose defines DockerImage.expose function in lua
func LUAFuncDockerImageTargetExpose(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	if L.GetTop() < 2 {
		L.ArgError(2, "Require ports")
		return 0
	}
	var ports []int
	for i := 2; i <= L.GetTop(); i++ {
		ports = append(ports, L.CheckInt(i))
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandExpose{Ports: ports})
	// Done
	return 0
}

// LUAFuncDockerImageTargetVolume defines DockerImage.volume function in lua
func LUAFuncDockerImageTargetVolume(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	if L.GetTop() < 2 {
		L.ArgError(2, "Require path")
		return 0
	}
	var paths []string
	for i := 2; i <= L.GetTop(); i++ {
		paths = append(paths, L.CheckString(i))
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandVolume{Paths: paths})
	// Done
	return 0
}

// LUAFuncDockerImageTargetUser defines DockerImage.user function in lua
func LUAFuncDockerImageTargetUser(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
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
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandUser{Name: name})
	// Done
	return 0
}

// LUAFuncDockerImageTargetWorkdir defines DockerImage.workdir function in lua
func LUAFuncDockerImageTargetWorkdir(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
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
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandWorkdir{Path: path})
	// Done
	return 0
}

// LUAFuncDockerImageTargetEnv defines DockerImage.env function in lua
func LUAFuncDockerImageTargetEnv(L *lua.LState) int {
	target := LUADockerImageTargetSelf(L)
	if target == nil {
		return 0
	}
	if len(target.DockerImage.Commands) == 0 {
		L.ArgError(0, "From command must be the first one")
		return 0
	}
	// Get key
	key := L.CheckString(2)
	if key == "" {
		L.ArgError(2, "Require key")
		return 0
	}
	// Get value
	value := L.CheckString(3)
	if value == "" {
		L.ArgError(3, "Require value")
		return 0
	}
	// Add command
	target.DockerImage.Commands = append(target.DockerImage.Commands, &DockerImageCommandEnv{Key: key, Value: value})
	// Done
	return 0
}
