// Author: lipixun
// File Name: runner.go
// Description:

package runner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	cli "gopkg.in/urfave/cli.v1"

	pbSpec "github.com/ops-openlight/openlight/protoc-gen-go/spec"

	"github.com/ops-openlight/openlight/pkg/repository"
	"github.com/ops-openlight/openlight/pkg/runner"
)

const (
	// CommandListFormat define the format to list commands
	CommandListFormat = "%-20v %-20v %-60v %v"
	// ProcessListFormat define the format to list processes
	ProcessListFormat = "%-20v%-64v%-24v%-10v%-10v"
)

// Runner implements the runner command line functions
type Runner struct{}

// GetCommands returns the commands
func (r *Runner) GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "runner",
			Usage: "Command Runner",
			Subcommands: []cli.Command{
				{
					Name:   "cmds",
					Usage:  "List commands",
					Action: r.listCommands,
				},
				{
					Name:   "start",
					Usage:  "Start the application",
					Action: r.start,
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
					Action: r.stop,
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
					Action: r.restart,
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
					Action: r.listProcesses,
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
					Action: r.showlogs,
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
					Action: r.clean,
				},
			},
		},
	}
}

func (r *Runner) init(c *cli.Context) (*repository.LocalRepository, *runner.Manager, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Failed to init: %v", err), 1)
	}
	repo, err := repository.NewLocalRepository(root)
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Failed to init: %v", err), 1)
	}
	manager, err := runner.NewManageByRepository(repo)
	if err != nil {
		return nil, nil, cli.NewExitError(fmt.Sprintf("Failed to init: %v", err), 1)
	}
	return repo, manager, nil
}

func (r *Runner) getRunnerSpec(repo *repository.LocalRepository) (*pbSpec.RunFile, error) {
	ruleFiles, err := repo.RootRuleFiles()
	if err != nil {
		return nil, cli.NewExitError(err.Error(), 1)
	}
	return ruleFiles.GetRun(), nil
}

func (r *Runner) listCommands(c *cli.Context) error {
	repo, _, err := r.init(c)
	if err != nil {
		return err
	}
	spec, err := r.getRunnerSpec(repo)
	if err != nil {
		return err
	}
	fmt.Printf(CommandListFormat, "ID", "Name", "Args...", "Comment")
	fmt.Println()
	for id, cmd := range spec.GetCommands() {
		fmt.Printf(CommandListFormat, id, cmd.Name, strings.Join(cmd.Args, " "), cmd.Comment)
		fmt.Println()
	}
	return nil
}

func (r *Runner) start(c *cli.Context) error {
	id, detach, isLog := c.String("id"), c.Bool("detach"), c.Bool("log")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	repo, manager, err := r.init(c)
	if err != nil {
		return err
	}
	spec, err := r.getRunnerSpec(repo)
	if err != nil {
		return err
	}
	cmd := spec.Commands[id]
	if cmd == nil {
		return cli.NewExitError(fmt.Sprintf("Command [%v] not found", id), 1)
	}
	if len(c.Args()) > 0 {
		cmd.Args = append(cmd.Args, c.Args()...)
	}
	// Start it
	proc, err := manager.Start(repo.Path(), cmd, detach)
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
			return r.showProcLogs(proc, false, true, 1000)
		}
	}
	// Done
	return nil
}

func (r *Runner) stop(c *cli.Context) error {
	id := c.String("id")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	_, manager, err := r.init(c)
	if err != nil {
		return err
	}
	if err := manager.Stop(id); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to stop process: %v", err), 1)
	}
	if err := manager.Clean(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to clean: %v", err), 1)
	}
	return nil
}

func (r *Runner) restart(c *cli.Context) error {
	id, isLog := c.String("id"), c.Bool("log")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	_, manager, err := r.init(c)
	if err != nil {
		return err
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
		return r.showProcLogs(proc, false, true, 1000)
	}
	return nil
}

func (r *Runner) listProcesses(c *cli.Context) error {
	showAll := c.Bool("all")
	_, manager, err := r.init(c)
	if err != nil {
		return err
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

func (r *Runner) showlogs(c *cli.Context) error {
	id, stdout, follow, lines := c.String("id"), c.Bool("stdout"), c.Bool("follow"), c.Int("lines")
	if id == "" {
		return cli.NewExitError("Require id", 1)
	}
	_, manager, err := r.init(c)
	if err != nil {
		return err
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
	return r.showProcLogs(proc, stdout, follow, lines)
}

func (r *Runner) showProcLogs(proc *runner.Process, stdout, follow bool, lines int) error {
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

func (r *Runner) clean(c *cli.Context) error {
	_, manager, err := r.init(c)
	if err != nil {
		return err
	}
	// Clean
	if err := manager.Clean(); err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to clean: %v", err), 1)
	}
	// Done
	return nil
}
