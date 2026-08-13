---
name: adr-review
description: Review all Architecture Decision Records (ADRs) in docs/adr/ and identify which ones are relevant to the current task, flag conflicts, and surface constraints before implementation begins. Use when starting a new feature, reviewing a proposal, checking if a design matches prior decisions, or when the user mentions "check ADRs", "review decisions", or "what did we decide about X".
---

# ADR Review

Read existing ADRs and surface what matters for the current task.

## When to Run This

- Before writing any code for a new feature or phase
- Before creating a new API proposal in `docs/api-proposals/`
- When a proposed design feels uncertain — "does this conflict with something we decided?"
- When Codex opens a PR that touches areas covered by an ADR

## Process

### 1. Read all ADRs

Read every file in `docs/adr/`. Note:
- Status (`Accepted` / `Proposed` / `Deprecated`)
- What decision was made
- The key rules section (if present)

Skip `Deprecated` ADRs unless the user is asking about history.

### 2. Map ADRs to the current task

For each `Accepted` ADR, ask: *does the current task touch the area this ADR governs?*

Flag an ADR as **relevant** if:
- The task creates or modifies an endpoint covered by the ADR's route table
- The task changes a schema shape defined in the ADR
- The task involves an entity or flow the ADR describes
- The task proposes something that contradicts a Key Rule in the ADR

### 3. Report findings

Output a concise report:

```
## ADR Review for: <current task description>

### Relevant ADRs

**ADR-0001: API Surface Design**  ← example
- Rule in play: "Policy masking is server-side. DisplayView fields are filtered before leaving the backend."
- Implication for this task: The Display App component must not conditionally hide title/organizer — trust the server.

### Conflicts or Tensions

<List any proposed design choices that contradict an accepted ADR. Be specific.>

### No conflicts found in

ADR-0002, ADR-0003 (not relevant to this task)
```

### 4. Recommend next steps

If a conflict is found:
- Flag it to the user before any code is written
- Suggest whether the right fix is (a) change the implementation to respect the ADR, or (b) open a new ADR to supersede the old decision

If no conflicts:
- State clearly: "No ADR conflicts found — proceed."

## What to Look For (Common Patterns)

| Task type | ADRs likely relevant |
|-----------|---------------------|
| New API endpoint | ADR covering API surface design |
| Schema change | ADR covering entity model or API surface |
| Auth change | ADR covering authentication strategy |
| Real-time / SSE | ADR covering delivery mechanism |
| Multi-tenancy | ADR covering tenant isolation rules |
| Display App rendering | ADR covering policy masking rules |

## Tone

Be direct. A review that says "nothing relevant" is as valuable as one that flags five conflicts. Don't pad the report.
