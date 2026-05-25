// Package main is a demo of the puljefordeling assignment algorithm.
package main

import (
	"fmt"
	"strings"

	"puljefordeling/internal/model"
	"puljefordeling/internal/solver"
)

func main() {
	weekend := buildWeekend()

	st := solver.NewState(2026)

	for i, slot := range weekend.Slots {
		// The organiser enables the late boost for the last slot in this demo.
		lateBoost := i == len(weekend.Slots)-1

		result := st.SolveSlot(slot, weekend.Players, lateBoost)
		printResult(slot, result, lateBoost)

		fmt.Printf("  Satisfaction: %d / %d players have received a score-5 event\n\n",
			st.SatisfiedCount(), countPlayersWithFive(weekend.Players))
	}
}

// buildWeekend returns a small but illustrative weekend scenario.
func buildWeekend() model.Weekend {
	slots := []model.Slot{
		{
			ID:   "fri-eve",
			Name: "Friday Evening",
			Events: []model.Event{
				{ID: "dnd", Name: "D&D: Curse of Strahd", Capacity: 4},
				{ID: "pf", Name: "Pathfinder: Kingmaker", Capacity: 3},
			},
		},
		{
			ID:   "sat-morn",
			Name: "Saturday Morning",
			Events: []model.Event{
				{ID: "coc", Name: "Call of Cthulhu", Capacity: 3},
				{ID: "bitd", Name: "Blades in the Dark", Capacity: 4},
				{ID: "vamp", Name: "Vampire: the Masquerade", Capacity: 3},
			},
		},
		{
			ID:   "sat-eve",
			Name: "Saturday Evening",
			Events: []model.Event{
				{ID: "ti4", Name: "Twilight Imperium", Capacity: 6, MinPlayers: 4},
				{ID: "gloom", Name: "Gloomhaven", Capacity: 4, MinPlayers: 2},
			},
		},
		{
			ID:   "sun-morn",
			Name: "Sunday Morning",
			Events: []model.Event{
				{ID: "oneshot", Name: "One-Shot Spectacular", Capacity: 4},
				{ID: "mage", Name: "Mage: the Ascension", Capacity: 3},
			},
		},
	}

	players := []model.Player{
		{
			ID: "alice", Name: "Alice",
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"dnd": 5, "pf": 2},
				"sat-morn": {"coc": 5, "vamp": 3},
				"sun-morn": {"oneshot": 4},
			},
		},
		{
			ID: "bob", Name: "Bob",
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"dnd": 5},
				"sat-morn": {"bitd": 5, "coc": 3},
				"sat-eve":  {"gloom": 4},
			},
		},
		{
			ID: "carol", Name: "Carol",
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"dnd": 5, "pf": 4},
				"sat-eve":  {"ti4": 5, "gloom": 3},
				"sun-morn": {"mage": 5},
			},
		},
		{
			ID: "dave", Name: "Dave",
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"dnd": 5},
				"sat-morn": {"vamp": 5, "bitd": 2},
				"sat-eve":  {"ti4": 3, "gloom": 2},
			},
		},
		{
			ID: "eve", Name: "Eve",
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"pf": 5},
				"sat-morn": {"coc": 4, "bitd": 5},
				"sun-morn": {"oneshot": 5},
			},
		},
		{
			ID: "frank", Name: "Frank",
			// Frank has no score-5 anywhere — will never be "satisfied"
			// but should still get good assignments.
			Prefs: map[string]map[string]model.Score{
				"fri-eve":  {"pf": 4, "dnd": 3},
				"sat-morn": {"bitd": 4},
				"sat-eve":  {"gloom": 4, "ti4": 3},
				"sun-morn": {"oneshot": 4},
			},
		},
		{
			ID: "grace", Name: "Grace",
			Prefs: map[string]map[string]model.Score{
				"sat-morn": {"coc": 5, "vamp": 4},
				"sat-eve":  {"ti4": 5},
				"sun-morn": {"mage": 4},
			},
		},
		{
			ID: "henry", Name: "Henry",
			// Henry only shows up Sunday — missed the whole weekend prior.
			Prefs: map[string]map[string]model.Score{
				"sun-morn": {"oneshot": 5, "mage": 3},
			},
		},
	}

	return model.Weekend{Slots: slots, Players: players}
}

// countPlayersWithFive returns how many players gave a score of 5 to at least
// one event anywhere in the weekend — these are the only players the primary
// objective applies to.
func countPlayersWithFive(players []model.Player) int {
	n := 0
	for _, p := range players {
		for _, slotPrefs := range p.Prefs {
			for _, score := range slotPrefs {
				if score == model.MaxScore {
					n++
					goto next
				}
			}
		}
	next:
	}
	return n
}

func printResult(slot model.Slot, result model.SlotResult, lateBoost bool) {
	boost := ""
	if lateBoost {
		boost = " [LATE BOOST ON]"
	}
	fmt.Printf("═══ %s%s ═══\n", slot.Name, boost)

	for _, ev := range slot.Events {
		players := result.Assignments[ev.ID]
		if len(players) == 0 {
			fmt.Printf("  %-30s  (empty)\n", ev.Name)
		} else {
			fmt.Printf("  %-30s  %s\n", ev.Name, strings.Join(players, ", "))
		}
	}

	if len(result.CancelledEvents) > 0 {
		fmt.Printf("  Cancelled (undersubscribed): %s\n", strings.Join(result.CancelledEvents, ", "))
	}
	if len(result.Unassigned) > 0 {
		fmt.Printf("  Unassigned: %s\n", strings.Join(result.Unassigned, ", "))
	}
	if len(result.NewlySatisfied) > 0 {
		fmt.Printf("  Newly satisfied: %s\n", strings.Join(result.NewlySatisfied, ", "))
	}
	fmt.Printf("  Score total: %d  |  Tie-breaking seed: %d\n", result.TotalScore, result.Seed)
}
