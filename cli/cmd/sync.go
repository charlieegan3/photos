package cmd

import (
	"github.com/charlieegan3/photos/internal/pkg/media"
	"github.com/charlieegan3/photos/internal/pkg/sync"
	"github.com/spf13/cobra"
)

func init() {
	syncCmd := cobra.Command{
		Use:   "sync",
		Short: "Refreshes and saves profile data",
	}

	syncCmd.AddCommand(sync.CreateSyncCmd())
	syncCmd.AddCommand(media.CreateSyncCmd())

	rootCmd.AddCommand(&syncCmd)
}
