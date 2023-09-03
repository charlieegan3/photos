package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base for the Photos application
var rootCmd = &cobra.Command{
	Use:   "photos",
	Short: "Photos is a web app to manage a photo library",
	Long: `Photos has been built to replace a photo library system that once was
	populated using Instagram data.

	It is designed to offer that functionality without the social features and
	a few minor features in addition to the core Instagram feature set.`,
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
		"config file (default is $HOME/.photos.yaml)",
	)
}

// initConfig reads in config file and errors if fails
func initConfig() {
	// if cfgFile is not set by flag or CONFIG_STRING logic, then use the HOME dir default
	if cfgFile == "" {
		cfgFile = "$HOME/.photos.yaml"
	}

	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed using config file %q: %s", viper.ConfigFileUsed(), err)
		return
	}

	log.Printf("Using config file: %s", viper.ConfigFileUsed())
}
