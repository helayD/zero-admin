# 验证报告格式参考

---

## 总体报告

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━
Epic Validation: <epic_name>
Generated: {timestamp}

Overall Score: {score}/100

Tasks: {total}
✅ Passed: {count}
⚠️  Warnings: {count}
❌ Failed: {count}

Critical Issues: {count}
━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 单任务报告

### 通过示例
```
Task 001 - User Authentication
Score: 95/100 ✅
✅ Frontmatter: Complete
✅ Task Metadata: Complete (File, Purpose, Leverage, Requirements, Prompt)
✅ Technical Details: API specs included
✅ Self-Containment: No external refs
⚠️  Acceptance: Missing metrics
```

### 失败示例
```
Task 003 - Permission Management  
Score: 45/100 ❌
❌ CRITICAL: Missing API specifications
❌ CRITICAL: External reference found ("see Epic.md")
❌ CRITICAL: Missing Prompt field

Action Required:
1. Add complete API specifications
2. Remove external references
3. Add Prompt with Role/Task/Restrictions/Success
```

---

## 分数阈值

| Score | Status | Action |
|-------|--------|--------|
| 90-100 | ✅ Excellent | Ready for implementation |
| 75-89 | ⚠️ Acceptable | Minor fixes recommended |
| 60-74 | ⚠️ Needs Work | Fix before implementation |
| <60 | ❌ Failed | Must fix before proceeding |

---

## Critical Violations (自动失败)

- 缺少 API 规格（API 任务）
- 缺少 Technical Details section
- 外部引用 Epic.md（pattern: `(see|refer|check).*epic\.md`）
- 不完整规格（TBD/TODO in required fields）
- 缺少 Prompt 字段或格式无效
- 无 Definition of Done
- 无效 YAML frontmatter
- 缺少必需 metadata 字段（File, Purpose, Requirements）
