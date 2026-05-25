# Puljefordeling

Automated participant assignment for TTRPG/boardgame conventions.

Given a set of time slots, events, player capacities, and player preference scores, this tool computes an optimal assignment of players to events — maximizing the number of players who get at least one of their top-ranked events across the weekend.

## Problem Summary

- A convention weekend has **4 sequential slots**.
- Each slot contains several **events** with fixed seat capacities and optional minimum-players thresholds.
- Each **player** rates events they are interested in (score 1–5) per slot.
- Some players are **DMs** running one or more events — they can't play during slots they run, but receive priority for the slots they can play.
- The algorithm assigns at most one event per player per slot.
- **Primary goal**: as many players as possible get assigned to at least one event they scored 5.
- **Secondary goal**: maximize total preference satisfaction across all assignments.

See [PROBLEM.md](./PROBLEM.md) for the full specification and [ALGORITHM.md](./ALGORITHM.md) for the algorithm, including the **Tweaks** section explaining all numeric constants and their fairness tradeoffs.

## Technology

- **Language**: [Go](https://go.dev/)
- **Module**: `puljefordeling`

## Algorithm Approach

The assignment is solved slot-by-slot in order. Each slot is treated as a min-cost max-flow problem with carry-forward state:

1. Build a bipartite graph of players ↔ events, edges weighted by adjusted preference scores.
2. Score adjustments encode several priorities:
   - **Unsatisfied players** (no score-5 assignment yet) get a boost on their score-5 edges.
   - **Scarcity awareness**: players with fewer remaining score-5 opportunities outrank those with many.
   - **DMs** get a flat bonus on every edge to reward contribution.
   - **Late boost** (organizer-toggled) doubles all scores for unsatisfied players in the final slots.
3. Respect event capacity constraints and minimum-player thresholds (events below their minimum are cancelled, one at a time, and re-solved).
4. **Seeded lottery** breaks ties between equal-score players, deterministic per convention year.
5. Finalize the slot assignment before proceeding to the next slot.

## Usage

```
go run . --input data.json
```

Input format: JSON describing slots, events (with capacities), players, and their per-slot preference scores.

## Project Structure

```
.
├── PROBLEM.md       # Detailed problem specification
├── README.md        # This file
├── go.mod
├── main.go          # Entry point
└── internal/
    ├── model/       # Data types: Slot, Event, Player, Preference
    └── solver/      # Assignment algorithm
```
