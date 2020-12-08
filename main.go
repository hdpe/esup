package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "migrate" {
		usage()
	}

	envName := os.Args[2]

	config := newConfig()

	es, err := newES(config.server)

	if err != nil {
		fatalError("couldn't create elasticsearch client: %v", err)
	}

	schema, err := getSchema(config, envName)

	if err != nil {
		fatalError("couldn't get schema: %v", err)
	}

	changelog := &Changelog{
		config: config.changelog,
		es:     es,
	}

	plan, err := makePlan(es, config.preprocess, changelog, schema, envName)

	if err != nil {
		fatalError("couldn't plan update: %v", err)
	}

	logPlan(plan)

	if len(plan) == 0 {
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nConfirm [Y/n]: ")
	text, _ := reader.ReadString('\n')

	if strings.ToLower(text) != "y\n" {
		println("Cancelled")
		os.Exit(0)
	}

	for _, item := range plan {
		if err = item.execute(); err != nil {
			fatalError("couldn't execute %v: %v", item, err)
		}
	}

	println("Complete")
}

func usage() {
	println(fmt.Sprintf("Usage: %v migrate ENVIRONMENT", os.Args[0]))
	os.Exit(0)
}

func logPlan(plan []planAction) {
	if len(plan) == 0 {
		println("No changes")
		return
	}

	println("Planned changes:")

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
