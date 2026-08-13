# 3V ROOMS Booking API — Research Notes

Research target: primary-source facts for the Go server / PuKK room-display integration.
Date of research: 2026-08-13.

## 1. Sources fetched

| # | URL | Status | Notes |
|---|---|---|---|
| 1 | https://docs.3vrooms.app/3vrooms/api/ | 200 OK | Auth overview / getting-started guide (German). Fetched via WebFetch and raw curl for full text. |
| 2 | https://docs.3vrooms.app/3vrooms/api/v2/ | 200 OK | Static page is a stub; body only embeds a Swagger-UI widget (`window.onload` → `SwaggerUIBundle({url:"/api/v2.json", ...})`). No prose content of its own. |
| 3 | https://docs.3vrooms.app/3vrooms/api/v1/ | 200 OK | Same pattern; embeds Swagger-UI pointed at `/api/v1.json`. No prose content of its own. |
| 4 | https://docs.3vrooms.app/api/v2.json | 200 OK | **Actual v2.0 OpenAPI/Swagger 2.0 spec**, loaded by page #2. `host: vnext.book.3vrooms.app`. Treated as primary source for all v2 endpoint/field facts below. |
| 5 | https://docs.3vrooms.app/api/v1.json | 200 OK | **Actual v1.0 OpenAPI/Swagger 2.0 spec**, loaded by page #3. `host: vnext.book.3vrooms.app`. Treated as primary source for all v1 endpoint/field facts below. |
| 6 | https://docs.3vrooms.app/3vrooms/api/jobmanager-usage/ | 200 OK | JobManager async batch usage guide (v2 job endpoints), includes polling-interval guidance. |
| 7 | https://docs.3vrooms.app/betrieb/synchronisation/zold/oauth/ | 200 OK | Describes OAuth2 for **Exchange Online synchronization** (not the rooms_api client auth flow). Not relevant to A. |
| 8 | https://docs.3vrooms.app/betrieb/synchronisation/zold/oauth/oauth2_support/ | 200 OK | Describes OAuth2 **Identity Provider (Azure AD) integration** for ROOMS login, unrelated to the rooms_api client-credential/apikey flow. Not relevant to A. |
| 9 | https://docs.3vrooms.app/betrieb/tuerschilder/ | 200 OK | Door-sign (Türschild) product overview — high-level, no API/checkin-window specifics of its own. |
| 10 | https://docs.3vrooms.app/betrieb/tuerschilder/konfiguratonressource/ | 200 OK | Door-sign **resource configuration** settings (admin UI), including check-in/checkout options and polling intervals for the door-sign device itself. |
| 11 | https://docs.3vrooms.app/3vrooms/einstellungen/ressourcen/ | 200 OK | Resource settings admin-UI reference page — contains the "Checkin aktiviert" and "NoShow Delay" field descriptions (key finding for D). |
| 12 | https://docs.3vrooms.app/3vrooms/einstellungen/basisdatenfuerressourcen/workflows/ | 200 OK | Generic status-workflow admin doc; checked for checkin/grace-period terms — none found relevant. |
| 13 | Sitemap: https://docs.3vrooms.app/sitemap.xml | 200 OK | Used to enumerate the full site (274 URLs) and confirm no other `/3vrooms/api/*` sub-pages exist beyond `v1`, `v2`, `jobmanager-usage`, and the index. |

No 404s encountered. All fetched URLs returned HTTP 200.

**Important structural note:** the documentation site (`docs.3vrooms.app`) does not host prose API reference content for v1/v2 beyond an embedded Swagger UI. The actual endpoint/field truth is in the Swagger 2.0 JSON specs at `https://docs.3vrooms.app/api/v1.json` and `https://docs.3vrooms.app/api/v2.json`, which were fetched and parsed directly (sources #4 and #5). All endpoint paths below are quoted verbatim from these specs, including the literal `{mandator}` path placeholder and the `ressources` (double-s) vs `resources`/`resourceId` spelling as they actually appear.

---

## A. Token endpoint / auth flow

**NOT FOUND IN DOCS as described in the task context** (i.e., no page documents `grant_type=basic` / `client_id=basic-auth`). The only client-credential/token flow documented is a *different* grant type. Documented facts:

- Source: https://docs.3vrooms.app/3vrooms/api/ (section "Authentisierung" → "ROOMS mit IDP Authentisierung").
- Token endpoint (as literally shown in the example): `POST https://vnext.idp.3vrooms.local/connect/token` (host is an example/placeholder — `vnext.idp.3vrooms.local`, not `vnext.idp.3vrooms.app`).
- Request format: form-encoded (`Content-Type: application/x-www-form-urlencoded`), body:
  `grant_type=apikey&scope=rooms_api&client_id=xxx&apikey={{apikey}}&client_secret=xxx`
  — i.e., the documented grant type is **`apikey`**, not `basic`, and the credential is a PIN/API-key value obtained from `Person → Bearbeiten → Logons → Erstellen → PIN → Code Erstellen`, not a username/password pair.
- `scope=rooms_api` **does match** the context given (confirmed).
- `client_id` in the doc's example is a placeholder (`xxx`), not `basic-auth`.
- IDP client config requirement (from the same page), shown as literal JSON:
  ```json
  {
    "Enabled": true,
    "ClientId": "xxx",
    "ClientName": "xxx",
    "ClientSecrets": [{"Value": "xxx"}],
    "AllowedGrantTypes": ["apikey"],
    "AllowedScopes": ["rooms_api"],
    "RequireConsent": false,
    "RequirePkce": true,
    "RequireClientSecret": "false"
  }
  ```
- Response shape: the doc only shows the field `access_token` being extracted (`{{postToken.response.body.access_token}}`). **NOT FOUND IN DOCS**: no mention of `expires_in`, `refresh_token`, `token_type`, or any other response field. No example JSON response body is shown at all — only the field-access expression.
- Token lifetime: **NOT FOUND IN DOCS**. No lifetime, expiry, or refresh-token guidance is given anywhere in the fetched pages.
- Usage: the resulting token is passed as `Authorization: Bearer {{token}}` on subsequent requests (confirmed, same page, with example `GET https://vnext.book.3vrooms.local/Default/api/v1.0/bookings/1177`).
- A second, separate legacy auth mode exists: **PIN/API-key header auth** (no token exchange) — pass `APIKEY: {{apikey}}` directly as an HTTP header on every request, gated by the global parameter "Darf PIN Authentisierung nutzen" = `True`. Source: same page, section "ROOMS ohne IDP Authentisierung".
- Two related but unrelated OAuth2 flows also exist in the docs and were checked but do not describe the rooms_api client auth: (1) Exchange Online sync OAuth2 (https://docs.3vrooms.app/betrieb/synchronisation/zold/oauth/) and (2) ROOMS login via external Identity Provider/Azure AD (https://docs.3vrooms.app/betrieb/synchronisation/zold/oauth/oauth2_support/) — this second page documents generic IDP fields (`ClientId`, `Secret`, `TokenEndpoint`, `MetadataEndpoint`, `UserInformationEndpoint`) for *user login into the ROOMS UI*, not for machine/API clients, and does not mention `grant_type=basic`.
- **NOT FOUND IN DOCS**: no page or spec documents a `grant_type=basic` value or a `client_id=basic-auth` value anywhere in the fetched sources. This may be an environment-specific/internal variant not published in the public docs, or may need to be confirmed against the live IDP's discovery/metadata endpoint rather than the docs site.

---

## B. Create a new ad-hoc booking

Source: https://docs.3vrooms.app/api/v2.json (Swagger v2 spec, fetched via v2 doc page).

- **Endpoint**: `POST /{mandator}/api/v2.0/bookings/quickbooking`
- `operationId: Bookings2_QuickBooking`, tag `Bookings2`, summary: "Creates a new booking with the data of the model."
- Consumes: `application/json`, `text/json`, `text/javascript`, `application/xml`, `text/xml`, `application/x-www-form-urlencoded`.
- Request body: `QuickBookingModel` (JSON body param name `quickBookingModel`), with these literal properties (per the spec's `definitions.QuickBookingModel`):
  - `ressourceId` (integer, int32) — "The id of the desired ressource." **Note the double-s spelling**, matching the confirmed freebusy endpoint spelling.
  - `start` (string, date-time) — "Start of the booking."
  - `end` (string, date-time) — "End of the booking."
  - `title` (string) — "The title of the booking."
  - `comment` (string) — "Holds the additional comment."
  - `headCount` (integer, int32) — "The amount of people."
  - `isPrivate` (boolean) — "Flag, ig this booking is private or not." [sic, typo in source spec]
  - `equipmentList` (array of `EquipmentDto`) — "Holds the additional bookable Equipment (GliederungId)."
  - `participantList` (array of `ParticipantDto`) — "Holds the additional participants."
- The Swagger 2.0 schema for `QuickBookingModel` has **no `required` array** — the spec does not mark any of these fields as strictly required. **NOT FOUND IN DOCS**: which fields are actually mandatory at runtime (the spec is silent on this).
- Response: `200 OK` → `BookingDto` (full booking object; see section D for a description of this schema's fields).
- Auth required: not explicitly stated per-endpoint in the spec (no `security` block present on this operation in the raw JSON); general doc guidance (source: https://docs.3vrooms.app/3vrooms/api/, "Authentisierung" section) states "Die Kommunikation mit dem REST-API erfolgt immer authentisiert mit einem Benutzer" (all REST API communication is always authenticated with a user) — i.e., some form of Bearer token or `APIKEY` header per section A is required, though this specific endpoint's spec entry does not itself declare a security scheme.
- A separate, older creation endpoint also exists in **v1**: `POST /{mandator}/api/v1.0/reservation/create` (and a variant `POST /{mandator}/api/v1.0/reservation/create/{gliederungId}/{standortId}/{resourceTypeId}`), and a v1 `POST /{mandator}/api/v1.0/bookings/{ressourceId}` endpoint. These were found in the v1 spec's path list but were **not inspected in field-level detail** since v2 `quickbooking` directly matches the "ad-hoc booking with start/end/resourceId=44" use case from context. Source: https://docs.3vrooms.app/api/v1.json.

---

## C. Modify/extend an existing booking's end time

Source: https://docs.3vrooms.app/api/v2.json.

- **Endpoint**: `PUT /{mandator}/api/v2.0/bookings/{id}/update`
- `operationId: Bookings2_Update` (note: the spec's `summary` text literally reads "Finds a booking with the specified filter." — apparently a copy-paste leftover from another operation's docstring in the source spec, but the path/method/schema make clear this is the update endpoint).
- Path parameter: `id` (integer, int32) — the booking id.
- Request body: `BookingUpdateModel`, which includes (among many other fields) a literal `end` property (string, date-time, description "End of the booking.") and `start` (string, date-time, description "Start of the booking."). This confirms the update endpoint can change a booking's end time without deleting/recreating it.
- Other notable `BookingUpdateModel` fields: `applyTempChangesToSeries` (bool), `title`, `comment`, `organizerId`, `creatorId`, `participantCount`, `externalParticipantCount`, `seatingId`, `classificationIds` (array of int), `costUnits`, `bestellungItems`, `gliederungEquipment`, `customFields`, `useTempdataContext` (bool, "True if the booking is being edited."), and **`resourceId`** (integer, int32, description "The new resource of the booking.") — **note this field is spelled `resourceId` (single-s)**, inconsistent with `QuickBookingModel.ressourceId` (double-s) and `BookingDto.ressourceId` (double-s) elsewhere in the same spec. This spelling inconsistency is verbatim from the fetched JSON, not a transcription error.
- Response: `200 OK` → `BookingDto`.
- No separate "extend booking" or "PATCH end-time only" endpoint was found; `bookings/{id}/update` is a general-purpose update covering the end time along with other fields.
- Related v1 endpoints found in the same spec family (not detailed further): `PUT /{mandator}/api/v1.0/bookings/{id}` (general update), `POST /{mandator}/api/v1.0/reservation/edit/{reservationId}`. Source: https://docs.3vrooms.app/api/v1.json.

---

## D. Check-in endpoint(s) and checkin-window/policy fields

### D.1 Check-in action endpoints (v1 only — not present in v2)

Source: https://docs.3vrooms.app/api/v1.json. These operations exist **only in the v1 spec's path list**; the v2 spec has no equivalent check-in/checkout/no-show paths.

| Method | Path | operationId | Summary (verbatim) |
|---|---|---|---|
| PUT | `/{mandator}/api/v1.0/bookings/{id}/checkin` | `Bookings_Checkin` | "Checkin the booking with the specific id." |
| GET | `/{mandator}/api/v1.0/bookings/{id}/cancheckin` | `Bookings_CanCheckin` | "Returns a boolean indicating if a checkin for the given booking is possible." |
| PUT | `/{mandator}/api/v1.0/bookings/{id}/checkout` | `Bookings_Checkout` | "Terminates the booking with the specific id.\r\nE.g. PUT api/bookings/4/checkout?sendMail=false." |
| GET | `/{mandator}/api/v1.0/bookings/{id}/cancheckout` | `Bookings_CanCheckout` | "Returns a boolean indicating if a checkin for the given booking is possible." [sic — description literally duplicated from CanCheckin in the source spec] |
| PUT | `/{mandator}/api/v1.0/bookings/{id}/checkout/release` | `Bookings_CheckoutRelease` | "Releases the booking with the specific id.\r\nE.g. PUT api/bookings/4/checkout/release?sendMail=false." |
| PUT | `/{mandator}/api/v1.0/bookings/{id}/release` | `Bookings_Release` | "Releases the booking with the specific id." |
| PUT | `/{mandator}/api/v1.0/bookings/{id}/noshow/{art}` | `Bookings_NoShow` | "Marks a booking as no-show. After the bookings/{id}/noshow/ the type of the no-show can be added (Manually, Smartrooms, MissingCheckin).\r\nE.g. PUT api/bookings/4/noshow or api/bookings/4/noshow/Manually." (`art` is an integer path param with enum values `[0, 1, 2]`, unlabeled in the spec — the three named types Manually/Smartrooms/MissingCheckin presumably map to these three enum values in that order, but the spec does not explicitly bind the labels to the numbers.) |

- `checkin` accepts an optional query parameter `pin` (string, not required).
- `checkin` and `checkout` both return `200 OK` → `BookingDto` on success.
- **NOT FOUND IN DOCS**: no `security`/auth-scheme annotation specific to these operations beyond the general API-wide auth requirement described in section A.

### D.2 "Check-in required" flag and checkin-window fields exposed on booking/resource objects

Source: https://docs.3vrooms.app/api/v2.json and https://docs.3vrooms.app/api/v1.json (`definitions.BookingDto` and `definitions.RessourceDto` — **identical field sets in both v1 and v2 specs**).

`BookingDto` (returned by booking create/update/checkin/checkout/etc.) includes these literal properties:
- `checkinConfirmed` (boolean) — no further description text in the spec, but the name strongly implies "has this booking been checked in."
- `checkinLeadTime` (string, date-time) — an absolute UTC timestamp field (not a duration).
- `checkinFollowupTime` (string, date-time) — an absolute UTC timestamp field (not a duration).
- `endFollowUp` (string, date-time) — a separate absolute UTC timestamp field.

None of these three date-time fields have description text in the spec beyond their name and type — **NOT FOUND IN DOCS**: no prose explains exactly what "lead time" vs "followup time" vs "endFollowUp" mean operationally (e.g., which one is the check-in deadline vs. which one is when the room auto-releases).

`RessourceDto` (returned by resource endpoints and embedded as `BookingDto.resource`) includes:
- `checkinEnabled` (boolean) — whether check-in is enabled for this resource.
- `followUpDuration` (integer, int32) — no unit specified in the schema itself (no `description` field on this property in the spec).
- `preparationDuration` (integer, int32) — likewise no description/unit in the schema.

`MapFeatureDto` (used by the floor-plan/maps endpoints) also exposes:
- `bookingCheckedIn` (boolean) — appears to be a read-only summary flag for map/floor-plan display.
- `resourceLinkCheckinEnabled` (boolean).

There is **no dedicated endpoint to write/configure `followUpDuration` or any grace-period value** in either spec — it only appears as a read-only-looking property on `RessourceDto` responses; no request body in the spec (e.g., `QuickBookingModel`, `BookingUpdateModel`) includes a settable `followUpDuration`, `checkinLeadTime`, or `checkinFollowupTime` field. **NOT FOUND IN DOCS**: no endpoint that lets a client set/change these values via the API — they read as organization/resource-configured policy values, consistent with the admin-UI finding below.

### D.3 Admin-UI documentation of the "NoShow Delay" policy

Source: https://docs.3vrooms.app/3vrooms/einstellungen/ressourcen/ (page section covering resource settings fields "Checkin aktiviert" and "NoShow Delay").

Verbatim (German) text:
> **Checkin aktiviert** — "Diese Option schaltet die Check-in-Funktion auf dem Türschild ein. Durch die Benutzergruppe wird entschieden, ob ein Check-in für die Benutzende notwendig ist oder nicht." (This option turns on the check-in function on the door sign. The user group decides whether check-in is required for the user or not.)
>
> **NoShow Delay** — "Minuten auswählen. Definiert nach wie vielen Minuten ohne Checkin der Raum wieder freigegeben wird. Standardwert = 15 Minuten" (Select minutes. Defines after how many minutes without check-in the room is released again. Default value = 15 minutes.)

This is the closest documented description of an "auto-cancellation-if-no-checkin" policy and its duration (default 15 minutes), but:
- It is documented **only as an admin-configuration-UI setting** (`Einstellungen → Ressourcen`), not as an API field name or endpoint. The doc text does not state that "NoShow Delay" is the same value as the Swagger spec's `RessourceDto.followUpDuration` — this mapping is a plausible inference from the field names/semantics but is **not explicitly confirmed anywhere in the fetched docs**, so it should be treated as unconfirmed.
- Whether an API client can read this per-resource NoShow-Delay value at all (e.g., is it `followUpDuration`?) is **NOT FOUND IN DOCS** as an explicit statement — only inferable by proximity of concept, not by an explicit cross-reference in either the prose docs or the Swagger descriptions.
- Whether the value is discoverable/configurable via the API, versus being purely an org-level policy set only through the ROOMS admin UI, is therefore **NOT FOUND IN DOCS** as a definitive answer. The Swagger spec shows `followUpDuration` as a readable integer property on `RessourceDto` (returned when fetching resource data), which — if it is indeed the same setting — would make it discoverable (read-only) via the API, but no fetched source explicitly states this equivalence.

### D.4 Door-sign (Türschild) specific check-in configuration

Source: https://docs.3vrooms.app/betrieb/tuerschilder/konfiguratonressource/ ("Konfigurationsparameter" table).

Relevant literal settings (admin UI, not API):
- **Unauthorisierte Aktionen erlauben**: "Aktionen werden im Namen des konfigurierten Users ausgeführt. Ermöglicht anonymes Check-in/Check-out sowie anonymes Ad-hoc-Buchen" (Actions are performed in the name of the configured user. Enables anonymous check-in/check-out and anonymous ad-hoc booking.)
- **Nur Checkin**: "Nur Check-in Funktion erlauben" (Only allow the check-in function.)
- **Checkout Modus**: two options — *Terminieren* ("Buchung wird auf jetzt gekürzt, Nachlaufzeit bleibt" — booking is shortened to now, follow-up time remains) or *Freigeben* ("Buchung wird auf jetzt gekürzt, Raum ist sofort frei" — booking is shortened to now, room is immediately free).
- **Anfrage Intervall (in Minuten)**: "Intervall, bei welchem das Türschild neue Buchungsdaten abruft (min. 1, max. 1440, Standard 5)" — this is the door-sign device's own booking-data polling interval (default 5 min, min 1, max 1440), a device-configuration setting, not an API-documented rate limit (see section F).

No checkin lead-time/grace-window value is stated on this page beyond what's covered in D.3.

---

## E. FreeBusyData bitstring semantics and `interval` parameter

Source: https://docs.3vrooms.app/api/v2.json, endpoint `GET /{mandator}/api/v2.0/ressources/{resourceId}/freebusy` (and the equivalent multi-resource variant `GET /{mandator}/api/v2.0/ressources/freebusy`).

Verbatim `summary` field from the spec (identical wording on both endpoints):
> "Retrieves the free/busy data for the specified resource(s) within a given time range, and returns the availability information as a binary string, where '1' indicates a busy interval and '0' indicates a free interval. The method calculates the total number of time intervals within the given time range based on the specified interval."

This **confirms**: `1` = busy, `0` = free (matching the confirmed unauthenticated example `"FreeBusyData": "000000000000"` for an all-free window).

`interval` query parameter (from the same spec entry):
- `name: interval`, `in: query`, `required: false`, `type: integer`, `format: int32`
- Description: "The length of each time interval in minutes (default is 15 minutes)."

This **confirms** the `interval` parameter controls the bucket size in minutes, and additionally establishes the **default value is 15 minutes** when `interval` is omitted (the confirmed example in the task context passed `interval=5` explicitly).

Response schema `FreeBusyDto` (from `definitions.FreeBusyDto`):
```json
{
  "type": "object",
  "properties": {
    "freeBusyData": { "type": "string" },
    "resourceId": { "type": "integer", "format": "int32" }
  }
}
```
Note: the schema property is camelCase `freeBusyData` (matching the `X-JsonPropertyFormat: camel` header behavior described in section A); the confirmed example response used PascalCase `FreeBusyData`/`ResourceId`, which is the default (no `X-JsonPropertyFormat` header sent). Source for the format-switching header: https://docs.3vrooms.app/3vrooms/api/ ("Anpassen der JSON-Response-Formatierung" section).

No further breakdown of what a single character encodes beyond busy/free per interval-bucket is given (e.g., no distinction between "busy from this room's own bookings" vs. "busy from a blocker" is documented in the spec text) — **NOT FOUND IN DOCS** beyond the single busy/free bit per bucket.

---

## F. Rate limits / throttling / polling guidance

- **No rate-limit or throttling documentation of any kind was found** for the `freebusy` or `bookings` REST endpoints in either the v1 or v2 Swagger specs (https://docs.3vrooms.app/api/v1.json, https://docs.3vrooms.app/api/v2.json) — searched for "rate limit," "throttle," "quota," "X-RateLimit," "per minute/second" with zero matches in both spec files. **NOT FOUND IN DOCS.**
- The only documented polling guidance found anywhere in the fetched docs applies to the **JobManager batch API**, not freebusy/bookings directly. Source: https://docs.3vrooms.app/3vrooms/api/jobmanager-usage/:
  > "Start with 2–5 Seconds Polling-Intervall. Danach Backoff (z.B. auf 10s/20s/30s erhöhen), um unnötige Last zu vermeiden." (Start with a 2–5 second polling interval. Afterward, back off — e.g., increase to 10s/20s/30s — to avoid unnecessary load.)
  This applies to polling `GET /{mandator}/api/v2.0/jobs/batches/{batchId}` for async job status, not to `freebusy` or synchronous booking calls.
- A separate, tangential polling-interval data point exists for the **physical door-sign device's own configuration**, not the API itself. Source: https://docs.3vrooms.app/betrieb/tuerschilder/konfiguratonressource/ — "Anfrage Intervall (in Minuten)": default 5 minutes, min 1, max 1440 (i.e., 3volutions' own door-sign product defaults to polling booking data every 5 minutes, with a configurable range of 1–1440 minutes). This is guidance for 3volutions' own device, not a documented rate limit imposed by the server, and is not stated to apply generally to third-party API clients like PuKK.
- **NOT FOUND IN DOCS**: no HTTP response headers (e.g., `Retry-After`, `X-RateLimit-*`), no documented per-minute/per-hour call caps, and no explicit guidance for how frequently a third-party integrator should poll `freebusy` or `bookings/find`.

---

## Open questions / not found — summary

- **A**: The exact `grant_type=basic` / `client_id=basic-auth` / username+password flow described in the task context is **not documented anywhere in the fetched docs**. The only documented client-credential flow uses `grant_type=apikey` with a PIN, against `POST https://vnext.idp.3vrooms.local/connect/token` (example/placeholder host). Token response shape beyond `access_token`, and token lifetime/expiry/refresh-token support, are **NOT FOUND IN DOCS**.
- **B**: Whether any `QuickBookingModel` fields are strictly required at request time is **NOT FOUND IN DOCS** (Swagger schema has no `required` array for this model). Explicit per-endpoint auth/security scheme annotation is also **NOT FOUND IN DOCS** at the operation level (only general prose guidance exists).
- **C**: Fully answered — `PUT /{mandator}/api/v2.0/bookings/{id}/update` with `BookingUpdateModel.end`. Note the `resourceId` (single-s) vs. `ressourceId` (double-s) spelling inconsistency between this model and others in the same v2 spec.
- **D**: Check-in endpoints exist only in v1, not v2. Booking-level `checkinConfirmed`/`checkinLeadTime`/`checkinFollowupTime`/`endFollowUp` fields and resource-level `checkinEnabled`/`followUpDuration` fields exist in both specs but are undocumented beyond field name/type (no prose description of exact semantics in the Swagger spec itself). The admin-UI "NoShow Delay" (default 15 minutes) is the closest documented description of an auto-release-after-no-checkin policy, but its equivalence to the API's `followUpDuration` field is **not explicitly confirmed** in any fetched source — this mapping remains unconfirmed/open. Whether this window is discoverable via the API at all (vs. purely an admin-configured, API-invisible policy) is therefore **NOT FOUND IN DOCS** as a definitive statement.
- **E**: Fully answered — `1` = busy, `0` = free; `interval` = bucket size in minutes, default 15.
- **F**: No documented rate limits for freebusy/bookings endpoints. Only the unrelated JobManager batch-polling guidance (2–5s initial, backoff to 10s/20s/30s) and the door-sign device's own default polling interval (5 min) were found — neither is a documented rate limit for third-party API clients.
