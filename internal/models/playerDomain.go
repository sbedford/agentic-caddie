package models

import (
	"fmt"
	"time"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

type Player struct {
	ID       int64
	Name     string
	Handicap float64
	Clubs    []Club
	Rounds   []Round
}

func (this *Player) GetClub(clubName string) (*Club, error) {
	for i, c := range this.Clubs {
		if c.ClubName == clubName {
			return &this.Clubs[i], nil
		}
	}
	return nil, fmt.Errorf("Error retrieving Club [%v] - club not in Players clubs", clubName)
}

type Club struct {
	Player         Player
	ClubName       string
	AddedDate      time.Time
	RemovedDate    time.Time
	CarryAvg       float64
	CarryReliable  float64
	CarryMax       float64
	DispersionAvgM float64
	DispersionBias string
	SampleSize     int64
	CalculatedAt   time.Time
}

func (this *Club) Load(c db.PlayerClub, p Player) {
	this.Player = p
	this.ClubName = c.ClubName
	this.CarryAvg = helpers.Float64(c.CarryAvg)
	this.CarryReliable = helpers.Float64(c.CarryReliable)
	this.CarryMax = helpers.Float64(c.CarryMax)
	this.DispersionAvgM = helpers.Float64(c.DispersionAvgM)
	this.DispersionBias = helpers.String(c.DispersionBias)
}
