package main

import (
	"fmt"
	"os"
	"regexp"
)

var cmds = map[string]func(string, []string) cmdline{
	"migrate": parseMigrateCmd,
}

func parseCmd(args []string) cmdline {
	if len(args) < 2 {
		return invalidCmdLine(nil)
	}

	cfunc, ok := cmds[args[1]]

	if !ok {
		return invalidCmdLine(nil)
	}

	return cfunc(args[1], args[2:])
}

func parseMigrateCmd(name string, args []string) cmdline {
	if len(args) == 0 {
		return invalidCmdLine(nil)
	}

	envName := args[0]

	if err := validateEnv(envName); err != nil {
		return invalidCmdLine(fmt.Errorf("ENVIRONMENT should be valid Elasticsearch resource prefix: %w", err))
	}

	approve := false

	if len(args) == 2 {
		if args[1] == "-approve" {
			approve = true
		} else {
			return invalidCmdLine(nil)
		}
	}

	return validCmdline(name, envName, approve)
}

func validateEnv(str string) error {
	p := `^[\pLl\pN][\pLl\pN\-_.]*$`
	if ok := regexp.MustCompile(p).MatchString(str); !ok {
		return fmt.Errorf("wanted string matching %v", p)
	}
	return nil
}

type cmdline struct {
	valid   bool
	err     error
	cmd     string
	envName string
	approve bool
}

func (c *cmdline) usage() string {
	msg := fmt.Sprintf("Usage: %v %v ENVIRONMENT [-approve]", os.Args[0], c.cmd)
	if c.err != nil {
		msg = fmt.Sprintf("%v\n\n%v", msg, c.err)
	}
	return msg
}

func validCmdline(cmd string, envName string, approve bool) cmdline {
	return cmdline{
		valid:   true,
		err:     nil,
		cmd:     cmd,
		envName: envName,
		approve: approve,
	}
}

func invalidCmdLine(err error) cmdline {
	return cmdline{
		valid:   false,
		err:     err,
		cmd:     "",
		envName: "",
		approve: false,
	}
}
