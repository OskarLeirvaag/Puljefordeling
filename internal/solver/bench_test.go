package solver_test

import (
	"fmt"
	"testing"

	"puljefordeling/internal/model"
	"puljefordeling/internal/scenario"
	"puljefordeling/internal/solver"
)

// BenchmarkSolveWeekend pre-generates a pool of distinct large weekends per
// size (generation is excluded from the timer), then times solving the whole
// weekend. b.Loop() autoscales the iteration count.
func BenchmarkSolveWeekend(b *testing.B) {
	sizes := []int{100, 250, 500, 1000, 2000}
	const pool = 8

	for _, n := range sizes {
		// Pre-generate `pool` distinct weekends at this size. Games scale with
		// players so the weekend stays roughly balanced (seats ≈ attendees).
		weekends := make([]model.Weekend, pool)
		for i := range weekends {
			cfg := scenario.DefaultConfig()
			cfg.Players = n
			cfg.GamesPerPulje = n/6 + 6
			cfg.Seed = uint64(1000 + i)
			weekends[i] = scenario.Generate(cfg)
		}

		b.Run(fmt.Sprintf("players=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				w := weekends[i%pool]
				i++
				st := solver.NewState(2026, w)
				for _, slot := range w.Slots {
					st.SolveSlot(slot, w.Players)
				}
			}
		})
	}
}

// BenchmarkSolveSlot times a single (largest) slot in isolation — the core
// min-cost-flow run, without the per-weekend setup.
func BenchmarkSolveSlot(b *testing.B) {
	cfg := scenario.DefaultConfig()
	cfg.Players = 1000
	cfg.GamesPerPulje = 1000/6 + 6
	cfg.Seed = 4242
	w := scenario.Generate(cfg)
	slot := w.Slots[0]

	b.ReportAllocs()
	for b.Loop() {
		// Fresh state each iteration: SolveSlot mutates fairness state.
		solver.NewState(2026, w).SolveSlot(slot, w.Players)
	}
}
