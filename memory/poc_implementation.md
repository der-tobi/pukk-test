---
name: poc-implementation
description: Current implementation state of the PuKK Go PoC server
metadata:
  type: project
---

Implemented on 2026-08-13 as a single Go module in `/workspaces/pukk-test`.

## What exists

- `main.go` starts one `net/http` server on port 5000, prompts once for the 3V Rooms password for `tobiapi4`, prints local URLs for PuKK configuration, and refreshes Rooms data every 30 seconds for live PoC testing.
- The app reads Rooms `freebusy`, not a Rooms event feed. External bookings appear on the next background refresh; the 30s cadence is intended to make that fast without calling Rooms on every PuKK poll.
- The app also reads upcoming bookings via v2 `bookings/find` during refresh. These exact start/end ranges are now authoritative for where booked LEDs appear, so a midnight booking should be placed at its real future slot instead of trusting potentially misaligned `freebusy` busy bits.
- Startup no longer blocks on the initial 3V Rooms refresh. The HTTP listener starts immediately; unknown availability renders busy until refresh succeeds.
- The startup refresh now fetches 3V Rooms `freebusy` first and independently from active-booking lookup, so the availability bitstring is available as early as possible even if booking lookup is slow or fails.
- `server.go` implements the PuKK remote server surface:
  - `GET /?action=poll&mac=...` returns a `set_leds_individual` command.
  - `POST /?action=...&mac=...` accepts button and NFC events.
  - `GET /healthz` returns plain `ok` for network reachability tests.
  - `GET /debug/status` reports request counts, last request, last refresh status, latest availability bitstring/window, pending button state, and observed device IPs.
- `renderer.go` recomputes the 12 LEDs statelessly on every poll from cached availability, active booking state, and provisional button selection.
- Per live-device simplification, the normal ring uses red `#FF0000` for busy/booked and bright green `#00FF00` for free. Pending button selections render blue `#006DFF` during the commit window, then turn red after commit. The earlier orange/violet/gradient/pulsing colors are disabled for the PoC.
- `device_client.go` pushes the rendered ring to the PuKK's local REST API (`POST http://<device-ip>/api/setLeds/individual`) on every poll when a real device IP is known. This is needed because live testing indicated the PuKK may not consume LED commands returned in the poll response.
- `cache.go` over-fetches `[now bucket, now bucket + 75min]` by default, requests Rooms `freebusy` at 1-minute resolution, infers the actual returned source interval from bitstring length when Rooms returns 1/5/15-minute buckets, aggregates that into 12 rolling five-minute LEDs from the current poll time, trims coarse fallback busy-run edges to avoid premature red LEDs, and treats unknown slices as busy. In normal successful refreshes, exact booking ranges from `bookings/find` decide red booked slots; freebusy remains fallback/diagnostic data.
- The display window is rolling from `now`, not fixed to wall-clock `:00/:05` buckets. Future booking/freebusy slots are marked busy by slot midpoint, while active/current ranges still use overlap. Example: at 23:53, a 00:00-00:15 booking renders `GRRRGGGGGGGG`.
- `app.go` implements short-press extend/ad-hoc selection with a 5s commit window, NFC checkin, and 3s longpress release/checkout.
- Button provisional commits are guarded by a generation token so stale timers from earlier button presses cannot commit after a later press reset the 5s window.
- Poll has a fallback for expired provisional selections: if the 5s deadline has passed but the timer callback has not run, the poll removes the pending overlay, renders the selection optimistically as busy, and commits in the background.
- Successful local ad-hoc/extend commits are rendered optimistically as the active busy booking immediately, even if 3V Rooms `freebusy` or active-booking lookup lags behind.
- Empty-resource button presses render blue in 15-minute pending steps: 1 press = 15min, 2 = 30min, 3 = 45min. For active bookings, one press extends only to the visible `now+60min` edge if less than 15 minutes remain in that visible hour. Button add frames animate blue clockwise, undo frames animate green counter-clockwise, and commit frames animate blue to red clockwise.
- When exact `bookings/find` ranges are known, button extension/ad-hoc selection is capped by the exact next booking start, not by coarse 5/15-minute `freebusy` slots. A 23-minute free gap can now be filled exactly to the next meeting boundary.
- `rooms_client.go` implements the 3V Rooms HTTP adapter:
  - token: `POST /connect/token` with `grant_type=basic`, `client_id=basic-auth`, `scope=rooms_api`, `user=tobiapi4`
  - freebusy: v2 `/ressources/44/freebusy`
  - active/upcoming booking lookup: v2 `/bookings/find`
  - ad-hoc booking: v2 `/bookings/quickbooking`
  - extend: v1 `PUT /bookings/{id}?end=...`
  - checkin: v1 `PUT /bookings/{id}/checkin?pin=...`
  - checkout/release: v1 `cancheckout`, then `checkout?sendMail=false` or `release`
- The Rooms booking decoder accepts the documented `BookingDto` `Begin`/`End` fields, the app's internal `start`/`end` names, and `RessourceId`/`resourceId` spellings. This prevents exact bookings from being discarded while `freebusy` still shows broad busy intervals.

## Verification

`go test ./...`, `go build ./...`, `GOOS=windows GOARCH=amd64 go build -o pukk-test.exe .`, and `git diff --check` pass in the devcontainer.

Smoke run inside the devcontainer showed only `172.17.0.3`, a Docker bridge address. A physical PuKK on WiFi probably cannot reach that address. For real-device testing, run `pukk-test.exe` on the Windows host and use the Windows WiFi IPv4 address.

## Known risks

- PuKK device-local REST push is implemented for the ambient ring on every poll. Dedicated faster NFC/longpress animations are still not implemented.
- The active-booking lookup uses the documented v2 `bookings/find` filters, but this still needs live-tenant verification with resource 44.
- The Windows executable is a local build artifact and is ignored by git.
- Windows Firewall or WiFi client isolation can still block the PuKK even when the app runs on the Windows host. Use `http://<windows-wifi-ip>:5000/healthz` from another device on WiFi as the first test.
- `/debug/status` includes `pendingUntil`, `pendingStart`, `pendingEnd`, and `pendingBlocks` while a button selection is still provisional.
- `/debug/status` includes `availabilityBits`, `availabilityStart`, `availabilityEnd`, `availabilityIntervalMinutes`, and `lastAvailabilityRefreshOk` after the FreeBusy cache refresh succeeds.
- `/debug/status` includes `exactBusyKnown`, `exactBusyCount`, `exactBusyRanges`, `lastBookingsRefreshOk`, and `lastBookingsRefreshError` for diagnosing whether exact booking ranges or coarse freebusy fallback drove the ring.
- `/debug/status` includes `lastLedHex`, `lastBlueLeds`, and `lastRedLeds` from the last poll response, plus `lastDevicePushOk`, `lastDevicePushError`, and `lastDevicePushIp` for the PuKK local REST push.
