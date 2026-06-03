---
title: Puljefordeling
sub_title: Who sits where, fairly (ish)
author: Oskar --> Regncon
---

```typst +render
#set page(fill: rgb("#1a1b26"))
#set text(size: 20pt, fill: rgb("#c0caf5"))

#align(center)[
  #text(36pt, weight: "black", fill: rgb("#f7768e"))[THE MINIMUM-COST FLOW PROBLEM]
]

#v(14pt)

#grid(
  columns: 2,
  gutter: 1.4em,
  align: (right + horizon, left + horizon),

  text(fill: rgb("#7aa2f7"), weight: "bold")[minimize],
  $ sum_((u,v) in E) a(u,v) dot f(u,v) $,

  text(fill: rgb("#7dcfff"), weight: "bold")[subject to], [],

  $ f(u,v) <= c(u,v) $, text(fill: rgb("#8a93b8"))[capacity],
  $ f(u,v) = -f(v,u) $, text(fill: rgb("#8a93b8"))[skew symmetry],
  $ sum_(w in V) f(u,w) = 0 $, text(fill: rgb("#8a93b8"))[conservation #h(0.5em) ($u != s, t$)],
  $ sum_(w in V) f(s,w) = d = sum_(w in V) f(w,t) $, text(fill: rgb("#8a93b8"))[required flow],
)
```

<!-- pause -->

## … breathe.

<!-- pause -->

The whole talk is that — in human:

> **who sits where, fairly.**

<!-- end_slide -->

# The way we all picture it

A spreadsheet. Players down the side, games across the top, a mark where
someone's interested:

```text
              D&D     CoC    Blades   Vampire
   Alice       ✓                ✓
   Bob         ✓       ✓
   Carol       ✓                         ✓
   Dave                ✓        ✓
   …
```

<!-- pause -->

Now just… fill it in. One game per person. Don't overfill a table.
Keep it fair.

And there are **four** of these grids — one per pulje — every weekend.

<!-- end_slide -->

# Could a computer just try everything?

Each player picks one of their games. Multiply it out:

```text
   150 players  ×  ~5 options each   ≈   5¹⁵⁰   ≈   10¹⁰⁴ combinations
```

That's more than the number of atoms in the universe (~10⁸⁰).

<!-- pause -->

So… is this one of those *impossible* problems?

<!-- pause -->

**No** — and that's the nice part. There's an exact, *fast* method.
You just have to stop staring at the grid and **draw it as a graph.**

> (For the nerds: this is an assignment / min-cost-flow problem — it's in **P**,
> polynomial. Not NP-hard.)

<!-- end_slide -->

# First, why the "obvious" ways are bad

**First-come-first-served** → the fastest clicker wins. That's not fair, it's
a reflex test.

<!-- pause -->

**Greedy** (seat each person in their favourite, in turn):

```text
   Bob wants D&D or CoC   →  greedy seats Bob in D&D (the last seat)
   Alice wants only D&D   →  full!  Alice gets nothing.

   Better:  Bob → CoC,  Alice → D&D   →  both happy.
```

Greedy can't **undo** an early choice. Order decides the outcome. Bad.

<!-- pause -->

> (There's a classic method, the *Hungarian algorithm*, for matching — but it
> assumes one-person-to-one-job. Our games seat many, and not everyone gets in.)

<!-- end_slide -->

# Draw it as a graph

People on top, games on the bottom; a line = "interested".
Lines only ever cross between the two groups — that's a **bipartite** graph.

```bash +exec_replace
cat <<'ANSIEOF'
┌───────┐            ┌─────┐            ┌─────────┐            ┌────────┐
│       │            │     │            │         │            │        │
│ [38;2;158;206;106mA[0m[38;2;158;206;106ml[0m[38;2;158;206;106mi[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m │      ┌─────┤ [38;2;158;206;106mB[0m[38;2;158;206;106mo[0m[38;2;158;206;106mb[0m │      ┌─────┤  [38;2;158;206;106mC[0m[38;2;158;206;106ma[0m[38;2;158;206;106mr[0m[38;2;158;206;106mo[0m[38;2;158;206;106ml[0m  │      ┌─────┤  [38;2;158;206;106mD[0m[38;2;158;206;106ma[0m[38;2;158;206;106mv[0m[38;2;158;206;106me[0m  │
│       │      │     │     │      │     │         │      │     │        │
└───┬───┘      │     └──┬──┘      │     └────┬────┘      │     └────┬───┘
    │          │        │         │          │           │          │    
    ├──────────┴────────┼─────────┴──────────┼───────────┘          │    
    ▼                   ▼                    ▼                      ▼    
┌───────┐            ┌─────┐            ┌─────────┐            ┌────────┐
│       │            │     │            │         │            │        │
│  [38;2;122;162;247mD[0m[38;2;122;162;247m&[0m[38;2;122;162;247mD[0m  │            │ [38;2;122;162;247mC[0m[38;2;122;162;247mo[0m[38;2;122;162;247mC[0m │            │ [38;2;122;162;247mV[0m[38;2;122;162;247ma[0m[38;2;122;162;247mm[0m[38;2;122;162;247mp[0m[38;2;122;162;247mi[0m[38;2;122;162;247mr[0m[38;2;122;162;247me[0m │            │ [38;2;122;162;247mB[0m[38;2;122;162;247ml[0m[38;2;122;162;247ma[0m[38;2;122;162;247md[0m[38;2;122;162;247me[0m[38;2;122;162;247ms[0m │
│       │            │     │            │         │            │        │
└───────┘            └─────┘            └─────────┘            └────────┘
ANSIEOF
```

<!-- end_slide -->

# How many can we even seat?

Add a **source** feeding every person, and a **sink** every game drains into.
Each game holds as many people as it has **seats**.

```bash +exec_replace
cat <<'ANSIEOF'
┌────────┐                      ┌───────┐                      ┌─────┐                      ┌──────┐
│        │                      │       │                      │     │                      │      │
│ [38;2;158;206;106ms[0m[38;2;158;206;106mo[0m[38;2;158;206;106mu[0m[38;2;158;206;106mr[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ Alice ├───────────┬─────────►│ [38;2;122;162;247mD[0m[38;2;122;162;247m&[0m[38;2;122;162;247mD[0m ├─────────────────────►│ [38;2;247;118;142ms[0m[38;2;247;118;142mi[0m[38;2;247;118;142mn[0m[38;2;247;118;142mk[0m │
│        │                      │       │           │          │     │                      │      │
└────┬───┘                      └───────┘           │          └─────┘                      └──────┘
     │                                              │                                           ▲   
     │                              ┌───────────────┘                                           │   
     │                          ┌───────┐                      ┌─────┐                          │   
     │                          │   ┴   │                      │     │                          │   
     ├─────────────────────────►│  Bob  ├─────────────────────►│ [38;2;122;162;247mC[0m[38;2;122;162;247mo[0m[38;2;122;162;247mC[0m ├──────────────────────────┘   
     │                          │       │                      │     │                              
     │                          └───────┘                      └─────┘                              
     │                                                            ▲                                 
     │                                                            │                                 
     │                          ┌───────┐                         │                                 
     │                          │       │                         │                                 
     └─────────────────────────►│ Carol ├─────────────────────────┘                                 
                                │       │                                                           
                                └───────┘                                                           
ANSIEOF
```

**Max-flow:** push as many "person→seat" units as fit. The narrowest cut = the
most people we can possibly seat.

<!-- end_slide -->

# Let's run the smallest tricky case

Alice wants **only D&D**. Bob wants **D&D or CoC**. One seat each.

```bash +exec_replace
cat <<'ANSIEOF'
┌────────┐                      ┌───────┐                      ┌───────┐                      ┌──────┐
│        │                      │       │                      │       │                      │      │
│ [38;2;158;206;106ms[0m[38;2;158;206;106mo[0m[38;2;158;206;106mu[0m[38;2;158;206;106mr[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ Alice ├───────────┬─────────►│ [38;2;122;162;247mD[0m[38;2;122;162;247m&[0m[38;2;122;162;247mD[0m[38;2;122;162;247m [0m[38;2;122;162;247m1[0m ├─────────────────────►│ [38;2;247;118;142ms[0m[38;2;247;118;142mi[0m[38;2;247;118;142mn[0m[38;2;247;118;142mk[0m │
│        │                      │       │           │          │       │                      │      │
└────┬───┘                      └───────┘           │          └───────┘                      └──────┘
     │                                              │                                             ▲   
     │                              ┌───────────────┘                                             │   
     │                          ┌───────┐                      ┌───────┐                          │   
     │                          │   ┴   │                      │       │                          │   
     └─────────────────────────►│  Bob  ├─────────────────────►│ [38;2;122;162;247mC[0m[38;2;122;162;247mo[0m[38;2;122;162;247mC[0m[38;2;122;162;247m [0m[38;2;122;162;247m1[0m ├──────────────────────────┘   
                                │       │                      │       │                              
                                └───────┘                      └───────┘                              
ANSIEOF
```

<!-- end_slide -->

# Step 1 — push any path

`source → Bob → D&D → sink`. Bob is seated (orange). **1 seated.**

```bash +exec_replace
cat <<'ANSIEOF'
┌────────┐                      ┌───────┐                      ┌───────┐                      ┌──────┐
│        │                      │       │                      │       │                      │      │
│ [38;2;158;206;106ms[0m[38;2;158;206;106mo[0m[38;2;158;206;106mu[0m[38;2;158;206;106mr[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ Alice ├───────────┬─────────►│ [38;2;224;175;104mD[0m[38;2;224;175;104m&[0m[38;2;224;175;104mD[0m[38;2;224;175;104m [0m[38;2;224;175;104m1[0m ├─────────────────────►│ [38;2;247;118;142ms[0m[38;2;247;118;142mi[0m[38;2;247;118;142mn[0m[38;2;247;118;142mk[0m │
│        │                      │       │           │          │       │                      │      │
└────┬───┘                      └───────┘           │          └───────┘                      └──────┘
     │                                              │                                             ▲   
     │                              ┌───────────────┘                                             │   
     │                          ┌───────┐                      ┌───────┐                          │   
     │                          │   ┴   │                      │       │                          │   
     └─────────────────────────►│  [38;2;224;175;104mB[0m[38;2;224;175;104mo[0m[38;2;224;175;104mb[0m  ├─────────────────────►│ [38;2;122;162;247mC[0m[38;2;122;162;247mo[0m[38;2;122;162;247mC[0m[38;2;122;162;247m [0m[38;2;122;162;247m1[0m ├──────────────────────────┘   
                                │       │                      │       │                              
                                └───────┘                      └───────┘                              
ANSIEOF
```

<!-- end_slide -->

# Step 2 — Alice is stuck… or is she?

D&D is full. So we **undo**: push Alice into D&D and bump Bob back along the
`undo` edge over to CoC. (green = already placed, orange = the new path)

```bash +exec_replace
cat <<'ANSIEOF'
┌────────┐                      ┌───────┐                      ┌───────┐                      ┌──────┐
│        │                      │       │                      │       │                      │      │
│ [38;2;158;206;106ms[0m[38;2;158;206;106mo[0m[38;2;158;206;106mu[0m[38;2;158;206;106mr[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ [38;2;224;175;104mA[0m[38;2;224;175;104ml[0m[38;2;224;175;104mi[0m[38;2;224;175;104mc[0m[38;2;224;175;104me[0m ├───────────┬──undo───►┤ [38;2;158;206;106mD[0m[38;2;158;206;106m&[0m[38;2;158;206;106mD[0m[38;2;158;206;106m [0m[38;2;158;206;106m1[0m ├─────────────────────►│ [38;2;247;118;142ms[0m[38;2;247;118;142mi[0m[38;2;247;118;142mn[0m[38;2;247;118;142mk[0m │
│        │                      │       │           │          │       │                      │      │
└────┬───┘                      └───────┘           │          └───────┘                      └──────┘
     │                                              │                                             ▲   
     │                              ▼───────────────┘                                             │   
     │                          ┌───────┐                      ┌───────┐                          │   
     │                          │   ┴   │                      │       │                          │   
     └─────────────────────────►│  [38;2;158;206;106mB[0m[38;2;158;206;106mo[0m[38;2;158;206;106mb[0m  ├─────────────────────►│ [38;2;224;175;104mC[0m[38;2;224;175;104mo[0m[38;2;224;175;104mC[0m[38;2;224;175;104m [0m[38;2;224;175;104m1[0m ├──────────────────────────┘   
                                │       │                      │       │                              
                                └───────┘                      └───────┘                              
ANSIEOF
```

<!-- end_slide -->

# Both seated

Alice → D&D, Bob → CoC. **2 / 2.** That's the **undo** greedy couldn't do —
and it's what makes the answer *optimal*, not just decent.

```bash +exec_replace
cat <<'ANSIEOF'
┌────────┐                      ┌───────┐                      ┌───────┐                      ┌──────┐
│        │                      │       │                      │       │                      │      │
│ [38;2;158;206;106ms[0m[38;2;158;206;106mo[0m[38;2;158;206;106mu[0m[38;2;158;206;106mr[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ [38;2;158;206;106mA[0m[38;2;158;206;106ml[0m[38;2;158;206;106mi[0m[38;2;158;206;106mc[0m[38;2;158;206;106me[0m ├─────────────────────►│ [38;2;158;206;106mD[0m[38;2;158;206;106m&[0m[38;2;158;206;106mD[0m[38;2;158;206;106m [0m[38;2;158;206;106m1[0m ├─────────────────────►│ [38;2;247;118;142ms[0m[38;2;247;118;142mi[0m[38;2;247;118;142mn[0m[38;2;247;118;142mk[0m │
│        │                      │       │                      │       │                      │      │
└────┬───┘                      └───────┘                      └───────┘                      └──────┘
     │                                                                                            ▲   
     │                                                                                            │   
     │                          ┌───────┐                      ┌───────┐                          │   
     │                          │       │                      │       │                          │   
     └─────────────────────────►│  [38;2;158;206;106mB[0m[38;2;158;206;106mo[0m[38;2;158;206;106mb[0m  ├─────────────────────►│ [38;2;158;206;106mC[0m[38;2;158;206;106mo[0m[38;2;158;206;106mC[0m[38;2;158;206;106m [0m[38;2;158;206;106m1[0m ├──────────────────────────┘   
                                │       │                      │       │                              
                                └───────┘                      └───────┘                              
ANSIEOF
```

<!-- end_slide -->

# …but it's not "interested / not"

People didn't tick boxes. They told us **how much**:

- 🔥 **Veldig** — a top choice
- 👍 **Middels** — would enjoy it
- 🤷 **Litt** — sure, why not

<!-- pause -->

Plain max-flow can't tell a 🔥 from a 🤷 — a seat is a seat to it.
**The edges have weights now.**

<!-- end_slide -->

# Min-cost max-flow

Put a **cost** on each person→game line: `cost = −(how much they want it)`.

1. **Max-flow** — seat as many people as possible.
2. **Min-cost** — of those fullest seatings, pick the **cheapest** (= happiest).

<!-- pause -->

The augmenting-path + **reverse-edge undo** you just watched is the whole
engine — we simply always walk the *cheapest* path first.

<!-- end_slide -->

# The one idea to remember

## Weights *are* the policy.

The solver is **preference-blind**. It only ever maximises total weight.

Every fairness rule — top choices first, fairness across the weekend,
rewarding GMs, helping the unlucky — is just **one number per line.**

<!-- pause -->

Change the numbers → change the policy. The algorithm never changes.

<!-- end_slide -->

# The priority bands

| Category                       | Weight |
| ------------------------------ | -----: |
| Unsatisfied · 🔥 (top choice)  |    800 |
| Satisfied · 🔥                 |    600 |
| 👍 Middels (either way)        |    400 |
| 🤷 Litt (either way)           |    200 |

Bonuses: **+miss** (unlucky), **+10** never-seated, **+60** DM.

```bash +exec
go run . weights
```

<!-- end_slide -->

# Spreading the joy

Once you've had a 🔥 you drop 800 → 600: you can still get a *second* great
game, but the unsatisfied now outrank you. Carried across all 4 puljer.

```bash +exec
go run . weekend
```

<!-- end_slide -->

# The hard tension

"Seat as many as possible" can actively **hurt**.

Three people want game **X** (top choice, 800). **Cleo** also threw a 🤷 at
quiet game **Y**. **Dan** only mildly wants X.

<!-- pause -->

To squeeze Dan in, we'd bump Cleo off her top choice down to Y:

> lose **600**, to gain Dan's **200**  →  net **−400**.

Pure "fill every chair" does it anyway. That feels wrong — and it is.

<!-- end_slide -->

# Participation, priced

Give "one more person seated" an explicit price **B** (= 300).
Make the trade only if it's worth it. The solver then **refuses** the bad one:

```bash +exec
go run . reroute
```

<!-- pause -->

`B = ∞` → fill every chair. `B = 0` → never trade. `B = 300` → sensible.

<!-- end_slide -->

# Can people game it?

We seal one pulje, assign it, then wait a day before the next — so players
**can re-submit** for later puljer after seeing the results.

<!-- pause -->

Any naive "boost the unsatisfied" invites one exploit:

> **Don't cash your 🔥 until the last day** — look unlucky on paper, then show
> up to the final slot with maximum priority.

We need a boost that **can't be farmed.**

<!-- end_slide -->

# Un-gameable scarcity

Make the boost **backward-looking**: +20 for each earlier pulje you *wanted* a
top choice and **missed**.

- You can't fake a miss — the assignment decides it, on locked results.
- The only way to farm it is to genuinely lose games you wanted. Nobody does.

```bash +exec
go run . scarcity
```

<!-- end_slide -->

# Game masters

GMs run games for us — they deserve priority when they *do* get to play.

- **+60** on every line — a strong bump…
- …but it stays **inside the band**: a regular player's unmet 🔥 (800) still
  beats a GM's 👍 (≤ 470).

Reward contributors **without** dragging them into a game they barely want.

<!-- end_slide -->

# Does it scale?

A realistic con, generated and solved live:

```bash +exec
go run . generate
```

<!-- pause -->

| Players | Time / weekend |
| ------: | -------------: |
|     150 |        ~4 ms   |
|     500 |       ~16 ms   |
|    1000 |       ~54 ms   |
|    2000 |      ~190 ms   |

Real scale is a non-issue.

<!-- end_slide -->

# In the product

The admin **"Emulér puljefordeling"** page previews a whole weekend —
read-only, nothing saved:

- 🔥 / 👍 / 🤷 — what each person *got*
- **turquoise** name — a GM playing elsewhere
- **red bar** — bumped below their top wish to make room

Slides explain the *why*; the page shows the *what*.

<!-- end_slide -->

# Recap

- A grid *looks* impossible; as a **graph** it's easy and fast.
- **Bipartite** people↔games, with a source and sink = **flow**.
- Max-flow seats the most; **min-cost** makes them happiest; reverse edges
  let it **undo**.
- **Weights are the policy** — top choices, fairness, GMs, the unlucky.
- **Price participation**; make scarcity **un-gameable**.
- Solves a weekend in milliseconds.

Every number is just a tunable constant.

**Takk!** — questions?
