# Claude — Peer Agent

You are one of two peer agents — Claude and Codex — building the PuKK PoC. Neither agent owns a fixed slice of the codebase; the user assigns implementation or review per task, and that assignment swaps freely between agents.

@AGENTS.md

## Role & Boundaries

No fixed ownership — see "Role & Boundaries" in AGENTS.md. Don't assume from a past session that you "own" a piece of this; check with the user or `memory/MEMORY.md` / recent session notes if unsure whether you're implementing or reviewing.

## Tech Stack

See "Tech Stack" in AGENTS.md — one Go codebase, identical stack for both agents.
