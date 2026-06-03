// Package scenario generates synthetic-but-plausible weekends for stress tests
// and scale demos. Demand follows a tunable power-law ("a few blockbuster games,
// a long tail of niche ones") — the skew is what creates real contention for the
// solver. It is NOT a claim that convention demand is "scale-free"; it is just an
// easy, controllable heavy-tailed weight. Everything is deterministic from Seed.
package scenario

import (
	"fmt"
	"math"
	"math/rand/v2"

	"puljefordeling/internal/model"
)

// Config controls weekend generation. Zero values are not sensible; start from
// DefaultConfig and tweak.
type Config struct {
	Seed               uint64
	Players            int
	Puljer             int
	GamesPerPulje      int
	MinCapacity        int
	MaxCapacity        int
	MaxInterests       int     // max games a player rates in a pulje they attend
	AttendanceProb     float64 // chance a player shows up to a given pulje
	PopularityExponent float64 // Zipf α: 0 = uniform demand, ~1 = strong blockbuster skew
	GMFraction         float64 // fraction of players who run games
	GMProbPerGame      float64 // chance a given game has a GM assigned
}

// DefaultConfig is a convention-sized weekend with realistic skew, roughly
// balanced so most attendees can be seated (≈ seats vs attendees per pulje):
// 24 games × ~5 seats ≈ 120 seats/pulje vs ≈ 120 attendees.
func DefaultConfig() Config {
	return Config{
		Seed:               1,
		Players:            150,
		Puljer:             4,
		GamesPerPulje:      24,
		MinCapacity:        4,
		MaxCapacity:        6,
		MaxInterests:       3,
		AttendanceProb:     0.8,
		PopularityExponent: 1.0,
		GMFraction:         0.12,
		GMProbPerGame:      0.4,
	}
}

// Generate builds a deterministic weekend from cfg.
func Generate(cfg Config) model.Weekend {
	rng := rand.New(rand.NewPCG(cfg.Seed, cfg.Seed^0x9e3779b97f4a7c15))

	players := make([]model.Player, cfg.Players)
	for i := range players {
		id := fmt.Sprintf("p%03d", i)
		players[i] = model.Player{ID: id, Name: id, Prefs: map[string]map[string]model.Score{}}
	}

	// GM pool: a random subset of players.
	nGM := int(cfg.GMFraction * float64(cfg.Players))
	gmPool := make([]string, 0, nGM)
	for _, idx := range rng.Perm(cfg.Players)[:max0(nGM)] {
		gmPool = append(gmPool, players[idx].ID)
	}

	levels := []model.Score{5, 3, 1} // Veldig, Middels, Litt (descending)
	slots := make([]model.Slot, cfg.Puljer)

	for s := 0; s < cfg.Puljer; s++ {
		slotID := fmt.Sprintf("pulje%d", s)

		// Popularity: each game gets a shuffled rank, weight = 1/rank^α.
		ranks := rng.Perm(cfg.GamesPerPulje)
		weights := make([]float64, cfg.GamesPerPulje)
		games := make([]model.Event, cfg.GamesPerPulje)
		gmUsedHere := map[string]bool{}

		for g := range games {
			capacity := cfg.MinCapacity + rng.IntN(cfg.MaxCapacity-cfg.MinCapacity+1)
			ev := model.Event{
				ID:       fmt.Sprintf("%s-g%d", slotID, g),
				Name:     fmt.Sprintf("Game %s-%d", slotID, g),
				Capacity: capacity,
			}
			weights[g] = 1.0 / math.Pow(float64(ranks[g]+1), cfg.PopularityExponent)
			if len(gmPool) > 0 && rng.Float64() < cfg.GMProbPerGame {
				gm := gmPool[rng.IntN(len(gmPool))]
				if !gmUsedHere[gm] { // a GM can't run two games in the same pulje
					ev.DMID = gm
					gmUsedHere[gm] = true
				}
			}
			games[g] = ev
		}
		slots[s] = model.Slot{ID: slotID, Name: fmt.Sprintf("Pulje %d", s), Events: games}

		// Each attending player rates a popularity-weighted handful of games.
		for i := range players {
			if rng.Float64() > cfg.AttendanceProb {
				continue
			}
			k := 1 + rng.IntN(cfg.MaxInterests)
			picks := pickWeighted(rng, weights, k)
			pm := make(map[string]model.Score, len(picks))
			for n, g := range picks {
				score := model.Score(1)
				if n < len(levels) {
					score = levels[n]
				}
				pm[games[g].ID] = score
			}
			if len(pm) > 0 {
				players[i].Prefs[slotID] = pm
			}
		}
	}

	return model.Weekend{Slots: slots, Players: players}
}

// pickWeighted draws up to k distinct indices [0,len(weights)) with probability
// proportional to weight, without replacement.
func pickWeighted(rng *rand.Rand, weights []float64, k int) []int {
	idx := make([]int, len(weights))
	w := make([]float64, len(weights))
	for i := range weights {
		idx[i] = i
		w[i] = weights[i]
	}

	out := make([]int, 0, k)
	for len(out) < k && len(idx) > 0 {
		total := 0.0
		for _, x := range w {
			total += x
		}
		r := rng.Float64() * total
		pick := 0
		for pick < len(w)-1 {
			if r < w[pick] {
				break
			}
			r -= w[pick]
			pick++
		}
		out = append(out, idx[pick])
		idx = append(idx[:pick], idx[pick+1:]...)
		w = append(w[:pick], w[pick+1:]...)
	}
	return out
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}
