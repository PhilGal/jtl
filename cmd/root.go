// Copyright © 2020 Philipp Galichkin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

var rootCmd = &cobra.Command{
	Use:   "jtl",
	Short: "Jtl is a command-line tool for posting worktime logs to a Jira server",
	Long: `These are common Jtl commands used in various situations:

  # log your work as you go to a local file (see: 'jtl help log'),
  # display summary report (see: 'jtl help report'),
  # finally, push all data from file to your company's remote server (see: 'jtl help push')
  
For better experience, it is recommended to add a valid configuration file $HOME/.jtl/config.yaml. Type 'jtl help push' for more details.

When you call any command, a programm is trying to locate a data file $HOME/data/<month-year>.csv Thus, each month you'll have a new data file.
There is, however, a possibility to force programm to use a particular data file with '--data' global option. If you use decide to use --data, use it with every command, because it is a runtime option.
Same goes for the config file with '--config' option.
`}

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
	rootCmd.PersistentFlags().StringVar(&dataFile, "data", "", "data file (default is $HOME/data/<month-year>.csv)")
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

	createNewDataFile := func() {
		dataDir := path.Join(homeDir(), ".jtl", "data")
		f := path.Join(dataDir, dataFileName())
		createDirIfNotExists(dataDir)
		createFileIfNotExists(f, dataFileHeader)
		//upgrade file to full path in context
		dataFile = f
	}

	if dataFile == "" {
		createNewDataFile()
	} else {
		if !fileExists(dataFile) {
			fmt.Println("Provided data file doesn't exist. Default one will be used.")
			createNewDataFile()
		}
	}
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
