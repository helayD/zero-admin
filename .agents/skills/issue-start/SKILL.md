---
name: issue-start
description: Begin work on a GitHub issue by clarifying coding standards and starting implementation. Use when starting development on a specific issue number, beginning implementation work, or when the user says "start issue".
allowed-tools: Bash, Read, Write, Grep, Glob, Edit, LS
argument-hint: "[issue_number]"
disable-model-invocation: true
---

# Issue Start

Begin work on a GitHub issue by clarifying coding standards and starting implementation.

## Scripts

| Script | Purpose |
|--------|---------|
| `quick-check.py` | 环境校验、Issue 查找、Epic 发现 |
| `init-analysis.py` | 从模板创建 analysis.md |
| `init-progress.py` | 从模板创建 progress.md |
| `detect-stack.py` | 自动检测项目技术栈（linter/test/formatter） |
| `log-progress.py` | 向 progress.md 追加带时间戳的日志 |
| `verify-docs.py` | 验证必需文档是否完整 |

## Templates

| Template | Purpose |
|----------|---------|
| `analysis-template.md` | 技术分析文档模板 |
| `progress-template.md` | 开发进度日志模板 |

## Additional Resources

- 各步骤标准输出格式详见 [output-formats.md](output-formats.md)

---

## MANDATORY COMPLIANCE

**CRITICAL**: 严格按 9 步工作流顺序执行。不得跳步、合并步骤或未经验证即继续。

### 必需文档（无例外）

1. **`analysis.md`** - 技术分析文档（Step 1 创建，Steps 2-4 更新）
2. **`progress.md`** - 开发进度日志（Step 5 创建，Step 6 持续更新）
3. **文档交叉引用** - 最终总结（Step 9 执行）

缺少任何文档 = 工作流终止。

---

## Role Definition

**You are a senior software engineer with 10 years of development experience (NOT a system architect).** Focus on practical, implementable technical solutions: code structure, API design, data models, and concrete implementation details.

---

## Core Principles

- **扩展优先于创建**: 优先扩展现有文件，新建需强有力理由
- **业务上下文感知**: 理解业务领域和用户价值后再实现
- **跨模块影响分析**: 识别依赖和集成点
- **所有任务类型都可能需要编码**: 即使标题含"验证"，也需执行全部步骤判断是否需要编码

---

## Quick Check

运行环境校验脚本：
```bash
python3 .Codex/skills/issue-start/scripts/quick-check.py $ARGUMENTS
```

此脚本会输出 `TASK_FILE`、`EPIC_NAME`、`ISSUE_DIR`、`ANALYSIS_EXISTS` 等变量。后续步骤中使用这些值。

如果脚本失败，报告错误并终止。

---

## 9-Step Workflow

### Step 1: Read Analysis and Task Details

**1.1: 读取任务文件**（`$ARGUMENTS.md`）：
- 理解验收标准和用户需求
- 注意特定技术要求
- 识别受影响的模块和文件

**1.2: 读取或创建 analysis.md**

如果 Quick Check 显示 `ANALYSIS_EXISTS=false`：
```bash
python3 .Codex/skills/issue-start/scripts/init-analysis.py $ARGUMENTS <epic_name> "任务概述"
```

如果已存在：读取技术方案和架构决策。

**1.3: 业务上下文分析（MANDATORY）**

评估并更新 analysis.md 的 "Business Context" section：
- **业务价值**: 解决什么业务问题？
- **目标用户**: 哪些用户角色受影响？
- **用户工作流影响**: 涉及哪些用户流程？UI/UX 变更？
- **成功指标**: 如何衡量成功？KPI？
- **业务规则与约束**: 关键业务规则、合规要求

**✅ STEP 1 验证** - 详见 [output-formats.md](output-formats.md) Step 1 section

❌ **验证全部通过前不可进入 Step 2！**

---

### Step 2: Code Review

**2.1: 定位相关文件**
```bash
grep -r "related_keyword" lib/ src/
```

**2.2: 架构适配评估**
- 哪些现有文件实现类似功能？
- 这些文件的职责是否覆盖新需求？
- 扩展是否违反单一职责原则？
- 是否有可复用的 services/components？

**2.3: 跨模块依赖分析**
- 识别所有相互依赖的模块/服务
- 映射模块间数据流（输入/输出契约）
- 检查 API 契约和接口定义
- 识别潜在的破坏性变更
- 审查认证/授权要求
- 评估跨模块性能影响

**2.4: 读取和审查代码**
- 理解当前实现、业务逻辑和数据流
- 审查错误处理和边界情况
- 检查编码标准、命名规范
- 识别冗余代码

**🔴 MANDATORY: 更新 analysis.md**：
- "Technical Approach" section
- "Affected Files" section
- "Dependencies & Integration" section

**✅ STEP 2 验证** - 详见 [output-formats.md](output-formats.md) Step 2 section

❌ **验证全部通过前不可进入 Step 3！**

---

### Step 3: Verify Code Quality

检测技术栈：
```bash
python3 .Codex/skills/issue-start/scripts/detect-stack.py .
```

使用检测到的命令运行：
- **Linter**: 零错误/警告
- **测试覆盖率**: 使用语言特定框架
- **命名规范**: 审查目标文件
- **编码模式**: 识别需遵循的模式
- **安全与性能**: 漏洞、认证、数据验证、数据库查询优化

**✅ STEP 3 验证** - 详见 [output-formats.md](output-formats.md) Step 3 section

❌ **验证通过前不可进入 Step 4！**

---

### Step 4: Decide on Implementation Approach

**决策标准：**
- 功能已存在且正常 → "no-change-needed"
- 小修复（1-2 文件，<50 行）→ 定向修复
- 新功能且架构支持扩展 → 扩展现有文件
- 需要新文件 → 给出合理理由
- 需要大规模重构 → 标记待审查

**新文件创建检查：**
1. 没有现有文件具有匹配职责
2. 不会造成职责重叠或逻辑分散
3. 有效理由：独立实体、文件 >500 行、独立生命周期

**风险评估与缓解（MANDATORY）**：
- 技术风险 + 业务风险 + 影响等级 + 缓解策略
- 回滚策略、数据库迁移回滚、Feature flag、监控告警
- 验证计划：预部署检查、部署后监控、冒烟测试

**🔴 MANDATORY: 更新 analysis.md**：
- "Implementation Plan" section
- "Risk Mitigation" section

如果决策为 "No-change" 或 "Review-needed"，在此停止并更新 GitHub issue。

**✅ STEP 4 验证** - 详见 [output-formats.md](output-formats.md) Step 4 section

❌ **验证全部通过前不可进入 Step 5！**

---

### Step 4.5: Fast Path (小改动/Hotfix)

仅当 **全部** 满足以下条件时使用简化路径：
- 范围 ≤2 文件且 <50 LOC
- 无 schema 迁移或公共 API 契约变更
- 低风险，无跨模块破坏性影响
- 测试已添加/更新，覆盖率 ≥90%

简化步骤：Step 2（仅审查直接相关文件）→ Step 5 → Step 6 → Step 7（完整验证）→ Step 9

注意：analysis.md 仍需记录 Technical Approach 和 Affected Files（简要版）。

---

### Step 5: Setup Progress Tracking

**5.1: 更新任务文件** frontmatter 的 `updated` 字段。

**5.2: 创建 progress.md**
```bash
python3 .Codex/skills/issue-start/scripts/init-progress.py $ARGUMENTS <epic_name>
```

**5.3: GitHub 分配**
```bash
gh issue edit $ARGUMENTS --add-assignee @me --add-label "in-progress"
```

**✅ STEP 5 验证清单**
- ✅ Issue 目录已创建
- ✅ progress.md 已创建且包含所有必需 section
- ✅ 任务文件 frontmatter 已更新
- ✅ GitHub issue 已分配并标记 "in-progress"

❌ **验证全部通过前不可进入 Step 6！**

---

### Step 6: Execute Coding Work

实现原则、编码标准和工作流详见 [coding-standards.md](coding-standards.md)。

**编码工作流**: 文件定位 → Domain 层 → Data 层 → Presentation 层 → 测试（≥90%）→ 文档

**🔴 MANDATORY: 每次重要变更后记录进度**
```bash
python3 .Codex/skills/issue-start/scripts/log-progress.py $ARGUMENTS <epic_name> "完成了什么"
```

**Note:** 使用 `issue-commit` skill 提交变更。

**✅ STEP 6 验证清单**
- ✅ 所有文件按 analysis.md 计划实现
- ✅ 测试已编写（覆盖率 ≥90%）
- ✅ progress.md 持续更新

❌ **验证全部通过前不可进入 Step 7！**

---

### Step 7: Verify Implementation Results

详细验证清单参见 [coding-standards.md](coding-standards.md) Step 7 section。

核心验证：
1. **运行测试**: Unit → Integration → E2E → Coverage（≥90%）
2. **Linter**: 零错误/警告
3. **功能**: 验收标准满足，无回归
4. **业务逻辑**: 规则正确、数据验证、错误处理、边界情况
5. **性能**: API <2s、无 N+1 查询

**🔴 MANDATORY: 在 progress.md 中记录验证结果**

**同步 GitHub Issue** - 详见 [output-formats.md](output-formats.md) Step 7 section

**✅ STEP 7 验证** - 详见 [output-formats.md](output-formats.md)

❌ **验证全部通过前不可进入 Step 8！**

---

### Step 8: Code Review

**自审清单：**
- ✅ 验收标准满足
- ✅ Linter 零错误
- ✅ 测试通过（覆盖率 ≥90%）
- ✅ 无 bug 或回归
- ✅ 业务规则正确执行
- ✅ 错误处理完善
- ✅ 安全已验证
- ✅ 性能可接受
- ✅ 标准合规
- ✅ 文档完整
- ✅ 跨模块集成已验证

自审通过后，使用 `issue-commit` skill 提交变更，然后创建 PR。

**✅ STEP 8 验证:** 自审完成，准备提交和 PR

---

### Step 9: Documentation Cross-Reference and Final Summary

**9.1: 更新任务文件 frontmatter**

在 `.Codex/epics/<epic>/$ARGUMENTS.md` 中添加：
```yaml
documentation:
  analysis: .Codex/epics/<epic>/issues/$ARGUMENTS/analysis.md
  progress: .Codex/epics/<epic>/issues/$ARGUMENTS/progress.md
  tests: .Codex/epics/<epic>/issues/$ARGUMENTS/tests/
```

**9.2: GitHub Issue 评论**
```bash
gh issue comment $ARGUMENTS --body "## 📋 Development Documentation

**Technical Analysis**: \`.Codex/epics/<epic>/issues/$ARGUMENTS/analysis.md\`
- Technical decisions, architecture fit, implementation plan

**Development Progress**: \`.Codex/epics/<epic>/issues/$ARGUMENTS/progress.md\`
- Complete timeline with all major changes

**Tests**: \`.Codex/epics/<epic>/issues/$ARGUMENTS/tests/\` (Coverage: {X%})

**Traceability Chain**: Task File → Analysis → Progress Log → Code Changes → Tests

This assists with: debugging, knowledge transfer, maintenance context, quick diagnosis"
```

**9.3: 验证文档完整性**
```bash
python3 .Codex/skills/issue-start/scripts/verify-docs.py $ARGUMENTS <epic_name>
```

**9.4: 输出最终总结** - 详见 [output-formats.md](output-formats.md) Step 9 section

**✅ STEP 9 验证清单**
- ✅ 任务文件 frontmatter 包含文档路径
- ✅ GitHub Issue 有文档评论和链接
- ✅ analysis.md 存在且完整
- ✅ progress.md 存在且完整
- ✅ 最终总结已生成
- ✅ 可追溯链已建立

**🎉 Workflow Complete!**

---

## Error Handling

如果任何步骤失败：
- 清晰报告："ERROR {失败内容}: {修复方法}"
- 不要留下部分状态
- 关键错误需更新 GitHub issue

**快速恢复：**
- **GitHub 认证错误**: `gh auth login`
- **权限错误**: `ls -la .Codex/epics/`
- **Git 冲突**: `git status && git pull --rebase`
- **磁盘空间**: `df -h .`
- **部分状态**: `git reset --hard HEAD`（先备份！）

---

## Important Notes

### 文件存在 ≠ 功能实现（CRITICAL）
- 创建文件只是第一步，必须关注实际实现
- 修改后验证代码逻辑正确实现需求
- 避免空壳文件，完成完整业务逻辑
- 通过单元测试和集成测试确认功能真正可用

### Development Standards
- 遵循 `.Codex/rules/datetime.md` 的时间戳格式
- 所有文件操作在当前仓库
- 优先扩展现有文件
- 复用现有代码模式
- 为每个函数编写测试
- 频繁提交并附描述性消息
- **使用中文进行所有交互**
