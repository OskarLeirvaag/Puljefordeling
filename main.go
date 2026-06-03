// Package main is a runnable demo of the puljefordeling assignment algorithm.
//
// Usage:
//
//	go run .            # full weekend walkthrough (default)
//	go run . weekend    # same
//	go run . weights    # the priority-weight bands, live from the solver
//	go run . reroute    # participation pricing: a strong wish is NOT sacrificed
//	go run . scarcity   # backward-looking miss bonus boosts the unlucky
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"puljefordeling/internal/model"
	"puljefordeling/internal/scenario"
	"puljefordeling/internal/solver"
)

func main() {
	scenario := "weekend"
	if len(os.Args) > 1 {
		scenario = os.Args[1]
	}

	switch scenario {
	case "weekend":
		runWeekend()
	case "weights":
		runWeights()
	case "reroute":
		runReroute()
	case "scarcity":
		runScarcity()
	case "generate":
		runGenerate()
	default:
		fmt.Printf("unknown scenario %q — try: weekend | weights | reroute | scarcity | generate [players] [seed]\n", scenario)
		os.Exit(1)
	}
}

// --- scenario: generate (realistic scale) -----------------------------------

func runGenerate() {
	cfg := scenario.DefaultConfig()
	if len(os.Args) > 2 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			cfg.Players = n
		}
	}
	if len(os.Args) > 3 {
		if s, err := strconv.ParseUint(os.Args[3], 10, 64); err == nil {
			cfg.Seed = s
		}
	}

	weekend := scenario.Generate(cfg)

	totalSeats := 0
	for _, sl := range weekend.Slots {
		for _, ev := range sl.Events {
			totalSeats += ev.Capacity
		}
	}

	st := solver.NewState(2026, weekend)
	start := time.Now()
	var seated, undersub, noSeat int
	gotVeldig, gotMiddels, gotLitt := 0, 0, 0
	for _, slot := range weekend.Slots {
		res := st.SolveSlot(slot, weekend.Players)
		undersub += len(res.UndersubscribedEvents)
		noSeat += len(res.Unassigned)
		for evID, pids := range res.Assignments {
			seated += len(pids)
			for _, pid := range pids {
				switch scoreFor(weekend.Players, pid, slot.ID, evID) {
				case 5:
					gotVeldig++
				case 3:
					gotMiddels++
				case 1:
					gotLitt++
				}
			}
		}
	}
	elapsed := time.Since(start)

	fmt.Printf("Generated weekend (seed %d, α=%.1f):\n", cfg.Seed, cfg.PopularityExponent)
	fmt.Printf("  %d players · %d puljer × %d games · %d seats total\n",
		cfg.Players, cfg.Puljer, cfg.GamesPerPulje, totalSeats)
	fmt.Println()
	fmt.Printf("Solved all puljer in %s\n", elapsed.Round(time.Microsecond))
	fmt.Printf("  seats filled:        %d / %d\n", seated, totalSeats)
	fmt.Printf("  got a top choice:    %d players (%.0f%% of %d)\n",
		st.SatisfiedCount(), 100*float64(st.SatisfiedCount())/float64(cfg.Players), cfg.Players)
	fmt.Printf("  of seats given:      %d Veldig · %d Middels · %d Litt\n", gotVeldig, gotMiddels, gotLitt)
	fmt.Printf("  left without a seat: %d (this slot's interested losers)\n", noSeat)
	fmt.Printf("  undersubscribed games flagged: %d\n", undersub)
}

func scoreFor(players []model.Player, pid, slotID, eventID string) model.Score {
	for _, p := range players {
		if p.ID == pid {
			return p.Prefs[slotID][eventID]
		}
	}
	return 0
}

// --- scenario: weights ------------------------------------------------------

func runWeights() {
	fmt.Println("Priority weight = which (player, game) edge wins a seat.")
	fmt.Println("Bands are spaced by 200; bumps stay inside a band.")
	fmt.Println()
	row := func(label string, w int) { fmt.Printf("  %-46s %4d\n", label, w) }

	row("Unsatisfied · Veldig (top choice)", solver.Weight(5, false, false, false, 0))
	row("  + 1 missed pulje", solver.Weight(5, false, false, false, 1))
	row("  + 3 missed puljer (cap)", solver.Weight(5, false, false, false, 3))
	row("  + never seated yet", solver.Weight(5, false, true, false, 0))
	row("  + is a DM elsewhere", solver.Weight(5, false, false, true, 0))
	fmt.Println()
	row("Satisfied · Veldig (a 2nd top choice)", solver.Weight(5, true, false, false, 0))
	fmt.Println()
	row("Middels", solver.Weight(3, false, false, false, 0))
	row("  + DM", solver.Weight(3, false, false, true, 0))
	row("Litt", solver.Weight(1, false, false, false, 0))
	row("  + DM", solver.Weight(1, false, false, true, 0))
	fmt.Println()
	fmt.Printf("  Note: a regular unmet Veldig (%d) always beats the best a DM's\n", solver.Weight(5, false, false, false, 0))
	fmt.Printf("  Middels can be (%d) — the DM bump never crosses a band.\n", solver.Weight(3, false, true, true, 0))
	fmt.Println()
	fmt.Printf("  Participation bonus (value of seating one more person): %d\n", solver.ParticipationBonus)
}

// --- scenario: reroute (participation pricing) ------------------------------

func runReroute() {
	// X (3 seats) is wanted by three top-choice players AND one who barely wants
	// it. Y (1 seat) is wanted only by one of the top-choice players, as a
	// throwaway. Seating the barely-interested player would mean bumping a
	// top-choice player down to Y — a big welfare loss for a tiny gain.
	slot := model.Slot{
		ID: "s1", Name: "The contested slot",
		Events: []model.Event{
			{ID: "X", Name: "Popular game X", Capacity: 3},
			{ID: "Y", Name: "Quiet game Y", Capacity: 1},
		},
	}
	players := []model.Player{
		{ID: "ada", Name: "Ada", Prefs: prefs1("s1", map[string]model.Score{"X": 5})},
		{ID: "ben", Name: "Ben", Prefs: prefs1("s1", map[string]model.Score{"X": 5})},
		{ID: "cleo", Name: "Cleo", Prefs: prefs1("s1", map[string]model.Score{"X": 5, "Y": 1})},
		{ID: "dan", Name: "Dan (only mildly wants X)", Prefs: prefs1("s1", map[string]model.Score{"X": 1})},
	}

	res := solver.NewState(2026, model.Weekend{Slots: []model.Slot{slot}}).SolveSlot(slot, players)
	printSlotResult(slot, res)

	fmt.Println()
	fmt.Println("Why is Dan unseated and Y left empty?")
	fmt.Println("  To seat Dan in X we'd bump Cleo (top choice, 800) down to Y (Litt, 200):")
	fmt.Printf("  a 600-point loss to gain Dan's 200 — net −400. Participation is only\n")
	fmt.Printf("  worth %d, and 400 > %d, so the solver refuses the trade.\n", solver.ParticipationBonus, solver.ParticipationBonus)
	fmt.Println("  Nobody is dragged off a strong wish just to fill a chair.")
}

// --- scenario: scarcity (backward-looking miss bonus) -----------------------

func runScarcity() {
	// s1: a DM and an unlucky player both want A (1 seat). The DM bump wins it,
	// so the unlucky player records a MISS. s2: the unlucky player and a fresh
	// player both want B (1 seat). The miss bonus must tip it to the unlucky one.
	s1 := model.Slot{ID: "s1", Name: "Slot 1", Events: []model.Event{{ID: "A", Name: "Game A", Capacity: 1}}}
	s2 := model.Slot{ID: "s2", Name: "Slot 2", Events: []model.Event{{ID: "B", Name: "Game B", Capacity: 1}}}
	s3 := model.Slot{ID: "s3", Name: "Slot 3", Events: []model.Event{{ID: "Z", Name: "Game Z", Capacity: 4, DMID: "dm"}}}

	dm := model.Player{ID: "dm", Name: "Dee (DM elsewhere)", Prefs: prefs1("s1", map[string]model.Score{"A": 5})}
	unlucky := model.Player{ID: "unlucky", Name: "Uma (unlucky)", Prefs: map[string]map[string]model.Score{
		"s1": {"A": 5}, "s2": {"B": 5},
	}}
	fresh := model.Player{ID: "fresh", Name: "Finn (fresh)", Prefs: prefs1("s2", map[string]model.Score{"B": 5})}

	st := solver.NewState(2026, model.Weekend{Slots: []model.Slot{s1, s2, s3}})

	r1 := st.SolveSlot(s1, []model.Player{dm, unlucky})
	printSlotResult(s1, r1)
	fmt.Println("  → Uma wanted A but the DM bump won it. Uma records a miss.")
	fmt.Println()

	r2 := st.SolveSlot(s2, []model.Player{unlucky, fresh})
	printSlotResult(s2, r2)
	fmt.Println("  → Both want B equally, but Uma's earlier miss (+20) tips it to her.")
	fmt.Println("    The bonus is backward-looking on locked results — it can't be farmed.")
}

// --- scenario: weekend ------------------------------------------------------

func runWeekend() {
	weekend := buildWeekend()
	st := solver.NewState(2026, weekend)

	for _, slot := range weekend.Slots {
		result := st.SolveSlot(slot, weekend.Players)
		printSlotResult(slot, result)
		fmt.Printf("  Satisfaction so far: %d players have a top-choice seat\n\n", st.SatisfiedCount())
	}
}

func buildWeekend() model.Weekend {
	slots := []model.Slot{
		{ID: "fri-eve", Name: "Friday Evening", Events: []model.Event{
			{ID: "dnd", Name: "D&D: Curse of Strahd", Capacity: 4, DMID: "frank"},
			{ID: "pf", Name: "Pathfinder: Kingmaker", Capacity: 3},
		}},
		{ID: "sat-morn", Name: "Saturday Morning", Events: []model.Event{
			{ID: "coc", Name: "Call of Cthulhu", Capacity: 3, DMID: "frank"},
			{ID: "bitd", Name: "Blades in the Dark", Capacity: 4},
			{ID: "vamp", Name: "Vampire: the Masquerade", Capacity: 3},
		}},
		{ID: "sat-eve", Name: "Saturday Evening", Events: []model.Event{
			{ID: "ti4", Name: "Twilight Imperium", Capacity: 6},
			{ID: "gloom", Name: "Gloomhaven", Capacity: 4, DMID: "frank"},
		}},
		{ID: "sun-morn", Name: "Sunday Morning", Events: []model.Event{
			{ID: "oneshot", Name: "One-Shot Spectacular", Capacity: 4},
			{ID: "mage", Name: "Mage: the Ascension", Capacity: 3},
		}},
	}

	players := []model.Player{
		{ID: "alice", Name: "Alice", Prefs: map[string]map[string]model.Score{
			"fri-eve": {"dnd": 5, "pf": 1}, "sat-morn": {"coc": 5, "vamp": 3}, "sun-morn": {"oneshot": 3},
		}},
		{ID: "bob", Name: "Bob", Prefs: map[string]map[string]model.Score{
			"fri-eve": {"dnd": 5}, "sat-morn": {"bitd": 5, "coc": 3}, "sat-eve": {"gloom": 3},
		}},
		{ID: "carol", Name: "Carol", Prefs: map[string]map[string]model.Score{
			"fri-eve": {"dnd": 5, "pf": 3}, "sat-eve": {"ti4": 5, "gloom": 3}, "sun-morn": {"mage": 5},
		}},
		{ID: "dave", Name: "Dave", Prefs: map[string]map[string]model.Score{
			"fri-eve": {"dnd": 5}, "sat-morn": {"vamp": 5, "bitd": 1}, "sat-eve": {"ti4": 3, "gloom": 1},
		}},
		{ID: "eve", Name: "Eve", Prefs: map[string]map[string]model.Score{
			"fri-eve": {"pf": 5}, "sat-morn": {"coc": 3, "bitd": 5}, "sun-morn": {"oneshot": 5},
		}},
		{ID: "frank", Name: "Frank (DM ×3)", Prefs: map[string]map[string]model.Score{
			"sun-morn": {"oneshot": 5, "mage": 3},
		}},
		{ID: "grace", Name: "Grace", Prefs: map[string]map[string]model.Score{
			"sat-morn": {"coc": 5, "vamp": 3}, "sat-eve": {"ti4": 5}, "sun-morn": {"mage": 3},
		}},
		{ID: "henry", Name: "Henry (Sunday only)", Prefs: map[string]map[string]model.Score{
			"sun-morn": {"oneshot": 5, "mage": 3},
		}},
	}

	return model.Weekend{Slots: slots, Players: players}
}

// --- shared printing --------------------------------------------------------

func prefs1(slotID string, scores map[string]model.Score) map[string]map[string]model.Score {
	return map[string]map[string]model.Score{slotID: scores}
}

func printSlotResult(slot model.Slot, result model.SlotResult) {
	fmt.Printf("═══ %s ═══\n", slot.Name)
	for _, ev := range slot.Events {
		players := result.Assignments[ev.ID]
		dm := ""
		if ev.DMID != "" {
			dm = fmt.Sprintf("  [DM: %s]", ev.DMID)
		}
		seats := fmt.Sprintf("(%d/%d)", len(players), ev.Capacity)
		if len(players) == 0 {
			fmt.Printf("  %-28s %-7s %s  (empty)\n", ev.Name, seats, dm)
		} else {
			fmt.Printf("  %-28s %-7s %s  %s\n", ev.Name, seats, dm, strings.Join(players, ", "))
		}
	}
	if len(result.UndersubscribedEvents) > 0 {
		fmt.Printf("  ⚠ needs review (<3 players): %s\n", strings.Join(result.UndersubscribedEvents, ", "))
	}
	if len(result.Unassigned) > 0 {
		fmt.Printf("  no seat: %s\n", strings.Join(result.Unassigned, ", "))
	}
	if len(result.NewlySatisfied) > 0 {
		fmt.Printf("  newly satisfied: %s\n", strings.Join(result.NewlySatisfied, ", "))
	}
}
