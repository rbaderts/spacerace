/**
 * A Race represents a past, future or currrently active game sssions.
 * A race may have 0 or more user's registered to participate.  An active
 * Race has single associated Game instance (the Game,Id is the same as the Race.Id,  which contains the Races
 * runtime state.
 */
package core

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg"
	"time"
)

var ()

type RaceStatusType int

const (
	_ RaceStatusType = iota
	RacePending
	RaceUnderway
	RaceComplete
)

var RaceStatuses = [...]string{
	"None",
	"RacePending",
	"RaceUnderway",
	"RaceComplete",
	"PractiseRace",
}

func (b RaceStatusType) String() string {
	return RaceStatuses[b]
}

type Race struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	StartTime    time.Time `json:"startTime"`
	Status       string    `json:"status"`
	Regitrations []RaceRegistration
	Game         *Game
}

type RaceRegistration struct {
	userId     int       `json:"userId"`
	registerOn time.Time `json:"registeredOn"`
	placed     int       `json:"placed"`
}

func (this *Race) UpdateRaceStatus(db *pg.DB, status RaceStatusType) error {

	this.Status = status.String()
	_, err := db.Model(this).Column("status").Returning("*").Update()

	/*
		fmt.Printf("UpdateRaceStatus\n")
		_, err := db.Exec(`
			UPDATE races set status='?' where id = ?
		`, status.String(), raceId)
	*/
	if err != nil {
		return err
	}
	fmt.Printf("UpdateRaceStatus done\n")
	return nil
}

func AddRace(db *pg.DB, name string, startTime time.Time, status string) (*Race, error) {
	var race Race
	_, err := db.QueryOne(&race, `
		INSERT INTO races (name, start_time, status) VALUES (?, ?, ?)
		RETURNING id
	`, name, startTime, status)
	if err != nil {
		return nil, err
	}
	return &race, nil
}

func DeleteRace(db *pg.DB, id int) error {
	var race Race
	_, err := db.QueryOne(&race, `
		DELETE FROM races where id = ?
	`, id)
	if err != nil {
		return err
	}
	return nil
}

func PurgeRaces(db *pg.DB) error {
	_, err := db.Exec("DELETE FROM races;")
	if err != nil {
		return err
	}
	return nil
}

func LoadRaces(db *pg.DB) ([]Race, error) {
	var races []Race
	_, err := db.Query(&races, `SELECT * FROM races`)
	return races, err
}

func (this *Race) Register(db *pg.DB, userId int) error {
	r := &RaceRegistration{userId, time.Now(), 0}
	_, err := db.QueryOne(r, `
		INSERT INTO race_registration (user_id, register_on, placed) VALUES (?user_id, ?register_on, &placed))
		RETURNING id
	`, r)
	return err

}

func (this *Race) MarshalJSON() ([]byte, error) {

	hasGame := false
	if this.Game != nil {
		hasGame = true
	}
	b, err := json.Marshal(map[string]interface{}{
		"id":        this.Id,
		"name":      this.Name,
		"startTime": this.StartTime,
		"status":    this.Status,
		"hasGame":   hasGame,
	})

	if err != nil {
		panic("error marshall Stop\n")
	}
	return b, err

}

func (this *Race) String() string {
	return fmt.Sprintf("id: %v, name: %v, startTime: %v, static: %v, game: %v", this.Id, this.Name, this.StartTime, this.Status, this.Game)
}
