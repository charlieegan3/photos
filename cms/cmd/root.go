package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base for the CMS application
var rootCmd = &cobra.Command{
	Use:   "cms",
	Short: "CMS is a web app to manage a photo library",
	Long: `CMS has been built to replace a photo library system that once was
	populated using Instagram data.

	It is designed to offer that functionality without the social features and
	a few minor features in addition to the core Instagram featureset.`,
}

// Execute adds child commands to the root command and sets flags
// appropriately. This is called by main.main().
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// config is required and persistent so available to all commands
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"config file (default is $HOME/.cms.yaml)",
	)
	rootCmd.MarkPersistentFlagRequired("config")
}

// initConfig reads in config file and errors if fails
func initConfig() {
	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed using config file:", viper.ConfigFileUsed())
		return
	}
	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
}
