// Package model defines the core data types for the puljefordeling algorithm.
package model

// Score is a player's interest level in an event (1–5).
type Score int

// MinScore and MaxScore are the inclusive bounds for preference scores.
const (
	MinScore Score = 1
	MaxScore Score = 5
)

// Event is a game session within a slot with a fixed seat count.
// MinPlayers is the minimum number of players required for the event to run.
// Zero means no minimum (the event runs with any number of players).
type Event struct {
	ID         string
	Name       string
	Capacity   int
	MinPlayers int
}

// Slot is a time block containing one or more events.
type Slot struct {
	ID     string
	Name   string
	Events []Event
}

// Player is a convention attendee.
// Prefs[slotID][eventID] = score. A missing entry means no interest.
type Player struct {
	ID    string
	Name  string
	Prefs map[string]map[string]Score
}

// Weekend holds all slots and players for a convention.
type Weekend struct {
	Slots   []Slot
	Players []Player
}

// SlotResult is the assignment output for a single slot.
type SlotResult struct {
	SlotID          string
	Assignments     map[string][]string // eventID -> assigned playerIDs
	CancelledEvents []string            // eventIDs cancelled due to insufficient players
	Unassigned      []string            // playerIDs with interest but no seat
	NewlySatisfied  []string            // playerIDs satisfied for the first time this slot
	TotalScore      int                 // sum of actual (unadjusted) scores for all assignments
	Seed            int64               // seed used for tie-breaking shuffle this slot
}
