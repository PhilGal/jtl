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
	"os"

	"github.com/philgal/jtl/cmd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version string

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
	cobra.OnInitialize(config.Init)
	cobra.OnInitialize(config.InitDataFile)
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.jtl/config.yaml)")
	rootCmd.PersistentFlags().String("data", "", "data file (default is $HOME/data/<month-year>.csv)")
	viper.BindPFlag("data", rootCmd.PersistentFlags().Lookup("data"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
