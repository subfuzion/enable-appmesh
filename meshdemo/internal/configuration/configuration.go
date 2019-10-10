/*
 Package config is a utility wrapper over cobra and viper to keep streamline the logic inside the CLI implementation.
 */
package configuration

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/subfuzion/meshdemo/pkg/io"
)

var ConfigFile string

func Init() {
	cobra.OnInitialize(initConfig)
}

// Init reads in config file and ENV variables if set.
func initConfig() {
	if ConfigFile != "" {
		// Use config file from the --config option.
		viper.SetConfigFile(ConfigFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			io.Fatal(1, err)
		}

		// Search config in home directory with name ".colorapp" (with any supported extension, e.g., .yaml, etc)
		viper.AddConfigPath(home)
		viper.SetConfigName(".colorapp")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		io.Error("error reading %s:\n%s", viper.ConfigFileUsed(), err)
	}
}

func ConfigFileUsed() string {
	return viper.ConfigFileUsed()
}
