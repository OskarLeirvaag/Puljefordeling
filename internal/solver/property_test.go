package solver_test

import (
	"fmt"
	"sort"
	"testing"

	"puljefordeling/internal/model"
	"puljefordeling/internal/scenario"
	"puljefordeling/internal/solver"
)

// solveWeekend runs every slot and returns the per-slot results.
func solveWeekend(w model.Weekend) []model.SlotResult {
	st := solver.NewState(2026, w)
	out := make([]model.SlotResult, 0, len(w.Slots))
	for _, slot := range w.Slots {
		out = append(out, st.SolveSlot(slot, w.Players))
	}
	return out
}

// TestGeneratedInvariants fuzzes the solver over many generated weekends and
// asserts the hard guarantees hold on every one.
func TestGeneratedInvariants(t *testing.T) {
	for seed := uint64(0); seed < 60; seed++ {
		cfg := scenario.DefaultConfig()
		cfg.Seed = seed
		cfg.Players = 120
		w := scenario.Generate(cfg)

		ratedScore := func(pid, slotID, eventID string) model.Score {
			for _, p := range w.Players {
				if p.ID == pid {
					return p.Prefs[slotID][eventID]
				}
			}
			return 0
		}

		results := solveWeekend(w)
		for si, slot := range w.Slots {
			res := results[si]

			capacity := map[string]int{}
			dmHere := map[string]string{} // playerID -> event they DM this slot
			for _, ev := range slot.Events {
				capacity[ev.ID] = ev.Capacity
				if ev.DMID != "" {
					dmHere[ev.DMID] = ev.ID
				}
			}

			seenThisSlot := map[string]bool{}
			for evID, pids := range res.Assignments {
				if len(pids) > capacity[evID] {
					t.Fatalf("seed %d slot %s event %s: %d players > capacity %d",
						seed, slot.ID, evID, len(pids), capacity[evID])
				}
				for _, pid := range pids {
					if seenThisSlot[pid] {
						t.Fatalf("seed %d slot %s: %s assigned to more than one event", seed, slot.ID, pid)
					}
					seenThisSlot[pid] = true

					if ratedScore(pid, slot.ID, evID) == 0 {
						t.Fatalf("seed %d slot %s: %s seated in %s without any interest", seed, slot.ID, pid, evID)
					}
					if dmHere[pid] != "" {
						t.Fatalf("seed %d slot %s: GM %s (runs %s) was seated as a player in %s",
							seed, slot.ID, pid, dmHere[pid], evID)
					}
				}
			}
		}
	}
}

// TestGeneratedDeterminism: the same generated weekend solved twice must yield
// byte-identical assignments (seeded shuffle + seeded generation).
func TestGeneratedDeterminism(t *testing.T) {
	for seed := uint64(0); seed < 20; seed++ {
		cfg := scenario.DefaultConfig()
		cfg.Seed = seed
		cfg.Players = 80

		a := solveWeekend(scenario.Generate(cfg))
		b := solveWeekend(scenario.Generate(cfg))

		if fingerprint(a) != fingerprint(b) {
			t.Fatalf("seed %d: solving the same weekend twice gave different assignments", seed)
		}
	}
}

// fingerprint renders assignments into a stable string (event IDs sorted so the
// fingerprint itself doesn't depend on Go's map iteration order).
func fingerprint(results []model.SlotResult) string {
	s := ""
	for _, r := range results {
		s += r.SlotID + "|"
		evIDs := make([]string, 0, len(r.Assignments))
		for evID := range r.Assignments {
			evIDs = append(evIDs, evID)
		}
		sort.Strings(evIDs)
		for _, evID := range evIDs {
			s += fmt.Sprintf("%s=%v;", evID, r.Assignments[evID])
		}
		s += "\n"
	}
	return s
}
