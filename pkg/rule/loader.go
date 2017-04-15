// Author: lipixun
// File Name: loader.go
// Description:

package rule

import (
	"fmt"

	"github.com/yuin/gopher-lua"

	"github.com/ops-openlight/openlight/pkg/rule/modules"
	LUA "github.com/ops-openlight/openlight/pkg/rule/modules/lua"
)

// Ensure the interface is implemented
var _ Loader = (*_FileLoader)(nil)

// Loader defines the rule loader interface
type Loader interface {
	LUA.ModuleContext
	// Close loader
	Close()
}

// _Loader implements the loader basic functions
type _Loader struct {
	l       *lua.LState
	modules map[string]LUA.Module
}

func newLoader() _Loader {
	var loader _Loader
	// Create modules
	loader.modules = modules.NewModules(&loader)
	// Create lua engine
	var L = lua.NewState(lua.Options{IncludeGoStackTrace: true})
	// Register modules
	for _, module := range loader.Modules() {
		L.PreloadModule(module.Name(), module.InitLUAModule)
	}
	loader.l = L
	// Done
	return loader
}

// Modules return all modules
func (loader *_Loader) Modules() []LUA.Module {
	var mods []LUA.Module
	for _, module := range loader.modules {
		mods = append(mods, module)
	}
	return mods
}

// GetModule returns the module
func (loader *_Loader) GetModule(name string) LUA.Module {
	return loader.modules[name]
}

// Close loader
func (loader *_Loader) Close() {
	loader.l.Close()
}

// _FileLoader implements the loader interface by loading from file
type _FileLoader struct {
	_Loader
}

// NewFileLoader creates a new file loader
func NewFileLoader(filenames []string) (Loader, error) {
	var loader = _FileLoader{
		_Loader: newLoader(),
	}
	// Read files
	for _, filename := range filenames {
		if err := loader.l.DoFile(filename); err != nil {
			return nil, fmt.Errorf("Failed to load rule file [%v], error: %s", filename, err)
		}
	}
	// Done
	return &loader, nil
}
