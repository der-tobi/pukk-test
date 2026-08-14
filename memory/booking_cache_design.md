---
name: booking-cache-design
description: How the server caches 3V Rooms freebusy data to avoid polling on every PuKK request while keeping the rolling 60-min ring display accurate
metadata:
  type: project
---

The PuKK's LED ring continuously displays a sliding `[now, now+60min]` window. The server must not call the 3V Rooms freebusy API on every device poll ([[tech-stack]]: device default poll is 5s), but a naive "fetch exactly the display window, refresh on a timer" cache has a bug: by the time the next refresh fires, the display window has already slid past what was fetched, so the newly-entering far edge of the ring (the last 5-min slice becoming visible) has no data behind it.

**Current PoC override (2026-08-14):** live testing changed the upstream `freebusy` resolution to `interval=1` minute while keeping the physical display at 12 five-minute LEDs. The cache aggregates one-minute source buckets into rolling five-minute LED slots starting at the current poll time. This avoids a meeting that ended at e.g. 23:15 keeping the last LED red at 23:18. The live refresh cadence is 30 seconds, while the lookahead fetch window remains 75 minutes.

**Exact booking overlay (2026-08-14):** live testing around a `00:00` booking showed `freebusy` could place the first 15 minutes red even when the actual booking should appear later in the rolling hour. The app now uses exact `bookings/find` start/end ranges as the authoritative source for booked LED placement when that lookup succeeds. `freebusy` remains useful as fallback, for diagnostics, and for unknown-data fail-closed behavior.

**Future exact-time cache improvement (2026-08-14):** for the next iteration, cache exact meeting start/end times for at least the next rolling hour, not only as a render overlay. This is needed for interactions that fill a free gap between the current meeting and the next one: e.g. if there are 23 minutes until the next booking, button extension/ad-hoc selection should cap at that exact 23-minute boundary instead of relying on coarse `freebusy` buckets or fixed 15-minute steps. Do not implement now; preserve as follow-up context.

**LED quantization rule (2026-08-14):** both exact bookings and `freebusy` fallback use the same rolling LED slots: slot 0 is `[now, now+5min)`, slot 1 is `[now+5min, now+10min)`, and so on. Future bookings/freebusy are sampled at each slot's midpoint, not by "any overlap", so a booking at `00:00-00:15` seen at `23:53` renders `GRRRGGGGGGGG`: first LED green, next three LEDs red, remaining LEDs green. Active/current bookings still use overlap so a booking that is already in progress stays red until it actually ends.

**Returned interval inference (2026-08-14):** live testing around `00:40` with a `01:00-01:30` booking showed the app could render almost everything red if Rooms returned fewer `freebusy` bits than requested. The cache now infers the actual source interval from the fetched window length and returned bitstring length for plausible Rooms intervals (`1`, `5`, `15` minutes). Source bits are indexed relative to the fetched window start, not by wall-clock `time.Truncate`, because the bitstring intervals are defined inside the requested `[start,end]` range. Exact booking ranges remain authoritative when `bookings/find` succeeds; the inferred interval only affects `freebusy` fallback.

**Coarse fallback edge handling (2026-08-14):** live testing at `00:52` with a `01:00` meeting showed the 15-minute `freebusy` fallback can still turn the leading LEDs red too early because a source bucket such as `00:50-01:05` is busy if it overlaps the meeting at all. For source intervals coarser than the 5-minute LED display, the cache now trims the leading and trailing edges of contiguous busy runs before projecting them to 5-minute LEDs. This is only a fallback heuristic; exact `bookings/find` ranges remain the preferred path and should be checked first in `/debug/status`.

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
