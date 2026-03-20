# 各步骤输出格式参考

本文件定义了 epic-decompose 工作流各步骤的标准输出格式。

---

## Step 0: Preflight Check

```
Epic file: .claude/epics/<feature>/epic.md ✓
Existing tasks: 15
Orphaned tasks: 2 (will auto-deprecate)
Continue auto-execution...
```

---

## Step 2: Compare with Existing Tasks

```
Task 001: KEEP (similarity: 0.92, matched by number)
Task 002: UPDATE (similarity: 0.78, matched by title, missing 2 specs)
Task 003: CREATE (no match found, validated no duplicates)
Task 004: UPDATE (similarity: 0.65, matched by content, force update due to existence)
Task 099: ORPHAN (auto-deprecate)
```

---

## Step 3: Detect Conflicts

```
Task 004: conflicts with Task 002 (similarity: 0.82)
  title: 0.90, desc: 0.78, deps: 0.75
  Added to conflicts_with field
```

---

## Step 5: Validate & Report

### GitHub Mapping File

```markdown
# GitHub Issue Mapping for Epic: <feature_name>

## File-to-Issue Mapping
001-user-list.md → #123 ✅
002-user-detail.md → #456 ✅
003-user-search.md → pending ⏳

## Synced
- Last Updated: 2025-11-13T16:00:00Z
- Command: /epic-decompose <feature_name>
```

### Summary Report

```
Epic decompose complete: <feature_name>

Operation Summary:
- KEEP: 10 (content match)
- UPDATE: 3 (smart merge)
- CREATE: 2 (new in Epic)
- ORPHAN: 1 (auto-deprecate)
- Total: 15 tasks (12 open, 3 completed)

Technical Coverage:
- Frontend components: 8
- Backend APIs: 5
- UI/UX specs: complete
- Dependencies: mapped
- Task metadata: complete (File, Purpose, Leverage, Requirements, Prompt)

Data Integrity:
- Atomic writes: protected
- No data corruption
- Lock released

Next Steps:
1. Review auto-deprecated tasks
2. Run: /epic-sync <feature_name>
3. Commit: git add . && git commit -m 'docs: update tasks'
```

---

## Step 6: GitHub Sync Check

```
Epic decompose complete, GitHub sync required separately

Next Steps:
- Run: /epic-sync <feature_name>
- This creates/updates GitHub Issues
- Updates github-mapping.md file
```
