// Author: lipixun
// Created Time : 三 11/23 18:46:57 2016
//
// File Name: app.go
// Description:
//	The runner app
package runner

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const (
	StatusRunning = 0
	StatusExited  = 1
	StatusError   = 2

	InstanceInfoFileName  = "info.json"
	InstanceLogStderrName = "stderr.log"
	InstanceLogStdoutName = "stdout.log"

	SignalInt  = 2
	SignalQuit = 3
	SignalKill = 9
)

type AppRunner struct {
	rootPath string
}

func New(p string) *AppRunner {
	return &AppRunner{rootPath: p}
}

// List all instances
func (this *AppRunner) List(onlyRunning bool) ([]*AppInstance, error) {
	if !onlyRunning {
		return this.loadInstances(nil, nil)
	} else {
		return this.loadInstances(nil, func(instance *AppInstance) bool {
			status, _ := instance.GetStatus()
			return status == StatusRunning
		})
	}
}

// Cleanup will remove all stopped instances
func (this *AppRunner) CleanAll() error {
	instances, err := this.List(false)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		status, _ := instance.GetStatus()
		if status == StatusExited {
			// Remove it
			if err := this.RemoveInstance(instance.ID); err != nil {
				return err
			}
		}
	}
	// Done
	return nil
}

func (this *AppRunner) GetInstance(id string) (*AppInstance, error) {
	instances, err := this.loadInstances(func(_id string) bool {
		return _id == id
	}, nil)
	if err != nil {
		return nil, err
	} else if len(instances) == 0 {
		return nil, nil
	} else {
		return instances[0], nil
	}
}

func (this *AppRunner) RemoveInstance(id string) error {
	return os.RemoveAll(filepath.Join(this.rootPath, id))
}

func (this *AppRunner) GetInstancesByName(name string) ([]*AppInstance, error) {
	return this.loadInstances(nil, func(instance *AppInstance) bool {
		return instance.Name == name
	})
}

func (this *AppRunner) GetRunningInstancesByName(name string) ([]*AppInstance, error) {
	return this.loadInstances(nil, func(instance *AppInstance) bool {
		if instance.Name == name {
			status, _ := instance.GetStatus()
			return status == StatusRunning
		}
		return false
	})
}

type AppStartOptions struct {
	Args       []string `json:"args"`
	WorkDir    string   `json:"workdir"`
	Singleton  bool     `json:"singleton"`
	Background bool     `json:"background"`
}

func (this *AppRunner) Start(name string, comamnd string, options AppStartOptions) (*AppInstance, error) {
	// TODO: We may need a system-wide lock to ensure the singleton
	if options.Singleton {
		// Ensure all other apps are stopped
		instances, err := this.GetInstancesByName(name)
		if err != nil {
			return nil, err
		}
		// Check the status
		for _, instance := range instances {
			if status, err := instance.GetStatus(); err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to get the status of process [%d], error: %s", instance.Pid, err))
			} else if status == StatusRunning {
				// Stop it
				if err := instance.Stop(); err != nil {
					return nil, errors.New(fmt.Sprintf("Failed to stop process [%d], error: %s", instance.Pid, err))
				}
			}
		}
	}
	// Start this app
	id, err := this.getNextRandomID()
	if err != nil {
		return nil, err
	}
	instancePath := filepath.Join(this.rootPath, id)
	if err := os.MkdirAll(instancePath, os.ModePerm); err != nil {
		return nil, err
	}
	var succeed bool = false
	defer func() {
		if !succeed {
			// Remove the instance path
			os.RemoveAll(instancePath)
		}
	}()
	// Create the stderr / stdout
	var stdout, stderr io.Writer
	stderrLogFile, err := os.Create(filepath.Join(instancePath, InstanceLogStderrName))
	if err != nil {
		return nil, err
	}
	stdoutLogFile, err := os.Create(filepath.Join(instancePath, InstanceLogStdoutName))
	if err != nil {
		return nil, err
	}
	if options.Background {
		// Run in background
		stderr = stderrLogFile
		stdout = stdoutLogFile
	} else {
		// Run in foreground
		stderr = io.MultiWriter(os.Stderr, stderrLogFile)
		stdout = io.MultiWriter(os.Stdout, stdoutLogFile)
	}
	// Generate the command
	cmd := exec.Command(comamnd, options.Args...)
	cmd.Dir = options.WorkDir
	cmd.Env = os.Environ()
	if options.Background {
		cmd.Stdin = nil
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.ExtraFiles = []*os.File{stderrLogFile, stdoutLogFile}
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
	}
	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	pid := cmd.Process.Pid
	// Background
	if options.Background {
		if err := cmd.Process.Release(); err != nil {
			return nil, err
		}
	}
	// Good the command is started, write the info
	instance := AppInstance{
		ID:      id,
		Time:    time.Now(),
		Name:    name,
		Command: comamnd,
		Options: options,
		Pid:     pid,
	}
	data, err := json.Marshal(&instance)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(instancePath, InstanceInfoFileName), data, os.ModePerm); err != nil {
		return nil, err
	}
	// Done
	succeed = true
	return &instance, nil
}

func (this *AppRunner) Stop(id string, clean bool) error {
	instance, err := this.GetInstance(id)
	if err != nil {
		return err
	}
	if instance == nil {
		return errors.New("Application instance not found")
	}
	if err := instance.Stop(); err != nil {
		return err
	}
	if clean {
		if err := this.Clean(instance.ID); err != nil {
			return err
		}
	}
	// Done
	return nil
}

func (this *AppRunner) Restart(id string, clean bool) (*AppInstance, error) {
	instance, err := this.GetInstance(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, errors.New("Application instance not found")
	}
	if err := instance.Stop(); err != nil {
		return nil, err
	}
	if clean {
		if err := this.Clean(instance.ID); err != nil {
			return nil, err
		}
	}
	return this.Start(instance.Name, instance.Command, instance.Options)
}

func (this *AppRunner) Clean(id string) error {
	return os.RemoveAll(filepath.Join(this.rootPath, id))
}

func (this *AppRunner) loadInstances(idFilterFunc func(string) bool, instanceFilterFunc func(*AppInstance) bool) ([]*AppInstance, error) {
	infos, err := ioutil.ReadDir(this.rootPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var instances []*AppInstance
	for _, info := range infos {
		// Check the id
		if idFilterFunc != nil && !idFilterFunc(info.Name()) {
			continue
		}
		// Read the info file
		data, err := ioutil.ReadFile(filepath.Join(this.rootPath, info.Name(), InstanceInfoFileName))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to read Info file [%s] not found in instance [%s], error: %s\n", InstanceInfoFileName, info.Name(), err))
		}
		var instance AppInstance
		err = json.Unmarshal(data, &instance)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Info file [%s] in instance [%s] is broken, error: %s\n", InstanceInfoFileName, info.Name(), err))
		}
		if instanceFilterFunc == nil || instanceFilterFunc(&instance) {
			instances = append(instances, &instance)
		}
	}
	return instances, nil
}

func (this *AppRunner) getNextRandomID() (string, error) {
	for {
		idBytes := make([]byte, 8)
		if _, err := rand.Read(idBytes); err != nil {
			return "", err
		}
		idStr := hex.EncodeToString(idBytes)
		if _, err := os.Stat(filepath.Join(this.rootPath, idStr)); err != nil {
			if os.IsNotExist(err) {
				return idStr, nil
			}
		}
	}
}

type AppInstance struct {
	ID      string          `json:"id"`
	Time    time.Time       `json:"time"`
	Name    string          `json:"name"`
	Command string          `json:"command"`
	Options AppStartOptions `json:"options"`
	Pid     int             `json:"pid"`
}

// Wait t
func (this *AppInstance) Wait() {
	proc, err := os.FindProcess(this.Pid)
	if err != nil {
		return
	}
	_, err = proc.Wait()
	if err != nil {
		return
	}
}

// Get the app instance status
// Returns: Status, error
func (this *AppInstance) GetStatus() (int, error) {
	proc, err := os.FindProcess(this.Pid)
	if err != nil {
		return StatusExited, nil
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		if err.Error() == "os: process already finished" {
			return StatusExited, nil
		} else {
			return StatusError, err
		}
	} else {
		return StatusRunning, nil
	}
}

// Stop this instance
func (this *AppInstance) Stop() error {
	proc, err := os.FindProcess(this.Pid)
	if err != nil {
		return nil
	}
	if err := proc.Signal(syscall.Signal(SignalInt)); err != nil {
		if err.Error() == "os: process already finished" {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

// Quit this instance
func (this *AppInstance) Quit() error {
	proc, err := os.FindProcess(this.Pid)
	if err != nil {
		return nil
	}
	if err := proc.Signal(syscall.Signal(SignalQuit)); err != nil {
		if err.Error() == "os: process already finished" {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}

// Kill this instance
func (this *AppInstance) Kill() error {
	proc, err := os.FindProcess(this.Pid)
	if err != nil {
		return nil
	}
	if err := proc.Signal(syscall.Signal(SignalKill)); err != nil {
		if err.Error() == "os: process already finished" {
			return nil
		} else {
			return err
		}
	} else {
		return nil
	}
}
