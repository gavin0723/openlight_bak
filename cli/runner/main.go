// Author: lipixun
// Created Time : 四 10/20 10:04:23 2016
//
// File Name: main.go
// Description:
//
package runner

import (
	"fmt"
	opcli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/pkg/log"
	"github.com/ops-openlight/openlight/pkg/runner"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/exec"
)

const (
	LogHeader = "CLI.Runner"

	StatusFormat = "%-24s%-32s%-48s%-10s%s\n"
)

func GetCommand() []cli.Command {
	return []cli.Command{
		{
			Category: "Runner",
			Name:     "start",
			Usage:    "Start the application",
			Action:   start,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "background,b",
					Usage: "Start the application in background",
				},
				cli.StringFlag{
					Name:  "app,p",
					Usage: "The application to start",
				},
				cli.StringFlag{
					Name:  "command,c",
					Usage: "The command to start",
				},
				cli.StringFlag{
					Name:  "wd,w",
					Usage: "The command work dir",
				},
				cli.BoolFlag{
					Name:  "singleton",
					Usage: "Make the application as a singleton (for current user)",
				},
				cli.BoolFlag{
					Name:  "ignore-config-args,i",
					Usage: "Ignore the arguments defined in config",
				},
			},
		},
		{
			Category: "Runner",
			Name:     "logs",
			Usage:    "Show the log of an application instance",
			Action:   showlogs,
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
			Category: "Runner",
			Name:     "stop",
			Usage:    "Stop the application instance",
			Action:   stop,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "id,i",
					Usage: "The application id to stop",
				},
				cli.BoolFlag{
					Name:  "clean,c",
					Usage: "Clean the application after stopped",
				},
			},
		},
		{
			Category: "Runner",
			Name:     "restart",
			Usage:    "Restart the application instance ",
			Action:   restart,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id,i",
					Usage: "The application id to restart",
				},
				cli.BoolFlag{
					Name:  "clean,c",
					Usage: "Clean the application after stopped",
				},
			},
		},
		{
			Category: "Runner",
			Name:     "status",
			Usage:    "List the application instance status",
			Action:   status,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all,a",
					Usage: "List all application instances",
				},
			},
		},
		{
			Category: "Runner",
			Name:     "clean-runner",
			Usage:    "Clean the application runners, this is remove all runner data of stopped application instances",
			Action:   clean,
		},
	}
}

func start(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Check parameters
	appName := c.String("app")
	background := c.Bool("background")
	command := c.String("command")
	workdir := c.String("wd")
	singleton := c.Bool("singleton")
	ignoreConfigArgs := c.Bool("ignore-config-args")
	args := c.Args()
	if appName == "" {
		logger.LeveledPrintln(log.LevelError, "Require application name")
		return cli.NewExitError("", 1)
	}
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	// Start
	instance, err := r.Start(appName, command, runner.AppStartOptions{
		Args:           args,
		WorkDir:        workdir,
		Singleton:      singleton,
		Background:     background,
		IgnoreSpecArgs: ignoreConfigArgs,
	})
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to start application, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	if !background {
		instance.Wait()
	}
	// Done
	return nil
}

func showlogs(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Check parameters
	id := c.String("id")
	apps := c.Args()
	isStdout := c.Bool("stdout")
	follow := c.Bool("follow")
	lines := c.Int("lines")
	if id == "" && len(apps) == 0 {
		logger.LeveledPrintln(log.LevelError, "Require either id or app")
		return cli.NewExitError("", 1)
	} else if id != "" && len(apps) > 0 {
		logger.LeveledPrintln(log.LevelError, "Require either id or app but not both")
		return cli.NewExitError("", 1)
	}
	if len(apps) > 1 {
		logger.LeveledPrintln(log.LevelError, "Cannot specify more than 1 apps")
		return cli.NewExitError("", 1)
	}
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	// Get the id
	if len(apps) > 0 {
		// Get id by app
		instances, err := r.GetInstancesByName(apps[0])
		if err != nil {
			logger.LeveledPrintf(log.LevelError, "Failed to get instances by name, error: %s\n", err)
			return cli.NewExitError("", 1)
		}
		if len(instances) == 0 {
			logger.LeveledPrintln(log.LevelError, "No instance found for this application")
			return cli.NewExitError("", 1)
		} else if len(instances) > 1 {
			// Filter out the running one
			var runningInstances []*runner.AppInstance
			for _, instance := range instances {
				if status, _ := instance.GetStatus(); status == runner.StatusRunning {
					runningInstances = append(runningInstances, instance)
				}
			}
			if len(runningInstances) == 1 {
				// Use this running instance
				id = runningInstances[0].ID
			} else {
				logger.LeveledPrintln(log.LevelError, "More than 1 instance found, cannot show log by application name")
				return cli.NewExitError("", 1)
			}
		} else {
			id = instances[0].ID
		}
	}
	// Get the instance log file
	filename := r.GetLogFile(id, isStdout)
	var args []string
	if lines > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, filename)
	// Run the tail command
	cmd := exec.Command("tail", args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	// Done
	return nil
}

func stop(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Get parameters
	ids := c.StringSlice("id")
	apps := c.Args()
	clean := c.Bool("clean")
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	// Stop the application
	for _, id := range ids {
		logger.Printf("Stopping [%s] ...... ", id)
		if err := r.Stop(id, clean); err != nil {
			logger.LeveledHeadedPrint("", log.LevelError, "Error: %s\n", err)
		} else {
			logger.LeveledHeadedPrint("", log.LevelSuccess, "Done\n")
		}
	}
	for _, name := range apps {
		instances, err := r.GetInstancesByName(name)
		if err != nil {
			logger.LeveledPrintf(log.LevelError, "Failed to get instances by name [%s], error: %s\n", name, err)
			continue
		}
		for _, instance := range instances {
			s, _ := instance.GetStatus()
			if s != runner.StatusExited {
				logger.Printf("Stopping [%s] ...... ", instance.ID)
				if err := r.Stop(instance.ID, clean); err != nil {
					logger.LeveledHeadedPrint("", log.LevelError, "Error: %s\n", err)
				} else {
					logger.LeveledHeadedPrint("", log.LevelSuccess, "Done\n")
				}
			}
		}
	}
	// Done
	return nil
}

func status(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Get parameters
	statusAll := c.Bool("all")
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	// Get the application status
	instances, err := r.List(!statusAll)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to list instances, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// List it
	fmt.Printf(StatusFormat, "ID", "Name", "Start Time", "Status", "Error")
	for _, instance := range instances {
		var status string
		s, err := instance.GetStatus()
		if s == runner.StatusRunning {
			status = "Running"
		} else if s == runner.StatusExited {
			status = "Exited"
		} else {
			status = "Error"
		}
		var errmsg string
		if err != nil {
			errmsg = err.Error()
		}
		fmt.Printf(StatusFormat, instance.ID, instance.Name, instance.Time, status, errmsg)
	}
	// Done
	return nil
}

func restart(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Get parameters
	id := c.String("id")
	apps := c.Args()
	clean := c.Bool("clean")
	if id == "" && len(apps) == 0 {
		logger.LeveledPrintln(log.LevelError, "Require either id or app")
		return cli.NewExitError("", 1)
	} else if id != "" && len(apps) > 0 {
		logger.LeveledPrintln(log.LevelError, "Require either id or app but not both")
		return cli.NewExitError("", 1)
	}
	if len(apps) > 1 {
		logger.LeveledPrintln(log.LevelError, "Cannot specify more than 1 apps")
		return cli.NewExitError("", 1)
	}
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	if len(apps) > 0 {
		// Get id by app
		instances, err := r.GetRunningInstancesByName(apps[0])
		if err != nil {
			logger.LeveledPrintf(log.LevelError, "Failed to get instances by name, error: %s\n", err)
			return cli.NewExitError("", 1)
		}
		if len(instances) == 0 {
			logger.LeveledPrintln(log.LevelError, "No instance found for this application")
			return cli.NewExitError("", 1)
		} else if len(instances) > 1 {
			logger.LeveledPrintln(log.LevelError, "More than 1 instance found, cannot restart by application name")
			return cli.NewExitError("", 1)
		} else {
			id = instances[0].ID
		}
	}
	// Restart the application
	logger.Printf("Restarting [%s] ...... ", id)
	if instance, err := r.Restart(id, clean); err != nil {
		logger.LeveledHeadedPrintf("", log.LevelError, "Error: %s\n", err)
	} else {
		logger.LeveledHeadedPrintf("", log.LevelSuccess, "Done. New Instance ID [%s]\n", instance.ID)
		if !instance.Options.Background {
			instance.Wait()
		}
	}
	// Done
	return nil
}

func clean(c *cli.Context) error {
	ws, err := opcli.GetWorkspace(c)
	if err != nil {
		return err
	}
	logger := ws.Logger.GetLoggerWithHeader(LogHeader)
	// Create runner
	r, err := runner.New(ws)
	if err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to create runner, error: %s", err)
		return cli.NewExitError("", 1)
	}
	// Clean
	if err := r.CleanAll(); err != nil {
		logger.LeveledPrintf(log.LevelError, "Failed to clean, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Done
	return nil
}
