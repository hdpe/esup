package cmd

import (
	"bufio"
	"fmt"
	"github.com/hdpe.me/esup/config"
	"github.com/hdpe.me/esup/plan"
	"github.com/hdpe.me/esup/util"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		envName := args[0]

		ctx := newContext(envName)

		planner := plan.NewPlanner(ctx.Es, ctx.Conf, ctx.Changelog, ctx.Schema, ctx.Proc, &util.DefaultClock{})

		getLock(ctx, envName)
		defer releaseLock(ctx, envName)

		resPlan, err := planner.Plan()

		if err != nil {
			return fmt.Errorf("couldn't plan update: %w", err)
		}

		logPlan(resPlan, ctx.Conf.Server)

		if !approve {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("\nConfirm [Y/n]: ")
			text, _ := reader.ReadString('\n')

			if strings.ToLower(text) != "y\n" {
				println("Cancelled")
				return nil
			}
		}

		coll := plan.NewCollector()

		for _, item := range resPlan {
			if err = item.Execute(ctx.Es, ctx.Changelog, coll); err != nil {
				return fmt.Errorf("couldn't execute %v: %v", item, err)
			}
		}

		println("Complete")
		return nil
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
