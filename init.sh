#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# init.sh — initialise a new project from this template
# Run once after cloning: ./init.sh
# ---------------------------------------------------------------------------

normalize() {
  echo "$1" | tr '[:upper:]' '[:lower:]' | tr '_' '-' | tr ' ' '-'
}

# Auto-detect project name
detect_name() {
  # 1. folder name
  local folder
  folder=$(normalize "$(basename "$PWD")")
  if [[ "$folder" != "vibe-template" && "$folder" =~ ^[a-z0-9-]+$ ]]; then
    echo "$folder"; return
  fi

  # 2. git remote name
  if git remote get-url origin &>/dev/null; then
    local repo
    repo=$(normalize "$(git remote get-url origin | sed 's/.*\///' | sed 's/\.git$//')")
    if [[ "$repo" != "vibe-template" && "$repo" =~ ^[a-z0-9-]+$ ]]; then
      echo "$repo"; return
    fi
  fi

  # 3. ask
  echo ""
}

PROJECT="${1:-}"

if [[ -z "$PROJECT" ]]; then
  PROJECT=$(detect_name)
fi

if [[ -z "$PROJECT" ]]; then
  read -rp "Project name (lowercase, hyphens): " PROJECT
  PROJECT=$(normalize "$PROJECT")
fi

if [[ ! "$PROJECT" =~ ^[a-z0-9-]+$ ]]; then
  echo "Error: invalid project name '$PROJECT' — use lowercase letters, numbers, hyphens only."
  exit 1
fi

echo "Initialising project: $PROJECT"

FILES=(
  "project.md"
  "api-contract/openapi.yaml"
  "docs/collaboration/project-summary.md"
  "docs/collaboration/workflow.md"
  "AGENTS.md"
  "CLAUDE.md"
  ".devcontainer/devcontainer.json"
  ".devcontainer/codex-config.toml"
)

for f in "${FILES[@]}"; do
  if [[ -f "$f" ]]; then
    sed -i "s/<projectname>/$PROJECT/g" "$f"
    echo "  updated $f"
  fi
done

if [[ ! -f ".claude/settings.local.json" && -f ".claude/settings.local.json.template" ]]; then
  cp ".claude/settings.local.json.template" ".claude/settings.local.json"
  echo "  created .claude/settings.local.json from template"
fi

echo ""
echo "Done. Still to fill in manually:"
echo "  CLAUDE.md  — <role>, <role description>, Tech Stack"
echo "  AGENTS.md  — <role>, Tech Stack"
echo "  project.md — vision, scope, domain model"
