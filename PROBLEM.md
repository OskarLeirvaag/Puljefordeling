# Problem Specification: Puljefordeling

## Context

A TTRPG/boardgame convention runs over a weekend with the following structure:

- **4 SLOTs** (sequential, fixed order):
  1. Friday evening
  2. Saturday morning
  3. Saturday evening
  4. Sunday morning

- Each SLOT contains a number of **EVENTs** (varies per slot, e.g. 5 or 7).
- Each EVENT has a fixed **capacity** (maximum number of PLAYERs it can hold).
- Each EVENT optionally has a **DM** (Dungeon Master / Game Master) — the player who runs it. A DM cannot also be assigned as a participant in any event during a SLOT they are running.
- Each PLAYER may express interest in zero or more EVENTs per SLOT, with a **score from 1–5** (5 = most interested).

---

## Definitions

| Term        | Meaning |
|-------------|---------|
| SLOT        | A time block during the weekend. SLOTs are processed sequentially. |
| EVENT       | A game session within a SLOT, with a fixed capacity. |
| PLAYER      | A convention attendee who registers interest in EVENTs. |
| SCORE       | A PLAYER's interest level in an EVENT: integer 1–5. Unranked = no interest. |
| DM          | A PLAYER who runs one or more EVENTs during the weekend. They cannot play in their own SLOTs but receive priority everywhere else. |
| SATISFIED   | A PLAYER who has been assigned at least one EVENT they scored 5 across the weekend so far. |
| UNSATISFIED | A PLAYER who has not yet received any EVENT they scored 5. |

---

## Constraints

1. A PLAYER can be assigned to **at most one EVENT per SLOT**.
2. An EVENT cannot exceed its **capacity**.
3. If a PLAYER has expressed **no interest** in any EVENT for a SLOT, they are **ignored** for that SLOT.
4. SLOTs must be **processed in order** (SLOT-1 is finalized before SLOT-2 begins, etc.).
5. Any EVENT assigned **2 or fewer PLAYERs** is **flagged for organiser review** in the report. The solver does not change or remove the assignment automatically.

---

## Objectives

### First objective: participation
> Maximize the number of interested PLAYERs who get to play in each SLOT.

The solver uses max-flow for this objective. A PLAYER may therefore be assigned to a lower-scored fallback if that is what allows the slot to seat more PLAYERs overall.

### Second objective: greatest wishes
> Among max-participation assignments, maximize the number of distinct PLAYERs who receive **at least one assignment to an EVENT they scored 5** across the entire weekend.

A PLAYER may score 5 on multiple EVENTs across any SLOT — all are equally desirable. Getting assigned to any one of them fulfils the "at-least-once" rule for that PLAYER.

- Once fulfilled, the PLAYER is **satisfied** and deprioritized when competing against unsatisfied PLAYERs for score-5 seats in later SLOTs.
- Being satisfied does not exclude a PLAYER from further score-5 assignments if capacity remains.

### Third objective: total preference
> Maximize the **total preference score sum** across all assignments.

Among solutions equal on participation and greatest-wish fairness, prefer the one where PLAYERs are assigned to events they care most about.

---

## Scoring Mechanic

To steer the assignment algorithm toward greatest-wish fairness inside the max-participation assignment, raw preference scores are adjusted before solving each SLOT:

| Condition | Adjusted score |
|---|---|
| UNSATISFIED player, event scored 5 | **10 + max(0, 5 − opportunities)** |
| UNSATISFIED player, score 1–4, **boost-enabled SLOT** | score × 2 |
| All other cases | actual score (1–5) |
| **DM bonus** (any edge for a DM player) | **+10 on top of the above** |

DMs receive a flat +10 bonus on every edge they own, stacking with all other adjustments. This compensates them for the time they contribute as game masters.

Where `opportunities` is the number of remaining SLOTs (including the current one) in which this player has at least one score-5 event. Fewer remaining chances ⇒ higher priority. Concretely:

| Opportunities | Adjusted score |
|---|---|
| 1 (only chance, now-or-never) | 14 |
| 2 | 13 |
| 3 | 12 |
| 4 | 11 |
| 5 or more | 10 |

**Key property**: within a fixed number of seated PLAYERs, a player's score-5 edge (adjusted to ≥ 10) outweighs any lower-scored non-DM edge (max 8 with boost, max 4 without). If their score-5 EVENT has free capacity in the final assignment, the solver should prefer that EVENT over a lower-scored fallback. If the score-5 EVENT is full, the solver may place the PLAYER in a fallback so another PLAYER can also play.

**Scarcity bonus** ensures that players who have only one or two score-5 opportunities across the weekend are not crowded out by players with many other chances. This is the cross-slot look-ahead used for greatest-wish fairness.

The late boost (all scores × 2 for unsatisfied players) applies to a configurable trailing window of SLOTs, set by the organizers at runtime.

### Operational model

The algorithm is run **once per SLOT**, not all at once. After each run the organizers see a satisfaction report (how many players are still unsatisfied) and decide whether to enable the late boost for the upcoming SLOTs. For example:

- After SLOT-2 results come in, organizers notice many players are still unsatisfied and enable the boost for SLOT-3 and SLOT-4.
- If satisfaction looks fine after SLOT-3, they may choose to only boost SLOT-4 (the final slot).

This means the boost window is a **live parameter** set before each slot is processed, not a fixed value decided upfront.

---

## Tie-breaking

When multiple PLAYERs have equal adjusted scores for the same EVENT, the winner is decided by a **seeded lottery**. The seed is derived from the convention year:

```
seed = year × 1000 + slotIndex
```

This means:
- Results are **fully reproducible** within a year — re-running with the same input gives the same output.
- Results are **different next year** — no organizer can predict or game the draw in advance.
- The seed is **printed in the report** so any player can verify the draw was fair.

---

## Capacity edge cases

Both of the following outcomes are valid and expected:

| Situation | Outcome |
|---|---|
| More interested PLAYERs than total event capacity in a SLOT | Some PLAYERs are left **unassigned** for that SLOT. The algorithm fills the most valuable seats first. |
| Fewer interested PLAYERs than an event's capacity | The event runs with a **partial fill**. Empty seats are not a problem. |
| 2 or fewer PLAYERs assigned to an event | The event is **flagged for organiser review** in the report. Assigned PLAYERs stay assigned. |

---

## Challenges

- **Contention**: Many PLAYERs may score 5 on the same EVENT — capacity limits mean not everyone can get their top pick.
- **Sequential commitment**: Decisions made in SLOT-1 are finalized before SLOT-2 runs and affect what options remain for SLOT-2 onward.
- **Look-ahead trade-off**: The solver does not globally optimize the whole weekend at once. That is intentional: organizers need to review and adjust each slot before continuing. The satisfaction-based priority weighting approximates cross-slot fairness without removing that manual control point.
- **Uneven interest distribution**: Some PLAYERs may score 5 on many events; others may have a single high-priority pick and nothing else.
- **Partial coverage**: A PLAYER who has no interests in some SLOTs is not penalized — they only participate in SLOTs where they expressed interest.
- **Undersubscribed events**: An event with very few assigned PLAYERs may need organiser attention. The solver flags these for review; it does not change assignments or re-solve automatically.

---

## Output

For each SLOT, the algorithm produces:

- **Assignment map**: `EVENT → [PLAYERs]`, respecting all capacity and exclusivity constraints.
- **Undersubscribed events**: EVENTs with 2 or fewer assigned PLAYERs, flagged for organiser review.
- **Unassigned PLAYERs**: PLAYERs who had interest but could not be placed in any rated event.
- **Newly satisfied PLAYERs**: PLAYERs who received their first score-5 assignment this SLOT.
- **Total score**: sum of actual (unadjusted) preference scores across all assignments.
- **Tie-breaking seed**: the seed used for the lottery shuffle this SLOT.
