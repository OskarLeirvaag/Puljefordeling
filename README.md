# Puljefordeling

Automated participant assignment for TTRPG/boardgame conventions.

Given a set of time slots, events, player capacities, and player preference scores, this tool assigns as many players as possible to games while strongly prioritizing players who have not yet received one of their top-ranked events.

## Problem Summary

- A convention weekend has **4 sequential slots**.
- Each slot contains several **events** with fixed seat capacities.
- Each **player** rates events they are interested in (score 1–5) per slot.
- Some players are **DMs** running one or more events — they can't play during slots they run, but receive priority for the slots they can play.
- The algorithm assigns at most one event per player per slot.
- **First goal**: seat as many interested players as possible in each slot.
- **Second goal**: among max-seat assignments, maximize the number of players who get at least one event they scored 5 across the weekend.
- **Third goal**: maximize total preference satisfaction across all assignments.

See [PROBLEM.md](./PROBLEM.md) for the full specification and [ALGORITHM.md](./ALGORITHM.md) for the algorithm, including the **Tweaks** section explaining all numeric constants and their fairness tradeoffs.

## Technology

- **Language**: [Go](https://go.dev/)
- **Module**: `puljefordeling`

## Algorithm Approach

The assignment is solved slot-by-slot in order. This is deliberate: organizers can review each slot, make manual adjustments if needed, then run the next slot with updated state. Each slot is treated as a min-cost max-flow problem with carry-forward state:

1. Build a bipartite graph of players ↔ events, edges weighted by adjusted preference scores.
2. Score adjustments encode several priorities:
   - **Unsatisfied players** (no score-5 assignment yet) get a boost on their score-5 edges.
   - **Scarcity awareness**: players with fewer remaining score-5 opportunities outrank those with many.
   - **DMs** get a flat bonus on every edge to reward contribution.
   - **Late boost** (organizer-toggled) doubles all scores for unsatisfied players in the final slots.
3. Respect event capacity constraints and flag events with 2 or fewer assigned players for organizer review.
4. **Seeded lottery** breaks ties between equal-score players, deterministic per convention year.
5. Finalize the slot assignment before proceeding to the next slot.

## Usage

```
go run .
```

The current executable runs the built-in demo weekend from `main.go`.

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
