package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appName     = "satisfaction-survey"
	appLongName = "Simple Satisfaction Survey"
)

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appLongName,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
