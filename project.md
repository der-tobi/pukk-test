# pukk-test — Project Definition

> **Working name:** PUKK Demo
> **Type:** Proof of concept — single-binary Go server for one proDVX PuKK device
> **Status:** Design phase complete (2026-08-13, via `/grilling`) — polling/caching, LED ring rendering, button/NFC/longpress state machines, and 3V Rooms API integration all resolved; see "Resolved Design Decisions" below. Implementation not yet started.
> **Tech stack:** Go, cross-compiled to a standalone Windows executable. Rationale: `memory/tech_stack.md`

---

## Resolved Design Decisions (2026-08-13)

The bullets below are the original brief, kept as-is for traceability. A full `/grilling` session worked through every open question and ambiguity in it; the resolved design lives in `memory/`, not inline here:

- [led_ring_design.md](memory/led_ring_design.md) — ring orientation, wall-clock-aligned rotation, gradient scope (only the "now" LED fades), pulse/blink rendering
- [button_state_machine.md](memory/button_state_machine.md) — extend/undo cycle, 5s provisional-commit, longpress-checkout redefined to match actual device capabilities, NFC
- [booking_cache_design.md](memory/booking_cache_design.md) — freebusy caching/buffer design, resolved stale/cold-start fallback (assume busy)
- [rooms_api_integration.md](memory/rooms_api_integration.md) — which 3V Rooms endpoints/versions, booking-ID discovery, auth/token fallback
- `docs/research/3v-rooms-api.md` and `docs/research/pukk-device-api.md` — primary-source facts backing the above

Two specifics from the brief below were changed during design review: the PuKK poll interval question ("i think of configuring it so, that we have a polling of 5 seconds. challenge this") was resolved to **keep 5s**, decoupled from the 3V Rooms refresh cadence (5min) rather than tied to it. The longpress bonus feature's 2500ms threshold was changed to **3s** to match the device's actual firmware, which only reports pre-classified long-press buckets (3s/5s/15s) — see `button_state_machine.md` for why the live progressive-animation idea also had to be dropped.

---

## What about
- Just a POC of the proDVX PUKK - so we hardcode everything. no config. quick and dirty -- except for the password for the 3v rooms api.
- you are free in the implementaton. the best way would be a single executable that you can build inside the devcontainer and i can then run it on my windows machine from the commandline.
- proDVX pukk inot s a simple device that communicates via restapi. it knows its server and a pollingrate. it is round and has 12 leds like the 5minutes parts of a clock. it has a nfc reader and a button. connectivity is via wifi.
- what to build
    - small poc server: 
        - log output from incomming requests on the commandline
        - app must show the ip to which the pukk can connect
        - app must connect to 3v ROOMS api: Auth: https://docs.3vrooms.app/3vrooms/api/
            - ressourceId: 44
            - grant_type=basic
            - client_id=basic-auth
            - scope=rooms_api
            - user=tobiapi4
            - password: ask in the app at startup
        - 3v rooms must not be called every thime the pukk calls. find a good way to make the pukk displaying the remaining time as precise as possible and also the uppcoming events in the next hour but without polling every minute. find a good compromise
    - config pukk
      - i think of configuring it so, that we have a polling of 5 seconds. challenge this
    - Led:
        - resolution for led display is 5 minutes
          - a running block of 5 minutes turns over 5 minutes from red to green - choose a decent gradient 1 minute steps
            - 5 minutes in the block are fully remaining: Hex: #ED2938
            - 5 minutes > remaining >= 4 minutes: Hex: #ee5b0f
            - 4 minutes > remaining >= 3 minutes: Hex: #e48400
            - 3 minutes > remaining >= 2 minutes: Hex: #cfa800
            - 2 minutes > remaining >= 1 minutes: Hex: #b0c800
            - 1 minutes > remaining >= 0 minutes: Hex: #82e53a
            - 0 minutes and 0 seconds in the block are remainig: Hex: #00FF7F
          - example
            - if a current meeting has 23 minutes remaining there are 4 red leds and the 5th is #e48400
        - current booking - busy = red or orange if checkin necessary red: Hex: #ED2938 - Orange: Hex: #FF9D09 and pulsing
        - upcomming booking - busy = #7d00b3
        - free = green - Hex: #00FF7F
        - select to book = blue
        - last 5 minutes of the booking red and pulsing / breathing
        - the ring allways displays the next hour
        - examples
            - at the start of a meeting that is >= 1h the ring is red.
            - if there is only half a hour left, the right half is red.
            - if there are 15 mins left the upper right quarter is red.
            - if the current meeting has 15 min left and after that is 15 free and after that is a booking of 30 mins the colors are like following
              - top right quarter 15 red, bottom right quarter green, left half dark violet
            - the led ring "turns" counter clockwise and allways displays the next hour.
        - actions on button press:
            - if we have a current meeting the meeting gets elonggated by 15 minuts by each buttonpress if status for this quarter is free. If there are 20 minutes until the next meeting, first buttonpress elonggates by 15 minuts and the second by 5. on a third press the 5 minuts get removed and the 4th removes the first extension of 15 minutes. each extended block first gets blue. if after 5 seconds no further buttonpress habens, the extension gets green. so the booking is delayed by 5 seconds after the buttonpress. the maximal elongation is allways now until plus 1 hour.
            - if we have no current meeting the mechanism works the same but from now on. so we can adhoc book for max. 1h.
                - happypath: no meeting in the next hour
                  - buttonpress 1 - 0-15 min get blue (top right)
                  - buttonpress 2 - 15-30 min get blue (bottom right)
                  - buttonpress 3 - 30-45 min get blue (bottom left)
                  - buttonpress 4 - 45-60 min get blue (top left)
                  - buttonpress 5 - 45-60 min get green (top left)
                  - and so on.
                - max booking is until the next booking or the next visible 1h
            - bonusfeature
              - longpress of 2500ms starts to turn the red lights each after another into green until the last led (first slow, then fast - like all leds are green in less than 2 seconds). the last led starts to blink for 1500ms in red then turns green . if the last led gets green the system checks the booking out. if user releases the button before the last led turns green the circle spins the same way back to the current remaining time.
    - NFC - accept all nfc readings and turn orange pulsing Leds of current booking into red and checkin the booking in rooms. use a soft transition from orange to red. confirm with lighting up to max brightnes for a short bliz.

Necessary docs:
- Example: https://github.com/ProDVX/PuKK_Overview
- 3V ROOMS api v2: https://docs.3vrooms.app/3vrooms/api/v2/
- 3V ROOMS api v1: https://docs.3vrooms.app/3vrooms/api/v1/
- ProDVX PuKK Remote Server API: https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html
- ProDVX PuKK REST API Specification: https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html

- Free Busy example (for simplicity without auth) : GET https://book.vnext.3vrooms.app/default/api/v2.0/ressources/44/freebusy?start=2026-08-13T10:00:00.000Z&end=2026-08-13T11:00:00.000Z&interval=5
- this returns:
{
  "FreeBusyData": "000000000000",
  "ResourceId": 44
}


