# Issue Commit 最佳实践与常见问题

---

## Best Practices

1. **频繁提交** — 小而聚焦的提交更易于 review
2. **一个 story 一次提交** — 不要混合不同 story 的变更
3. **自动执行优先** — 标准工作流零人工交互
4. **标准格式** — 统一的 commit message 格式，包含 story 引用
5. **更新任务状态** — 保持本地 story file 与工作同步
6. **输入验证** — 始终检查 story ID 和文件路径的安全性
7. **并行预检** — Quick Check 步骤可并发执行提升速度
8. **备份保护** — 关键文件操作包含回滚能力
9. **自动推送** — 保持远程仓库同步，作为自动备份

---

## Common Issues

### Story Not Found
```
✗ Cannot find story file for ID: 1-4
Solution: 确保 story 文件在 _opcos/implementation-artifacts/ 或 _opcos/planning-artifacts/ 目录
```

### Invalid Story ID Format
```
✗ Story ID 格式无效: 123（必须为 X-Y 格式，如 1-4）
Solution: 使用正确格式的 story ID
```

### No Changes to Commit
```
✗ No changes to commit. Working tree is clean.
Solution: Make changes first, or check: git status
```

### Commit on Main Branch
```
✗ Cannot commit directly to main/master branch
Solution: Create feature branch: git checkout -b story/1-4
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
feature/1-4-license-management  → Story 1-4
story/2-1-auth                 → Story 2-1
bugfix/1-3-factory-account     → Story 1-3
```

### From Commit Messages
```
"Story 1-4: Add license APIs"  → Story 1-4
"[1-3] Factory status fix"    → Story 1-3
"story 2-1 RBAC refactor"       → Story 2-1
```

---

## Output Summary Format

```
✓ Commit completed and pushed to remote

Story: 1-4 - 授权点额度包经营
Branch: ai_flutter_client
Commit: 1702e8fac
Remote: origin/ai_flutter_client

Files committed:
  Source: 12
  Tests: 7
  Config: 5
  Docs: 7

Next steps:
- Continue work: Make more changes and run this skill again
- Complete story: Mark story as done in sprint-status.yaml
- View changes: git log --oneline -5
```

---

## Important Notes

- **Zero interaction**: Fully automated workflow from staging to push
- Always commit with story reference for traceability
- Never commit directly to main/master
- Auto-push keeps remote repository in sync
- Update story status to reflect progress
- **Security first**: Input validation prevents injection attacks
- **Error recovery**: All critical operations include backup/rollback
- **Performance**: Use parallel Quick Check execution when possible
- **File safety**: Always validate file paths and permissions before writing
