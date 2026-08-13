# Memory Index

All agents read this file at session start. Add pointers here when creating new memory files.

## Shared

<!-- Example entries — replace with real ones as the project evolves:
- [project_context.md](project_context.md) — project goals, decisions, current phase
- [user_role.md](user_role.md) — who the user is, preferences, collaboration style
- [feedback_style.md](feedback_style.md) — do/don't from past sessions
- [api_decisions.md](api_decisions.md) — agreed API contract decisions
-->

- [tooling_preferences.md](tooling_preferences.md) — local Codex/agent tooling preferences and configuration decisions
- [tech_stack.md](tech_stack.md) — why pukk-test is Go (single-binary Windows cross-compile), alternatives considered
- [collaboration_model.md](collaboration_model.md) — Claude/Codex are equal-rights peers, no fixed ownership split
- [booking_cache_design.md](booking_cache_design.md) — 3V Rooms freebusy caching: buffer-beyond-display-window design, default 5min refresh, resolved stale/cold-start fallback
- [led_ring_design.md](led_ring_design.md) — LED ring orientation, wall-clock rotation, gradient scope, pulse/blink rendering approach
- [button_state_machine.md](button_state_machine.md) — button extend/undo cycle, 5s commit window, longpress-checkout redefined to device capabilities, NFC
- [rooms_api_integration.md](rooms_api_integration.md) — which 3V Rooms endpoints/versions we use, booking-ID discovery, auth/token fallback, known gaps to verify
- `docs/research/3v-rooms-api.md` — primary-source research on the 3V Rooms API (auth, endpoints, freebusy semantics)
- `docs/research/pukk-device-api.md` — primary-source research on the PuKK device's own protocol (LED commands, button/NFC events, local REST API)

## Agent session notes

Each agent maintains `memory/<agent-name>/session.md`. Add a link here when a new agent joins.

- [memory/claude/session.md](claude/session.md) — Claude
- [memory/codex/session.md](codex/session.md) — Codex
