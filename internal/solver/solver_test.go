package solver

import (
	"fmt"
	"slices"
	"testing"

	"puljefordeling/internal/model"
)

// --- helpers ----------------------------------------------------------------

func slot(id string, events ...model.Event) model.Slot {
	return model.Slot{ID: id, Name: id, Events: events}
}

func event(id string, capacity int) model.Event {
	return model.Event{ID: id, Name: id, Capacity: capacity}
}

func player(id string, prefs map[string]map[string]model.Score) model.Player {
	return model.Player{ID: id, Name: id, Prefs: prefs}
}

func prefs(slotID string, scores map[string]model.Score) map[string]map[string]model.Score {
	return map[string]map[string]model.Score{slotID: scores}
}

func assigned(result model.SlotResult, eventID string) []string {
	return result.Assignments[eventID]
}

// --- adjustScore ------------------------------------------------------------

func TestAdjustScore_SatisfiedAlwaysRaw(t *testing.T) {
	for _, score := range []model.Score{1, 2, 3, 4, 5} {
		got := adjustScore(score, true, true)
		if got != int(score) {
			t.Errorf("satisfied score %d: want %d, got %d", score, score, got)
		}
	}
}

func TestAdjustScore_UnsatisfiedFiveIsAlwaysTen(t *testing.T) {
	if got := adjustScore(5, false, false); got != 10 {
		t.Errorf("unsatisfied 5 no boost: want 10, got %d", got)
	}
	if got := adjustScore(5, false, true); got != 10 {
		t.Errorf("unsatisfied 5 with boost: want 10, got %d", got)
	}
}

func TestAdjustScore_LateBoostDoublesLowerScores(t *testing.T) {
	for _, score := range []model.Score{1, 2, 3, 4} {
		got := adjustScore(score, false, true)
		want := int(score) * 2
		if got != want {
			t.Errorf("unsatisfied score %d with boost: want %d, got %d", score, want, got)
		}
	}
}

func TestAdjustScore_NoBoostLowerScoresUnchanged(t *testing.T) {
	for _, score := range []model.Score{1, 2, 3, 4} {
		got := adjustScore(score, false, false)
		if got != int(score) {
			t.Errorf("unsatisfied score %d no boost: want %d, got %d", score, score, got)
		}
	}
}

// --- SolveSlot --------------------------------------------------------------

func TestSolveSlot_BasicAssignment(t *testing.T) {
	// Two players each prefer a different event — both should get their pick.
	sl := slot("s1", event("A", 1), event("B", 1))
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 5})),
		player("bob", prefs("s1", map[string]model.Score{"B": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if !slices.Contains(assigned(result, "A"), "alice") {
		t.Error("alice should be assigned to A")
	}
	if !slices.Contains(assigned(result, "B"), "bob") {
		t.Error("bob should be assigned to B")
	}
	if len(result.Unassigned) != 0 {
		t.Errorf("expected no unassigned, got %v", result.Unassigned)
	}
}

func TestSolveSlot_CapacityRespected(t *testing.T) {
	// Three players all want the same event with capacity 2.
	sl := slot("s1", event("A", 2))
	players := []model.Player{
		player("p1", prefs("s1", map[string]model.Score{"A": 5})),
		player("p2", prefs("s1", map[string]model.Score{"A": 5})),
		player("p3", prefs("s1", map[string]model.Score{"A": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if len(assigned(result, "A")) != 2 {
		t.Errorf("event A capacity 2: want 2 assigned, got %d", len(assigned(result, "A")))
	}
	if len(result.Unassigned) != 1 {
		t.Errorf("want 1 unassigned, got %d", len(result.Unassigned))
	}
}

func TestSolveSlot_NoInterestSkipped(t *testing.T) {
	// One player has interest, one does not.
	sl := slot("s1", event("A", 2))
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 4})),
		player("bob", map[string]map[string]model.Score{}), // no interest in s1
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if !slices.Contains(assigned(result, "A"), "alice") {
		t.Error("alice should be assigned")
	}
	if slices.Contains(assigned(result, "A"), "bob") {
		t.Error("bob has no interest and should not be assigned")
	}
}

func TestSolveSlot_HighScoreWinsContention(t *testing.T) {
	// Capacity 1: player with score 5 should beat player with score 3.
	sl := slot("s1", event("A", 1))
	players := []model.Player{
		player("low", prefs("s1", map[string]model.Score{"A": 3})),
		player("high", prefs("s1", map[string]model.Score{"A": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if !slices.Contains(assigned(result, "A"), "high") {
		t.Error("high scorer should win the seat")
	}
	if !slices.Contains(result.Unassigned, "low") {
		t.Error("low scorer should be unassigned")
	}
}

func TestSolveSlot_SatisfactionTracked(t *testing.T) {
	// Player gets a score-5 event → should appear in NewlySatisfied.
	sl := slot("s1", event("A", 2))
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 5})),
		player("bob", prefs("s1", map[string]model.Score{"A": 3})),
	}

	st := NewState(2026)
	result := st.SolveSlot(sl, players, false)

	if !slices.Contains(result.NewlySatisfied, "alice") {
		t.Error("alice should be newly satisfied")
	}
	if slices.Contains(result.NewlySatisfied, "bob") {
		t.Error("bob scored 3, not 5 — should not be satisfied")
	}
	if !st.IsSatisfied("alice") {
		t.Error("state should mark alice as satisfied")
	}
}

func TestSolveSlot_SatisfiedPlayerDeprioritised(t *testing.T) {
	// Slot 1: alice gets her score-5 event → satisfied.
	// Slot 2: same event, capacity 1. alice (satisfied, score 5)
	// vs charlie (unsatisfied, score 5). Charlie should win.
	sl1 := slot("s1", event("A", 1))
	sl2 := slot("s2", event("A", 1))

	alice := model.Player{
		ID:   "alice",
		Name: "alice",
		Prefs: map[string]map[string]model.Score{
			"s1": {"A": 5},
			"s2": {"A": 5},
		},
	}
	charlie := model.Player{
		ID:   "charlie",
		Name: "charlie",
		Prefs: map[string]map[string]model.Score{
			"s2": {"A": 5},
		},
	}

	st := NewState(2026)
	st.SolveSlot(sl1, []model.Player{alice}, false)

	if !st.IsSatisfied("alice") {
		t.Fatal("alice should be satisfied after slot 1")
	}

	result := st.SolveSlot(sl2, []model.Player{alice, charlie}, false)

	if !slices.Contains(assigned(result, "A"), "charlie") {
		t.Error("unsatisfied charlie should win the seat over satisfied alice")
	}
	if slices.Contains(assigned(result, "A"), "alice") {
		t.Error("satisfied alice should lose the seat to unsatisfied charlie")
	}
}

func TestSolveSlot_LateBoostPrioritisesUnsatisfied(t *testing.T) {
	// Without boost: satisfied alice (score 5) beats unsatisfied bob (score 4).
	// With boost:    unsatisfied bob (score 4→8) beats satisfied alice (score 5).
	sl := slot("s1", event("A", 1))
	alice := model.Player{
		ID:   "alice",
		Name: "alice",
		Prefs: map[string]map[string]model.Score{
			"s1": {"A": 5},
		},
	}
	bob := model.Player{
		ID:   "bob",
		Name: "bob",
		Prefs: map[string]map[string]model.Score{
			"s1": {"A": 4},
		},
	}

	// Without boost: alice wins.
	stNoBoost := NewState(2026)
	stNoBoost.satisfied["alice"] = true // pre-satisfy alice
	resultNoBoost := stNoBoost.SolveSlot(sl, []model.Player{alice, bob}, false)
	if !slices.Contains(assigned(resultNoBoost, "A"), "alice") {
		t.Error("without boost: satisfied alice (score 5) should beat unsatisfied bob (score 4)")
	}

	// With boost: bob wins (4×2=8 > 5).
	stBoost := NewState(2026)
	stBoost.satisfied["alice"] = true // pre-satisfy alice
	resultBoost := stBoost.SolveSlot(sl, []model.Player{alice, bob}, true)
	if !slices.Contains(assigned(resultBoost, "A"), "bob") {
		t.Error("with boost: unsatisfied bob (score 4→8) should beat satisfied alice (score 5)")
	}
}

func TestSolveSlot_TotalScoreIsUnadjusted(t *testing.T) {
	// TotalScore must reflect actual preference scores, not adjusted ones.
	sl := slot("s1", event("A", 2))
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 5})),
		player("bob", prefs("s1", map[string]model.Score{"A": 3})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if result.TotalScore != 8 {
		t.Errorf("TotalScore: want 8 (5+3), got %d", result.TotalScore)
	}
}

func TestSolveSlot_EmptySlot(t *testing.T) {
	sl := slot("s1", event("A", 4))
	result := NewState(2026).SolveSlot(sl, []model.Player{}, false)

	if len(result.Assignments) != 0 {
		t.Error("no players means no assignments")
	}
}

func eventMin(id string, capacity, minPlayers int) model.Event {
	return model.Event{ID: id, Name: id, Capacity: capacity, MinPlayers: minPlayers}
}

func TestSolveSlot_UndersubscribedEventCancelled(t *testing.T) {
	// Event A requires 3 players but only 1 is interested.
	// Event B has no minimum. The 1 player should end up on B after A is cancelled.
	sl := model.Slot{
		ID:   "s1",
		Name: "s1",
		Events: []model.Event{
			eventMin("A", 4, 3),
			event("B", 4),
		},
	}
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 5, "B": 3})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if !slices.Contains(result.CancelledEvents, "A") {
		t.Error("event A should be cancelled (1 player < MinPlayers 3)")
	}
	if !slices.Contains(assigned(result, "B"), "alice") {
		t.Error("alice should be reassigned to B after A is cancelled")
	}
}

func TestSolveSlot_CancellationCascades(t *testing.T) {
	// Event A requires 2 players, B requires 2 players.
	// Only 3 players total, all preferring A.
	// First run: A gets 3, B gets 0 → B cancelled.
	// Re-run with only A: A gets 3 ≥ 2 → stable.
	sl := model.Slot{
		ID:   "s1",
		Name: "s1",
		Events: []model.Event{
			eventMin("A", 4, 2),
			eventMin("B", 4, 2),
		},
	}
	players := []model.Player{
		player("p1", prefs("s1", map[string]model.Score{"A": 5})),
		player("p2", prefs("s1", map[string]model.Score{"A": 5})),
		player("p3", prefs("s1", map[string]model.Score{"A": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if !slices.Contains(result.CancelledEvents, "B") {
		t.Error("event B should be cancelled (no players)")
	}
	if !slices.Contains(result.CancelledEvents, "A") {
		// A has 3 players ≥ MinPlayers 2, should survive
		if len(assigned(result, "A")) < 2 {
			t.Error("event A should run with at least 2 players")
		}
	}
}

func TestSolveSlot_TopScoreAlwaysBeatsFallback(t *testing.T) {
	// 10 players all have X=5 (adjusted to 10). X has capacity for all 10.
	// 4 of them also have Y=4. Nobody should be routed to Y while X has room —
	// the MCMF must never trade a score-10 edge for a score-4 edge.
	sl := model.Slot{
		ID:   "s1",
		Name: "s1",
		Events: []model.Event{
			event("X", 10),
			event("Y", 4),
		},
	}
	players := make([]model.Player, 10)
	for i := range players {
		p := model.Player{
			ID:   fmt.Sprintf("p%d", i),
			Name: fmt.Sprintf("p%d", i),
			Prefs: map[string]map[string]model.Score{
				"s1": {"X": 5},
			},
		}
		if i < 4 {
			p.Prefs["s1"]["Y"] = 4
		}
		players[i] = p
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if len(assigned(result, "X")) != 10 {
		t.Errorf("all 10 players should be on X (capacity allows it), got %d", len(assigned(result, "X")))
	}
	if len(assigned(result, "Y")) != 0 {
		t.Errorf("no player should be routed to Y while X has capacity, got %v", assigned(result, "Y"))
	}
}

func TestSolveSlot_TopScoreLoserGetsFallback(t *testing.T) {
	// X has capacity 4, 6 players want it at score 5. The 2 who lose the
	// lottery and also have Y=4 should be reassigned to Y, not left unassigned.
	sl := model.Slot{
		ID:   "s1",
		Name: "s1",
		Events: []model.Event{
			event("X", 4),
			event("Y", 4),
		},
	}
	players := []model.Player{
		player("p1", prefs("s1", map[string]model.Score{"X": 5, "Y": 4})),
		player("p2", prefs("s1", map[string]model.Score{"X": 5, "Y": 4})),
		player("p3", prefs("s1", map[string]model.Score{"X": 5})),
		player("p4", prefs("s1", map[string]model.Score{"X": 5})),
		player("p5", prefs("s1", map[string]model.Score{"X": 5})),
		player("p6", prefs("s1", map[string]model.Score{"X": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if len(assigned(result, "X")) != 4 {
		t.Errorf("X capacity 4: want 4 assigned, got %d", len(assigned(result, "X")))
	}
	// p1 and p2 have a fallback; whichever of them lost the X lottery should
	// end up on Y, not unassigned.
	onY := assigned(result, "Y")
	onX := assigned(result, "X")
	for _, pid := range []string{"p1", "p2"} {
		if !slices.Contains(onX, pid) && !slices.Contains(onY, pid) {
			t.Errorf("%s lost the X lottery but was not reassigned to Y", pid)
		}
	}
}

func TestSolveSlot_MinPlayersZeroMeansNoMinimum(t *testing.T) {
	// MinPlayers == 0: event should run even with 1 player.
	sl := model.Slot{
		ID:   "s1",
		Name: "s1",
		Events: []model.Event{
			event("A", 4), // MinPlayers defaults to 0
		},
	}
	players := []model.Player{
		player("alice", prefs("s1", map[string]model.Score{"A": 5})),
	}

	result := NewState(2026).SolveSlot(sl, players, false)

	if len(result.CancelledEvents) != 0 {
		t.Errorf("no cancellations expected, got %v", result.CancelledEvents)
	}
	if !slices.Contains(assigned(result, "A"), "alice") {
		t.Error("alice should be assigned to A")
	}
}
