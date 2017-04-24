// Author: lipixun
// File Name: process.go
// Description:

package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"
)

// Process is the process
type Process struct {
	// Run process
	pbSpec.RunProcess
	// Command
	cmd *exec.Cmd
	// Load error
	loadError error
}

// LoadProcess loads process from file
func LoadProcess(filename string) *Process {
	_, err := os.Stat(filename)
	if err != nil {
		return &Process{
			loadError: fmt.Errorf("Failed to check data file: %v", err),
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		return &Process{
			loadError: fmt.Errorf("Failed to open data file: %v", err),
		}
	}
	defer file.Close()
	var proc Process
	if err := jsonpb.Unmarshal(file, &proc.RunProcess); err != nil {
		return &Process{
			loadError: fmt.Errorf("Failed to load data file: %v", err),
		}
	}
	return &proc
}

// Loaded returns if the process is loaded
func (p *Process) Loaded() bool {
	if p == nil {
		return false
	}
	return p.loadError == nil
}

// LoadError returns the load error
func (p *Process) LoadError() error {
	if p == nil {
		return errors.New("Nil returned")
	}
	return p.loadError
}

// IsAlive returns if the process is alive or not
func (p *Process) IsAlive() bool {
	if p.State == nil {
		return false
	}
	proc, err := os.FindProcess(int(p.State.Pid))
	if err != nil {
		return false
	}
	if err := proc.Signal(syscall.Signal(SignalNone)); err != nil {
		return false
	}
	return true
}

// Cmd returns the command
func (p *Process) Cmd() *exec.Cmd {
	return p.cmd
}

// Start the process
func (p *Process) start(detach bool) error {
	if p.Command == nil {
		return errors.New("Require command")
	}
	if len(p.Command.Args) == 0 {
		return errors.New("Require at least 1 command args")
	}
	if p.State == nil {
		return errors.New("Require state")
	}
	if p.Log == nil {
		return errors.New("Require log")
	}
	// Create start command
	stderrLogFile, err := os.Create(p.Log.Stderr)
	if err != nil {
		return fmt.Errorf("Failed to open stderr log file: %v", err)
	}
	stdoutLogFile, err := os.Create(p.Log.Stdout)
	if err != nil {
		return fmt.Errorf("Failed to open stdout log file: %v", err)
	}
	// Generate the command
	cmd := exec.Command(p.Command.Args[0], p.Command.Args[1:]...)
	cmd.Dir = p.State.Workdir
	// Env
	var env = make([]string, len(os.Environ()))
	copy(env, os.Environ())
	for _, e := range p.Command.Envs {
		env = append(env, e)
	}
	cmd.Env = env
	// Detach or not
	if detach {
		cmd.Stdin = nil
		cmd.Stdout = stdoutLogFile
		cmd.Stderr = stderrLogFile
		cmd.ExtraFiles = []*os.File{stderrLogFile, stdoutLogFile}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
			//Setpgid: true,
		}
		defer stderrLogFile.Close()
		defer stdoutLogFile.Close()
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = io.MultiWriter(os.Stdout, stdoutLogFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, stderrLogFile)
	}
	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}
	// Update data
	p.cmd = cmd
	p.State.Pid = int32(cmd.Process.Pid)
	p.State.StartTime, err = ptypes.TimestampProto(time.Now())
	if err != nil {
		return err
	}
	// Done
	return nil
}

// Wait for the process
func (p *Process) Wait() error {
	if p.cmd == nil {
		if !p.IsAlive() {
			return nil
		}
		proc, err := os.FindProcess(int(p.State.Pid))
		if err != nil {
			return err
		}
		_, err = proc.Wait()
		return err
	}
	return p.cmd.Wait()
}

// Stop process
func (p *Process) Stop() error {
	if p.IsAlive() {
		// Kill it
		proc, err := os.FindProcess(int(p.State.Pid))
		if err != nil {
			return err
		}
		if err := proc.Signal(syscall.Signal(SignalInt)); err != nil {
			return err
		}
	}
	// Done
	return nil
}

// Restart process
func (p *Process) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	return p.start(true)
}

// Dump data to file
func (p *Process) Dump(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	m := jsonpb.Marshaler{}
	return m.Marshal(file, &p.RunProcess)
}
