package cmd

import (
	"encoding/base64"
	"io/ioutil"
	"log"
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
}

// initConfig reads in config file and errors if fails
func initConfig() {
	if cfgFile == "" {
		cfgFile = "$HOME/.cms.yaml"
	}

	// if there is a CONFIG_STRING var, then dump that to the config file location
	if configString := os.Getenv("CONFIG_STRING"); configString != "" {
		yamlConfig, err := base64.StdEncoding.DecodeString(configString)
		if err != nil {
			log.Fatalf("Failed to decode CONFIG_STRING: %s", err)
			return
		}
		err = ioutil.WriteFile(cfgFile, yamlConfig, 0644)
		if err != nil {
			log.Fatalf("Failed writing CONFIG_STRING: %s", err)
			return
		}
	}

	viper.SetConfigFile(cfgFile)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed using config file %q: %s", viper.ConfigFileUsed(), err)
		return
	}
	log.Printf("Using config file: %s", viper.ConfigFileUsed())
}
