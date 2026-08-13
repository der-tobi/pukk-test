# proDVX PuKK Device API — Research Notes

Research target: primary-source facts about the PuKK device's own protocol (what it sends to our server, what our server can send back, what the device's local REST API offers), needed to design the Go server's HTTP handlers and LED rendering.
Date of research: 2026-08-13.

**Confidence note:** this pass is less exhaustive than `docs/research/3v-rooms-api.md` — it's a single research pass over the GitHub README and the two embedded Swagger specs, not a full read of every linked page (e.g. the Dropbox-hosted "User and Integration manual" PDF linked from the README was not fetched). Treat field-level details as a strong starting point to verify against the live device during implementation, not as verbatim-quoted spec text throughout.

## Sources fetched

- https://github.com/ProDVX/PuKK_Overview (README.md) — repo also contains `.gitignore`, `LICENSE`, `assets/graphic.drawio`, `package.json`, `server.js`; no config file (no `default_config.jsonc` or similar) exists in this repo.
- https://prodvx.github.io/docs/PuKK_Workspace/PuKK_Event_Specification.html — titled "ProDVX PuKK Remote Server API": describes requests the **PuKK sends to our server**.
- https://prodvx.github.io/docs/PuKK_Workspace/PuKK_API_Specification.html — titled "ProDVX PuKK REST API Specification": describes the REST API **on the device itself**, base `http://{ip_address}/api`.

Both spec pages are Swagger-UI shells with the OpenAPI YAML inlined in a `<script>` block — a plain HTML→markdown fetch renders blank; the raw HTML had to be pulled to read the actual spec.

## 1. LED control (server → device, via poll response)

`set_leds_individual` (`CmdSetLedsIndividual`) takes `led_values.colors`: an array of exactly 12 (`minItems`/`maxItems`: 12) `LedColor` objects, each `{brightness: 0-100, red: 0-255, green: 0-255, blue: 0-255}` — full arbitrary color per LED, no palette restriction. LEDs addressed "starting 12h" (i.e. index 0 = 12 o'clock). Plus `duration_ms` (0 = always on).

Other command variants in the same discriminated union: `set_leds` (static, all LEDs same color), `set_leds_marquee` / `set_leds_rainbow` (rotating), `set_leds_breathe` (`LedDynamic`: `color, duration_ms, speed_ms`), `set_leds_quadrant` (4 zones), `set_leds_clock` (ticking-clock animation), `set_leds_off`, plus an `ota` command.

**Not confirmed**: whether `set_leds_breathe` (or any dynamic command) can target a subset of the 12 LEDs while the rest stay on a static `set_leds_individual` array in the same response, or whether a poll response can only carry one command type covering the whole ring. This is why [[led-ring-design]] chose to simulate pulsing/blinking via alternating `set_leds_individual` calls rather than depend on this.

## 2. Poll mechanism (device → server, and our response)

Event spec defines `GET /` on our server with query params `action=poll&mac=<device-mac>`. Our 200 response body **is** the `Command` object (one of the LED commands above, or `ota`) — the poll response itself carries the display update. `204` = no new command. There is no separate push/webhook channel for LED state; it only ever travels in a poll response.

**Not found**: poll-interval configurability from either spec or the README — the README only documents setting the device's "CMS Server Address" (`http://<ip>:3001`), nothing about interval. (The user separately confirmed in this session that poll interval *is* configurable on the device, overriding this gap.)

## 3. Button events (device → server)

`POST /` on our server, `action` ∈ `short_press | double_press | multiple_press | long_press_3s | long_press_5s | long_press_15s`. The device buckets press duration into **three fixed long-press thresholds itself** — there is no raw millisecond duration field and no separate down/up event pair. A single event fires once the press is fully classified; there is no live "still held" signal and no distinct "released early" event. No debounce behavior is documented.

This directly shaped [[button-state-machine]]'s longpress-checkout redesign (project.md's original 2500ms threshold isn't an available bucket; 3s is the nearest, and the live progressive-animation idea isn't implementable against this event model).

## 4. NFC events (device → server)

Same `POST /` endpoint, `action=nfc`, body = `NfcData`: `dataType` (Text/JSON/URI/VCard), `nfcId` (UID), `nfcValue`, `nfcSpecs {cardType, sensRes, selRes, bitrate, afi, dsfid, ndefUpdated}`.

## 5. Breathing/pulsing

Built-in device-side via `set_leds_breathe` (see §1) and the marquee/rainbow rotation commands — the device can animate without our server pushing successive frames. [[led-ring-design]] deliberately doesn't rely on this (see the "not confirmed" note in §1) and instead simulates pulse/blink by alternating colors across our own successive updates.

## 6. Bonus finding: device's own local REST API

The device also exposes a REST API on itself (`http://{device-ip}/api/setLeds`, `/api/setLeds/individual`, etc. — same payload shapes as §1, minus the `command` discriminator field since the endpoint implies it) for **direct proactive control outside the poll cycle**. This is the mechanism [[led-ring-design]] and [[button-state-machine]] rely on for feedback that's faster than the poll interval (NFC soft-transition + blitz, longpress-checkout confirmation). Requires knowing the device's IP — the working assumption is to read it from the source address of the device's own poll requests, not confirmed against real hardware yet.

## Open questions / not found — summary

- Poll-interval configurability: not documented in these sources (user confirmed out-of-band that it is configurable).
- Whether dynamic LED commands (breathe/marquee) can be scoped to a subset of LEDs alongside static individual colors elsewhere: not found — avoided by design rather than resolved.
- Device-local-API auth/pairing requirements: not checked.
- The Dropbox-hosted "User and Integration manual" PDF (linked from the README) was not fetched — may contain more detail on poll-interval config and the local API.
