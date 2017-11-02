package core

import (
	_ "database/sql"
	"github.com/go-pg/pg"
	"golang.org/x/crypto/bcrypt"

	"fmt"
	"time"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "test"
)

type User struct {
	Id             int
	Email          string
	Name           string
	PasswordDigest []byte
	Provider       string
	LastLogin      time.Time
}

const insertUserSQL = `insert into "users" (id, email, password_hash, lastLogin) value 
 ($1, $2, $3, $4);
 `

func AddProvidedUser(db *pg.DB, email string, provider string) (*User, error) {

	var user User

	_, err := db.QueryOne(&user, `
		INSERT INTO users (email, provider, last_login) VALUES 
		(?, ?, ?) RETURNING id
	`, email, provider, nil)

	//_, err := DBConn.Exec("insert into users(id, name, password_digest, password_salt, lastLogin) values($1, $2, $3)",
	// 	 user, hash, nil)
	//	err = DB.Insert(&User{user, string(hash), "", nil})
	return &user, err
}

func AddUser(db *pg.DB, email string, password string) (*User, error) {

	fmt.Printf("AddUser %s, %s\n", email, password)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	//	user := &User{Email: email, Name: email, Provider: "LOCAL", PasswordDigest: hash, LastLogin: time.Now()}
	var user User

	_, err = db.QueryOne(&user, `
		INSERT INTO users (email, name, password_digest, provider, last_login) VALUES 
		(?, ?, ?, ?, ?) RETURNING id 
		`, email, email, hash, "LOCAL", time.Now())

	//_, err := DBConn.Exec("insert into users(id, name, password_digest, password_salt, lastLogin) values($1, $2, $3)",
	// 	 user, hash, nil)
	//	err = DB.Insert(&User{user, string(hash), "", nil})
	return &user, err
}

func LoadUserByEmail(db *pg.DB, email string) (*User, error) {

	var user User
	_, err := db.QueryOne(&user, "SELECT * from users where email = ?", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func LoadUser(db *pg.DB, id int) (*User, error) {

	var user User
	_, err := db.QueryOne(&user, "SELECT * from users where id = ?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func Auth(db *pg.DB, email string, password string) (*User, error) {

	passwordBytes := []byte(password)

	fmt.Printf("querying for user %s\n", email)
	var user User
	_, err := db.QueryOne(&user, "SELECT * from users where email = ?", email)

	if err != nil {
		fmt.Printf("queried user error = %v\n", err)
		return nil, err
	}
	fmt.Printf("queried user: name = %s\n", user.Email)

	result := bcrypt.CompareHashAndPassword(user.PasswordDigest, passwordBytes)
	if result == nil {
		fmt.Printf("Passwords matched \n")
		return &user, nil
	}
	fmt.Printf("Passwords didn't match \n")
	return nil, result

	//	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	//		DB_USER, DB_PASSWORD, DB_NAME)
	//	db, err := sql.Open("postgres", dbinfo)
	//	if err != nil {
	//		return nil, err
	//	}
	//	defer db.Close()

}
