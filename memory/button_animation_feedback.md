---
name: button-animation-feedback
description: Requested immediate PuKK button feedback animation for provisional extension and undo
metadata:
  type: feedback
---

On 2026-08-14, after the red/green rolling ring and Rooms extension calls worked on the live PuKK, the user requested richer immediate feedback for button presses.

Desired active-booking scenario:

- If 15 minutes remain, LEDs 0-2 stay red.
- First button press extends by 15 minutes and immediately turns LEDs 3-5 blue, ideally animated in order 3, then 4, then 5.
- Second press turns LEDs 6-8 blue in the same clockwise sequence.
- Third press turns LEDs 9-11 blue in the same sequence if there is no blocking next meeting.
- The next press begins undo and turns LEDs 11, then 10, then 9 green again.
- If no further press arrives before the commit delay, the selected blue LEDs turn red from the smallest index to the largest, indicating the clock is extending while the Rooms request is committed.

Implementation direction:

- Keep the normal ambient/poll-rendered ring red/green only.
- Implement this as a device-local REST animation using the latest PuKK IP observed from poll requests.
- Pass the event `mac` into app event handling so a POST button event can resolve the last known device IP for that PuKK.
- Reuse the existing pending-selection generation token so stale animations stop when a later press changes the pending selection.
