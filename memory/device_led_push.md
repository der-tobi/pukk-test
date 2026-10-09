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

## Firmware button feedback (2026-10-08)

The user reports a white ring on button press and prefers controlling the feedback or retaining the existing ring. The [ProDVX user manual, FAQ section 9](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Manual.pdf) confirms firmware button feedback sets the ring to light yellow and can retain it if the CMS is unavailable. The app never sends white. The current [configuration schema](https://prodvx.github.io/docs/PuKK_Workspace/configuration/configuration_schema.json) and [local REST API](https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html) expose no button-feedback color/disable option. Keeping colors unchanged throughout the physical hold cannot be guaranteed from the server: the documented event API has no down/up signals.

The app now re-pushes the rendered ring even when a short press makes no visual change (for example, the active booking already fills the visible hour). Existing selection/checkout animations still control feedback once the classified event arrives. This replaces firmware feedback without waiting for the next poll in the unchanged-selection case. Live verification remains necessary; no device was connected during this investigation.

Later feedback on 2026-10-08: the user wants a clear indication that the 3s hold threshold has been reached and finds longpress animations unhelpful. Checkout now sends one static blue processing frame and the final ambient frame, with no sweeps. Polls keep the blue processing overlay while Rooms responds. Documented events lack down/up/raw-duration signals; a reliable server-timed threshold indicator while held requires additional firmware support or verified live event timing. The previous after-release assertion is a design assumption, not an explicitly documented timing guarantee.

The app still refreshes Rooms via `freebusy`, not an event feed. For the PoC, the background refresh cadence was reduced to 30 seconds so externally-created bookings appear quickly without calling Rooms on every 5-second PuKK poll. The freebusy lookahead window remains 75 minutes to preserve the buffered 5-minute-slot display model.

## Rückkanal-Korrektur (2026-10-09)

Der Nutzer sah beim ersten Poll keine Marketing-Sequenz und bei Tastendruck
Weiß. Regressionstests fanden einen echten URL-Fehler für Windows-typische
IPv4-mapped-IPv6-Adressen (`::ffff:192.168.…`) und native IPv6-Adressen.
`remoteIP` normalisiert nun gemappte Adressen; der Geräteclient baut für
native IPv6 korrekt geklammerte HTTP-Ziele. Palette und lokale API-Payload
wurden gegen die aktuelle Hersteller-Spezifikation geprüft.

Vor einem notwendigen langsamen Rooms-Booking-Lookup beim kurzen Tastendruck
wird die vorhandene Belegungsanzeige synchron wieder gesendet. So bleibt
die Firmware-Rückmeldung nicht während der gesamten Rooms-Anfrage stehen.
Bei bekannten exakten Buchungen folgt weiterhin direkt die Auswahlanimation.
Der konkrete Live-Rückkanal bleibt ohne Nutzerlogs unbestätigt.
