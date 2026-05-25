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
- Each PLAYER may express interest in zero or more EVENTs per SLOT, with a **score from 1–5** (5 = most interested).

---

## Definitions

| Term     | Meaning |
|----------|---------|
| SLOT     | A time block during the weekend. SLOTs are processed sequentially. |
| EVENT    | A game session within a SLOT, with a fixed participant capacity. |
| PLAYER   | A convention attendee who registers interest in EVENTs. |
| SCORE       | A PLAYER's interest level in an EVENT: integer 1–5. Unranked = no interest. |
| SATISFIED   | A PLAYER who has been assigned at least one EVENT they scored 5 across the weekend so far. |
| UNSATISFIED | A PLAYER who has not yet received any EVENT they scored 5. |

---

## Constraints

1. A PLAYER can be assigned to **at most one EVENT per SLOT**.
2. An EVENT cannot exceed its **capacity**.
3. If a PLAYER has expressed **no interest** in any EVENT for a SLOT, they are **ignored** for that SLOT.
4. SLOTs must be **processed in order** (SLOT-1 is finalized before SLOT-2 begins, etc.).

---

## Objectives

### Primary (most important)
> Maximize the number of distinct PLAYERs who receive **at least one assignment to an EVENT they scored 5** across the entire weekend.

A PLAYER may score 5 on multiple EVENTs across any SLOT — all are equally desirable. Getting assigned to any one of them fulfils the "at-least-once" rule for that PLAYER.

- Once fulfilled, the PLAYER is **satisfied** and deprioritized when competing against unsatisfied PLAYERs for score-5 seats in later SLOTs.
- Being satisfied does not exclude a PLAYER from further score-5 assignments if capacity remains.

### Secondary
> Maximize the **total preference score sum** across all assignments.

Among solutions equal on the primary objective, prefer the one where PLAYERs are assigned to events they care most about.

---

## Scoring Mechanic

To steer the assignment algorithm toward the primary objective, raw preference scores are adjusted before solving each SLOT:

| Condition | Adjusted score |
|---|---|
| UNSATISFIED player, event scored 5, any SLOT | 10 |
| UNSATISFIED player, any event, **boost-enabled SLOT** | score × 2 |
| All other cases | actual score (1–5) |

The late boost (all scores × 2 for unsatisfied players) applies to a configurable trailing window of SLOTs, set by the organizers at runtime.

### Operational model

The algorithm is run **once per SLOT**, not all at once. After each run the organizers see a satisfaction report (how many players are still unsatisfied) and decide whether to enable the late boost for the upcoming SLOTs. For example:

- After SLOT-2 results come in, organizers notice many players are still unsatisfied and enable the boost for SLOT-3 and SLOT-4.
- If satisfaction looks fine after SLOT-3, they may choose to only boost SLOT-4 (the final slot).

This means the boost window is a **live parameter** set before each slot is processed, not a fixed value decided upfront.

---

## Challenges

- **Contention**: Many PLAYERs may rank the same EVENT highest — capacity limits mean not everyone can get their top pick.
- **Sequential commitment**: Decisions made in SLOT-1 are irreversible and affect what options remain for SLOT-2 onward.
- **Look-ahead trade-off**: Greedily maximizing SLOT-1 may starve PLAYERs of top-pick opportunities in later SLOTs. The algorithm must balance immediate and future satisfaction.
- **Uneven interest distribution**: Some PLAYERs may rank many events equally high; others may have a single high-priority pick and nothing else.
- **Partial coverage**: A PLAYER who has no interests in some SLOTs should not be penalized — they only participate in SLOTs where they expressed interest.

---

## Output

For each SLOT, the algorithm produces an **assignment map**: `EVENT → [PLAYERs]`, respecting all capacity and exclusivity constraints.

Additionally, a summary should report:
- How many PLAYERs received at least one top-score assignment.
- Total preference score across all assignments.
- Unassigned PLAYERs per SLOT (had interest but did not fit any event).
