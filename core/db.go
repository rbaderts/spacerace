package core

import (
	"fmt"
	_ "github.com/go-pg/pg"
	"github.com/gocraft/dbr/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"log"
	_ "log"
	"os"
	"time"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
)

//var DB *sql.DB
var DB *dbr.Connection
var DBSession *dbr.Session

func SetupDB() *dbr.Connection {

	fmt.Printf("SetupDB\n")
	pg_user := os.Getenv("POSTGRES_USER")
	pg_pw := os.Getenv("POSTGRES_PASSWORD")
	pg_host := os.Getenv("SPACERACE_DB_HOST")
	pg_db := "spacerace"

	if pg_host == "" {
		pg_host = "localhost"
	}

	addr := fmt.Sprintf("%s:%s", pg_host, "5432")

	fmt.Printf("addr = %v\n", addr)

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=verify-full", pg_user, pg_pw, pg_host, pg_db)

	//	func Open(driver, dsn string, log EventReceiver) (*Connection, error)

	conn, err := dbr.Open("postgres", connStr, nil)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		log.Fatal(err)
	}
	DB = conn

	return DB

}

func MigrateDB(assetDir string) {

	pg_user := os.Getenv("POSTGRES_USER")
	pg_pw := os.Getenv("POSTGRES_PASSWORD")
	pg_host := os.Getenv("SPACERACE_DB_HOST")
	pg_db := "spacerace"

	fmt.Printf("pg_user = %v, pg_pw = %v, pg_host = %v\n", pg_user, pg_pw, pg_host)

	db_url := fmt.Sprintf("postgres://%s:'%s'@%s:5432/%s?sslmode=disable", pg_user, pg_pw, pg_host, pg_db)

	fmt.Printf("dburl = %s\n", db_url)

	for {
		var err error
		//db, err = dbr.Open("postgres", db_url, nil)
		conn, err := dbr.Open("postgres", db_url, nil)
		if err != nil {
			fmt.Printf("err = %v\n", err)
		}

		DB = conn
		DBSession = conn.NewSession(nil)
		if err != nil {
			fmt.Printf("Error = %v\n", err)
			time.Sleep(time.Second * 2)
		} else {
			for {
				if err = DBSession.Ping(); err != nil {
					//continue
					time.Sleep(time.Second * 2)
				} else {
					break
				}
			}
			break
		}
	}

	m, err := migrate.New("file://"+assetDir, db_url)


	fmt.Printf("m = %v, err = %v\n", m, err)

	if err != nil {
		fmt.Errorf("err = %v", err)
	}

	err = m.Up()
	if err != nil {
		fmt.Errorf("Up err = %v", err)
	}


	v, _, _:= m.Version()
	fmt.Printf("database name = %v\n", v)
}
