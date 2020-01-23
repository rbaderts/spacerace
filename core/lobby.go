package core

import (
	"bytes"
	"fmt"
	"html/template"
)

type Lobby struct {
	Races []Race
	Email string
	Img   string
	Auth0ClientId     string
	Auth0ClientSecret string
	Auth0Domain       string
	Auth0CallbackURL  template.URL

}

func NewLobby() *Lobby {

	lobby := new(Lobby)
	return lobby

}

func (this *Lobby) RefreshRaces() *Lobby {
	races, err := LoadRaces(Environment.DB)
	if err != nil {
		fmt.Printf("NewLobby error - %v\n", err)
	}
	lobby.Races = races
	return lobby
}

func (this *Lobby) String() string {
	b := new(bytes.Buffer)
	for _, r := range this.Races {
		fmt.Fprintf(b, "%v\n", r)
	}
	return b.String()
}

func (lobby *Lobby) AddRace(race *Race) {
	lobby.Races = append(lobby.Races, *race)
}
