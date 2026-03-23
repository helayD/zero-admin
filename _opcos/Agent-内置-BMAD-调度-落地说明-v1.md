# zero-admin Agent 内置 BMAD 调度落地说明 v1

## 目的

将 BMAD 自动推进能力直接落到现有 `zero-admin` agent 上，而不是另建平行 agent 或硬编码脚本流程。

该 agent 必须同时承担：
- **对外层（Feishu 可见）**：状态判断、最优解决策、阶段进展回传、阻塞说明
- **执行层协调**：调用 codex-autopilot / tmux / watchdog 去执行 BMAD 节点

---

## 核心要求

### 1. 动作动态生成，不允许硬编码 story
系统中不允许把以下内容写死：
- `dev-story:2-2`
- `code-review:2-2`
- `refine-story:2-2`

允许固定的是：
- 动作类型（create-story / dev-story / analyze-current-state / code-review）

必须动态推导的是：
- `story key`

### 2. 真实状态优先
每次调度时必须优先读取：
- `_opcos/implementation-artifacts/sprint-status.yaml`
- 当前 story 文件
- `git status --short`
- `git log --oneline -10`
- `~/.autopilot/state/zero-admin.json`
- `~/.autopilot/logs/watchdog.log`
- 必要时 `autopilot:zero-admin` 的 tmux pane 输出

### 3. 每个 BMAD 节点一个新会话语义
必须保持：
- create-story 一个节点
- story refine 一个节点
- dev-story 一个节点
- code-review 一个节点
- review 修复 一个节点

---

## 推荐调度顺序

### 状态 A：无 ready-for-dev story
动作：
- `create-story`

### 状态 B：存在 ready-for-dev story，且无明显实现痕迹
动作：
- `dev-story:<derived_story_key>`

### 状态 C：代码已明显推进，但 story 工件未同步
动作：
- `analyze-current-state:<derived_story_key>`

### 状态 D：实现基本完成，可准备 review
动作：
- `code-review:<derived_story_key>`

### 状态 E：检测到阻塞
动作：
- `blocked:<reason>`

---

## Feishu 回传要求

在以下时机必须回传进展：
1. 判定出新的最优动作
2. 已派发执行任务
3. 关键里程碑完成
4. 遇到阻塞
5. 一个 BMAD 节点完成

推荐消息结构：
- 当前节点
- 当前状态
- 当前最优下一步
- 是否已派发
- 当前阻塞 / 风险

---

## 与 codex-autopilot 的关系

- `zero-admin agent` 是调度中枢
- `codex-autopilot` 是执行器
- watchdog 是运行时守护
- Feishu 是外部可见反馈面

这四者关系必须固定为：

> zero-admin 判断 -> codex-autopilot 执行 -> zero-admin 回传 Feishu

而不是：

> autopilot 自己决定流程并静默推进

---

## 当前落地状态

已落地到：
- `AGENTS.md`：写入动态调度、非硬编码 story、飞书进展回传规则

待继续落地：
1. 将 dispatcher 改造成“供 zero-admin agent 调用的底层执行器”
2. 让周期触发入口服务于现有 agent，而不是独立平行系统
3. 为 Feishu 进展消息增加固定模板与节流规则
