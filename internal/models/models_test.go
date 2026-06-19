package models

import (
	"testing"
)

func Test_CurrentHoleReturnsCorrect(t *testing.T) {

	hole2 := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}

	subject := Round{
		PlayedHoles: []PlayedHole{
			PlayedHole{Hole: Hole{HoleNumber: 1}, Completed: true},
			hole2,
		},
	}

	nextHole, _ := subject.CurrentHole()
	if nextHole.Hole.HoleNumber != hole2.Hole.HoleNumber {
		t.Errorf("Result was incorrect, got: %v, want: %v.", nextHole.Hole.HoleNumber, hole2.Hole.HoleNumber)
	}
}

func Test_CantAddShotToCompletedHole(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: true}
	club := Club{ClubName: "Driver"}

	_, err := hole.RecordShot("tee", club, "fairway", "", "clean", "")

	if err == nil {
		t.Errorf("Didnt receive error")
	}
}

func Test_AddShotAppendsToShotsTaken(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, err := hole.RecordShot("tee", club, "fairway", "", "clean", "")

	if err != nil {
		t.Error(err)
	}

	if shot.ShotNumber != 1 {
		t.Errorf("[Test_AddShotAppendsToShotsTaken] - Expected ShotNumber [1] Got [%v]", shot.ShotNumber)
	}

	shots := len(hole.ShotsTaken)
	if shots != 1 {
		t.Errorf("Result was incorrect, got %v, want %v", shots, 1)
	}
}

func Test_CurrentLocationRespectsLastShot(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	hole.RecordShot("tee", club, "fairway", "", "clean", "")
	hole.RecordShot("tee", club, "green", "", "clean", "")

	location := hole.CurrentLocation()
	if location != "green" {
		t.Errorf("[TestCurrentLocation] - Expected [fairway] Got [%v]", location)
	}
}

func Test_LookupShotByNumber(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, _ := hole.RecordShot("tee", club, "fairway", "", "clean", "")

	s, _ := hole.LookupShot(int(shot.ShotNumber))
	if s == nil {
		t.Errorf("[Test_LookupShotByNumber] - Expected [Shot 1] Got [nil]")
	}

	_, err := hole.LookupShot(763)
	if err == nil {
		t.Errorf("[Test_LookupShotByNumber] - Expected [Shot 1] Got [nil]")
	}

}

/*
   ('shot_result', 'fairway', 'Fairway',       1),
   ('shot_result', 'rough',   'Rough',         2),
   ('shot_result', 'bunker',  'Bunker',        3),
   ('shot_result', 'hazard',  'Hazard',        4),
   ('shot_result', 'ob',      'Out of Bounds', 5),
   ('shot_result', 'lost',    'Lost Ball',     6),
   ('shot_result', 'green',   'Green',         7),
   ('shot_result', 'holed',   'Holed',         8),
   ('shot_result', 'unknown', 'Unknown',       9);

*/

func Test_CurrentLocationUsesLastShotIfOB(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	hole.RecordShot("tee", club, "OB", "", "clean", "")

	location := hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}

	hole.RecordShot("tee", club, "lost", "", "clean", "")
	location = hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}
}

func Test_LastValidShot(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	if hole.GetLastValidShot() != nil {
		t.Errorf("[Test_LastValidShot] - dont know how this is possible")
	}

	hole.RecordShot("tee", club, "fairway", "", "clean", "")
	lastValidShot := hole.GetLastValidShot()
	if lastValidShot.Result != "fairway" {
		t.Errorf("[Test_LastValidShot] - expected [fairway]] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot("approach", club, "OB", "", "clean", "")
	hole.RecordShot("approach", club, "OB", "", "clean", "")
	lastValidShot = hole.GetLastValidShot()
	if lastValidShot.Result != "fairway" {
		t.Errorf("[Test_LastValidShot] - expected [fairway]] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot("approach", club, "green", "", "clean", "")
	lastValidShot = hole.GetLastValidShot()
	if lastValidShot.Result != "green" {
		t.Errorf("[Test_LastValidShot] - expected [green] got [%v]", lastValidShot.Result)
	}
}

func Test_ValidShots(t *testing.T) {
	shot := Shot{
		Result: "fairway",
	}

	checkValid := func(s Shot, expected bool) {
		if s.ValidShot() != expected {
			t.Errorf("[Test_ValidShots] - %v - Expected [%v] Got [%v]", s.Result, expected, !expected)
		}
	}

	checkValid(shot, true)

	shot.Result = "rough"
	checkValid(shot, true)

	shot.Result = "bunker"
	checkValid(shot, true)

	shot.Result = "OB"
	checkValid(shot, false)

	shot.Result = "green"
	checkValid(shot, true)

}

func Test_CurrentLocationReturnsTeeForNewHole(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}

	location := hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}
}
