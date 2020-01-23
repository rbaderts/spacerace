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

package main

import (
	"fmt"
	"github.com/gobuffalo/packr"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"path/filepath"

	//	"github.com/mattes/migrate/database/postgres"
	//	"golang-migrate/migrate"

	/*
		_"github.com/golang-migrate/migrate/v4/database/postgres"
	*/

	//	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	//_ "github.com/mattes/migrate/source/file"
	"github.com/rbaderts/spacerace/cmd"
	"github.com/rbaderts/spacerace/core"

	//"github.com/rbaderts/spacerace/migrations"
	_ "flag"
	"os"
	_ "os/signal"
)

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")


func main() {

	///        cpufile, memfile, err := cmd.StartProfile()

	/*
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func(){
			for _ := range c {
				cmd.StopProfile(cpufile, memfile)
			}
		}()
	*/
	//	defer cmd.StopProfile(cpufile, memfile)

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file")
	}

	core.ConfigureLogging()

	dir := setupMigrationAssets()
	core.MigrateDB(dir)

	fmt.Printf("deleting path %v\n", dir)
	defer os.RemoveAll(dir);
	if err != nil {
		fmt.Errorf("Error removing migration tmp dir: %v\n", err)
	}

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}



	/*
		fmt.Printf("CPUProfile = %s\n", cmd.CPUProfile)
		if cmd.CPUProfile != "" {
			f, err := os.Create(cmd.CPUProfile)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	*/
}

func setupMigrationAssets() string {

	tmpDir, err := ioutil.TempDir("./", "migrations_tmp")

	//err := os.Mkdir("./migrations_tmp", 0755)
	if err != nil {
		log.Fatal(err)
	}

	box := packr.NewBox("./migrations")

	for _, m := range box.List() {
		fmt.Printf("box item %s\n", m)
		bytes, err := box.Find(m)
		if err != nil {
			fmt.Printf("err = %v", err)
			continue
		}
		fmt.Printf("bytes = %v", string(bytes))
		tmpFile := filepath.Join(tmpDir, m)
		fmt.Printf("tmpFn = %v\n", tmpFile)
		err = ioutil.WriteFile(tmpFile, bytes, 0644)
		if err != nil {
			fmt.Errorf("err = %v", err)
			continue
		}
	}
	return tmpDir;
}
