---
name: device-led-push
description: Live-device finding that PuKK LED state should be pushed through the device-local REST API, not only returned in poll responses
metadata:
  type: project
---

On 2026-08-14 live testing showed the physical PuKK ring could stay blue even when the server-rendered `lastLedHex` state was red/green.

The ProDVX Remote Server API spec is internally inconsistent: it defines command responses for poll, but its introduction says response info will not be consumed and device control should be handled through the PuKK REST API. The implementation now keeps returning the LED command in `GET /?action=poll` for compatibility, but also pushes the same individual LED values to the polling device via:

`POST http://<device-ip>/api/setLeds/individual`

The local REST payload intentionally contains only `led_values`, without the `command` discriminator, because the endpoint itself implies individual LED control.

The device IP comes from the source address of the PuKK poll request. `/debug/status` reports `lastDevicePushOk`, `lastDevicePushError`, and `lastDevicePushIp` so live tests can distinguish:

- Rooms did not provide the expected freebusy bits.
- The server rendered the wrong LED colors.
- The local PuKK REST push failed or was not accepted.

The app still refreshes Rooms via `freebusy`, not an event feed. For the PoC, the background refresh cadence was reduced to 30 seconds so externally-created bookings appear quickly without calling Rooms on every 5-second PuKK poll. The freebusy lookahead window remains 75 minutes to preserve the buffered 5-minute-slot display model.
