#!/usr/bin/env bash
set -euo pipefail

if [ "$(basename "$(pwd)")" != "vibe-template" ] && grep -rl '<projectname>' project.md CLAUDE.md AGENTS.md .devcontainer/ api-contract/ docs/ 2>/dev/null | grep -q .; then
  bash "$(dirname "$0")/../init.sh"
fi                                                                                   

echo "Checking base tools..."

command -v python >/dev/null && echo "python ok"
command -v node >/dev/null && echo "node ok"
command -v npm >/dev/null && echo "npm ok"
command -v codex >/dev/null && echo "codex ok" || echo "codex missing"
command -v claude >/dev/null && echo "claude ok" || echo "claude missing"

mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/.claude"
touch "$HOME/.claude.json"
chmod 700 "$HOME/.claude" || true
chmod 600 "$HOME/.claude.json" || true
if command -v sudo >/dev/null 2>&1; then
  sudo chown -R "$(id -un)":"$(id -gn)" "$HOME/.claude" "$HOME/.claude.json" 2>/dev/null || true
fi

echo
echo "Tool check:"
command -v claude >/dev/null && echo "claude ok ($(command -v claude))" || echo "claude missing"
command -v codex >/dev/null && echo "codex ok ($(command -v codex))" || echo "codex missing"