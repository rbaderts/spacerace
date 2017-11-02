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
	"database/sql"
	_ "github.com/lib/pq"

	"fmt"
	"github.com/joho/godotenv"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"

	_ "github.com/mattes/migrate/source/file"
	"github.com/rbaderts/spacerace/cmd"
	"github.com/rbaderts/spacerace/core"

	//"github.com/rbaderts/spacerace/migrations"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file")
	}

	core.ConfigureLogging()

	migrateDB()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func migrateDB() {

	pg_user := os.Getenv("POSTGRES_USER")
	pg_pw := os.Getenv("POSTGRES_PASSWORD")
	pg_host := os.Getenv("SPACERACE_DB_HOST")
	pg_db := "spacerace"

	//pg_d::= os.Getenv("POSTGRES_PASSWORD")

	//	s := bindata.Resource(migrations.AssetNames(),
	//		func(name string) ([]byte, error) {
	//			return migrations.Asset(name)
	//		})

	//	d, err := bindata.WithInstance(s)
	//	if err != nil {
	//		fmt.Printf("Migratione err %v\n", err)
	//	}

	//	m, err := migrate.NewWithSourceInstance("go-bindata", d, "database://foobar")

	fmt.Printf("pg_user = %v, pg_pw = %v, pg_host = %v\n", pg_user, pg_pw, pg_host)

	db_url := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", pg_user, pg_pw, pg_host, pg_db)

	var db *sql.DB
	for {
		var err error
		db, err = sql.Open("postgres", db_url)
		if err != nil {
			fmt.Printf("Error = %v\n", err)
			time.Sleep(time.Second * 2)
		} else {
			for {
				if err = db.Ping(); err != nil {
					//continue
					time.Sleep(time.Second * 2)
				} else {
					break
				}
			}
			break
		}
	}

	//	    db, err := sql.Open("postgres", "postgres://localhost:5432/database?sslmode=enable")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		fmt.Printf("Error 1 = %v\n", err)
	}

	//m, err := migrate.NewWithSourceInstance("go-bindata", d, db_url)
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", "postgres", driver)

	if err != nil {
		fmt.Printf("Error 2 = %v\n", err)
	}

	err = m.Up()
	if err != nil {
		fmt.Printf("Error 3 = %v\n", err)
	}

	//	s := bindata.Resource(migrations.AssetNames(),
	//		func(name string) ([]byte, error) {
	//			return migrations.Asset(name)
	//		})
	//	driver, err := postgres.WithInstance(db, &postgres.Config{})
	//	m, err := migrate.NewWithDatabaseInstance(
	//		"go-bindata", "postgres", driver)
	//	m.Steps(2)

	//	d, err := bindata.WithInstances(s)
	//	m, err := migrate.NewWithSourceInstance("go-bindata", c, "postgres://postgres:postgres@localhost:5432/spacerace")

}
