---
name: ring-poc-simplification
description: Live-device simplification of PuKK ring colors and button feedback for the PoC
metadata:
  type: project
---

On 2026-08-14, after live PuKK testing, the ring UX was simplified for the PoC.

## Decision

Use only two ambient colors:

- Busy/booked = red `#FF0000`
- Pending/provisional button selection = blue `#006DFF`
- Free = bright green `#00FF00`

Do not use the earlier orange check-in color, violet upcoming-booking color, red-to-green gradient, or brightness pulsing in the normal poll-rendered ring. The original `#00FF7F` free color looked blue/turquoise on the physical device and was replaced with pure green `#00FF00`; the original red was replaced with pure red `#FF0000`. Current booked time remains red and free time remains green. Pending button selections are the exception: they render blue during the commit window and animate via device-local REST push.

The live PoC ring is rolling from the current poll time, not fixed to wall-clock `:00/:05/...` buckets. Future bookings are assigned to LEDs by the midpoint of each 5-minute LED slot. Example: at 23:44, a 00:00-00:15 meeting should render as first quarter green, second quarter red, left half green. At 23:53, that same 00:00-00:15 meeting should render `GRRRGGGGGGGG`: first LED green, next three LEDs red, remaining LEDs green.

## Button UX

On button presses against an empty resource:

- 1 press renders the next 15 minutes blue until commit, then red.
- 2 presses render the next 30 minutes blue until commit, then red.
- 3 presses render the next 45 minutes blue until commit, then red.
- Remaining visible time stays green.

On button presses during an active booking, the already-booked remaining time stays red and the provisional extension renders blue until commit, then red. Extension is capped by `now + 60min`, not by a wall-clock block from the booking's original start. Example: a 17:00-18:00 booking pressed at 17:10 extends to 18:10 when no later booking blocks the hour.

## Reason

The live device showed confusing yellow/orange/blue behavior. For the demo, clear binary state is more important than nuanced check-in/provisional/fading colors.
