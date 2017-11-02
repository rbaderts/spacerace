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

package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	_ "github.com/spf13/cobra"

	"github.com/sirupsen/logrus"

	"html/template"
	"net/http"
	"strconv"

	"time"
)

var (
	gameTemplate     *template.Template
	lobbyTemplate    *template.Template
	loginTemplate    *template.Template
	registerTemplate *template.Template
	upgrader         = websocket.Upgrader{}
	Games            map[int]*Game
	Players          map[int]Player
	lobby            *Lobby
)

var Log = logrus.New()
var PerfLog = logrus.New()

const (
	TIME_FORMAT = "02/06/2002 3:04PM"
)

func FormatAsDate(t time.Time) string {
	return t.Format(TIME_FORMAT)
}

/*
func FormatAsInt(i int) string {
	return string(i)
}
*/

var Store sessions.Store

var SessionSecret string

type GameParams struct {
	PlayerID string
}

type GameData struct {
	GameID   string
	PlayerID string
}

func Server() {

	SessionSecret = os.Getenv("SESSION_SECRET")
	fmt.Printf("SessionSecret = %v\n", SessionSecret)

	r := mux.NewRouter()

	//Store the cookie store which is going to store session data in the cookie
	Store = sessions.NewCookieStore([]byte(SessionSecret))

	SetupAuth(r)
	_ = SetupDB()

	PurgeRaces(DB)
	lobby = NewLobby()

	fmap := template.FuncMap{
		"FormatAsDate": FormatAsDate,
		"eq": func(a, b interface{}) bool {
			return a == b
		},
	}

	gameTemplate = template.New("game")
	gameTemplate.Funcs(fmap).ParseFiles(
		"templates/game.tmpl", "templates/header.tmpl")

	lobbyTemplate = template.New("lobby")
	lobbyTemplate.Funcs(fmap).ParseFiles(
		"templates/lobby.tmpl", "templates/header.tmpl")

	loginTemplate = template.New("login")
	loginTemplate.Funcs(fmap).ParseFiles(
		"templates/login.tmpl", "templates/header.tmpl")

	Games = make(map[int]*Game)
	Players = make(map[int]Player)

	//r.HandleFunc("/updates", serveWs).Queries("gameId", "", "playerId", "")
	r.HandleFunc("/updates/{gameId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		gameIdStr, _ := vars["gameId"]
		gameId, err := strconv.Atoi(gameIdStr)
		if err != nil {
			fmt.Printf("Can't get gameId\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		serveWs(w, r, gameId)
	}).Methods("GET", "POST")

	r.HandleFunc("/lobby", func(w http.ResponseWriter, r *http.Request) {

		lobby.RefreshRaces()
		fmt.Printf("Games: = %v\n", Games)
		for i, _ := range lobby.Races {
			g, present := Games[lobby.Races[i].Id]
			if present {
				lobby.Races[i].Game = g
			}
		}

		fmt.Printf("lobby: %v\n", lobby)
		user := IsLoggedIn(r)
		if user == nil {
			http.Redirect(w, r, "/login", 302)
		} else {
			lobby.Email = user.Email
			err := lobbyTemplate.Execute(w, lobby)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}).Methods("GET")

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginFunc(w, r)
	}).Methods("GET", "POST")

	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

		fmt.Printf("r.Method = %v\n", r.Method)
		if r.Method == "GET" {
			registerTemplate = template.New("register")
			registerTemplate.Funcs(fmap).ParseFiles(
				"templates/register.tmpl", "templates/header.tmpl")

			err := registerTemplate.Execute(w, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			r.ParseForm()
			// logic part of log in
			email := r.Form["email"][0]
			password := r.Form["password"][0]
			AddUser(DB, email, password)
			http.Redirect(w, r, "/login", 302)

		}
	}).Methods("GET", "POST")

	r.HandleFunc("/newrace", func(w http.ResponseWriter, r *http.Request) {

		user := IsLoggedIn(r)
		if user == nil {
			http.Redirect(w, r, "/login", 302)
		} else {
			race, err := NewPractiseRace()
			if err != nil {
				fmt.Printf("Error = %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			game := CreateGame(race)
			race.Game = game
			game.Race = race

			fmt.Printf("new race - %v\n", race)
			response, _ := json.Marshal(race)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			w.Write(response)

		}
	}).Methods("POST")

	r.HandleFunc("/races/{id}", func(w http.ResponseWriter, r *http.Request) {

		user := IsLoggedIn(r)
		if user == nil {
			http.Redirect(w, r, "/login", 302)
		} else {
			vars := mux.Vars(r)
			idstr, _ := vars["id"]
			fmt.Printf("id = %s\n", idstr)
			id, err := strconv.Atoi(idstr)
			if err != nil {
				fmt.Printf("Error = %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			game := Games[id]
			fmt.Printf("opening game: %v\n", game)

			if game.IsRunning() == false {
				game.Start()
			}

			data := &GameData{strconv.Itoa(game.Id), ""}

			err = gameTemplate.Execute(w, data)

			if err != nil {
				fmt.Printf("Error = %v\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	}).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/lobby", 302)
	}).Methods("GET")

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("resources"))))
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("templates"))))
	http.ListenAndServe(":8080", r)

}

func NewPractiseRace() (*Race, error) {
	race, err := AddRace(DB, "PractiseRace", time.Now(), "RaceUnderway")
	return race, err
}

func CreateGame(race *Race) *Game {

	game := NewGame(race.Id)
	Games[race.Id] = game

	return game
}

func setCookie(w http.ResponseWriter, gameId string) {
	cookie := new(http.Cookie)
	cookie.Name = "GameID"
	cookie.Value = gameId
	http.SetCookie(w, cookie)
}

func getCurrentGame(r *http.Request) *Game {

	cookie, err := r.Cookie("GameID")
	if err != nil {
		return nil
	}
	raceId, err := strconv.Atoi(cookie.Value)
	v, has := Games[raceId]
	if has {
		return v
	}
	return nil
}

func serveWs(w http.ResponseWriter, r *http.Request, gameId int) {

	fmt.Printf("Server - serveWs\n")

	game := Games[gameId]

	player := NewHumanPlayer(game)
	Players[player.GetPlayerId()] = player

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("upgrade:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hp := player.(*HumanPlayer)

	hp.setWebsocket(ws)
	game.Join(player)
	go hp.ProcessCommands()

}

func IsLoggedIn(r *http.Request) *User {
	session, _ := Store.Get(r, "session")

	loggedIn := session.Values["loggedin"]
	if loggedIn != nil {
		fmt.Printf("loggedIn = %v\n", loggedIn)
		userId := session.Values["loggedin"].(int)
		user, err := LoadUser(DB, userId)
		if err == nil {
			return user
		}
	}
	return nil
}

func LoginFunc(w http.ResponseWriter, r *http.Request) {

	session, err := Store.Get(r, "session")

	if err != nil {
		fmt.Printf("LoginFunc err = %v\n", err)
		loginTemplate.Execute(w, nil)
		// in case of error during
		// fetching session info, execute login template
	} else {

		user := IsLoggedIn(r)
		if user == nil {
			fmt.Printf("LoginFunc: Not Logged In\n")
			if r.Method == "POST" {
				fmt.Printf("LoginFunc: POST\n")

				pw := r.FormValue("password")
				email := r.FormValue("email")

				user, result := Auth(DB, email, pw)

				fmt.Printf("User auth result = %v\n", result)

				if result == nil {
					session.Values["loggedin"] = user.Id
					session.Values["loggedin_email"] = user.Email
					session.Save(r, w)
					http.Redirect(w, r, "/lobby", 302)
					return
				}
			} else if r.Method == "GET" {
				fmt.Printf("LoginFunc: GET\n")
				err := loginTemplate.Execute(w, nil)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			http.Redirect(w, r, "/lobby", 302)
		}
	}
}

func LogoutFunc(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "session")
	if err == nil {
		//If there is no error, then remove session
		session.Values["loggedin"] = 0
		session.Save(r, w)
	}
	http.Redirect(w, r, "/login", 302)
}

func ConfigureLogging() {

	file, err := os.OpenFile("spacerace.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		Log.Out = file
		Log.Formatter = &logrus.JSONFormatter{}
		Log.Level = logrus.InfoLevel

	} else {
		fmt.Printf("Failed setup spacerace.log: %v\n", err)
	}

	file, err = os.OpenFile("perf.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		PerfLog.Out = file
		PerfLog.Formatter = &logrus.JSONFormatter{}
		PerfLog.Level = logrus.InfoLevel
	} else {
		fmt.Printf("Failed setup perf.log: %v\n", err)
	}

}
