package cmd

import (
	"bufio"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/es"
	"github.com/hdpe.me/esup/plan"
	"github.com/hdpe.me/esup/schema"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
)

var approve bool

func init() {
	migrateCmd.Flags().BoolVarP(&approve, "approve", "a", false, "approve this migration without prompting")
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate ENVIRONMENT",
	Short: "Migrate an esup schema",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		return validateEnv(args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		envName := args[0]

		conf, err := config.NewConfig()

		if err != nil {
			fatalError("couldn't read configuration: %v", err)
		}

		esClient, err := es.NewClient(conf.Server)

		if err != nil {
			fatalError("couldn't create elasticsearch client: %v", err)
		}

		resSchema, err := schema.GetSchema(conf, envName)

		if err != nil {
			fatalError("couldn't get schema: %v", err)
		}

		changelog := plan.NewChangelog(conf.Changelog, esClient)

		planner := plan.NewPlanner(esClient, conf, changelog, resSchema, envName)
		resPlan, err := planner.Plan()

		if err != nil {
			fatalError("couldn't plan update: %v", err)
		}

		logPlan(resPlan, conf.Server)

		if len(resPlan) == 0 {
			os.Exit(0)
		}

		if !approve {
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
	},
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

func validateEnv(str string) error {
	p := `^[\pLl\pN][\pLl\pN\-_.]*$`
	if ok := regexp.MustCompile(p).MatchString(str); !ok {
		return fmt.Errorf("wanted string matching %v", p)
	}
	return nil
}

func fatalError(format string, a ...interface{}) {
	println(fmt.Sprintf(format, a...))
	os.Exit(1)
}
