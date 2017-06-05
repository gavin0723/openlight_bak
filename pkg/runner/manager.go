// Author: lipixun
// File Name: manager.go
// Description:

package runner

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/repository"
	"github.com/ops-openlight/openlight/pkg/utils"
)

// Filenames
const (
	RunProcessDataFileName  = "data.json"
	RunProcessLogStderrName = "stderr.log"
	RunProcessLogStdoutName = "stdout.log"
)

// Signals
const (
	SignalNone = 0
	SignalInt  = 2
	SignalQuit = 3
	SignalKill = 9
)

// Manager is used to manage all run processes
type Manager struct {
	root string
}

// NewManager creates a new Manager
func NewManager(root string) (*Manager, error) {
	root, err := utils.GetRealPath(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(root)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if os.IsNotExist(err) {
		// Create new directory
		if err := os.MkdirAll(root, os.ModePerm); err != nil {
			return nil, err
		}
	} else if !info.IsDir() {
		return nil, fmt.Errorf("Root [%v] is not a directory", root)
	}
	// Done
	return &Manager{
		root: root,
	}, nil
}

// NewManageByRepository creates a new RepositoryManager
func NewManageByRepository(repo *repository.LocalRepository) (*Manager, error) {
	if repo == nil {
		return nil, errors.New("Require repository")
	}
	dirname, err := repo.InitOutputDir("", _OutputDirname)
	if err != nil {
		return nil, err
	}
	return NewManager(dirname)
}

// Start a new command
func (manager *Manager) Start(workdir string, cmd *pbSpec.RunCommand, detach bool) (*Process, error) {
	if cmd == nil {
		return nil, errors.New("Require command")
	}
	// Init
	id, err := manager.initNextProc()
	if err != nil {
		return nil, err
	}
	// Create process
	var p Process
	p.Id = id
	p.Command = cmd
	p.State = &pbSpec.RunProcessState{
		Workdir: filepath.Join(workdir, cmd.Workdir),
	}
	p.Log = &pbSpec.RunProcessLog{
		Stdout: filepath.Join(manager.root, id, RunProcessLogStdoutName),
		Stderr: filepath.Join(manager.root, id, RunProcessLogStderrName),
	}
	// Start it
	if err := p.start(detach); err != nil {
		return nil, err
	}
	// Dump process
	if err := p.Dump(filepath.Join(manager.root, id, RunProcessDataFileName)); err != nil {
		return nil, err
	}
	// Done
	return &p, nil
}

// Restart a process
func (manager *Manager) Restart(id string) (*Process, error) {
	proc, err := manager.Get(id)
	if err != nil {
		return nil, err
	}
	if !proc.Loaded() {
		return nil, fmt.Errorf("Failed to load process: %v", proc.LoadError())
	}
	if err := proc.Restart(); err != nil {
		return nil, err
	}
	// Dump new process
	if err := proc.Dump(filepath.Join(manager.root, id, RunProcessDataFileName)); err != nil {
		return nil, err
	}
	// Done
	return proc, nil
}

// Stop a process
func (manager *Manager) Stop(id string) error {
	proc, err := manager.Get(id)
	if err != nil {
		return err
	}
	if !proc.Loaded() {
		return fmt.Errorf("Failed to load process: %v", proc.LoadError())
	}
	// Stop it
	return proc.Stop()
}

// initNextProc initialized for next process
func (manager *Manager) initNextProc() (string, error) {
	for {
		var idBytes = make([]byte, 8, 8)
		_, err := rand.Read(idBytes)
		if err != nil {
			return "", err
		}
		id := hex.EncodeToString(idBytes)
		dirname := filepath.Join(manager.root, id)
		if _, err := os.Stat(dirname); os.IsNotExist(err) {
			if err = os.MkdirAll(dirname, os.ModePerm); err != nil {
				return "", err
			}
			return id, nil
		}
	}
}

// Get a process
func (manager *Manager) Get(id string) (*Process, error) {
	dirname := filepath.Join(manager.root, id)
	info, err := os.Stat(dirname)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Path [%v] is not a directory", dirname)
	}
	// Load process
	return LoadProcess(filepath.Join(dirname, RunProcessDataFileName)), nil
}

// List processes
func (manager *Manager) List() ([]*Process, error) {
	infos, err := ioutil.ReadDir(manager.root)
	if err != nil {
		return nil, err
	}
	var procs []*Process
	for _, info := range infos {
		if info.IsDir() {
			procs = append(procs, LoadProcess(filepath.Join(manager.root, info.Name(), RunProcessDataFileName)))
		}
	}
	return procs, nil
}

// Clean useless data
func (manager *Manager) Clean() error {
	procs, err := manager.List()
	if err != nil {
		return err
	}
	for _, proc := range procs {
		if !proc.Loaded() || !proc.IsAlive() {
			if err := os.RemoveAll(filepath.Join(manager.root, proc.Id)); err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	_OutputDirname = "runner"
)
