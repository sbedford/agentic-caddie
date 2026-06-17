package agent

import (
	"strconv"
	"strings"
)

// Structure:
//
//	## Player Profile
//	Name: <name>
//	Current handicap: <handicap>
//	Handicap trend: <direction and magnitude over last N rounds>
//	Rounds in database: <count>
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
	sb.WriteString("\n Daily Handicap: ")
	sb.WriteString(strconv.FormatInt(agentInput.CurrentRound.DailyHandicap, 10))

	return sb.String()
}
