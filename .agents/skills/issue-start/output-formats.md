# 各步骤输出格式参考

本文件定义了 issue-start 工作流各步骤的标准输出格式。Claude 在执行每个步骤后必须按照此格式输出。

---

## Step 1: 业务上下文分析

```
Business Context Analysis - Issue #$ARGUMENTS:
- Business value: {description}
- User personas: {list}
- Affected workflows: {list}
- Success metrics: {KPIs}
- Critical business rules: {list}
```

### Step 1 验证清单

- ✅ analysis.md 存在于 `.claude/epics/<epic>/issues/$ARGUMENTS/analysis.md`
- ✅ analysis.md 包含所有必需 section
- ✅ 所有验收标准已从任务文件中识别
- ✅ 业务价值和用户影响已理解
- ✅ 成功指标已定义
- ✅ 业务规则已记录
- ✅ analysis.md 的 Business Context 已更新

---

## Step 2: 代码审查

```
Code Review - Issue #$ARGUMENTS:
- Feature status: {exists/missing/partial}
- Code quality: {good/needs-improvement/defective}
- Main issues found: {list}
- Current implementation: {what code actually does}
- Architecture fit: {can-extend-existing/need-new-files}
- Files to extend: {list existing files}
- Files to create: {list with justification or "none"}
- Required action: {implement/enhance/fix/refactor}

Dependencies & Integration - Issue #$ARGUMENTS:
- Dependent modules: {list modules that depend on this}
- Required modules: {list modules this depends on}
- API contracts: {list affected APIs and breaking changes}
- Data flows: {describe input/output between modules}
- Integration points: {list integration interfaces}
- Performance impact: {high/medium/low - details}
- Breaking changes: {yes/no - list changes}
- Migration required: {yes/no - strategy}
```

### Step 2 验证清单

- ✅ Code Review 输出已生成（完全匹配格式）
- ✅ Dependencies & Integration 输出已生成（完全匹配格式）
- ✅ 所有相关文件已读取和分析
- ✅ 架构适配性已评估
- ✅ 跨模块依赖已映射
- ✅ analysis.md 的 Technical Approach、Affected Files、Dependencies & Integration 已更新

---

## Step 3: 代码质量检查

```
Code Quality Check - Issue #$ARGUMENTS:
- Linter status: {Clean/X errors/warnings}
- Test coverage: {X%}
- Applicable rules: {list key rules from .claude/rules/}
- Architecture: {e.g., DDD with Riverpod/Bloc, MVC, Clean Architecture}
- File naming: {e.g., snake_case, camelCase, PascalCase}
- Standards compliance: {compliant/needs-improvement}
- File headers: {added/missing/not-applicable}
- Redundant code: {none/cleaned/needs-cleanup}

Security & Performance Check - Issue #$ARGUMENTS:
- Security vulnerabilities: {none/list issues}
- Authentication/Authorization: {properly-implemented/needs-review}
- Data validation: {comprehensive/missing}
- Input sanitization: {present/missing}
- Performance bottlenecks: {none/list concerns}
- Database query optimization: {optimized/needs-review}
- API response time: {acceptable/needs-optimization}
- Memory/resource leaks: {none/detected}
```

### Step 3 验证清单

- ✅ Linter 已运行并有实际结果
- ✅ 测试覆盖率已检查
- ✅ Quality Check 输出已生成
- ✅ Security & Performance Check 输出已生成
- ✅ 无关键安全漏洞
- ✅ 无主要性能问题

---

## Step 4: 实现方案决策

```
Implementation Decision - Issue #$ARGUMENTS:
- Decision: {Proceed/No-change/Review-needed}
- Reason: {brief explanation}
- Approach: {extend-existing/create-new/refactor}
- Files to modify: {list}
- Files to create: {list with justification or "none"}
- Estimated changes: {number of lines/files}
```

```
File Operation Plan - Issue #$ARGUMENTS:
Extend existing:
  - path/to/file1.ext: add method X, add field Y
  - path/to/file2.ext: add method Z

Create new: {list with strong justification or "none"}
  - path/to/new_file.ext: {reason why existing files cannot be extended}

Rationale: {explain why this is the minimal approach}
```

```
Risk Assessment - Issue #$ARGUMENTS:
Technical Risks:
  - Risk: {description} | Impact: {high/medium/low} | Mitigation: {strategy}

Business Risks:
  - Risk: {description} | Impact: {high/medium/low} | Mitigation: {strategy}

Rollback Strategy:
  - Rollback plan: {description}
  - Database migration rollback: {applicable/not-applicable - details}
  - Feature flag: {yes/no - flag name}
  - Monitoring alerts: {list critical metrics to monitor}

Validation Plan:
  - Pre-deployment checks: {list}
  - Post-deployment monitoring: {metrics and duration}
  - Smoke tests: {list critical paths}
```

### Step 4 验证清单

- ✅ Implementation Decision 输出
- ✅ File Operation Plan 输出
- ✅ Risk Assessment & Mitigation 输出
- ✅ analysis.md 的 Implementation Plan 和 Risk Mitigation 已更新
- ✅ 回滚策略已定义
- ✅ 监控计划已建立

---

## Step 7: 验证实现结果

GitHub Issue 评论（通过时）：
```bash
gh issue comment $ARGUMENTS --body "Implementation complete

**Changes**: {list}
**Tests**: {results}
**Verification**: All criteria met, tests passing, linter clean
**Standards**: Code formatted, zero errors, headers added, no redundant code, reuse confirmed
**Business Validation**: All business rules enforced, acceptance criteria met, user workflows verified
**Performance**: API response times acceptable, database queries optimized, no memory leaks
**Metrics**: {list instrumented metrics for tracking success}"

gh issue edit $ARGUMENTS --add-label "ready-for-review" --remove-label "in-progress"
```

GitHub Issue 评论（有问题时）：
```bash
gh issue comment $ARGUMENTS --body "Implementation needs attention

**Issues**: {list}
**Next steps**: {fixes needed}"
```

---

## Step 9: 最终总结

```
✅ Completed Issue #$ARGUMENTS

Epic: <epic_name>
Task file: .claude/epics/<epic_name>/$ARGUMENTS.md
Issue dir: .claude/epics/<epic_name>/issues/$ARGUMENTS/

Documentation Files (ALL CREATED):
  ✅ analysis.md - Technical analysis and decisions
  ✅ progress.md - Development timeline
  ✅ tests/ - Test files ({X%} coverage)

Cross-References:
  ✅ Task file frontmatter updated
  ✅ GitHub Issue commented
  ✅ Traceability chain established

Workflow: Analysis → Code Review → Quality Check → Decision → Progress Setup → Coding → Verification → Peer Review → Documentation

Verification: Tests {X/Y passed}, Linter {status}, Functionality {status}, Cohesion {centralized/scattered}

File Structure:
  .claude/epics/<epic_name>/issues/$ARGUMENTS/
  ├── analysis.md (MANDATORY)
  ├── progress.md (MANDATORY)
  └── tests/

GitHub Issue: https://github.com/{repo}/issues/$ARGUMENTS
```
