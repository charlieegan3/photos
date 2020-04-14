package cmd

import "github.com/charlieegan3/photos/internal/pkg/sync"

func init() {
	rootCmd.AddCommand(sync.CreateSyncCmd())
}
