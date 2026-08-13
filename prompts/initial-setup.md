# Initial Setup

- Familiarise yourself with this repository
- Read `project.md` — this is the project idea
- Interview me about any placeholders and unclear points
- Review the tech stack, Dockerfile, and devcontainer — challenge anything that looks off
- Update all necessary files so we can start working
- Don't guess — ask if something is unclear

## Session Goals

1. `/grill-me` — stress-test the project idea in `project.md`
2. `/write-a-prd` — turn it into a PRD
3. `/prd-to-plan` — break into phases
4. `/prd-to-issues` — create GitHub issues
5. `/design-an-interface` — design the interface

## Memory — persist everything in the repo

Write durable memory into `memory/` so any agent can pick up where you left off:

- **Shared knowledge** (project decisions, user preferences, feedback): create topic files like `memory/project_context.md`, `memory/user_role.md`, `memory/feedback_style.md` and add them to `memory/MEMORY.md`.
- **Session notes**: update your own `memory/claude/session.md` (or `memory/codex/session.md`) at the end of every session — current focus, decisions made, open questions, next steps.

Do not rely on external agent memory. Everything needed to resume must be in this repo.

---

## Multi-agent mode

Add this if splitting work between Claude and Codex:

- Agree on the API contract before any implementation
- Each agent works in its own worktree (see `CLAUDE.md` / `AGENTS.md`)
- Before every task: check for open PRs from the other agent, review contract changes, then proceed
- Frontend mocks the backend via Prism: `npx @stoplight/prism-cli mock api-contract/openapi.yaml --port 4010`
- Write introductions for all participants so everyone has shared context
- Maintain a shared todo list and session notes in the repo
