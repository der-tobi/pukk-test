# PuKK Remote Server API

Date: 2026-08-13
Status: Implemented for PoC

## Context

The existing `api-contract/openapi.yaml` was a placeholder. The PuKK device talks to this Go server through the ProDVX remote server API:

- `GET /?action=poll&mac=...` for polling LED commands.
- `POST /?action=...&mac=...` for button and NFC events.

## Proposal

Document the two PuKK-facing endpoints in `api-contract/openapi.yaml`:

- `GET /` with `action=poll`, returning a `set_leds_individual` command with 12 LED colors.
- `POST /` with `action` values from the PuKK event spec, returning `204` when accepted.
- `GET /healthz` and `GET /debug/status` as local diagnostics for checking network reachability and whether PuKK requests have arrived.

The contract intentionally documents only the local Go server surface. Outbound 3V Rooms calls are implementation details and are not represented here.

## Notes

The response schema mirrors the PuKK LED command shape from `docs/research/pukk-device-api.md`: 12 `LedColor` values, each carrying brightness and RGB channels, inside `led_values.colors`.

`/debug/status` is intentionally allowed to grow with PoC diagnostics. On 2026-08-14, exact booking refresh fields were added (`exactBusyKnown`, `exactBusyCount`, `lastBookingsRefreshOk`, `lastBookingsRefreshError`) so live-device tests can distinguish exact booking-range rendering from coarse FreeBusy fallback rendering.
