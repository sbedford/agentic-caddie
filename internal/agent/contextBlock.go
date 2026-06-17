package agent

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sbedford/agentic-caddie/internal/helpers"
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

func toContextString(agentInput GetAdviceRequest) string {

	var sb strings.Builder

	sb.WriteString("## Player Profile")
	sb.WriteString("Name: ")
	sb.WriteString(agentInput.Player.Name)

	holesPlayed := 3

	sb.WriteString("\n## Current Round")
	sb.WriteString("\nDaily Handicap: ")
	sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.DailyHandicap, 10))
	sb.WriteString("\nCompetition Type: ")
	sb.WriteString(agentInput.CurrentRound.CompetitionType.String)
	sb.WriteString("\nHoles Played: ")
	sb.WriteString(strconv.Itoa(holesPlayed))

	if *helpers.NullableString(agentInput.CurrentRound.CompetitionType) == "stableford" {
		sb.WriteString("\nPoints: ")
		if helpers.NullableInt64(agentInput.CurrentRound.TotalPoints) == nil {
			sb.WriteString("0")
			sb.WriteString("\nPoints Behind: 0")
		} else {
			sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.TotalPoints.Int64, 10))
			sb.WriteString("\nPoints Behind: ")
			sb.WriteString(strconv.Itoa((holesPlayed * 2) - int(agentInput.CurrentRound.TotalPoints.Int64)))
		}
	}

	if *helpers.NullableString(agentInput.CurrentRound.CompetitionType) == "stroke" {
		sb.WriteString("\nStrokes: ")
		if helpers.NullableInt64(agentInput.CurrentRound.TotalScore) == nil {
			sb.WriteString("0")
			sb.WriteString("\nTo Par: 0")
		} else {
			sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.TotalScore.Int64, 10))
			sb.WriteString("\nTo Par: 0")
		}
	}

	sb.WriteString("Available Clubs: \n")
	sb.WriteString("Format: C:Club RC:Reliable Carry AC:Average Carry D:Dispersion \n")
	for _, club := range agentInput.Clubs {
		fmt.Fprintf(&sb, "C:%v RC: %v D: %vm(%v)\n", club.ClubName, club.CarryReliable, club.DispersionAvgM, club.DispersionBias)
	}

	return sb.String()
}
