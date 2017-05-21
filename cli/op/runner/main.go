// Author: lipixun
//
// File Name: main.go
// Description:

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/golang/protobuf/ptypes"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/cli/op/common"
	"github.com/ops-openlight/openlight/pkg/rule"
	runnerRule "github.com/ops-openlight/openlight/pkg/rule/modules/runner"
	"github.com/ops-openlight/openlight/pkg/runner"
)

const (
	// DirName defines the directory name of ops runner
	DirName = "runner"

	// DefaultRunnerSpecFileName defines the default runner spec filename
	DefaultRunnerSpecFileName = "runfile"

	// CommandListFormat define the format to list commands
	CommandListFormat = "%-20v %-20v %-60v %v"
	// ProcessListFormat define the format to list processes
	ProcessListFormat = "%-20v%-64v%-24v%-10v%-10v"
)

// GetCommands gets runner commands
func GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "rn",
			Usage: "Command Runner",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "spec",
					Usage: "Spec file",
				},
				cli.StringFlag{
					Name:  "workspace",
					Usage: "Workspace path",
					Value: common.DefaultWorkspacePath,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:   "cmds",
					Usage:  "List commands",
					Action: listCommands,
				},
				{
					Name:   "start",
					Usage:  "Start the application",
					Action: start,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "id,i",
							Usage: "The command (id) to start",
						},
						cli.BoolFlag{
							Name:  "detach,d",
							Usage: "Start the command and detach",
						},
						cli.BoolFlag{
							Name:  "log",
							Usage: "Show log after started. This flag will be available if detach is set",
						},
					},
				},
				{
					Name:   "stop",
					Usage:  "Stop a process",
					Action: stop,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "id,i",
							Usage: "The application id to stop",
						},
					},
				},
				{
					Name:   "restart",
					Usage:  "Restart a process",
					Action: restart,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "id,i",
							Usage: "The application id to restart",
						},
						cli.BoolFlag{
							Name:  "log",
							Usage: "Show log after started. This flag will be available if detach is set",
						},
					},
				},
				{
					Name:   "ps",
					Usage:  "List processes",
					Action: listProcesses,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "all,a",
							Usage: "List all processes (Includes error and stopped)",
						},
					},
				},
				{
					Name:   "logs",
					Usage:  "Show the log of a process",
					Action: showlogs,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "id,i",
							Usage: "The application id to show logs",
						},
						cli.BoolFlag{
							Name:  "stdout",
							Usage: "Show the stdout log instead of stderr",
						},
						cli.BoolFlag{
							Name:  "follow,f",
							Usage: "Follow the log",
						},
						cli.IntFlag{
							Name:  "lines,n",
							Usage: "Output the last n lines of log instead of all",
						},
					},
				},
				{
					Name:   "clean",
					Usage:  "Clean useless data",
					Action: clean,
				},
			},
		},
	}
}

type _RunnerSpec struct {
	*pbSpec.RunFile
	Dirname string
}

func loadSpec(c *cli.Context) (*_RunnerSpec, error) {
	filename := c.GlobalString("spec")
	if filename == "" {
		filename = DefaultRunnerSpecFileName
	}
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	// Load spec file
	engine := rule.NewEngine()
	ctx, err := engine.ParseFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to load spec file [%v]: %v", filename, err)
	}
	// Return spec
	return &_RunnerSpec{
		ctx.GetModule("runner").(runnerRule.Module).Spec(),
		filepath.Dir(filename),
	}, nil
}

func newManager(c *cli.Context) (*runner.Manager, error) {
	workspacePath := c.String("workspace")
	if workspacePath == "" {
		workspacePath = common.DefaultWorkspacePath
	}
	return runner.NewManager(filepath.Join(workspacePath, DirName))
}

func listCommands(c *cli.Context) error {
	spec, err := loadSpec(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to load spec: %v", err), 1)
	}
	fmt.Printf(CommandListFormat, "ID", "Name", "Args...", "Comment")
	fmt.Println()
	for id, cmd := range spec.Commands {
		fmt.Printf(CommandListFormat, id, cmd.Name, strings.Join(cmd.Args, " "), cmd.Comment)
		fmt.Println()
	}
	return nil
}

func start(c *cli.Context) error {
	id, detach, isLog := c.String("id"), c.Bool("detach"), c.Bool("log")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	spec, err := loadSpec(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to load spec: %v", err), 1)
	}
	cmd := spec.Commands[id]
	if cmd == nil {
		return cli.NewExitError(fmt.Sprintf("Command [%v] not found", id), 1)
	}
	if len(c.Args()) > 0 {
		cmd.Args = append(cmd.Args, c.Args()...)
	}
	// Start it
	proc, err := manager.Start(spec.Dirname, cmd, detach)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to start process: %v", err), 1)
	}
	if !detach {
		if err := proc.Wait(); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to wait process: %v", err), 1)
		}
		if err := manager.Clean(); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to clean: %v", err), 1)
		}
	} else {
		go func() {
			proc.Wait()
		}()
		// Check log
		if isLog {
			return showProcLogs(proc, false, true, 1000)
		}
	}
	// Done
	return nil
}

func stop(c *cli.Context) error {
	id := c.String("id")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	if err := manager.Stop(id); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to stop process: %v", err), 1)
	}
	if err := manager.Clean(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to clean: %v", err), 1)
	}
	return nil
}

func restart(c *cli.Context) error {
	id, isLog := c.String("id"), c.Bool("log")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	proc, err := manager.Restart(id)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to restart process: %v", err), 1)
	}
	go func() {
		proc.Wait()
	}()
	// Check log
	if isLog {
		return showProcLogs(proc, false, true, 1000)
	}
	return nil
}

func listProcesses(c *cli.Context) error {
	showAll := c.Bool("all")
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	procs, err := manager.List()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to list processes: %v", err), 1)
	}
	fmt.Printf(ProcessListFormat, "ID", "Name", "Start Time", "PID", "Alive")
	fmt.Println()
	for _, proc := range procs {
		if !proc.Loaded() && showAll {
			fmt.Printf(ProcessListFormat, proc.Id, proc.LoadError(), "", "", "")
			fmt.Println()
			continue
		}
		var name, pid, startTimeStr string
		if proc.Command != nil {
			name = proc.Command.Name
		}
		if proc.State != nil {
			pid = fmt.Sprintf("%v", proc.State.Pid)
			startTime, err := ptypes.Timestamp(proc.State.StartTime)
			if err != nil {
				fmt.Printf(ProcessListFormat, proc.Id, err, "", "", "")
				fmt.Println()
				continue
			}
			startTimeStr = startTime.Format(time.RFC3339)
		}
		isAlive := proc.IsAlive()
		if isAlive || showAll {
			fmt.Printf(ProcessListFormat, proc.Id, name, startTimeStr, pid, isAlive)
			fmt.Println()
		}
	}
	// Done
	return nil
}

func showlogs(c *cli.Context) error {
	id, stdout, follow, lines := c.String("id"), c.Bool("stdout"), c.Bool("follow"), c.Int("lines")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	proc, err := manager.Get(id)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get process: %v", err), 1)
	}
	if !proc.Loaded() {
		return cli.NewExitError(fmt.Sprintf("Failed to load process: %v", proc.LoadError()), 1)
	}
	if proc.Log == nil {
		return cli.NewExitError("Process log info is not defined", 1)
	}
	// Show logs
	return showProcLogs(proc, stdout, follow, lines)
}

func showProcLogs(proc *runner.Process, stdout, follow bool, lines int) error {
	// Run tail
	var logFileName string
	if stdout {
		logFileName = proc.Log.Stdout
	} else {
		logFileName = proc.Log.Stderr
	}
	var args []string
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, logFileName)
	cmd := exec.Command("tail", args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to run tail: %v", err), 1)
	}
	// Done
	return nil
}

func clean(c *cli.Context) error {
	manager, err := newManager(c)
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create manager: %v", err), 1)
	}
	// Clean
	if err := manager.Clean(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to clean: %v", err), 1)
	}
	// Done
	return nil
}
