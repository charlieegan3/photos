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
	// if there is a CONFIG_STRING var, then dump that to the config file location
	if configString := os.Getenv("CONFIG_STRING"); configString != "" {
		// CONFIG_STRING is used in serverless environments where /tmp is generally the only writable location
		cfgFile = "/tmp/secrets/config.yaml"
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

	if viper.GetString("google.service_account_key") != "" {
		// place google credentials on disk
		tmpfile, err := ioutil.TempFile("", "google.*.json")
		if err != nil {
			log.Fatal(err)
		}
		content := []byte(viper.GetString("google.service_account_key"))
		if _, err := tmpfile.Write(content); err != nil {
			tmpfile.Close()
			log.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			log.Fatal(err)
		}

		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpfile.Name())
	}
}
