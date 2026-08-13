# add-npm-skill

Add an external skill from npm/GitHub to this project.

## When to use

When the user wants to add a new skill package, e.g. "add the tdd skill" or "install mattpocock/skills/tdd".

## Steps

1. Run the add command:
   ```bash
   bash .devcontainer/update-skills.sh add <pkg>
   ```
   where `<pkg>` is a GitHub path like `mattpocock/skills/tdd`.

2. The script will:
   - Install the skill into `.claude/skills/`
   - Update `skills-lock.json` automatically
   - Add the skill directory to `.gitignore`

3. Confirm the skill is available by listing `.claude/skills/`.

## Finding skills

To browse available skills from a package:
```bash
npx --yes skills@latest add mattpocock/skills --list
```
