package main

import (
	"bufio"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/plan"
	"github.com/hdpe.me/esup/schema"
	"os"
	"strings"
)

func main() {
	cmd := parseCmd(os.Args)

	if !cmd.valid {
		println(cmd.usage())
		os.Exit(0)
	}

	conf, err := config.NewConfig()

	if err != nil {
		fatalError("couldn't read configuration: %v", err)
	}

	esClient, err := es.NewClient(conf.Server)

	if err != nil {
		fatalError("couldn't create elasticsearch client: %v", err)
	}

	resSchema, err := schema.GetSchema(conf, cmd.envName)

	if err != nil {
		fatalError("couldn't get schema: %v", err)
	}

	changelog := plan.NewChangelog(conf.Changelog, esClient)

	planner := plan.NewPlanner(esClient, conf, changelog, resSchema, cmd.envName)
	resPlan, err := planner.Plan()

	if err != nil {
		fatalError("couldn't plan update: %v", err)
	}

	logPlan(resPlan, conf.Server)

	if len(resPlan) == 0 {
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

	for _, item := range resPlan {
		if err = item.Execute(); err != nil {
			fatalError("couldn't execute %v: %v", item, err)
		}
	}

	println("Complete")
}

func logPlan(plan []plan.PlanAction, serverConfig config.ServerConfig) {
	if len(plan) == 0 {
		println("No changes")
		return
	}

	println(fmt.Sprintf("Planned changes on %s:\n", serverConfig.Address))

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
