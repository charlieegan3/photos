package cmd

import (
	"github.com/spf13/cobra"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Commands for running adhoc or recurring jobs",
}

func init() {
	rootCmd.AddCommand(jobsCmd)
}
