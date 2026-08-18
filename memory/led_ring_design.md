---
name: led-ring-design
description: How the PuKK's 12-LED ring maps to the rolling 60-min window — orientation, rotation, gradient scope, and how pulsing/blinking feedback is produced
metadata:
  type: project
---

**Current PoC override (updated 2026-08-17):** live-device testing simplified the ring palette. See [ring_poc_simplification.md](ring_poc_simplification.md). The orientation below still applies, but the current implementation now uses a rolling window from the current poll time instead of wall-clock-aligned buckets. Unchecked current-booking LEDs are orange; checked-in current-booking LEDs are red; later/future busy ranges are violet; free time is green; pending/termination feedback is blue. Gradient and pulsing are disabled in the current implementation.

**Live-device transport update (2026-08-14):** the server still returns a poll command, but also pushes the ambient red/green ring through the PuKK local REST API on every poll. See [device_led_push.md](device_led_push.md). This supersedes the older assumption below that the regular ambient ring state is delivered purely through poll responses.

The PuKK has 12 LEDs arranged like the 5-minute marks of a clock face. Design settled via a full `/grilling` session on 2026-08-13 (see `memory/claude/session.md` for that date's session notes).

**Orientation:** LED index 0 (12 o'clock) is always the current `0-5min` bucket. Time increases clockwise: index 1 is `5-10min`, index 2 is `10-15min`, ..., index 11 is `55-60min`. A future booking first appears in the `55-60min` slot at index 11 (top-left, immediately before 12 o'clock), then moves counter-clockwise toward index 0 as time passes. This matches the intended "like a clock" behavior: the current booking ending soon is shown in `0-5min`, and newly visible future state enters at `55-60min`. Verified against all four worked examples in `project.md` (half-hour-left -> right half red; 15-min-left -> upper-right quarter red).

**Superseded rotation model:** the original design used fixed wall-clock 5-minute boundaries (:00, :05, ... :55). Live testing showed users expected the ring to mean "from now" instead. Current PoC implementation uses slot 0 = `[now, now+5min)`, slot 1 = `[now+5min, now+10min)`, and so on.

**Future-slot quantization:** for future bookings and `freebusy` fallback, a 5-minute LED is marked busy when the slot midpoint falls inside the busy range. Do not use "any overlap" for future ranges, because that makes edge-aligned bookings light one extra LED. Example: at 23:53, a 00:00-00:15 booking renders `GRRRGGGGGGGG`.

**Known rolling-window quantization artifact (2026-08-18):** live observation with bookings `00:00-00:30` and `00:45-01:00`: at `00:14`, the ring showed LEDs 0-4 red, then 2 green, then violet; at `00:15`, it showed LEDs 0-3 red, then 3 green. The user expected the clear/free gap between meetings to stay visually the same length. This is probably caused by the rolling `now` window plus 5-minute LED resolution and the current mix of active-overlap for current bookings with midpoint quantization for future bookings. Treat as an open design question, not yet a requested fix: possible approaches include wall-clock-aligned buckets again, a consistent midpoint rule for active ranges, preserving gap length by anchoring to booking boundaries, or accepting that exact gap lengths will jump at minute/bucket boundaries on a 12-LED ring.

**Gradient scope — deliberately narrow:** only the position-0 ("now") LED ever shows the 1-minute-step red→green gradient from `project.md` (countdown of the current 5-min block). Every other LED — including the far one that newly enters at each rotation, and any slot where a future booking's start/end doesn't land on a clean 5-min boundary — is solid-colored per whichever discrete state applies (busy/upcoming-violet/free/select-blue), no fading. This was an explicit simplification: the user initially wondered whether upcoming bookings needed their own fade-in (e.g. a meeting starting in 57min fading from green to violet as it enters the far edge), then chose the simpler rule instead — "keep it simple... only the now-slot gets a gradient."

**Pulsing/breathing:** used for checkin-required (orange) and meeting-ending (last block, red) states. It's a brightness/color animation *layered on top of* the gradient color, not a replacement — the gradient color persists underneath, pulsing just adds the animation. Distinct from **blinking**, which is only used for the longpress-checkout confirmation LED (see [[button-state-machine]]).

**How pulsing/blinking are produced:** the device supports native `set_leds_breathe`, but it's unconfirmed whether a dynamic command can be scoped to one LED while the rest stay static via `set_leds_individual` in the same response (see `docs/research/pukk-device-api.md` §1). Rather than depend on that, the server simulates pulse/blink by alternating the affected LED's color between two values across successive poll responses — consistent with the "recompute everything fresh, stateless" rule below.

**Color computation — stateless:** all 12 LED colors are recomputed fresh on every poll response, derived from the cached freebusy data ([[booking-cache-design]]) plus current poll time. No ticking goroutine maintains ring state independently of polls — this avoids drift/sync bugs between a background ticker and what's actually served, and the 5s poll cadence is fine-grained enough for 1-minute gradient steps.

**Fast feedback outside the poll cycle:** some effects need multiple frames faster than the 5s poll interval. Longpress checkout pushes a blue sweep over the current booking LEDs from the meeting end toward now, then a green sweep after Rooms release succeeds. NFC check-in pushes an orange-to-red sweep over the unchecked current-booking LEDs after Rooms check-in succeeds. Fast frames are pushed via the device's own local REST API (`http://{device-ip}/api/setLeds...`, see `docs/research/pukk-device-api.md` §6) rather than waiting for the next poll.

**Closed-hours fallback guard (2026-08-17):** if exact booking ranges are unavailable but an active booking is known, the active booking range constrains the rendered busy/check-in LEDs. Coarse `freebusy` can be all-busy outside opening hours, so it must not expand a 15-minute active booking into a full orange/red ring.

**Adjacent active/follow-up boundary (2026-08-18):** when a current booking is extended exactly to the next booking's start, the renderer must still keep the follow-up booking violet. Do not infer "first busy range" purely from contiguous busy LEDs when an active booking is known; active/current slots are red or orange, and non-active exact busy slots remain upcoming-violet even if they touch the active booking with no green gap.

**How to apply:** whichever agent implements LED rendering should build the "stateless, recompute every poll" renderer first, then layer the local-API push path on top for the handful of fast one-off animations — don't build them as one system.
