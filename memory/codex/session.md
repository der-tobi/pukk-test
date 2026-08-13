# Codex — Session Notes

Update this file at the end of every session. Keep it current so the next session can pick up without re-reading the whole chat.

## Current focus

- Reviewed Claude's design/research pass on the PuKK PoC docs and memory files, then fixed the accepted documentation/repo hygiene findings. No implementation work started.

## Decisions made this session

- Treated the review as the current uncommitted worktree against the project brief, because no branch/commit fixed point was supplied.
- Used the local `.claude/skills/code-review` guidance where applicable, but did not spawn sub-agents because the review target was uncommitted design documentation rather than a pinned branch diff.
- Applied the user's LED clarification: index 0 is the current `0-5min` slot and newly visible future state enters at index 11 (`55-60min`) before moving counter-clockwise toward index 0.

## Open questions / blockers

- Still needs empirical verification during implementation: v1 extend/update request shape, local PuKK REST API call-back behavior, and NFC checkin `pin` behavior.

## Next steps

- Populate `api-contract/openapi.yaml` with PuKK-facing `GET /` poll and `POST /` action endpoints before scaffolding the Go server.
