package core

import (
	"fmt"
	"github.com/gocraft/dbr/v2"
	_ "golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type UserId int

type User struct {
	Id        int
	Subject   string
	Email     string
	Provider  string
	LastLogin time.Time
}

func AddProvidedUser(db *dbr.Session, email string, provider string, subject string) (*User, error) {

	var id int64
	err := db.InsertInto("users").
		Pair("subject", subject).
		Pair("email", email).
		Pair("provider", provider).
		Returning("id").Load(&id)

	if err != nil {
		log.Fatalf("Insert User failed: %v", err)
		return nil, err
	}

	var user User
	err = db.Select("*").From("users").Where("id = ?", id).LoadOne(&user)

	if err != nil {
		log.Fatalf("Select User failed: %v", err)
		return nil, err
	}

	return &user, err
}

func AddUser(db *dbr.Session, email string, name string) (*User, error) {

	fmt.Printf("AddUser %s\n", email)

	subject := email
	result, err := db.InsertInto("users").
		Pair("subject", subject).
		Pair("email", email).
		Pair("provider", "").
		Returning("id").Exec()

	var id int64
	id, err = result.LastInsertId()

	if err != nil {
		fmt.Printf("err10: %v\n", err)
		log.Fatal(err)
	}

	var user User
	err = db.Select("*").From("users").Where("id = ?", id).LoadOne(&user)

	if err != nil {
		fmt.Printf("err11: %v\n", err)
		log.Fatal(err)
	}

	return &user, err
}

func LoadUserByEmail(db *dbr.Session, email string) (*User, error) {

	var user User
	err := db.Select("*").From("users").Where("email = ?", email).LoadOne(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func LoadUserBySubject(db *dbr.Session, subject string) (*User, error) {

	var user User
	err := db.Select("*").From("users").Where("subject = ?", subject).LoadOne(&user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
