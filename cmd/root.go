/*
Copyright © 2020 Philipp Galichkin <phil.gal@outlook.com>

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
package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	dataFile string
)

const (
	dataFileHeader         = "id,date,activity,hours,jira,category"
	defaultDateTimePattern = "02-01-2006 15:04"
)

var rec = logRecord{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jtl",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cobra.OnInitialize(initDataFile)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jtl/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&dataFile, "data", "", "data file (default is $HOME/data/<month-year>.<ext>)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".jtl" (without extension).
		viper.AddConfigPath(path.Join(home, ".jtl"))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Config file not loaded", err)
	}

	viper.SetDefault("DateTimePattern", defaultDateTimePattern)
}

func initDataFile() {
	dataDir := path.Join(homeDir(), ".jtl", "data")
	f := path.Join(dataDir, dataFileName())
	createDirIfNotExists(dataDir)
	createFileIfNotExists(f, dataFileHeader)
	//upgrade file to full path in context
	dataFile = f
}

func dataFileName() string {
	now := time.Now()
	return fmt.Sprintf("%v.csv", now.Format("Jan-2006"))
}

func createDirIfNotExists(path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func createFileIfNotExists(filename string, dataFileHeader string) {
	if !fileExists(filename) {
		f, err := os.Create(filename)
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		if err != nil {
			log.Fatal(err)
		}
		f.WriteString(dataFileHeader)
	}
}

func homeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return home
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
