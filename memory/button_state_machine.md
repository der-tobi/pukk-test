---
name: button-state-machine
description: Button-press extend/undo cycle, the 5s provisional-commit window, and the longpress-checkout gesture as redefined to match actual device capabilities
metadata:
  type: project
---

**Current PoC override (2026-08-14):** provisional button selections render blue during the commit window, with immediate device-local animation frames; after commit they become red booked time. See [ring_poc_simplification.md](ring_poc_simplification.md) and [button_animation_feedback.md](button_animation_feedback.md). The 5s delayed commit, LIFO undo cycle, and max `now+60min` cap below still apply.

Settled via `/grilling` on 2026-08-13 (see `memory/claude/session.md`). Applies equally to extending a current meeting and to building an ad-hoc booking from "now" — same mechanism either way.

**Extension algorithm:** each press either adds `min(15, headroom)` — where headroom is capped by the next meeting's start or the 60-min ring edge, whichever comes first — or, once no headroom remains, starts undoing previous additions in LIFO order back to baseline, then the cycle repeats (extend → extend → ... → undo → undo → ... → extend again). The "+5" example in `project.md` (20 minutes until the next meeting) is just an instance of a capped addition, not a distinct fixed step.

**Provisional commit:** each press turns the affected block blue immediately. If no further press happens for 5 seconds, the server commits the change to 3V Rooms ([[rooms-api-integration]]) and the block turns green. If the commit call fails, revert to the pre-press state and log loudly — no retry, no complicated failure handling. This is intentionally the simplest option for a PoC.

**Max reach = ring window, intentionally coupled:** the max elongation ("now + 1h") is deliberately the same as the ring's fixed 60-min display window — a 12-LED, 5-min-resolution ring has no way to represent state beyond what it displays anyway, so decoupling booking range from display range would create bookable-but-invisible time. Confirmed: the longest any booking can ever be (ad-hoc or extended) is 60 minutes from now.

**Longpress checkout — redefined from the original spec.** `project.md` originally described a 2500ms hold with a live progressive spin-to-green animation, cancelable by early release with a wind-back animation. The device API doesn't support this: it only reports a single classified event *after* release (`long_press_3s/5s/15s`, no raw duration, no down/up pair — see `docs/research/pukk-device-api.md` §3). Resolved design:
- Threshold snapped to the device's native **3s** bucket (nearest to 2500ms).
- No live animation while held — the server gets no signal until the press is fully classified.
- On receiving `long_press_3s`: look up the current booking on this resource (on-demand, see [[rooms-api-integration]]), push a blue sweep over the remaining current-booking LEDs from the meeting end back toward now via the device's local REST API (see `docs/research/pukk-device-api.md` §6), trigger checkout/release, then sweep those LEDs green on success.
- Early release (under 3s) may still be classified by the device as `short_press`, `double_press`, or `multiple_press`; the API simply gives us no raw "button down/up" signal and no way to animate or cancel a progressive hold before classification. To avoid an accidental checkout attempt turning into an extension/booking, implementation should suppress normal short-press handling when it can identify that the event was part of a near-long hold, or otherwise accept this as a hardware behavior to verify on the real device before demo.
- Longpress only does anything when there's a current booking; no-op otherwise.

**NFC:** accepts any tag, no UID/card validation (matches the PoC's "hardcode everything" framing), but currently does not mutate Rooms. The current booking renders orange regardless of `checkinConfirmed`, so NFC can still be demonstrated as a received device event without relying on a working Rooms check-in mutation. The tested v1 `bookings/{id}/checkin?pin=...` path returned "Mit der angegebenen Buchung kann kein Checkin durchgeführt werden"; local Rooms source shows that a non-empty `pin` is checked against reservation `DoormanList` PINs, not the startup API account password. The documented PuKK remote-server event model exposes a single `nfc` event with card data, not a tap duration or card-present/card-removed pair, so an NFC 3-second hold cannot be implemented unless live hardware provides repeated/undocumented signals.

**Concurrency:** a single in-memory mutex per resource serializes all mutation attempts (button extend, NFC checkin, longpress checkout) — no conflict detection or merge logic beyond "last one wins." Deliberately minimal: this is a one-device PoC where true simultaneous physical inputs can't happen anyway.
