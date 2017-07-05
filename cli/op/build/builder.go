// Author: lipixun
// File Name: builder.go
// Description:

package build

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"gopkg.in/urfave/cli.v1"

	"github.com/ops-openlight/openlight/pkg/artifact"
	opbuilder "github.com/ops-openlight/openlight/pkg/builder"
	"github.com/ops-openlight/openlight/pkg/repository"
)

// Builder implements build command line functions
type Builder struct{}

// GetCommands returns the commands
func (builder *Builder) GetCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "build",
			Usage:     "Build targets",
			ArgsUsage: "Target names. Use * for all targets otherwise default targets will be used if no target is specified",
			Action:    builder.build,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update,u",
					Usage: "Update the dependency or not",
				},
				cli.BoolFlag{
					Name:  "local-only-ref",
					Usage: "Only resolve reference locally",
				},
				cli.BoolFlag{
					Name:  "no-target",
					Usage: "Do not build target dependencies",
				},
				cli.BoolFlag{
					Name:  "no-pip",
					Usage: "Do not build pip dependencies",
				},
				cli.BoolFlag{
					Name:  "no-go",
					Usage: "Do not build go dependencies",
				},
				cli.BoolFlag{
					Name:  "no-go-install",
					Usage: "Do not install go binary",
				},
				cli.BoolFlag{
					Name:  "ignore-go-install-error",
					Usage: "Ignore go install error",
				},
			},
		},
		{
			Name:      "build-dep",
			Usage:     "Build or install dependencies",
			ArgsUsage: "Target names. Use * for all targets otherwise default targets will be used if no target is specified",
			Action:    builder.buildDeps,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "update,u",
					Usage: "Update the dependency or not",
				},
				cli.BoolFlag{
					Name:  "no-target",
					Usage: "Do not build target dependencies",
				},
				cli.BoolFlag{
					Name:  "no-pip",
					Usage: "Do not build pip dependencies",
				},
				cli.BoolFlag{
					Name:  "no-go",
					Usage: "Do not build go dependencies",
				},
			},
		},
		{
			Name:  "build-show",
			Usage: "Show build info",
			Subcommands: []cli.Command{
				{
					Name:      "ref",
					Usage:     "Show references info",
					ArgsUsage: "Target names. Use * for all targets otherwise default targets will be used if no target is specified",
					Action:    builder.showRefs,
				},
				{
					Name:   "targets",
					Usage:  "Show targets info",
					Action: builder.showTargets,
				},
				{
					Name:      "deps",
					Usage:     "Show dependency",
					ArgsUsage: "Target names. Use * for all targets otherwise default targets will be used if no target is specified",
					Action:    builder.showDeps,
				},
			},
		},
	}
}

func (builder *Builder) init(c *cli.Context) (*repository.LocalRepository, error) {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
	// Get local repository
	dirname, err := os.Getwd()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to get work directory: %v", err), 1)
	}
	repo, err := repository.NewLocalRepository(dirname)
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to load local repository: %v", err), 1)
	}
	return repo, nil
}

func (builder *Builder) getTargets(c *cli.Context, repo *repository.LocalRepository) ([]*repository.Target, error) {
	pkg, err := repo.RootPackage()
	if err != nil {
		return nil, cli.NewExitError(fmt.Sprintf("Failed to get root package: %v", err), 1)
	}
	var targets []*repository.Target
	if c.NArg() == 0 {
		// Default targets
		for _, name := range pkg.Spec().GetOptions().GetDefaultTargets() {
			target := pkg.GetTarget(name)
			if target == nil {
				return nil, cli.NewExitError(fmt.Sprintf("Target [%v] not found", name), 1)
			}
			targets = append(targets, target)
		}
	} else {
		var names []string
		var all bool
		for i := 0; i < c.NArg(); i++ {
			name := c.Args().Get(i)
			if name == "*" {
				all = true
				break
			}
			names = append(names, name)
		}
		if all {
			targets = pkg.Targets()
		} else {
			for _, name := range names {
				target := pkg.GetTarget(name)
				if target == nil {
					return nil, cli.NewExitError(fmt.Sprintf("Target [%v] not found", name), 1)
				}
				targets = append(targets, target)
			}
		}
	}
	return targets, nil
}

func (builder *Builder) build(c *cli.Context) error {
	repo, err := builder.init(c)
	if err != nil {
		return err
	}
	// Create new builder
	b, err := opbuilder.NewBuilder(repo, c.GlobalBool("verbose"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create builder: %v", err), 1)
	}
	// Create options
	var options []opbuilder.BuildTargetOption
	if c.Bool("update") {
		options = append(options, opbuilder.WithUpdateDependencyOption())
	}
	if c.Bool("no-target") {
		options = append(options, opbuilder.WithDonotBuildTargetDependencyOption())
	}
	if c.Bool("no-pip") {
		options = append(options, opbuilder.WithDonotBuildPipDependencyOption())
	}
	if c.Bool("no-go") {
		options = append(options, opbuilder.WithDonotBuildGoDependencyOption())
	}
	if c.Bool("no-go-install") {
		options = append(options, opbuilder.WithDonotInstallGoBinaryOption())
	}
	if c.Bool("ignore-go-install-error") {
		options = append(options, opbuilder.WithIgnoreInstallGoBinaryErrorOption())
	}
	// Write tag
	log.Infoln("Build Tag:", color.YellowString("%v", b.Tag()))
	// Build deps
	targets, err := builder.getTargets(c, repo)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := b.BuildTarget(target.Name(), options...); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to build target: %v", err), 1)
		}
	}
	// Write all artifacts
	var files, dockerImages []artifact.Artifact
	for _, result := range b.GetBuildResults() {
		if result.Artifact() == nil {
			continue
		}
		switch result.Artifact().GetType() {
		case artifact.ArtifactTypeFile:
			files = append(files, result.Artifact())
		case artifact.ArtifactTypeDockerImage:
			dockerImages = append(dockerImages, result.Artifact())
		}
	}
	if len(files) > 0 {
		log.Infoln("Generated files:")
		for _, f := range files {
			log.Info("\t", color.GreenString("%v", f.String()))
		}
	}
	if len(dockerImages) > 0 {
		log.Infoln("Generated docker images:")
		for _, image := range dockerImages {
			log.Info("\t", color.GreenString("%v", image.String()))
		}
	}
	// Done
	return nil
}

func (builder *Builder) buildDeps(c *cli.Context) error {
	repo, err := builder.init(c)
	if err != nil {
		return err
	}
	// Create new builder
	b, err := opbuilder.NewBuilder(repo, c.GlobalBool("verbose"))
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to create builder: %v", err), 1)
	}
	// Create options
	var options []opbuilder.BuildTargetOption
	if c.Bool("update") {
		options = append(options, opbuilder.WithUpdateDependencyOption())
	}
	if c.Bool("no-target") {
		options = append(options, opbuilder.WithDonotBuildTargetDependencyOption())
	}
	if c.Bool("no-pip") {
		options = append(options, opbuilder.WithDonotBuildPipDependencyOption())
	}
	if c.Bool("no-go") {
		options = append(options, opbuilder.WithDonotBuildGoDependencyOption())
	}
	// Build deps
	targets, err := builder.getTargets(c, repo)
	if err != nil {
		return err
	}
	for _, target := range targets {
		if err := b.BuildTargetDependencies(target.Name(), options...); err != nil {
			return cli.NewExitError(fmt.Sprintf("Failed to build target dependency: %v", err), 1)
		}
	}
	// Done
	return nil
}

func (builder *Builder) showRefs(c *cli.Context) error {
	repo, err := builder.init(c)
	if err != nil {
		return err
	}
	pkg, err := repo.RootPackage()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get root package: %v", err), 1)
	}
	for _, ref := range pkg.References() {
		fmt.Printf("%v: Remote [%v] Path [%v]\n",
			ref.Name(),
			ref.Spec().GetRemote(),
			ref.Spec().GetPath(),
		)
		if len(ref.Spec().GetFinders()) > 0 {
			fmt.Println("\tFinder:")
			for _, finderSpec := range ref.Spec().GetFinders() {
				if pythonFinderSpec := finderSpec.GetPython(); pythonFinderSpec != nil {
					fmt.Printf("\t\t%v: Python -> Module [%v] Parent [%v]\n", finderSpec.Name, pythonFinderSpec.Module, pythonFinderSpec.Parent)
				} else if goFinderSpec := finderSpec.GetGo(); goFinderSpec != nil {
					fmt.Printf("\t\t%v: GO -> Package [%v] Parent [%v] Root [%v]\n", finderSpec.Name, goFinderSpec.Package, goFinderSpec.Parent, goFinderSpec.Root)
				} else {
					fmt.Printf("\t\t%v: Invalid finder\n", finderSpec.Name)
				}
			}
		}
	}
	return nil
}

func (builder *Builder) showTargets(c *cli.Context) error {
	repo, err := builder.init(c)
	if err != nil {
		return err
	}
	pkg, err := repo.RootPackage()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get root package: %v", err), 1)
	}
	fmt.Printf("%-20v %v\n========================================================\n", "Name", "Description")
	for _, target := range pkg.Targets() {
		fmt.Printf("%-20v %v\n", target.Name(), target.Spec().GetDescription())
	}
	return nil
}

func (builder *Builder) showDeps(c *cli.Context) error {
	repo, err := builder.init(c)
	if err != nil {
		return err
	}
	pkg, err := repo.RootPackage()
	if err != nil {
		return cli.NewExitError(fmt.Sprintf("Failed to get root package: %v", err), 1)
	}
	// Get targets
	if c.NArg() > 0 {
		for i := 0; i < c.NArg(); i++ {
			name := c.Args().Get(i)
			fmt.Println(name)
			target := pkg.GetTarget(name)
			if target == nil {
				fmt.Println("\tTarget not found")
				continue
			}
			builder.printDeps(target)
		}
		return nil
	}
	// Print all targets
	for _, target := range pkg.Targets() {
		fmt.Println(target.Name())
		builder.printDeps(target)
	}
	return nil
}

func (builder *Builder) printDeps(target *repository.Target) {
	for _, dep := range target.Spec().Dependencies {
		if targetDep := dep.GetTarget(); targetDep != nil {
			fmt.Printf("\tTarget: Reference [%v] Path [%v] Target [%v] Build [%v]\n",
				targetDep.Reference,
				targetDep.Path,
				targetDep.Target,
				targetDep.Build,
			)
		} else if pipDep := dep.GetPip(); pipDep != nil {
			fmt.Printf("\tPip: Module [%v]\n", pipDep.Module)
		} else if goDep := dep.GetGo(); goDep != nil {
			fmt.Printf("\tGo: Package [%v]\n", goDep.Package)
		} else {
			fmt.Println("\tInvalid dependency")
		}
	}
}
