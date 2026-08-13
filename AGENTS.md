# Agent Instructions — <role>

You are <role description>.

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

<describe your role and what you own — e.g. "You own the backend. Never touch frontend/.">

## Tech Stack

<fill in>

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
git worktree add /tmp/<projectname>-<slug> -b feat/<role>-<slug> origin/main

# Clean up after PR is merged
git worktree remove /tmp/<projectname>-<slug>
git branch -d feat/<role>-<slug>
```

- Never commit or stage files in the main worktree (`/workspaces/<projectname>`).
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
5. Check out or create the feature branch: `git worktree add /tmp/<projectname>-<slug> -b feat/<slice>`

### API contract change proposal

If you need a new or changed endpoint:
1. Create `docs/api-proposals/<date>-<topic>.md` describing the change
2. Open a PR with only that file
3. Wait for the other agent to agree and the contract to be updated before implementing

### API mocking

The frontend mocks the backend using Prism against the agreed contract:
```bash
npx @stoplight/prism-cli mock api-contract/openapi.yaml --port 4010
```

### ADR process

- New ADRs are opened as PRs with status `Proposed`.
- Status updates to `Accepted` only on merge.
