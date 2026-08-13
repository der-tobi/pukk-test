# vibe-template

Vibecoding starter template with Claude Code and Codex, including a curated set of agent skills.

## Skills

Skills live in `.claude/skills/` and are available to all agents (see `AGENTS.md`).

External skills are managed via the [skills CLI](https://github.com/mattpocock/skills) and tracked in `skills-lock.json`. They are **not** committed to git — only the lock file is.

### Auto-update

Skills are automatically restored and refreshed weekly on container start (`postStartCommand` in `.devcontainer/devcontainer.json`).

### Add a new skill

```bash
bash .devcontainer/update-skills.sh add mattpocock/skills/tdd
```

This installs the skill, updates `skills-lock.json`, and adds the directory to `.gitignore`.

You can also invoke `/add-npm-skill` from within any agent session.

### Browse available skills

```bash
npx skills@latest add mattpocock/skills --list
```

### Manual refresh

```bash
bash .devcontainer/update-skills.sh
```
