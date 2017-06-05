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

-- You can include multiple modules. All instructions within these modules will be treated as in this file.

-- Declare the package info and get the package object
-- NOTE: You cannot create package object and `build.package` must be only called once
pkg = build.package({
    name = "Package name used for display",
    comment = "Some comment text may be useful...",
    author = "Author info",
    url = "Url of the package",
    })

-- Set options
pkg:options():set("defaultTargets", { "all" })

-- Define a reference. All dependent packages must define references.
-- Format: table with keys:
--  * remote (git address), required
--  * path, optional. Use this attribute to change the root the a repository
--    NOTE: Path is rarely used
ref = build.Reference.new({ remote = "github.com/ops-openlight/somepackage" })

-- Add reference
-- Format: table with keys:
--  * name
-- * Reference (object)
pkg:references():add("somepackage", ref)
-- You can get the reference in this way:
ref = pkg:references():get("somepackage")

-- Create finders for references
finder = build.PythonFinder.new({
    module = "module1",
    -- Use the parent directory (1 means use the directory of the python package directory. If set to 2, means use the directory of the directory of the python package directory)
    parent = 1,
    })
ref:finders():add("finder1", finder)

ref:finders():add("finder2", build.GoFinder.new({ package = "testgo" }))

-- Define a command target
-- Format: table with keys
--  * command
--  * args, optional
--  * workdir, optional
cmd = build.CommandTarget.new({ command = "make", args = { "build" }, workdir = "./build" })

-- Add it
pkg:targets():add("cmd", cmd)

-- Define a go binary target
-- Format: table with keys:
--  * package, go package (to build)
--  * output (name)
--  * goVersion, optional
cli = build.GoBinaryTarget.new({ package = "github.com/ops-openlight/cli/op", output = "op" })

-- Add it
pkg:targets():add("cli", cli)

-- Define the target dependency
-- Format: table with keys
--  * reference, optional
--  * path, optional
--  * target
--  * build, optional
cli:dependent(build.TargetDependency.new({
    reference = "somepackage",
    target = "sometarget",
    -- Build the target or just prepare the dependency
    build = true
    }))

-- Define a go dependency which will be resolved by go dependency rule (not by openlight target dependency rule).
-- By this approach, openlight build tool could be easily integrated with existing projects (which are not built by openlight)
-- You specify unlimited packages in one command, and also use multiple dependentOnGo command is ok as well.
cli:dependent(build.GoDependency.new({ package = "github.com/Sirupsen/logrus" }))

-- Define a python lib target
-- Python lib target requires python setup.py to build the package. Openlight will run setup.py sdist to build the python setup package.
-- Format: table with keys
--  * workdir, optional
--  * setup, optional
--  * output, optional
pylib = build.PythonLibTarget.new({ workdir = "./", setup = "setup.py", output = "pylib.tgz" })

-- Add it
pkg:targets():add("pylib", pylib)

-- Call dependent without arguments is ok
pylib:dependent()

-- Define a internal dependency
pylib:dependent(
    build.TargetDependency.new({ target = "cli" }),
    build.TargetDependency.new({ reference = "github.com/ops-openlight/somepackage", target = "some target" })
    )

-- Define a python pip dependency
pylib:dependent(build.PipDependency.new({ module = "module could be resolved by pip" }))

-- Define a docker image target
-- Format: Table with keys:
--  * repository
--  * image
docker = build.DockerImageTarget.new({ repository = "some-repository", image = "dockerimagename" })
docker:dependent(build.TargetDependency.new({
    target = "sometarget",
    -- Build the target or just prepare the dependency
    build = true
    }))

-- Add it
pkg:targets():add("image1", docker)

-- Docker: FROM
-- Format: base image
-- NOTE: From must come first before any other commands
docker:from("ubuntu:16.04")

-- Docker: LABEL
-- Format: key, value
docker:label("key", "value")

-- Docker: ADD
-- Format: source object, target path
docker:add(build.File.new({ reference = "some-reference", filenames = { "filename" }}), "/bin/cli")
-- File could also be used in this way:
docker:add(build.File.new({ filenames = { "filename" }}), "/bin/cli")

-- Docker: COPY
-- Format: source object, target path
docker:copy(build.Artifact.new({ reference = "some-reference", target = "target name", filenames = { "filename" }}), "/bin/cli")
-- Artifact could also be used in this way:
docker:copy(build.Artifact.new({ target = "target name" }), "/bin/cli")

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
all = build.Target.new()
all:dependent(
    build.TargetDependency.new({ target = "cmd", build = true }),
    build.TargetDependency.new({ target = "cli", build = true }),
    build.TargetDependency.new({ target = "pylib", build = true }),
    build.TargetDependency.new({ target = "image1", build = true })
    )
pkg:targets():add("all", all)
