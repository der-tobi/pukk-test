---
name: led-ring-design
description: How the PuKK's 12-LED ring maps to the rolling 60-min window — orientation, rotation, gradient scope, and how pulsing/blinking feedback is produced
metadata:
  type: project
---

The PuKK has 12 LEDs arranged like the 5-minute marks of a clock face. Design settled via a full `/grilling` session on 2026-08-13 (see `memory/claude/session.md` for that date's session notes).

**Orientation:** LED index 0 (12 o'clock) is always the current `0-5min` bucket. Time increases clockwise: index 1 is `5-10min`, index 2 is `10-15min`, ..., index 11 is `55-60min`. A future booking first appears in the `55-60min` slot at index 11 (top-left, immediately before 12 o'clock), then moves counter-clockwise toward index 0 as time passes. This matches the intended "like a clock" behavior: the current booking ending soon is shown in `0-5min`, and newly visible future state enters at `55-60min`. Verified against all four worked examples in `project.md` (half-hour-left -> right half red; 15-min-left -> upper-right quarter red).

**Rotation model — wall-clock-aligned, discrete:** the 12 slots are fixed to wall-clock 5-minute boundaries (:00, :05, ... :55), not continuously slid relative to "now." Every 5 minutes, at the boundary, the oldest slot (position 0) retires, everything shifts one step toward position 0, and a new solid-colored slot appears at position 11 reflecting whatever is newly visible 55-60 minutes out.

**Gradient scope — deliberately narrow:** only the position-0 ("now") LED ever shows the 1-minute-step red→green gradient from `project.md` (countdown of the current 5-min block). Every other LED — including the far one that newly enters at each rotation, and any slot where a future booking's start/end doesn't land on a clean 5-min boundary — is solid-colored per whichever discrete state applies (busy/upcoming-violet/free/select-blue), no fading. This was an explicit simplification: the user initially wondered whether upcoming bookings needed their own fade-in (e.g. a meeting starting in 57min fading from green to violet as it enters the far edge), then chose the simpler rule instead — "keep it simple... only the now-slot gets a gradient."

**Pulsing/breathing:** used for checkin-required (orange) and meeting-ending (last block, red) states. It's a brightness/color animation *layered on top of* the gradient color, not a replacement — the gradient color persists underneath, pulsing just adds the animation. Distinct from **blinking**, which is only used for the longpress-checkout confirmation LED (see [[button-state-machine]]).

**How pulsing/blinking are produced:** the device supports native `set_leds_breathe`, but it's unconfirmed whether a dynamic command can be scoped to one LED while the rest stay static via `set_leds_individual` in the same response (see `docs/research/pukk-device-api.md` §1). Rather than depend on that, the server simulates pulse/blink by alternating the affected LED's color between two values across successive poll responses — consistent with the "recompute everything fresh, stateless" rule below.

**Color computation — stateless:** all 12 LED colors are recomputed fresh on every poll response, derived from the cached freebusy data ([[booking-cache-design]]) plus wall-clock time at request time. No ticking goroutine maintains ring state independently of polls — this avoids drift/sync bugs between a background ticker and what's actually served, and the 5s poll cadence is fine-grained enough for 1-minute gradient steps.

**Fast feedback outside the poll cycle:** some effects need multiple frames faster than the 5s poll interval — the NFC "soft transition from orange to red, then a blitz flash," and the longpress-checkout confirmation flash / optional counter-clockwise sweep (see [[button-state-machine]]). These are pushed via the device's own local REST API (`http://{device-ip}/api/setLeds...`, see `docs/research/pukk-device-api.md` §6) rather than waiting for the next poll. The regular ambient ring state (gradient, busy/free/upcoming, simulated pulsing) continues to be delivered purely through poll responses.

**How to apply:** whichever agent implements LED rendering should build the "stateless, recompute every poll" renderer first, then layer the local-API push path on top for the handful of fast one-off animations — don't build them as one system.
