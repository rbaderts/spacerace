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
	"fmt"
	"github.com/rbaderts/spacerace/core"
	"os"
	"runtime/pprof"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var CPUProfile string
var MemProfile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "spacerace",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
/*
func Execute() error {

*/
	/*
	cpufile, memfile, err := startProfile()
	if err != nil {
		fmt.Printf("cannot setup profiling", err)
		//cmdExitCode = 254
		return err
	}
	defer stopProfile(cpufile, memfile)
	*/

/*
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
*/

func init() {

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spacerace.yaml)")

	RootCmd.PersistentFlags().StringVar(&CPUProfile, "cpuprofile", "", "write CPU profile to the file")
	RootCmd.PersistentFlags().StringVar(&MemProfile, "memprofile", "", "write CPU profile to the file")
	RootCmd.PersistentFlags().BoolVar(&core.SkipLogin, "skiplogin",  false, "SkipLogin")

	//	RootCmd.PersistentFlags().StringVar(&CPUProfile, "cpuprofile", "", "cpuprofile outputfile")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//	viper.BindPFlag("cpuprofile", RootCmd.PersistentFlags().Lookup("cpuprofile"))

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

		// Search config in home directory with name ".spacerace" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("spacerace")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func StartProfile() (cpufile *os.File, memfile *os.File, err error) {
	fmt.Printf("CPUProfile = %s, MemProfile = %s\n", CPUProfile, MemProfile)
	if CPUProfile != "" {
		cpufile, err = os.Create(CPUProfile)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create cpu profile file %q: %v", CPUProfile, err)
		}
		pprof.StartCPUProfile(cpufile)
	}
	if MemProfile != "" {
		memfile, err = os.Create(MemProfile)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create memory profile file %q: %v", MemProfile, err)
		}
	}
	return cpufile, memfile, nil
}

func StopProfile(cpuprofile, memprofile *os.File) {
	if CPUProfile != "" {
		pprof.StopCPUProfile()
		cpuprofile.Close()
	}
	if MemProfile != "" {
		pprof.WriteHeapProfile(memprofile)
		memprofile.Close()
	}
}

// runWrapper returns a func(cmd *cobra.Command, args []string) that internally
// will add command function return code and the reinsertion of the "--" flag
// terminator.
func runWrapper(cf func(cmd *cobra.Command, args []string) (exit int)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		cpufile, memfile, err := StartProfile()
		if err != nil {
			fmt.Printf("cannot setup profiling", err)
			//cmdExitCode = 254
			return
		}
		defer StopProfile(cpufile, memfile)

		//cmdExitCode = cf(cmd, args)
	}
}
