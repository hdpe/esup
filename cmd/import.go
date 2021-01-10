package cmd

import (
	"fmt"
	"github.com/hdpe.me/esup/imp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import RESOURCE_TYPE RESOURCE_IDENTIFIER ENVIRONMENT",
	Short: "Import an existing Elasticsearch resource into the esup changelog",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(3)(cmd, args); err != nil {
			return err
		}
		if err := validateResourceType(args[0]); err != nil {
			return err
		}
		return validateEnv(args[2])
	},
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		resourceIdentifier := args[1]
		envName := args[2]

		ctx := newContext(envName)

		i := imp.NewImporter(ctx.Changelog, ctx.Schema, ctx.Proc)

		err := i.ImportResource(resourceType, resourceIdentifier)

		if err != nil {
			fatalError("couldn't import resource: %v", err)
		}
	},
}

func validateResourceType(rt string) error {
	if rt != "index_set" && rt != "document" {
		return fmt.Errorf("invalid resource type for changelog: %q", rt)
	}
	return nil
}
