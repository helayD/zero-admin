# 任务比较算法详细参考

本文件定义了 epic-decompose 工作流中 Step 2 使用的任务比较算法。

---

## 存在性匹配规则（OR 逻辑，任一匹配即视为存在）

| 匹配类型 | 规则 | 示例 |
|----------|------|------|
| **编号匹配** | Epic 中 `001` → 检查 `001-*.md` 是否存在 | `001` → `001-user-list.md` |
| **标题匹配** | 标题相似度 ≥ 0.70 | "用户列表页" ≈ "用户列表管理页" |
| **内容匹配** | 功能描述相似度 ≥ 0.75 | 相似功能点描述 |

**任一匹配 → 文件存在，进入深度比较**
**全部不匹配 → 文件不存在，标记为 CREATE 候选**

---

## 深度内容比较（判定 KEEP vs UPDATE）

### 比较字段

| 字段 | 说明 |
|------|------|
| `name` | 标题文本 |
| `功能点` | 核心功能列表 |
| `用户操作流程` | 交互步骤 |
| `UI 元素清单` | 页面/组件/弹窗列表 |
| `Dependencies` | 任务和外部依赖 |
| `Technical Details` | 技术规格（Component/API/State/SQL/Model） |
| `Task Metadata` | File, Purpose, Leverage, Requirements, Prompt |

### 相似度公式

```
similarity = (title × 0.3) + (features × 0.2) + (workflow × 0.2) + (deps × 0.15) + (specs × 0.15)
```

### 各维度计算（Jaccard 系数）

```
dimension_sim = matched_count / max(epic_count, existing_count)
```

- **title**: 分词后关键词匹配（忽略停用词）
- **features**: 功能描述匹配（相同动词+宾语 = 匹配）
- **workflow**: 用户操作步骤匹配
- **deps**: 依赖项（Task ID / 外部服务）匹配
- **specs**: 技术规格（Component/API/State/Model）匹配

### 示例

```
Epic: [list, search, detail, delete]  Existing: [list, filter, detail]
Matched: [list, detail] = 2, max(4,3) = 4
features_sim = 2/4 = 0.50
```

### 技术规格差异检测

- 统计 Epic 中新增的技术规格数
- 如果缺少 ≥2 个规格 → 标记 UPDATE

---

## 决策逻辑

### 文件存在时（Step 1 匹配成功）

| 条件 | 操作 |
|------|------|
| similarity ≥ 0.85 **OR** status 为 completed/in-progress | **KEEP** |
| similarity < 0.85 **OR** 缺少 ≥2 specs **AND** status 为 open | **UPDATE** |
| ❌ 永远不会 | **CREATE**（文件存在时禁止创建） |

### 文件不存在时（Step 1 无匹配）

| 条件 | 操作 |
|------|------|
| 无匹配文件 | **CREATE** |
| ⚠️ 创建前 | 二次验证无类似文件存在（防重复） |

### 孤立任务

| 条件 | 操作 |
|------|------|
| 文件存在但 Epic 中未引用 | **ORPHAN**（自动标记 deprecated） |

---

## 冲突检测（Step 3）

| 相似度范围 | 处理 |
|-----------|------|
| 0.75 - 0.85 | 潜在冲突，标记 `conflicts_with` |
| ≥ 0.85 | 高度相似，建议合并 |

冲突处理原则：
- **不删除**: 保留所有现有任务
- **标记**: 在 frontmatter 添加 `conflicts_with: ["003", "007"]`
- **日志**: 输出冲突详情供人工审查
