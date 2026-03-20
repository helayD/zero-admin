# Issue Commit 最佳实践与常见问题

---

## Best Practices

1. **频繁提交** — 小而聚焦的提交更易于 review
2. **一个 issue 一次提交** — 不要混合不同 issue 的变更
3. **自动执行优先** — 标准工作流零人工交互
4. **标准格式** — 统一的 commit message 格式，包含 issue 引用
5. **更新任务状态** — 保持本地 task file 与工作同步
6. **输入验证** — 始终检查 issue 号和文件路径的安全性
7. **并行预检** — Quick Check 步骤可并发执行提升速度
8. **备份保护** — 关键文件操作包含回滚能力
9. **自动推送** — 保持远程仓库同步，作为自动备份

---

## Common Issues

### Issue Not Found
```
✗ Cannot access issue #$ARGUMENTS
Solution: Check issue number or run: gh auth login
```

### No Changes to Commit
```
✗ No changes to commit. Working tree is clean.
Solution: Make changes first, or check: git status
```

### Commit on Main Branch
```
✗ Cannot commit directly to main/master branch
Solution: Create feature branch: git checkout -b feature/$ARGUMENTS
```

### Push Rejected
```
✗ Push failed - remote has changes
Solution: Pull first: git pull --rebase origin {branch}
Then: git push origin {branch}
```

---

## Auto-Discovery Examples

### From Branch Name
```
feature/123-add-auth      → Issue #123
issue-456                 → Issue #456
bugfix/789-fix-login      → Issue #789
fix/42-typo               → Issue #42
```

### From Commit Messages
```
"Issue #123: Add authentication"  → Issue #123
"Closes #456"                     → Issue #456
"Fix #789"                        → Issue #789
```

---

## Output Summary Format

```
✓ Commit completed and pushed to remote

Issue: #$ARGUMENTS - {issue_title}
Branch: {current_branch}
Commit: {commit_hash}
Remote: origin/{current_branch}

Files committed:
  Source: {count}
  Tests: {count}
  Config: {count}
  Docs: {count}

Next steps:
- Continue work: Make more changes and run this skill again
- Complete task: When ready, run issue-close skill
- View changes: git log --oneline -5
```

---

## Important Notes

- **Zero interaction**: Fully automated workflow from staging to push
- Always commit with issue reference for traceability
- Never commit directly to main/master
- Auto-push keeps remote repository in sync
- Update task status to reflect progress
- **Security first**: Input validation prevents injection attacks
- **Error recovery**: All critical operations include backup/rollback
- **Performance**: Use parallel Quick Check execution when possible
- **File safety**: Always validate file paths and permissions before writing
