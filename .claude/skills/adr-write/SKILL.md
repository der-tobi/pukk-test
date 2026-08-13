---
name: adr-write
description: Write an Architecture Decision Record (ADR) by interviewing the user about a decision, exploring alternatives, and saving the result to docs/adr/. Use when the user wants to record an architectural decision, document a design choice, create an ADR, or capture why something was built a certain way.
---

# ADR Write

Create a new Architecture Decision Record in `docs/adr/`.

## Process

### 1. Check existing ADRs

Read `docs/adr/` to find the highest existing number and avoid collisions.
Also scan for any existing ADR that already covers this decision — update it instead of creating a duplicate.

### 2. Interview the user

Ask only what you don't already know from context:

- **What decision was made?** (One sentence.)
- **What problem or context forced this decision?** (What would have happened without it?)
- **What alternatives were considered?** (At least two; why were they rejected?)
- **What are the consequences?** (What gets easier? What gets harder? Any risks?)
- **Who are the deciders?** (Claude, Codex, both, or external?)

If the decision was just made in this conversation, you likely already have most of this — confirm rather than re-ask.

### 3. Write the ADR

Use the template below. Save to `docs/adr/NNNN-slug.md` where `NNNN` is zero-padded (e.g. `0002`).

```markdown
# ADR-NNNN: <Title>

**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-XXXX
**Date:** YYYY-MM-DD
**Deciders:** <Claude, Codex, or names>
**Relates to:** <ADR-XXXX, issue links, plan phases — optional>

---

## Context

What situation, constraint, or problem prompted this decision?
Include relevant forces: performance, team structure, existing code, deadlines.

## Decision

What was decided. Be specific — include routes, schema shapes, rules, or code patterns
as needed. If the decision produces a canonical list (route table, schema), include it here.

### Key Rules (if applicable)

Non-negotiable constraints that follow from this decision.

## Consequences

**Positive:**
- ...

**Negative / Watch-outs:**
- ...

## Alternatives Considered

### Alternative A — <Name>
Why it was considered. Why it was rejected (or what was kept from it).

### Alternative B — <Name>
...
```

### 4. Present and confirm

Show the draft to the user. Ask if anything is missing or wrong. Adjust on feedback.

### 5. Open a PR for review

ADR review happens via pull request — this is how Claude and Codex communicate about architectural decisions.

```bash
git checkout -b adr/NNNN-short-slug
git add docs/adr/NNNN-slug.md
git commit -m "docs: add ADR-NNNN <title> (status: Proposed)"
gh pr create --title "ADR-NNNN: <title>" --body "..."
```

- Status in the file stays `Proposed` until the PR is approved and merged
- Reviewers (Claude, Codex, user) leave feedback as PR comments
- Once all feedback is resolved, update status to `Accepted`, merge, close PR
- Do **not** mark `Accepted` before the PR is merged

### 6. Update related files

If the decision affects the API contract, a plan phase, or CLAUDE.md/AGENTS.md — flag it. Do not update those files automatically; ask the user.

## Naming Convention

`docs/adr/NNNN-short-slug.md`

- Numbers are sequential, zero-padded to 4 digits
- Slug is lowercase kebab-case, max 5 words
- Examples: `0002-auth-strategy.md`, `0003-real-time-delivery.md`, `0004-multi-tenancy-isolation.md`

## Status Values

| Status | Meaning |
|--------|---------|
| `Proposed` | Decision is under discussion — not yet binding |
| `Accepted` | Decision is final — both Claude and Codex must follow it |
| `Deprecated` | Was accepted but no longer applies |
| `Superseded by ADR-XXXX` | Replaced by a newer decision |
