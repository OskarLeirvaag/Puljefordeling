# Puljefordeling

Automated participant assignment for TTRPG/boardgame conventions.

Given a set of time slots, events, player capacities, and player preference scores, this tool computes an optimal assignment of players to events — maximizing the number of players who get at least one of their top-ranked events across the weekend.

## Problem Summary

- A convention weekend has **4 sequential slots**.
- Each slot contains several **events** with fixed seat capacities.
- Each **player** rates events they are interested in (score 1–5) per slot.
- The algorithm assigns at most one event per player per slot.
- **Primary goal**: as many players as possible get assigned to at least one event they rated at their personal maximum score.
- **Secondary goal**: maximize total preference satisfaction across all assignments.

See [PROBLEM.md](./PROBLEM.md) for the full specification.

## Technology

- **Language**: [Go](https://go.dev/)
- **Module**: `puljefordeling`

## Algorithm Approach

The assignment is solved slot-by-slot in order. Each slot is treated as a constrained matching problem:

1. Build a bipartite graph of players ↔ events (edges weighted by preference score).
2. Apply a priority-aware assignment that accounts for each player's global "top score" status — players who have not yet secured a top-score assignment across prior slots are given priority.
3. Respect event capacity constraints.
4. Finalize the slot assignment before proceeding to the next slot.

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
