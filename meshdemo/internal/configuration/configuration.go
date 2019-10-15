/*
Copyright © 2019 Tony Pujals <tpujals@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Package config is a utility wrapper over cobra and viper to keep streamline the logic inside the CLI implementation.
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
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// ignore expected error if user has not yet created a config
		default:
			io.Error("error reading %s:\n%s", viper.ConfigFileUsed(), err)
		}
	}
}

func ConfigFileUsed() string {
	return viper.ConfigFileUsed()
}
