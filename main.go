package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	cmd := parseCmd(os.Args)

	if !cmd.valid {
		println(cmd.usage())
		os.Exit(0)
	}

	config, err := newConfig()

	if err != nil {
		fatalError("couldn't read configuration: %v", err)
	}

	es, err := newES(config.server)

	if err != nil {
		fatalError("couldn't create elasticsearch client: %v", err)
	}

	schema, err := getSchema(config, cmd.envName)

	if err != nil {
		fatalError("couldn't get schema: %v", err)
	}

	changelog := &Changelog{
		config: config.changelog,
		es:     es,
	}

	planner := newPlanner(es, config, changelog, schema, cmd.envName)
	plan, err := planner.Plan()

	if err != nil {
		fatalError("couldn't plan update: %v", err)
	}

	logPlan(plan, config.server)

	if len(plan) == 0 {
		os.Exit(0)
	}

	if !cmd.approve {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nConfirm [Y/n]: ")
		text, _ := reader.ReadString('\n')

		if strings.ToLower(text) != "y\n" {
			println("Cancelled")
			os.Exit(0)
		}
	}

	for _, item := range plan {
		if err = item.execute(); err != nil {
			fatalError("couldn't execute %v: %v", item, err)
		}
	}

	println("Complete")
}

func logPlan(plan []planAction, serverConfig ServerConfig) {
	if len(plan) == 0 {
		println("No changes")
		return
	}

	println(fmt.Sprintf("Planned changes on %s:", serverConfig.address))

	msg := ""

	for _, item := range plan {
		msg += fmt.Sprintf(" - %v\n", item)
	}

	print(msg)
}

func fatalError(format string, a ...interface{}) {
	println(fmt.Sprintf(format, a...))
	os.Exit(1)
}
