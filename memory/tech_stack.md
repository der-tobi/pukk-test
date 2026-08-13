---
name: tech-stack
description: Tech stack decision for the pukk-test PoC server — Go, chosen for single-binary Windows cross-compilation
metadata:
  type: project
---

The pukk-test PoC server is written in **Go**, built inside the devcontainer and cross-compiled to a standalone Windows executable (`GOOS=windows GOARCH=amd64 go build`).

**Why:** [[project-context]] requires "a single executable that you can build inside the devcontainer and I can then run it on my windows machine from the commandline" with no config beyond the 3V Rooms password. Go was chosen over the alternatives considered:

- **Deno** — also produces a single cross-platform executable (`deno compile`), but wasn't already installed in the devcontainer, and `deno compile` needs to download the target runtime at build time.
- **Node.js** — already installed in the devcontainer (v20), but single-executable packaging on Windows (Node SEA, or the unmaintained `pkg`) is more fragile/experimental than Go's or Deno's story.
- **Rust** — smallest/fastest binary, but cross-compiling to Windows from Linux needs a mingw-w64 toolchain, adding setup complexity that doesn't fit a "quick and dirty" POC.

Go's stdlib (`net/http`, goroutines) covers both the PuKK-facing device API and the outbound 3V Rooms API client without a framework.

**How to apply:** Any implementation work (Claude or Codex) targets Go. The devcontainer's `.devcontainer/Dockerfile` installs the Go toolchain via the official tarball (pinned version, see `ARG GO_VERSION`); `.devcontainer/devcontainer.json` enables the `golang.go` VS Code extension. Verified in-session that `GOOS=windows GOARCH=amd64 go build` from this devcontainer produces a valid Windows PE binary.
