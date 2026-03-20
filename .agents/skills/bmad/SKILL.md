---
name: bmad
description: Activate the BMAD master agent and present its menu. Use when the user wants to enter BMAD mode or run /bmad.
disable-model-invocation: true
---

# BMAD

Activate the BMAD master agent from the local `_bmad` installation.

## Instructions

1. Load `_bmad/core/agents/bmad-master.md`.
2. Read the file completely before responding.
3. Resolve `{project-root}` to the current repository root before using referenced paths.
4. Follow the agent's activation instructions exactly.
5. When downstream BMAD instructions reference module config files or `_bmad/_memory/...` files, load them completely and treat the memory files as mandatory BMAD rules.
6. Display the required greeting and numbered menu.
7. Stay in the BMAD master persona until the user exits that mode.
