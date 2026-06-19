package models

import (
	"strconv"
	"testing"
	"time"
)

func Test_TeeGetHole_ReturnsHole(t *testing.T) {
	tee := Tee{
		ID:   1,
		Name: "white",
		Holes: []Hole{
			Hole{
				HoleNumber: 1,
				Par:        3,
				Distance:   1,
			},
			Hole{
				HoleNumber: 2,
				Par:        4,
				Distance:   2,
			},
			Hole{
				HoleNumber: 3,
				Par:        5,
				Distance:   3,
			},
		},
	}
	hole := tee.GetHole(1)

	if hole == nil {
		t.Errorf("Test_TeeGetHole_ReturnsHole - Nil Returned")
	}

	if hole.HoleNumber != 1 {
		t.Errorf("Test_TeeGetHole_ReturnsHole - Expected Hole [1] Got [%v]", strconv.FormatInt(hole.HoleNumber, 10))
	}

}

func Test_ProgressHole(t *testing.T) {

	course := Course{
		ID:   1,
		Name: "test",
		Tees: []Tee{
			Tee{
				ID:   1,
				Name: "white",
				Holes: []Hole{
					Hole{
						HoleNumber: 1,
						Par:        3,
						Distance:   1,
					},
					Hole{
						HoleNumber: 2,
						Par:        4,
						Distance:   2,
					},
					Hole{
						HoleNumber: 3,
						Par:        5,
						Distance:   3,
					},
				},
			},
		},
	}

	golfer := Player{ID: 1, Name: "Test", Handicap: 10}

	round := Round{
		ID:              1,
		Course:          course,
		Golfer:          golfer,
		Tee:             course.Tees[0],
		RoundDate:       time.Now(),
		DailyHandicap:   9,
		RoundType:       RoundTypeCompetition,
		CompetitionType: CompetitionTypeStableford,
	}

	hole1, err := round.ProgressHole()
	if err != nil || hole1 == nil {
		t.Fatalf("Test_ProgressHole - Hole 1 - Nil Returned")
	} else {
		if hole1.Hole.HoleNumber != 1 {
			t.Fatalf("Test_ProgressHole - Hole 1 - Expected Hole Number [1] Got [%v]", hole1.Hole.HoleNumber)
		}

		if hole1.DistanceToPin != 1 {
			t.Fatalf("Test_ProgressHole - Hole 1 - Expected DistanceToPin [1] Got [%v]", hole1.DistanceToPin)
		}
	}
	hole1.Completed = true

	hole2, err := round.ProgressHole()
	if err != nil {
		msg := err.Error()
		t.Fatalf("Got Error [%v]", msg)
	}
	if hole2 == nil {
		t.Fatalf("Test_ProgressHole - Hole 2 - Nil Returned")
	} else {
		if hole2.Hole.HoleNumber != 2 {
			t.Fatalf("Test_ProgressHole - Hole 2 - Expected Hole Number [2] Got [%v]", hole2.Hole.HoleNumber)
		}

		if hole2.DistanceToPin != 2 {
			t.Fatalf("Test_ProgressHole - Hole 2 - Expected DistanceToPin [2] Got [%v]", hole2.DistanceToPin)
		}
	}
	hole2.Completed = true

	hole3, err := round.ProgressHole()
	if err != nil {
		msg := err.Error()
		t.Fatalf("Got Error [%v]", msg)
	}
	if hole3 == nil {
		t.Fatalf("Test_ProgressHole - Hole 3 - Nil Returned")
	} else {
		if hole3.Hole.HoleNumber != 3 {
			t.Fatalf("Test_ProgressHole - Hole 3 - Expected Hole Number [3] Got [%v]", hole3.Hole.HoleNumber)
		}

		if hole3.DistanceToPin != 3 {
			t.Fatalf("Test_ProgressHole - Hole 3 - Expected DistanceToPin [3] Got [%v]", hole3.DistanceToPin)
		}
	}

}

func Test_InitialiseRound_ProgressesHole(t *testing.T) {
	golfer := Player{ID: 1, Name: "Test", Handicap: 10}
	course := Course{
		ID:   1,
		Name: "test",
		Tees: []Tee{
			Tee{
				ID:   1,
				Name: "white",
				Holes: []Hole{
					Hole{
						HoleNumber: 1,
						Par:        3,
					},
					Hole{
						HoleNumber: 2,
						Par:        4,
					},
					Hole{
						HoleNumber: 3,
						Par:        5,
					},
				},
			},
		},
	}

	round, err := InitialiseRound(course, golfer, course.Tees[0], time.Now(), 5, RoundTypeCompetition, CompetitionTypeStableford)

	if err != nil {
		msg := err.Error()
		t.Fatalf("Got Error [%v]", msg)
	}

	if len(round.PlayedHoles) != 1 {
		t.Errorf("Test_InitialiseRound_ProgressesHole - Expected PlayedHole [%v] Got [%v]", 1, len(round.PlayedHoles))
	}

}

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

	_, err := hole.RecordShot(250, "tee", club, "fairway", "", StrikeQualityClean, "")

	if err == nil {
		t.Errorf("Didnt receive error")
	}
}

func Test_AddShotAppendsToShotsTaken(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, err := hole.RecordShot(250, "tee", club, "fairway", "", StrikeQualityClean, "")

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

	hole.RecordShot(250, "tee", club, "fairway", "", StrikeQualityClean, "")
	hole.RecordShot(250, "tee", club, "green", "", StrikeQualityClean, "")

	location := hole.CurrentLocation()
	if location != "green" {
		t.Errorf("[TestCurrentLocation] - Expected [fairway] Got [%v]", location)
	}
}

func Test_LookupShotByNumber(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	shot, _ := hole.RecordShot(160, "tee", club, "fairway", "", StrikeQualityClean, "")

	s, _ := hole.LookupShot(int(shot.ShotNumber))
	if s == nil {
		t.Errorf("[Test_LookupShotByNumber] - Expected [Shot 1] Got [nil]")
	}

	_, err := hole.LookupShot(763)
	if err == nil {
		t.Errorf("[Test_LookupShotByNumber] - Expected [Shot 1] Got [nil]")
	}

}

func Test_CurrentLocationUsesLastShotIfOB(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}

	hole.RecordShot(160, "tee", club, "OB", "", StrikeQualityClean, "")

	location := hole.CurrentLocation()
	if location != "tee" {
		t.Errorf("[TestCurrentLocation] - Expected [tee] Got [%v]", location)
	}

	hole.RecordShot(160, "tee", club, "lost", "", StrikeQualityClean, "")
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

	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	lastValidShot := hole.GetLastValidShot()
	if lastValidShot.Result != LocationFairway {
		t.Errorf("[Test_LastValidShot] - expected [fairway] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot(160, ShotTypeApproach, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	lastValidShot = hole.GetLastValidShot()
	if lastValidShot.Result != LocationFairway {
		t.Errorf("[Test_LastValidShot] - expected [fairway] got [%v]", lastValidShot.Result)
	}

	hole.RecordShot(160, ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")
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

func Test_RoundTotalScore(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.CompleteHole(2)

	if hole.Score != 4 {
		t.Errorf("[Test_RoundTotalScore] - Expected [4] Got [%v]", hole.Score)
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 3}}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	hole.RecordPenalty(1)
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")
	hole.CompleteHole(2)

	if hole.Score != 6 {
		t.Errorf("[Test_RoundTotalScore] - Penalties - Expected [6] Got [%v]", hole.Score)
	}
}

func Test_Round_RecordPenalty(t *testing.T) {

	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordPenalty(4)

	if !hole.Penalty {
		t.Errorf("[Test_Round_RecordPenalty] - Expected [true] Got [false]")
	}

}

func Test_WipeHole(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeTee, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	hole.RecordWipe()

	if !hole.Wiped {
		t.Errorf("[Test_WipeHole] - Expected Wiped [true] Got [false]")
	}

	if !hole.Completed {
		t.Errorf("[Test_WipeHole] - Expected Completed [true] Got [false]")
	}

	if hole.Score != 0 {
		t.Errorf("[Test_WipeHole] - Expected Score [0] Got [%v]", hole.Score)
	}
}

func Test_FairwayHit(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")

	if hole.FairwayHit == false {
		t.Errorf("[Test_FairwayHit] - Tee in Fairway - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationOutOfBounds, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")

	if hole.FairwayHit == true {
		t.Errorf("[Test_FairwayHit] - After OB - Expected [false] Got [true]")
	}
}

func Test_GreenInRegulation(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3}, Completed: false}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Par 3 in 1 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeChip, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == true {
		t.Errorf("[Test_GreenInRegulation] - Par 3 Second Shot - Expected [false] Got [true]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 4}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Drove Par 4 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 4}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Par 4 in 2 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 4}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypePitch, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == true {
		t.Errorf("[Test_GreenInRegulation] - Par 4 in 3 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 5}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Drove Par 5 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 5}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Par 5 in 2 - Expected [true] Got [false]")
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 5}, Completed: false}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeLayup, club, LocationFairway, "", StrikeQualityClean, "")
	hole.RecordShot(160, ShotTypeApproach, club, LocationGreen, "", StrikeQualityClean, "")

	if hole.GreenInRegulation == false {
		t.Errorf("[Test_GreenInRegulation] - Par 5 in 2 - Expected [true] Got [false]")
	}
}

func Test_CalculatePoints(t *testing.T) {
	hole := PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3, StrokeIndex: 5}, Round: &Round{DailyHandicap: 5}}
	club := Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.CompleteHole(2)

	if hole.Points != 0 {
		t.Errorf("[Test_CalculatePoints] - No Points without Stableford - Expected [0] Got [%v]", hole.Points)
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3, StrokeIndex: 5}, Round: &Round{DailyHandicap: 5, CompetitionType: CompetitionTypeStableford}}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.CompleteHole(2)

	if hole.Points != 3 {
		t.Errorf("[Test_CalculatePoints] - Par on Hole with Shot - Expected [3] Got [%v]", hole.Points)
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3, StrokeIndex: 5}, Round: &Round{DailyHandicap: 5, CompetitionType: CompetitionTypeStableford}}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.CompleteHole(3)

	if hole.Points != 2 {
		t.Errorf("[Test_CalculatePoints] - Boegy on Hole with Shot - Expected [2] Got [%v]", hole.Points)
	}

	hole = PlayedHole{Hole: Hole{HoleNumber: 2, Par: 3, StrokeIndex: 6}, Round: &Round{DailyHandicap: 5, CompetitionType: CompetitionTypeStableford}}
	club = Club{ClubName: "Driver"}
	hole.RecordShot(160, ShotTypeTee, club, LocationGreen, "", StrikeQualityClean, "")
	hole.CompleteHole(3)

	if hole.Points != 1 {
		t.Errorf("[Test_CalculatePoints] - Boegy on Hole with No Shot - Expected [1] Got [%v]", hole.Points)
	}
}

func Test_ProgressHole_EstablishesCourseHoleId(t *testing.T) {
	golfer := Player{ID: 1, Name: "Test", Handicap: 10}
	course := Course{
		ID:   1,
		Name: "test",
		Tees: []Tee{
			Tee{
				ID:   1,
				Name: "white",
				Holes: []Hole{
					Hole{
						HoleNumber:   1,
						Par:          3,
						CourseHoleID: 1,
					},
					Hole{
						HoleNumber:   2,
						Par:          4,
						CourseHoleID: 2,
					},
					Hole{
						HoleNumber:   3,
						Par:          5,
						CourseHoleID: 3,
					},
				},
			},
		},
	}

	round, _ := InitialiseRound(course, golfer, course.Tees[0], time.Now(), 5, RoundTypeCompetition, CompetitionTypeStableford)

	hole1, err := round.CurrentHole()
	if err != nil {
		t.Errorf("Got Error - %v", err.Error())
	}
	if hole1 == nil {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 1 - Got Nil PlayedHole")
	}
	if hole1.Hole.CourseHoleID != 1 {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 1 - Expected CourseHoleID [1] Got [%v]", hole1.Hole.CourseHoleID)
	}
	hole1.Completed = true

	hole2, err := round.ProgressHole()
	if err != nil {
		t.Errorf("Got Error - %v", err.Error())
	}
	if hole2 == nil {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 2 - Got Nil PlayedHole")
	}
	if hole2.Hole.CourseHoleID != 2 {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 2 - Expected CourseHoleID [2] Got [%v]", hole2.Hole.CourseHoleID)
	}
	hole2.Completed = true

	hole3, err := round.ProgressHole()
	if err != nil {
		t.Errorf("Got Error - %v", err.Error())
	}
	if hole3 == nil {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 2 - Got Nil PlayedHole")
	}
	if hole3.Hole.CourseHoleID != 3 {
		t.Errorf("[Test_ProgressHole_EstablishesCourseHoleId] - Hole 2 - Expected CourseHoleID [3] Got [%v]", hole3.Hole.CourseHoleID)
	}
}
