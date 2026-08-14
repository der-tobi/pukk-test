# 3V Rooms Source Mount

The devcontainer mounts local 3V Rooms-related source/docs for implementation-time API research.

## Paths

- `/workspaces/rooms-q-and-a/basis` — mounted from `${localEnv:HOME}/projects/microfast`, read-only. Use this as the primary local source-code reference for 3V Rooms API behavior, request/response models, endpoint implementation details, auth handling, and edge cases that are unclear from public Swagger docs.

The previously mentioned `docs-template/examples/zkb-rfp` subtree is not relevant to this PuKK PoC and should be ignored if present.

## How to apply

After the container restart, inspect the Rooms source mount before implementing uncertain 3V Rooms behavior, especially:

- token endpoint/auth parameters for the user's `grant_type=basic`, `client_id=basic-auth`, `scope=rooms_api`, `user=tobiapi4` flow
- v1 booking update/extend request shape
- booking lookup endpoints and filters for finding the active booking by resource/time
- quickbooking runtime-required fields
- checkin/checkout/release semantics and `pin` handling

The public research in `docs/research/3v-rooms-api.md` remains useful, but local source wins where it reveals actual server behavior.
