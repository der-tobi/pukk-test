# Claude — Session Notes

Update this file at the end of every session. Keep it current so the next session (or a replacement agent) can pick up without re-interviewing the user.

## Current focus

- Full design pass on `project.md` complete via `/grilling` (2026-08-13). Every open question and ambiguity in the original brief is now resolved and written up in `memory/`. Implementation has not started yet.

## Decisions made this session

All decisions are written up as topic files, not repeated here in full — see `memory/MEMORY.md` for the index:

- [[led-ring-design]] — ring orientation (12 o'clock = now, clockwise = future), wall-clock-aligned discrete rotation every 5 min, gradient only on the "now" LED (simplified away from also fading upcoming-booking edges), pulse/blink simulated via alternating colors across stateless poll responses, fast one-off feedback pushed via the device's local REST API instead of waiting on the poll cycle.
- [[button-state-machine]] — extend/undo LIFO cycle (`min(15, headroom)` per press, undo in reverse once full), 5s provisional-commit window with simple revert-on-failure, max reach intentionally capped to the ring's 60-min window, longpress-checkout redefined from the brief's 2500ms to the device's native 3s bucket (the device only reports classified post-hoc press events — no live hold signal exists, so the originally-envisioned live progressive animation and early-release wind-back aren't implementable; simplified to "no event fires on early release, nothing to abort").
- [[rooms-api-integration]] — v2 for freebusy reads, v1 for all booking mutation (only v1 has checkin/checkout/release); `grant_type=basic` kept despite docs describing `apikey`, because the user has live-tested it against their tenant; booking IDs resolved on-demand via a narrow-window lookup, not prefetched; checkin-required (orange) reflects Rooms' own `checkinConfirmed` field rather than app-side policy; NFC checkin sends the startup password as `pin`.
- [[booking-cache-design]] — resolved its one open question: any time-slice with no confirmed data (dry buffer, or before the first successful fetch) renders as busy, not stale-marked or free.
- Wrote up `docs/research/pukk-device-api.md` (LED command shapes, button/NFC event shapes, the device's own local REST API) — this didn't exist before and was previously only in a sub-agent's output, not persisted anywhere.
- Found and reconciled with `docs/research/3v-rooms-api.md`, which already existed on disk (not written by me this session, likely from a Codex session) and was more rigorous than my own sub-agent's Rooms API research — used it as the primary source and flagged one direct conflict (see below).

## Open questions / blockers

- **Flagged discrepancy needing empirical verification during implementation**: whether the v1 "extend booking" call takes query params (`PUT bookings/{id}?end=&duration=`, per my sub-agent's research) or is just the general-purpose `PUT bookings/{id}` with a JSON body (per the more rigorous existing `docs/research/3v-rooms-api.md`, which found no dedicated extend op in v1 at all). Documented in [[rooms-api-integration]]. Whoever writes the Rooms client should check this against the live API first.
- Device-local-API IP discovery (needed for fast feedback animations) assumes the device's poll-request source IP is usable to call back — not verified against real hardware.
- NFC checkin's `pin=<startup password>` approach is untested — may need revisiting if the API rejects it or the account lacks rights.
- `api-contract/openapi.yaml` is still an empty placeholder — the design is now settled enough to populate it.

## Next steps

- Populate `api-contract/openapi.yaml` with the PuKK-facing endpoints (poll `GET /`, button/NFC `POST /`) now that the design settled what each response needs to contain.
- Scaffold the actual Go module (`go.mod`, entrypoint) when the user asks either agent to start implementing. Suggest verifying the flagged v1-extend-endpoint-shape question early, since it blocks the extend/checkout code path.
