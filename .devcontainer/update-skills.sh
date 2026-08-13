#!/usr/bin/env bash
# update-skills.sh — manage external skills in .claude/skills/
#
# Source of truth: skills-lock.json (auto-maintained by the skills CLI)
#
# Usage:
#   update-skills.sh              # restore missing skills + weekly refresh
#   update-skills.sh add <pkg>    # e.g. mattpocock/skills/tdd
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOCK_FILE="$REPO_ROOT/skills-lock.json"
SKILLS_DIR="$REPO_ROOT/.claude/skills"
GITIGNORE="$REPO_ROOT/.gitignore"
STAMP_FILE="$SKILLS_DIR/.skills-last-updated"
WEEK_SECONDS=$((7 * 24 * 60 * 60))

mkdir -p "$SKILLS_DIR"

# Read skills-lock.json and print one "source/name" per line
lock_packages() {
  node -e "
    const l = require('$LOCK_FILE');
    Object.entries(l.skills).forEach(([name, s]) => console.log(s.source + '/' + name));
  "
}

is_stale() {
  [ ! -f "$STAMP_FILE" ] && return 0
  local last now
  last=$(cat "$STAMP_FILE")
  now=$(date +%s)
  [ $((now - last)) -gt $WEEK_SECONDS ]
}

install_skill() {
  npx --yes skills@latest add "$1" --agent claude-code --yes < /dev/null
}

# Ensure .gitignore has an entry for every skill in skills-lock.json
sync_gitignore() {
  node -e "const l=require('$LOCK_FILE'); Object.keys(l.skills).forEach(n=>console.log(n));" \
  | while IFS= read -r name; do
    local entry=".claude/skills/${name}/"
    grep -qxF "$entry" "$GITIGNORE" 2>/dev/null || echo "$entry" >> "$GITIGNORE"
  done
}

cmd_add() {
  local pkg="$1"
  echo "Installing: ${pkg##*/}"
  install_skill "$pkg"
  sync_gitignore
}

cmd_update() {
  if [ ! -f "$LOCK_FILE" ]; then
    echo "No skills-lock.json found — nothing to install."
    return
  fi

  # Install any skills missing from disk
  while IFS= read -r pkg; do
    local name="${pkg##*/}"
    if [ ! -d "$SKILLS_DIR/$name" ]; then
      echo "Installing missing skill: $name"
      install_skill "$pkg"
    fi
  done < <(lock_packages)

  # Full refresh once a week
  if is_stale; then
    echo "Refreshing all skills..."
    while IFS= read -r pkg; do
      install_skill "$pkg"
    done < <(lock_packages)
    date +%s > "$STAMP_FILE"
    echo "Skills up to date."
  else
    echo "Skills up to date (next refresh in less than 7 days)."
  fi
}

case "${1:-}" in
  add)
    [ -z "${2:-}" ] && { echo "Usage: $0 add <pkg>  e.g. mattpocock/skills/tdd" >&2; exit 1; }
    cmd_add "$2"
    ;;
  *)
    cmd_update
    ;;
esac
