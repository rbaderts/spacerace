package core

import (
	"github.com/go-pg/pg"

	"fmt"
	_ "log"
	"os"
)

var DB *pg.DB

func SetupDB() *pg.DB {

	pg_user := os.Getenv("POSTGRES_USER")
	pg_pw := os.Getenv("POSTGRES_PASSWORD")
	pg_host := os.Getenv("SPACERACE_DB_HOST")
	pg_db := "spacerace"

	addr := fmt.Sprintf("%s:%s", pg_host, "5432")

	fmt.Printf("addr = %v\n", addr)
	options := pg.Options{
		User:     pg_user,
		Password: pg_pw,
		Database: pg_db,
		Addr:     addr,
	}

	DB = pg.Connect(&options)

	if DB == nil {

		DB = pg.Connect(&options)

	}

	return DB

}
