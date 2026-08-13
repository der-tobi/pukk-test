# Agent Instructions — PuKK PoC (Claude & Codex, peers)

You are one of two peer agents — Claude and Codex — building a proof-of-concept server for the proDVX PuKK device (see `project.md`). There is no frontend/backend split here; it's a single Go binary. Neither agent has a fixed ownership boundary — the user assigns implementation or review per task, and swaps that assignment freely between the two of you.

## Mode — decide at session start

**Default: Solo** — lean, fast, no ceremony.
**Opt-in: Multi-agent** — user will say "multi-agent mode" to activate the full workflow below.

---

## Always

- Read `project.md` to orient before starting.
- Read `memory/MEMORY.md` and all files it references — this is the shared project memory.
- At the end of every session update `memory/<your-agent-name>/session.md` with current focus, decisions, open questions, and next steps. Create the file if it doesn't exist yet.
- Write new shared knowledge (project decisions, user preferences, feedback) as topic files in `memory/` and add them to `memory/MEMORY.md`. Do not rely on external memory.
- Read `docs/adr/` and respect accepted ADRs.
- Prefer small, reviewable edits.
- Do not invent API fields not listed in `api-contract/openapi.yaml`.
- If something is unclear, ask or state assumptions explicitly.
- Write tests **before** implementation (red → green → refactor) — mandatory for all backend code and any frontend business logic. Skip for pure UI rendering, layout, and styling.
- Skills live in `.claude/skills/` — browse the folders, read the `SKILL.md` of anything relevant, and follow its instructions. Claude Code users can also invoke them as `/skill-name`.

## Role & Boundaries

- No fixed ownership split. The codebase is a single Go module; either agent may touch any file.
- Per-task assignment: the user tells you whether you're implementing or reviewing a given piece of work in a given session. Don't assume you own a slice from a previous session — if unsure, ask or check `memory/MEMORY.md` and recent session notes.
- In multi-agent mode, worktrees (below) exist purely to avoid merge conflicts when both agents work concurrently — they aren't ownership boundaries.

## Tech Stack

- **Language: Go.** Chosen because the deliverable is a single dependency-free executable, built inside the devcontainer, that the user runs standalone from the Windows command line — `GOOS=windows GOARCH=amd64 go build` produces that directly with no bundler, installer, or runtime on the target machine. See `memory/tech_stack.md` for the full comparison against Deno/Node/Rust.
- **HTTP:** stdlib `net/http` for both the PuKK-facing device API and the outbound 3V Rooms API client — no framework needed at this size.
- **Concurrency:** goroutines for the background 3V Rooms refresh/cache loop, decoupled from the PuKK device's own poll cadence (see `project.md`).
- **No config files, no database.** Per `project.md`, everything is hardcoded except the 3V Rooms API password, which is prompted for at startup.
- Go toolchain is installed in `.devcontainer/Dockerfile`; VS Code's `golang.go` extension is enabled in `.devcontainer/devcontainer.json`.

## Git hygiene

- Never commit secrets.
- Keep changes scoped to the requested task.

---

## Multi-agent mode

Activate when splitting work between two agents, or using one for review and the other for implementation.

### Worktrees — mandatory in multi-agent mode

Every agent works in its own git worktree to avoid conflicts:

```bash
git fetch origin
git worktree add /tmp/pukk-test-<slug> -b feat/<role>-<slug> origin/main

# Clean up after PR is merged
git worktree remove /tmp/pukk-test-<slug>
git branch -d feat/<role>-<slug>
```

- Never commit or stage files in the main worktree (`/workspaces/pukk-test`).
- Main worktree stays on `main` and is read-only.
- Use `/tmp/` for worktrees — `/workspaces/` is root-owned in devcontainer.

### Branch strategy (vertical slices)

```
main
└── feat/<slice-name>
```

- **Contract first:** API contract for the slice must be agreed and committed before any code.
- **One PR per slice** — both sides merged together when both are green.

### Before every task

1. `gh pr list` — check for open PRs from the other agent. Review and discuss first.
2. `git diff main -- api-contract/` — check for contract changes since last session.
3. Surface ADRs relevant to the task. Flag conflicts before writing code.
4. Discuss findings before writing any code.
5. Check out or create the feature branch: `git worktree add /tmp/pukk-test-<slug> -b feat/<slice>`

### API contract

There's no separate frontend service here, so the template's default "frontend mocks backend via Prism" flow doesn't apply as-is. `api-contract/openapi.yaml` is repurposed: it documents the PuKK-facing HTTP endpoints our server exposes (per the ProDVX PuKK REST API spec), so both agents implement the same request/response shapes regardless of who wrote which handler. Still contract-first — agree the endpoint/schema in the file before implementing it.

If you need a new or changed endpoint:
1. Create `docs/api-proposals/<date>-<topic>.md` describing the change
2. Open a PR with only that file
3. Wait for the other agent to agree and the contract to be updated before implementing

Once real endpoints are defined, Prism can still mock them for local testing without a physical PuKK device:
```bash
npx @stoplight/prism-cli mock api-contract/openapi.yaml --port 4010
```

### ADR process

- New ADRs are opened as PRs with status `Proposed`.
- Status updates to `Accepted` only on merge.
