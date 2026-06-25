package agent

import (
	"fmt"
	"strconv"
	"strings"
)

// Structure:
//
//	## Player Profile
//	Name: <name>
//	Current handicap: <handicap>
//	Date range: <earliest round> to <most recent round>
//
//	## Recent Form
//	Last 5 rounds: <scores and courses, most recent first>
//	Recent scoring average: <N>
//	Recent GIR%: <N>
//	Recent fairways hit%: <N>
//	Recent putts per round: <N>
//
//	Curent Round Performance
//
//	## Known Tendencies
//	<Derived from game model — populated progressively from Stage 2 onwards.>
//	<Empty or minimal in Stage 1.>
//
//	## Courses Played
//	<List of courses in the database with round count at each.>

func toContextString(agentInput GetHoleStrategyRequest) string {

	var sb strings.Builder

	sb.WriteString("## Player Profile")
	sb.WriteString("Name: ")
	sb.WriteString(agentInput.Player.Name)

	sb.WriteString("\n## Current Round")

	sb.WriteString("\nCourse Id: ")
	sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.Course.ID, 10))
	sb.WriteString("\nTees: ")
	sb.WriteString(agentInput.CurrentRound.Tee.Name)

	sb.WriteString("\nDaily Handicap: ")
	sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.DailyHandicap, 10))
	sb.WriteString("\nCompetition Type: ")
	sb.WriteString(string(agentInput.CurrentRound.CompetitionType))
	sb.WriteString("\nHoles Played: ")
	sb.WriteString(strconv.Itoa(len(agentInput.CurrentRound.PlayedHoles)))

	if agentInput.CurrentRound.IsStableford() {
		sb.WriteString("\nPoints: ")
		sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.TotalPoints, 10))
		sb.WriteString("\nPoints Behind: ")
		sb.WriteString(strconv.FormatInt(int64(agentInput.CurrentRound.PointsBehind()), 10))
	}

	if agentInput.CurrentRound.IsStroke() {
		sb.WriteString("\nStrokes: ")
		sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.TotalScore, 10))
		sb.WriteString("\nStrokes Above Par: ")
		sb.WriteString(strconv.FormatInt(int64(agentInput.CurrentRound.StrokesAbovePar()), 10))
	}

	sb.WriteString("\n## Available Clubs: \n")
	sb.WriteString("Format: C:Club RC:Reliable Carry AC:Average Carry D:Dispersion \n")
	for _, club := range agentInput.Player.Clubs {
		fmt.Fprintf(&sb, "C:%v RC: %v D: %vm(%v)\n", club.ClubName, club.CarryReliable, club.DispersionAvgM, club.DispersionBias)
	}

	return sb.String()
}
