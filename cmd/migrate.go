package cmd

import (
	"bufio"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/plan"
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

		ctx := newContext(envName)

		planner := plan.NewPlanner(ctx.es, ctx.conf, ctx.changelog, ctx.schema, ctx.proc)
		resPlan, err := planner.Plan()

		if err != nil {
			fatalError("couldn't plan update: %v", err)
		}

		logPlan(resPlan, ctx.conf.Server)

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
