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
runner = require("runner")

-- Create a command
-- format: Table with keys:
--  * name
--  * comment
--  * command
--  * args
--  * workdir
--  * envs
cmd = runner.Command.new({
    name = "test",
    comment = "This is a test run command",
    command = "python",
    args = {"-m", "test"},
    workdir = "./test/",
    envs = { "TEST1=VALUE1", "TEST2=VALUE2" },
})

cmds = runner.commands():add("cmd1", cmd)
