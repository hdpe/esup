package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "esup",
	Short: "esup is a schema migration tool for Elasticsearch",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if _, e := fmt.Fprintln(os.Stderr, err); e != nil {
			panic(e)
		}
		os.Exit(1)
	}
}

func fatalError(format string, a ...interface{}) {
	println(fmt.Sprintf(format, a...))
	os.Exit(1)
}
