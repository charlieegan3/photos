package cmd

import (
	"github.com/charlieegan3/photos/internal/pkg/data"
	"github.com/charlieegan3/photos/internal/pkg/locations"
	"github.com/charlieegan3/photos/internal/pkg/media"
	"github.com/spf13/cobra"
)

func init() {
	syncCmd := cobra.Command{
		Use:   "sync",
		Short: "Refreshes and saves profile data",
	}

	syncCmd.AddCommand(data.CreateSyncCmd())
	syncCmd.AddCommand(media.CreateSyncCmd())
	syncCmd.AddCommand(locations.CreateSyncCmd())

	rootCmd.AddCommand(&syncCmd)
}
