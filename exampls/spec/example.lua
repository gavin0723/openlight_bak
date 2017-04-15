-- The example build spec written in lua.
-- The build file should be named `.BUILD` when used as default build spec file.
--
--  For (a) build spec (files) in a repository, package must be defined and only be defined once
--
--  Targets could be defined any times but name could not be duplicated
--
--  So, it's obviously that target name and file name / path could not be the same. If so, will use target.
--
--  Variables are a string - string map which could be used in many options. This feature lets user could define build options according to dynamic infos (Such as version, build time, etc..)
--  All avaiable variables will be listed in the document
--

-- Require
build = require("build")

-- Define the package
-- Format: name, remote uri, options. options is optional.
pkg = build.Package.new("github.com/ops-openlight/openlight", "github.com/ops-openlight/openlight", {
    -- Define the package options
    defaultTargets = { "oldcli" },
})

-- Define the package options. This will overwrite the options defined in package.
pkg:options().defaultTargets = { "cli" }

-- OR: To replace all options
pkg:options({ defaultTargets = { "all" } })

-- Add the package to spec. Specify multiple packages is supported
build.addPackage(pkg)

-- You can get the package back by the following line
pkg = build.getPackage("github.com/ops-openlight/openlight")

-- You can also delete a package from spec. Specify multiple names is supported
-- build.deletePackage("github.com/ops-openlight/openlight")

-- Define a reference. All dependent packages must define references.
-- Format: name, remote uri, options. options is optional
pkg:reference("github.com/ops-openlight/somepackage", "github.com/ops-openlight/somepackage")
-- You can get the reference in this way:
ref = pkg:reference("github.com/ops-openlight/somepackage")

-- Declare local finders for references
ref:pythonLocalFinder("finder1", "somepkg", {
    -- Use the parent directory (1 means use the directory of the python package directory. If set to 2, means use the directory of the directory of the python package directory)
    parent = 1,
    })
-- You can also set local finder by ref variable and set options of the finder
finder = ref:localFinder("finder1")
finder:options().parent = 2

-- Define a command target
pkg:command("cmd", "make", "build", {
    -- Set options
    workdir = "./build",
    output = "./build/output",
})

-- Define a go binary target
-- Format: name, go package (to build), options. options is optional
cli = pkg:goBinary("cli", "github.com/ops-openlight/cli/op", {
    -- Options
    -- Output binary name
    output = "op",
})

-- We can also define the option values in this way. The name `cli` in the following code is the target name.
pkg:target("cli"):options().output = "op2"
-- We can also reference the target by the return value of `goBinary`
cli:options().goVersion = "1.7"

-- Define the target dependency
-- Format: package, target, options. options is optional
--          OR
--         target, options. options is optional.
cli:dependent(build.TargetDependency.new("github.com/ops-openlight/somepackage", "sometarget", {
    -- Build the target or just prepare the dependency
    build = true
}))

-- Define a go dependency which will be resolved by go dependency rule (not by openlight target dependency rule).
-- By this approach, openlight build tool could be easily integrated with existing projects (which are not built by openlight)
-- You specify unlimited packages in one command, and also use multiple dependentOnGo command is ok as well.
cli:dependent(build.GoDependency.new("github.com/Sirupsen/logrus", "more", "..."))

-- Define a python lib target
-- Name: pypkg
-- Python lib target requires python setup.py to build the package. Openlight will run setup.py sdist to build the python setup package.
pypkg = pkg:pythonLib("pypkg", {
    -- Work dir
    workdir = "./"
})

-- Call dependent without arguments is ok
pypkg:dependent()

-- Define a internal dependency
pypkg:dependent(
    build.TargetDependency.new("cli"),
    build.TargetDependency.new("github.com/ops-openlight/somepackage", "some target")
    )

-- Define a python pip dependency
pypkg:dependent(build.PipDependency.new("module could be resolved by pip", "more", "..."))

-- Define the target options

-- The setup file path (the relative path to repository root)
pypkg:options().setupPath = "python/setup.py"
-- The output package name. Could use variables string.
pypkg:options().output = "pypkg.tgz"

-- Define a docker image target
-- Format: name, repository, image name, options. options is optional
docker = pkg:dockerImage("image1", "some-repository", "dockerimagename")

-- Docker: FROM
-- Format: base image
-- NOTE: From must come first before any other commands
docker:from("ubuntu:16.04")

-- Docker: LABEL
-- Format: key, value
docker:label("key", "value")

-- Docker: ADD
-- Format: source object, target path
docker:add(build.File.new("some-repository", "filename"), "/bin/cli")
-- File could also be used in this way:
docker:add(build.File.new("filename"), "/bin/cli")

-- Docker: COPY
-- Format: source object, target path
pkg:target("image1"):copy(build.Artifact.new("some-repository", "target name"), "/bin/cli")
-- Artifact could also be used in this way:
pkg:target("image1"):copy(build.Artifact.new("target name"), "/bin/cli")

-- Docker: RUN
-- Format: command
docker:run("pip install somepkg")

-- Docker: ENTRYPOINT
-- Format: args, args, ...
docker:entrypoint("/bin/cli", "--start")

-- Docker: EXPOSE
-- Format: port, port, ...
docker:expose(1234, 5678, 9012)

-- Docker: VOLUME
-- Format: path, path, ...
docker:volume("/data/logs")

-- Docker: USER
-- Format: username
docker:user("www")

-- Docker: WORKDIR
-- Format: path
docker:workdir("/")

-- Docker: ENV
-- Format: key, value
docker:env("key", "value")

-- Define a general target
pkg:target("all"):dependent(
    build.TargetDependency.new("cmd", { build = true }),
    build.TargetDependency.new("cli", { build = true }),
    build.TargetDependency.new("pypkg", { build = true }),
    build.TargetDependency.new("image1", { build = true })
    )
