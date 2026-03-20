---
name: bmad-bmm-validate-prd
description: Validate a PRD against standards. Use when the user says "validate this PRD" or "run PRD validation"
---

# bmad-bmm-validate-prd

Run the BMAD command defined in `.claude/commands/bmad-bmm-validate-prd.md`, but do not stop at the command shim.

1. Load `.claude/commands/bmad-bmm-validate-prd.md`.
2. Read it completely, including YAML frontmatter and body.
3. Resolve `{project-root}` to the current repository root before using referenced paths.
4. Continue following every referenced BMAD file transitively until the concrete `_bmad` agent, workflow, or task is fully loaded.
5. Execute that concrete BMAD resource exactly as written.
6. When the BMAD instructions reference module config files such as `_bmad/core/config.yaml` or `_bmad/*/config.yaml`, load them completely before producing output.
7. When the BMAD instructions reference `_bmad/_memory/...` files, load them completely and treat them as mandatory BMAD memory rules, not optional background context.
8. Do not stop after reading the command shim if it points to deeper BMAD instructions.
