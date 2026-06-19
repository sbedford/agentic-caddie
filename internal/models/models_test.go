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

	_, err := hole.RecordShot("tee", club, "fairway", "", StrikeQualityClean, "")

	if err == nil {
		t.Errorf("Didnt receive error")
	}
}

func Test_AddShotAppendsToShotsTaken(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, err := hole.RecordShot("tee", club, "fairway", "", StrikeQualityClean, "")

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

	hole.RecordShot("tee", club, "fairway", "", StrikeQualityClean, "")
	hole.RecordShot("tee", club, "green", "", StrikeQualityClean, "")

	location := hole.CurrentLocation()
	if location != "green" {
		t.Errorf("[TestCurrentLocation] - Expected [fairway] Got [%v]", location)
	}
}

func Test_LookupShotByNumber(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, _ := hole.RecordShot("tee", club, "fairway", "", StrikeQualityClean, "")

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

	hole.RecordShot("tee", club, "OB", "", StrikeQualityClean, "")

	location := hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}

	hole.RecordShot("tee", club, "lost", "", StrikeQualityClean, "")
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

	hole.RecordShot(ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	lastValidShot := hole.GetLastValidShot()
	if lastValidShot.Result != LocationFairway {
		t.Errorf("[Test_LastValidShot] - expected [fairway] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot(ShotTypeApproach, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(ShotTypeApproach, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	lastValidShot = hole.GetLastValidShot()
	if lastValidShot.Result != LocationFairway {
		t.Errorf("[Test_LastValidShot] - expected [fairway] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot(ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")
	lastValidShot = hole.GetLastValidShot()
	if lastValidShot.Result != LocationGreen {
		t.Errorf("[Test_LastValidShot] - expected [green] got [%v]", lastValidShot.Result)
	}
}

func Test_ValidShots(t *testing.T) {
	shot := Shot{
		Result: LocationFairway,
	}

	checkValid := func(s Shot, expected bool) {
		if s.ValidShot() != expected {
			t.Errorf("[Test_ValidShots] - %v - Expected [%v] Got [%v]", s.Result, expected, !expected)
		}
	}

	checkValid(shot, true)

	shot.Result = LocationRough
	checkValid(shot, true)

	shot.Result = LocationBunker
	checkValid(shot, true)

	shot.Result = LocationOutOfBounds
	checkValid(shot, false)

	shot.Result = LocationGreen
	checkValid(shot, true)

}

func Test_CurrentLocationReturnsTeeForNewHole(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}

	location := hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}
}
