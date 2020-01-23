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
	"github.com/gocraft/dbr/v2"
	"log"
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
	Id           int64     `json:"id" db:"id"`
	UserId       int64     `json:"userId" db:"user_id"`
	Name         string    `json:"name" db:"race_name"`
	StartTime    time.Time `json:"startTime" db:"start_time"`
	Status       string    `json:"status" db:"status"`
///	Registrations []RaceRegistration
	Game         *Game
}

type RaceRegistration struct {
	UserId       int64     `json:"userId" db:"user_id"`
	RaceId       int64     `json:"raceId" db:"race_id"`
	RegisteredOn time.Time `json:"registeredOn" db:"registered_on"`
	Placed       int       `json:"placed" db:"placed"`
}

func (this *Race) UpdateRaceStatus(db *dbr.Session, status RaceStatusType) error {

	var err error
	_, err = db.Update("races").Set("status", status.String()).
		Where("id = ?", this.Id).Exec()

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

//func AddRace(db *pg.DB, name string, startTime time.Time, status string) (*Race, error) {
func AddRace(db *dbr.Session, userId int64, name string, startTime time.Time, status string) (*Race, error) {

	var id int64
	err := db.InsertInto("races").
		Pair("race_name", name).
		Pair("user_id", userId).
		Pair("start_time", startTime).
		Pair("status", status).
		Returning("id").Load(&id)

	if err != nil {
		log.Fatalf("Insert User failed: %v", err)
		return nil, err
	}

	var race Race
	err = db.Select("*").From("races").Where("id = ?", id).LoadOne(&race)

	if err != nil {
		log.Fatalf("Select User failed: %v", err)
		return nil, err
	}

	return &race, nil

}

/*
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
 */

func LoadRaces(db *dbr.Session) ([]Race, error) {

	result, err := db.Select("*").From("races").Rows()

	var races = make([]Race, 0)

	for {
		if result.Next() == false {
			if err := result.Close(); err != nil {
				return races, err
			} else {
				return races, nil
			}
		}
		var id int64
		var userId int64
		var name string
		var status string
		var startTime time.Time
		if err := result.Scan(&id, &userId, &name, &startTime, &status); err != nil {
			_ = result.Close()
			log.Fatalf("Scan tournament data failed: %v\n", err)
			return nil, err
		}
		rec := Race{id, userId, name, startTime, status, nil}
		races = append(races, rec)
	}
	return races, err

}


func (this *Race) Register(db *dbr.Session, userId int64) error {
	var id int64
	if this.Id == 0 {
		err := db.InsertInto("race_registration").
			Columns("user_id", "register_on", "placed").
			Values(userId, time.Now(), 0).
			Returning("id").
			Load(&id)
		if err != nil {
			log.Fatalf("Insert Tournaments failed: %v", err)
			return err
		}
		this.Id = id

	}

	return nil

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
