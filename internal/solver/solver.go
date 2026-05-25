package solver // solver is defined in flow.go

import (
	"math/rand/v2"
	"sort"

	"puljefordeling/internal/model"
)

// State carries satisfaction data and seeding parameters forward across slots.
// A player is satisfied once they have received at least one assignment
// to an event they scored 5.
type State struct {
	satisfied map[string]bool
	year      int
	slotIndex int
}

// NewState returns a fresh State for the start of a weekend.
// year is used to derive per-slot tie-breaking seeds (seed = year×1000 + slotIndex),
// making results deterministic within a year but different across years.
func NewState(year int) *State {
	return &State{
		satisfied: make(map[string]bool),
		year:      year,
	}
}

// IsSatisfied reports whether playerID has received a score-5 assignment.
func (s *State) IsSatisfied(playerID string) bool {
	return s.satisfied[playerID]
}

// SatisfiedCount returns the number of satisfied players.
func (s *State) SatisfiedCount() int {
	return len(s.satisfied)
}

// SolveSlot assigns players to events for one slot, updates the satisfaction
// state, and returns the result.
//
// lateBoost, when true, doubles all scores for unsatisfied players (in
// addition to the permanent score-5 doubling that always applies).
//
// Events whose MinPlayers threshold is not met are cancelled and removed
// iteratively: the slot is re-solved after each cancellation until no further
// events need to be cancelled.
func (s *State) SolveSlot(slot model.Slot, players []model.Player, lateBoost bool) model.SlotResult {
	seed := int64(s.year)*1000 + int64(s.slotIndex)
	s.slotIndex++

	result := model.SlotResult{
		SlotID:      slot.ID,
		Assignments: make(map[string][]string),
		Seed:        seed,
	}

	// Only consider players who expressed at least one interest in this slot.
	interested := make([]model.Player, 0, len(players))
	for _, p := range players {
		if len(p.Prefs[slot.ID]) > 0 {
			interested = append(interested, p)
		}
	}
	if len(interested) == 0 {
		return result
	}

	// Shuffle players before building the graph so that equal-score ties are
	// broken randomly but reproducibly. The seed is year-derived, so results
	// are consistent within one event year and different next year.
	rng := rand.New(rand.NewPCG(uint64(seed), 0)) //nolint:gosec
	rng.Shuffle(len(interested), func(i, j int) {
		interested[i], interested[j] = interested[j], interested[i]
	})

	// Iteratively solve and cancel undersubscribed events until stable.
	activeEvents := make([]model.Event, len(slot.Events))
	copy(activeEvents, slot.Events)

	var assignments map[string][]string
	for {
		assignments = s.runMCMF(slot.ID, activeEvents, interested, lateBoost)

		cancelled := false
		remaining := activeEvents[:0]
		for _, ev := range activeEvents {
			if ev.MinPlayers > 0 && len(assignments[ev.ID]) < ev.MinPlayers {
				cancelled = true
				result.CancelledEvents = append(result.CancelledEvents, ev.ID)
				continue
			}
			remaining = append(remaining, ev)
		}
		activeEvents = remaining

		if !cancelled {
			break
		}
	}

	result.Assignments = assignments

	// Update satisfaction state and collect totals.
	assigned := make(map[string]bool, len(interested))
	for evID, playerIDs := range assignments {
		for _, pid := range playerIDs {
			score := s.lookupScore(pid, slot.ID, evID, interested)
			result.TotalScore += int(score)
			if score == model.MaxScore && !s.satisfied[pid] {
				s.satisfied[pid] = true
				result.NewlySatisfied = append(result.NewlySatisfied, pid)
			}
			assigned[pid] = true
		}
	}

	// Collect unassigned players (had interest but no seat was available).
	for _, p := range interested {
		if !assigned[p.ID] {
			result.Unassigned = append(result.Unassigned, p.ID)
		}
	}

	// Sort all output slices for deterministic results.
	sort.Strings(result.NewlySatisfied)
	sort.Strings(result.Unassigned)
	sort.Strings(result.CancelledEvents)
	for evID := range result.Assignments {
		sort.Strings(result.Assignments[evID])
	}

	return result
}

// runMCMF builds and solves the flow network for the given events and players,
// returning the raw assignment map (eventID -> []playerID, unsorted).
func (s *State) runMCMF(slotID string, events []model.Event, players []model.Player, lateBoost bool) map[string][]string {
	assignments := make(map[string][]string)
	if len(events) == 0 {
		return assignments
	}

	// Node layout:
	//   0           → source
	//   1 .. P      → one node per player
	//   P+1 .. P+E  → one node per event
	//   P+E+1       → sink
	P := len(players)
	E := len(events)
	source := 0
	sink := P + E + 1
	g := newFlowGraph(sink + 1)

	for i := range players {
		g.addEdge(source, i+1, 1, 0)
	}

	eventIdx := make(map[string]int, E)
	for j, ev := range events {
		eventIdx[ev.ID] = j
		g.addEdge(P+1+j, sink, ev.Capacity, 0)
	}

	for i, p := range players {
		for evID, rawScore := range p.Prefs[slotID] {
			j, ok := eventIdx[evID]
			if !ok {
				continue
			}
			adj := adjustScore(rawScore, s.satisfied[p.ID], lateBoost)
			g.addEdge(i+1, P+1+j, 1, -adj)
		}
	}

	g.minCostFlow(source, sink)

	for i, p := range players {
		for _, eid := range g.adj[i+1] {
			e := g.edges[eid]
			if e.flow != 1 || e.to < P+1 || e.to > P+E {
				continue
			}
			evID := events[e.to-P-1].ID
			assignments[evID] = append(assignments[evID], p.ID)
		}
	}

	return assignments
}

// lookupScore returns the raw preference score for a player in a slot/event.
func (s *State) lookupScore(playerID, slotID, eventID string, players []model.Player) model.Score {
	for _, p := range players {
		if p.ID == playerID {
			return p.Prefs[slotID][eventID]
		}
	}
	return 0
}

// adjustScore returns the adjusted edge weight for a (player, event) pair.
// Unsatisfied players have their score-5 edges doubled (5→10) always.
// When lateBoost is true, all edges for unsatisfied players are doubled.
func adjustScore(score model.Score, satisfied, lateBoost bool) int {
	if satisfied {
		return int(score)
	}
	if score == model.MaxScore {
		return int(model.MaxScore) * 2
	}
	if lateBoost {
		return int(score) * 2
	}
	return int(score)
}
