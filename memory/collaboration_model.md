---
name: collaboration-model
description: Claude and Codex are peers on pukk-test with no fixed ownership split — the user assigns implement/review per task
metadata:
  type: feedback
---

Claude and Codex work on pukk-test as **equal-rights peers**, not as an owned frontend/backend split. The user assigns "implement this" or "review this" per task, to either agent, and swaps that assignment freely between sessions.

**Why:** The user stated this directly when setting up the project: "you and [the other agent] work on this project with equal rights. I will ask to implement one of you and maybe review the other and vice versa." This also fits the project shape — pukk-test is one small Go binary with no natural frontend/backend seam, so the template's default "own frontend, never touch backend" boundary doesn't apply.

**How to apply:** Don't assume ownership of a file or feature from a previous session just because you (this agent) wrote it. Check `memory/MEMORY.md` and the relevant `memory/<agent>/session.md` files, or ask the user, before assuming whether you're implementing or reviewing in a given session. Worktrees in multi-agent mode ([[tech-stack]] context: single Go module) are for avoiding merge conflicts, not for enforcing a boundary.
