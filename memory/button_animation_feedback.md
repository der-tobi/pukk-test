---
name: button-animation-feedback
description: Requested immediate PuKK button feedback animation for provisional extension and undo
metadata:
  type: feedback
---

On 2026-08-14, after the red/green rolling ring and Rooms extension calls worked on the live PuKK, the user requested richer immediate feedback for button presses. Implemented in the Go app after baseline commit `7fdeac9`.

Desired active-booking scenario:

- If 15 minutes remain, LEDs 0-2 stay red.
- First button press extends by 15 minutes and immediately turns LEDs 3-5 blue, ideally animated in order 3, then 4, then 5.
- Second press turns LEDs 6-8 blue in the same clockwise sequence.
- Third press turns LEDs 9-11 blue in the same sequence if there is no blocking next meeting.
- The next press begins undo and turns LEDs 11, then 10, then 9 green again.
- If no further press arrives before the commit delay, the selected blue LEDs turn red from the smallest index to the largest, indicating the clock is extending while the Rooms request is committed.

Implementation direction:

- Normal booked/free time remains red/green.
- Pending button selections render blue during the 5-second commit window so normal polls do not overwrite the temporary feedback.
- Button press frames are also pushed immediately through the PuKK device-local REST API using the POST event's source IP, with fallback to the latest PuKK IP observed from poll requests for the same `mac`.
- The existing pending-selection generation token cancels stale animations when a later press changes the pending selection.
- The default frame delay is 100ms; tests override it to 1ms.
