// Author: lipixun
// Created Time : 四 10/20 10:04:23 2016
//
// File Name: main.go
// Description:
//
package runner

import (
	"errors"
	"fmt"
	opcli "github.com/ops-openlight/openlight/cli"
	"github.com/ops-openlight/openlight/cli/config"
	"github.com/ops-openlight/openlight/log"
	"github.com/ops-openlight/openlight/runner"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
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
			Name:     "log",
			Usage:    "Show the log of an application instance",
			Action:   showlog,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id,i",
					Usage: "The application id to stop",
				},
				cli.StringFlag{
					Name:  "app,p",
					Usage: "The application name to stop",
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
				cli.StringSliceFlag{
					Name:  "app,p",
					Usage: "The application name to stop",
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
					Usage: "The application id to stop",
				},
				cli.StringFlag{
					Name:  "app,p",
					Usage: "The application name to stop",
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
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Check parameters
	appName := c.String("app")
	background := c.Bool("background")
	command := c.String("command")
	workdir := c.String("wd")
	singleton := c.Bool("singleton")
	ignoreConfigArgs := c.Bool("ignore-config-args")
	args := c.Args()
	if appName == "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require application name")
		return cli.NewExitError("", 1)
	}
	// Load application config
	appConfigs, err := loadRunnerApplications(opc)
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to load runner configs, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Set the application parameters
	appConfig, ok := appConfigs[appName]
	if ok {
		// Use the config value
		if appConfig.Name != "" {
			appName = appConfig.Name
		}
		if command == "" {
			command = appConfig.Command
		}
		if workdir == "" {
			workdir = appConfig.Workdir
		}
		if singleton == false && appConfig.Singleton {
			singleton = appConfig.Singleton
		}
		if !ignoreConfigArgs && len(appConfig.Args) > 0 {
			newArgs := make([]string, len(appConfig.Args))
			copy(newArgs, appConfig.Args)
			newArgs = append(newArgs, args...)
			args = newArgs
		}
	}
	// Check the command parameters
	if command == "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require command")
		return cli.NewExitError("", 1)
	}
	if opc.Debug() {
		opc.Logger().LeveledPrintf(log.LevelDebug, "Start with command: %s %s\n", command, strings.Join(args, " "))
	}
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	instance, err := r.Start(appName, command, runner.AppStartOptions{
		Args:       args,
		WorkDir:    workdir,
		Singleton:  singleton,
		Background: background,
	})
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to start application, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	if !background {
		instance.Wait()
	}
	// Done
	return nil
}

func showlog(c *cli.Context) error {
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Load application config
	appConfigs, err := loadRunnerApplications(opc)
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to load runner configs, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Get parameters
	id := c.String("id")
	app := c.String("app")
	isStdout := c.Bool("stdout")
	follow := c.Bool("follow")
	lines := c.Int("lines")
	if id == "" && app == "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require either id or app")
		return cli.NewExitError("", 1)
	} else if id != "" && app != "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require either id or app but not both")
		return cli.NewExitError("", 1)
	}
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	if app != "" {
		appConfig, ok := appConfigs[app]
		if ok {
			app = appConfig.Name
		}
		// Get id by app
		instances, err := r.GetInstancesByName(app)
		if err != nil {
			opc.Logger().LeveledPrintf(log.LevelError, "Failed to get instances by name, error: %s\n", err)
			return cli.NewExitError("", 1)
		}
		if len(instances) == 0 {
			opc.Logger().LeveledPrintln(log.LevelError, "No instance found for this application")
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
				opc.Logger().LeveledPrintln(log.LevelError, "More than 1 instance found, cannot show log by application name")
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
	cmd.Stdout = opc.OutputStream()
	cmd.Stderr = opc.ErrorStream()
	cmd.Run()
	// Done
	return nil
}

func stop(c *cli.Context) error {
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Load application config
	appConfigs, err := loadRunnerApplications(opc)
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to load runner configs, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Get parameters
	ids := c.StringSlice("id")
	apps := c.StringSlice("app")
	clean := c.Bool("clean")
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	// Stop the application
	for _, id := range ids {
		opc.Logger().Printf("Stopping [%s] ...... ", id)
		if err := r.Stop(id, clean); err != nil {
			opc.Logger().LeveledHeadedPrint("", log.LevelError, "Error: %s\n", err)
		} else {
			opc.Logger().LeveledHeadedPrint("", log.LevelSuccess, "Done\n")
		}
	}
	for _, name := range apps {
		// Get the actual name
		appConfig, ok := appConfigs[name]
		if ok {
			name = appConfig.Name
		}
		instances, err := r.GetInstancesByName(name)
		if err != nil {
			opc.Logger().LeveledPrintf(log.LevelError, "Failed to get instances by name [%s], error: %s\n", name, err)
			continue
		}
		for _, instance := range instances {
			s, _ := instance.GetStatus()
			if s != runner.StatusExited {
				opc.Logger().Printf("Stopping [%s] ...... ", instance.ID)
				if err := r.Stop(instance.ID, clean); err != nil {
					opc.Logger().LeveledHeadedPrint("", log.LevelError, "Error: %s\n", err)
				} else {
					opc.Logger().LeveledHeadedPrint("", log.LevelSuccess, "Done\n")
				}
			}
		}
	}
	// Done
	return nil
}

func status(c *cli.Context) error {
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Get parameters
	statusAll := c.Bool("all")
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	// Get the application status
	instances, err := r.List(!statusAll)
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to list instances, error: %s\n", err)
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
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Get parameters
	id := c.String("id")
	app := c.String("app")
	clean := c.Bool("clean")
	if id == "" && app == "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require either id or app")
		return cli.NewExitError("", 1)
	} else if id != "" && app != "" {
		opc.Logger().LeveledPrintln(log.LevelError, "Require either id or app but not both")
		return cli.NewExitError("", 1)
	}
	// Load application config
	appConfigs, err := loadRunnerApplications(opc)
	if err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to load runner configs, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	if app != "" {
		appConfig, ok := appConfigs[app]
		if ok {
			app = appConfig.Name
		}
		// Get id by app
		instances, err := r.GetRunningInstancesByName(app)
		if err != nil {
			opc.Logger().LeveledPrintf(log.LevelError, "Failed to get instances by name, error: %s\n", err)
			return cli.NewExitError("", 1)
		}
		if len(instances) == 0 {
			opc.Logger().LeveledPrintln(log.LevelError, "No instance found for this application")
			return cli.NewExitError("", 1)
		} else if len(instances) > 1 {
			opc.Logger().LeveledPrintln(log.LevelError, "More than 1 instance found, cannot restart by application name")
			return cli.NewExitError("", 1)
		} else {
			id = instances[0].ID
		}
	}
	// Restart the application
	opc.Logger().Printf("Restarting [%s] ...... ", id)
	if instance, err := r.Restart(id, clean); err != nil {
		opc.Logger().LeveledHeadedPrintf("", log.LevelError, "Error: %s\n", err)
	} else {
		opc.Logger().LeveledHeadedPrintf("", log.LevelSuccess, "Done. New Instance ID [%s]\n", instance.ID)
		if !instance.Options.Background {
			instance.Wait()
		}
	}
	// Done
	return nil
}

func clean(c *cli.Context) error {
	opc, err := opcli.NewFromCLI(c)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	// Create the runner
	r := runner.New(opc.WorkDir().GetRunnerPath())
	// Clean
	if err := r.CleanAll(); err != nil {
		opc.Logger().LeveledPrintf(log.LevelError, "Failed to clean, error: %s\n", err)
		return cli.NewExitError("", 1)
	}
	// Done
	return nil
}

func loadRunnerApplications(opc *opcli.OpenlightCLI) (map[string]config.RunnerApplicationConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// First load from global
	// Load from git current
	// Load from current directory
	var configFilenames []string
	for _, name := range config.RunnerConfigNames {
		configFilenames = append(configFilenames, filepath.Join(opc.WorkDir().GetConfigPath(), name))
	}
	if p, err := opcli.GetGitRootFromCurrentDirectory(); err == nil && p != wd {
		for _, name := range config.RunnerConfigNames {
			configFilenames = append(configFilenames, filepath.Join(p, name))
		}
	}
	for _, name := range config.RunnerConfigNames {
		configFilenames = append(configFilenames, filepath.Join(wd, name))
	}
	// Load one by one
	appConfigs := make(map[string]config.RunnerApplicationConfig)
	for _, filename := range configFilenames {
		if _, err := os.Stat(filename); err != nil {
			continue
		}
		c, err := config.LoadRunnerConfigFromFile(filename)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to load runner config file [%s], error: %s", filename, err))
		}
		for name, app := range c.Apps {
			appConfigs[name] = app
		}
	}
	// Check debug
	if opc.Debug() {
		for name, _ := range appConfigs {
			opc.Logger().LeveledPrintf(log.LevelDebug, "Loaded appliation: %s\n", name)
		}
	}
	// Done
	return appConfigs, nil
}
