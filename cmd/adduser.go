// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	_ "github.com/go-pg/pg"
	"github.com/rbaderts/spacerace/core"
	"github.com/spf13/cobra"
)

// adduserCmd represents the server command
var adduserCmd = &cobra.Command{
	Use:   "adduser [name] [password]",
	Short: "Setsup an empty DB schema",
	Long: `Runs the spacerace game serve

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		adduser_runit(args)
	},
}

func init() {
	RootCmd.AddCommand(adduserCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

var ()

func adduser_runit(args []string) {

	_ = core.SetupDB()

	core.AddUser(core.DB, args[0], args[1])

}
