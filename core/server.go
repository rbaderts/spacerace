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
	"github.com/go-chi/chi"
	"github.com/gocraft/dbr/v2"
	"strings"

	"fmt"
	"github.com/go-chi/render"
	"github.com/gobuffalo/packr"
	"log"
	"os"

	//	"github.com/gorilla/mux"
	//	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	_ "github.com/spf13/cobra"

	"github.com/sirupsen/logrus"

	"html/template"
	"net/http"
	"strconv"

	// "github.com/codegangsta/negroni"
	"github.com/rbaderts/spacerace/auth"
	// "github.com/rbaderts/spacerace/routes/callback"
	// "github.com/rbaderts/spacerace/routes/login"
	// "github.com/rbaderts/spacerace/routes/logout"
	/// "github.com/rbaderts/spacerace/routes/user"
	_ "log"
	_ "os/user"
	"time"
	// "github.com/rbaderts/spacerace/routes/home"
)

var (
	gameTemplate     *template.Template
	lobbyTemplate    *template.Template
	loginTemplate    *template.Template
	registerTemplate *template.Template
	upgrader         = websocket.Upgrader{WriteBufferSize: 1024, ReadBufferSize: 1024}

	Games     map[int64]*Game
	Players   map[int]Player
	lobby     *Lobby
	Users     map[string]*User
	SkipLogin bool
)

type Env struct {
	DB   *dbr.Session
	Port string
	Host string
}

var Log = logrus.New()
var PerfLog = logrus.New()

const (
	TIME_FORMAT = "02/06/2002 3:04PM"
)

func init() {
	Users = make(map[string]*User)

}

func FormatAsDate(t time.Time) string {
	return t.Format(TIME_FORMAT)
}

/*
func FormatAsInt(i int) string {
	return string(i)
}
*/

//var Store sessions.Store

var fmap = template.FuncMap{
	"FormatAsDate": FormatAsDate,
	"eq": func(a, b interface{}) bool {
		return a == b
	},
}

var SessionSecret string

type GameParams struct {
	PlayerID string
}

type GameData struct {
	GameID   string
	PlayerID string
}

/*
func StartAuth(r *mux.Router) {

	r.HandleFunc("/", home.HomeHandler)
	r.HandleFunc("/login", login.LoginHandler)
	r.HandleFunc("/logout", logout.LogoutHandler)
	r.HandleFunc("/callback", callback.CallbackHandler)

	amw := auth.NewAuthenticationMiddleware()
	amw.Populate()

	if SkipLogin {
	} else {
		r.Handle("/user", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(user.UserHandler)),
		)).Methods("GET")
	}

	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public/"))))
}

*/

/*
func Server() {

	SessionSecret = os.Getenv("SESSION_SECRET")
	fmt.Printf("SessionSecret = %v\n", SessionSecret)

	r := mux.NewRouter()

	//Store the cookie store which is going to store session data in the cookie
	auth.Store = sessions.NewCookieStore([]byte(SessionSecret))

	auth.AuthInit()
	StartAuth(r)

	_ = SetupDB()

	PurgeRaces(DB)
	lobby = NewLobby()

	gameTemplate = template.New("game")
	gameTemplate.Funcs(fmap).ParseFiles(
		"templates/game.tmpl", "templates/header.tmpl", "templates/footer.tmpl")

	lobbyTemplate = template.New("lobby")
	lobbyTemplate.Funcs(fmap).ParseFiles(
		"templates/lobby.tmpl", "templates/header.tmpl", "templates/footer.tmpl")

	loginTemplate = template.New("login")
	loginTemplate.Funcs(fmap).ParseFiles(
		"templates/login.tmpl", "templates/header.tmpl", "templates/footer.tmpl")

	Games = make(map[int]*Game)
	Players = make(map[int]Player)

	fmt.Printf("Setting up handlers\n")

	if SkipLogin {

		//		r.HandleFunc("/", home.HomeHandler).Methods("GET")
		r.HandleFunc("/", LobbyHandler).Methods("GET")
		r.HandleFunc("/lobby", LobbyHandler).Methods("GET")
		r.HandleFunc("/newrace", NewRaceHandler).Methods("POST")
		r.HandleFunc("/races/{id}", RaceHandler).Methods("GET")
		r.HandleFunc("/updates/{gameId}", GameUpdateHandler).Methods("GET", "POST")

		/*
			r.Handle("/lobby", LobbyHandler).Methods("GET")
			r.Handle("/newrace", NewRaceHandler).Methods("POST")
			r.Handle("/races/{id}", RaceHandler).Methods("GET")
			r.Handle("/updates/{gameId}", GameUpdateHandler).Methods("GET", "POST")
*/

/*
	} else {
		r.Handle("/updates/{gameId}", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(GameUpdateHandler)),
		)).Methods("GET", "POST")

		r.HandleFunc("/", home.HomeHandler).Methods("GET")

		r.Handle("/lobby", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(LobbyHandler)),
		)).Methods("GET")

		r.Handle("/register", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(RegisterHandler)),
		)).Methods("GET", "POST")

		r.Handle("/newrace", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(NewRaceHandler)),
		)).Methods("POST")

		r.Handle("/races/{id}", negroni.New(
			negroni.HandlerFunc(IsLoggedIn),
			negroni.Wrap(http.HandlerFunc(RaceHandler)),
		)).Methods("GET")
	}

//	r.PathPrefix("/resources/").Handler(http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	r.PathPrefix("/resources").Handler(http.StripPrefix("/resources", http.FileServer(http.Dir("resources"))))
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("templates"))))

	fmt.Printf("ListenAndServe:")
	http.ListenAndServe(":3000", r)


}
*/


var Environment *Env

func Server() {

	env := &Env{
		DB:   DBSession,
		Port: os.Getenv("PORT"),
		Host: os.Getenv("HOST"),
		// We might also have a custom log.Logger, our
		// template instance, and a config struct as fields
		// in our Env struct.
	}
	Environment = env

	assetBox := packr.NewBox("../web")

	Games = make(map[int64]*Game)
	Players = make(map[int]Player)

	lobby = NewLobby()
	loadTemplates()

	r := chi.NewRouter()

	r.Use(render.SetContentType(render.ContentTypeJSON))
	//	r.Use(jwtauth.Verifier(tokenAuth))

	FileServer(r, "/static", assetBox)

	/*
		r.HandleFunc("/", LobbyHandler).Methods("GET")
		r.HandleFunc("/lobby", LobbyHandler).Methods("GET")
		r.HandleFunc("/newrace", NewRaceHandler).Methods("POST")
		r.HandleFunc("/races/{id}", RaceHandler).Methods("GET")
	*/

	r.Group(func(r chi.Router) {
		if !SkipLogin {
			r.Use(auth.AuthenticationRequired)
		}
		r.Get("/lobby", Handler{env, HomeRenderHandler}.ServeHTTP)

		r.Post("/races", Handler{env, NewRaceHandler}.ServeHTTP)
		r.Route("/races/{gameID}", func(r chi.Router) {
			r.Get("/", Handler{env, RaceHandler}.ServeHTTP)
		})
		r.Route("/updates/{gameID}", func(r chi.Router) {
			r.Get("/", Handler{env, WebserviceHandler}.ServeHTTP)
			r.Post("/", Handler{env, WebserviceHandler}.ServeHTTP)
		})
		/*
					r.Route("/users", func(r chi.Router) {
						r.Get("/", Handler{env, GetPlayerListHandler}.ServeHTTP)
						r.Delete("/", Handler{env, DeletePlayersHandler}.ServeHTTP)
						r.Post("/", Handler{env, PostUserHandler}.ServeHTTP)
						r.Route("/{playerID}", func(r chi.Router) {
							r.Post("/paid", Handler{env, PlayerPaidHandler}.ServeHTTP)
						})
					})

			})
		*/
	})

	r.Get("/callback", AuthCallbackHandler)
	r.Get("/login", LoginHandler)
	r.Get("/logout", LogoutHandler)
	/*
		r.Route("/api", func(r chi.Router) {
			r.Route("/tournaments", func(r chi.Router) {
				r.Post("/", CreateTournamentHandler)
				r.Get("/", ListTournamentHandler)
				r.Route("/{tournamentID}", func(r chi.Router) {
	*/
	fmt.Printf("launching server on 3000\n")

	if err := http.ListenAndServe(":3000", r); err != nil {
		fmt.Printf("ListenAndServe error = %v\n", err)

	}

}

func WebserviceHandler(env *Env, w http.ResponseWriter, r *http.Request) error {

	gameIdStr := chi.URLParam(r, "gameID")
	gameId, err := strconv.ParseInt(gameIdStr, 10, 0)

	if err != nil {
		return StatusError{500, err}
	}

    serveWs(w, r, gameId)

	return nil
}

func NewPractiseRace(env *Env, userId int64) (*Race, error) {
	race, err := AddRace(env.DB, userId, "PractiseRace", time.Now(), "RaceUnderway")
	return race, err
}

func CreateGame(race *Race) *Game {

	game := NewGame(race.Id)
	fmt.Printf("new game = %v\n", game)
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
	raceId, err := strconv.ParseInt(cookie.Value, 10, 0)
	v, has := Games[int64(raceId)]
	if has {
		return v
	}
	return nil
}

func serveWs(w http.ResponseWriter, r *http.Request, gameId int64) {

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
	game.UpdatePlayer(player)

	go hp.ProcessCommands()

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

/*

func LobbyHandler(w http.ResponseWriter, r *http.Request) {

	lobby.RefreshRaces()
	fmt.Printf("Games: = %v\n", Games)
	for i, _ := range lobby.Races {
		g, present := Games[lobby.Races[i].Id]
		if present {
			lobby.Races[i].Game = g
		}
	}

	profile := GetProfile(w, r)

	fmt.Printf("prof=%v\n", profile)
	lobby.Email = profile["email"].(string)
	lobby.Img = profile["picture"].(string)

	lobby.Auth0ClientId = os.Getenv("AUTH0_CLIENT_ID")
	lobby.Auth0ClientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	lobby.Auth0Domain = os.Getenv("AUTH0_DOMAIN")
	lobby.Auth0CallbackURL = template.URL(os.Getenv("AUTH0_CALLBACK_URL"))
*/

//	lobby.Email = user.Email
/*
	fmt.Printf("lobby email: %v\n", lobby.Email)
	err := lobbyTemplate.Execute(w, lobby)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//		}

}
*/

/*
func GameUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	gameIdStr, _ := vars["gameId"]
	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		fmt.Printf("Can't get gameId\n")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	serveWs(w, r, gameId)
}
*/

/*
func RegisterHandler(w http.ResponseWriter, r *http.Request) {

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
		AddUserWithPassword(DB, email, password)
		http.Redirect(w, r, "/login", 302)

	}
}
*/

func NewRaceHandler(env *Env, w http.ResponseWriter, r *http.Request) error {

	var userId int64 = 0
	if (!SkipLogin) {
		userId = r.Context().Value("uid").(int64)
	}

	race, err := NewPractiseRace(env, userId)
	if err != nil {
		return StatusError{500, err}
	}
	game := CreateGame(race)
	race.Game = game
	game.Race = race

	fmt.Printf("new race - %v\n", race)

	render.JSON(w, r, race)

	return nil

}

func RaceHandler(env *Env, w http.ResponseWriter, r *http.Request) error {

	//_ := r.Context().Value("uid").(int)

	raceIdStr := chi.URLParam(r, "gameID")
	raceId, err := strconv.ParseInt(raceIdStr, 10, 0)
	if err != nil {
		return StatusError{500, err}
	}

	game := Games[raceId]
	fmt.Printf("opening game: %v\n", game)

	if game.IsRunning() == false {
		game.Start()
	}

	data := &GameData{strconv.FormatInt(game.Id, 10), ""}

	err = gameTemplate.Execute(w, data)

	if err != nil {
		return StatusError{500, err}
	}

	return nil
}

func loadTemplates() {

	gameTemplate = template.Must(template.New("game").Funcs(fmap).ParseFiles(
		"web/templates/game.tmpl",
		"web/templates/header.tmpl",
		"web/templates/footer.tmpl"))

	lobbyTemplate = template.Must(template.New("lobby").Funcs(fmap).ParseFiles(
		"web/templates/lobby.tmpl",
		"web/templates/header.tmpl",
		"web/templates/footer.tmpl")).Funcs(fmap)

	loginTemplate = template.Must(template.New("login").Funcs(fmap).ParseFiles(
		"web/templates/login.tmpl",
		"web/templates/header.tmpl",
		"web/templates/footer.tmpl")).Funcs(fmap)

}

/*
func HomeHandler(env *Env, w http.ResponseWriter, r *http.Request) error {

	data := struct {
		Auth0ClientId     string
		Auth0ClientSecret string
		Auth0Domain       string
		Auth0CallbackURL  template.URL
	}{
		os.Getenv("AUTH0_CLIENT_ID"),
		os.Getenv("AUTH0_CLIENT_SECRET"),
		os.Getenv("AUTH0_DOMAIN"),
		template.URL(os.Getenv("AUTH0_CALLBACK_URL")),
	}

	templates.RenderTemplate(w, "home", data)
}
*/

func HomeRenderHandler(env *Env, w http.ResponseWriter, r *http.Request) error {

	session, err := auth.AuthStore.Get(r, "auth-session")
	if err != nil {
		return StatusError{500, err}
	}

	fmt.Printf("lobby1\n")
	var name string
	if (!SkipLogin) {
		var ok bool
		var val interface{}
		if val, ok = session.Values["given_name"]; !ok {
			return StatusError{http.StatusSeeOther, err}
		}

		name = val.(string)
	} else {
		name = "noone"
	}

	data := struct {
		Email     string
		Img      string
		Races     []Race
	}{
		name,
		"static/img/bracketlogo.gif",
		lobby.Races,
	}

	fmt.Printf("lobby3\n")
	if err := lobbyTemplate.Execute(w, data); err != nil {
		return StatusError{500, err}
	}
	return nil
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))

}

type Handler struct {
	*Env
	H func(e *Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

type Username struct {
	Name string `json:"name"`
}

func (this Username) String() string {
	return fmt.Sprintf(`{"name": "%s"}`, this.Name)
}

type Error interface {
	error
	Status() int
}

// StatusError represents an error with an associated HTTP status code.
type StatusError struct {
	Code int
	Err  error
}

func (se StatusError) Error() string {
	return se.Err.Error()
}

// Returns our HTTP status code.
func (se StatusError) Status() int {
	return se.Code
}
