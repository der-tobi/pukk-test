---
name: booking-cache-design
description: How the server caches 3V Rooms freebusy data to avoid polling on every PuKK request while keeping the rolling 60-min ring display accurate
metadata:
  type: project
---

The PuKK's LED ring continuously displays a sliding `[now, now+60min]` window. The server must not call the 3V Rooms freebusy API on every device poll ([[tech-stack]]: device default poll is 5s), but a naive "fetch exactly the display window, refresh on a timer" cache has a bug: by the time the next refresh fires, the display window has already slid past what was fetched, so the newly-entering far edge of the ring (the last 5-min slice becoming visible) has no data behind it.

**Decision:** decouple the upstream lookahead range from the refresh cadence, and always over-fetch beyond the display window:

```
lookahead_buffer = display_window (60min, fixed) + rooms_refresh_interval + safety_margin
```

- `rooms_refresh_interval` — **configurable constant, default 5 minutes.** Governs how often the server re-fetches `/freebusy` from 3V Rooms in the background.
- `safety_margin` — assumed default 10 minutes (not yet confirmed with user — adjustable if they want tighter/looser staleness tolerance).
- With the defaults: fetch `[now, now+75min]` (interval=5, matching LED granularity) every 5 minutes. This guarantees ≥10 min of buffered margin ahead of the display window at all times, even right before a scheduled refresh.
- If `rooms_refresh_interval` is changed, `lookahead_buffer` recalculates automatically — the margin invariant (`buffer ≥ display_window + refresh_interval + safety_margin`) should never need manual re-tuning.
- **Immediate refresh trigger:** any local booking mutation (button-press extend, checkin, NFC checkin) forces an out-of-band refresh so self-caused changes appear on the PuKK's very next poll, not after waiting out the timer.
- Known trade-off: an externally-made booking (not via this PuKK's button) can take up to `rooms_refresh_interval` to appear on the ring.

**Why:** [[project-context]] — 3V Rooms must not be polled on every device request; the ring must still show a precise, gap-free rolling hour. This design was worked out by catching a real bug in an earlier draft (fixed-interval cache with no buffer) during setup discussion with the user.

**Resolved (2026-08-13, via `/grilling`):** if a background refresh fails repeatedly and the buffer runs dry before the display window is fully covered, or before the very first successful fetch completes at startup, any time-slice with no confirmed data renders as **busy** (red), not free or stale-marked. A PoC occasionally showing a free room as busy is a minor annoyance; showing a busy room as free risks someone interrupting a meeting — busy is the safer default. See [[led-ring-design]]. Also resolved: token-refresh failures ([[rooms-api-integration]]) fall back to the same "serve last-known-good cache" behavior rather than crashing or re-prompting for the password mid-run.

**How to apply:** whichever agent implements the 3V Rooms integration should build this as the caching layer from the start, not bolt it on later — it's load-bearing for LED correctness, not an optimization.
