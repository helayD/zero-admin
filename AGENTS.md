# AGENTS.md

请使用中文回答。

## Agent 任务感知与 codex-autopilot 绑定

- 当前 agent 只负责本仓库：`/Users/helay/Documents/GitHub/zero-admin`
- 对应 codex-autopilot tmux 窗口：`zero-admin`
- 当用户问“我有什么任务没完成 / 现在在做什么 / 还有什么待办 / 当前任务是什么”时，**不要只根据聊天记忆回答**，也不要优先说“记忆不可用”。
- 必须优先检查与本仓库相关的真实状态来源：
  1. `git status --short`
  2. `git log --oneline -10`
  3. `~/.autopilot/state/zero-admin.json`
  4. `~/.autopilot/logs/watchdog.log` 里与 `zero-admin` 相关的最近记录
  5. 必要时查看 tmux 窗口 `autopilot:zero-admin` 的 pane 输出
- 回答任务类问题时，优先输出：
  - 已完成
  - 进行中
  - 未完成/待处理
  - 卡点/风险
  - 下一步建议
- 如果 autopilot 状态读不到，要明确说明是“autopilot 状态不可读”或“tmux 状态不可读”，不要笼统说“没有任务”或“只有当前群聊上下文可见”。

## Codex 会话质量守则

- 用户说“codex 清理上下文”时：在当前 Codex 会话里执行 `/compact`。
- 用户说“codex 继续 / 恢复会话”时：优先执行 `codex resume --last`。
- 用户说“codex 新会话”时：不要 `resume --last`，直接启动一个全新的 Codex 会话：`codex --dangerously-bypass-approvals-and-sandbox`。
- 对本仓库来说，**避免 Codex 上下文污染优先于少开新会话**。
- 只要出现以下迹象，就优先考虑 `/compact` 或直接新开会话：
  - 对当前需求、分支、任务目标开始混淆
  - 反复引用过时计划或旧实现
  - 把别的项目/别的模块上下文带进来
  - 连续多轮修改后推理质量明显下降
- 一旦怀疑上下文被污染，宁可保守地新开会话，也不要硬接着做。
- 在上下文不干净时：
  - 不做架构决策
  - 不做大规模重构
  - 不删改核心逻辑
- Compact 或新会话之后，先重新阅读仓库内与当前任务直接相关的约束文件，例如：`AGENTS.md`、`CONVENTIONS.md`、`prd-todo.md`、`task_plan.md`、`findings.md`、`progress.md`（若存在），再继续编码。

## 绑定项目群回复规则

- 在已绑定的 Feishu 项目群里，**不需要 @ 才回复**。
- 只要消息来自当前绑定项目群，且内容属于以下类型，就应直接回复，不要过度克制：
  - 明确提问
  - 任务推进 / 继续执行
  - 状态询问 / 进度追问
  - 要求总结、汇报、给方案、做判断
  - 与当前仓库、当前任务、当前 autopilot/codex 状态直接相关的指令
- 不要因为“群聊礼貌”或“避免打扰”而默认沉默；在这些项目群里，及时反馈比克制更重要。
- 如果信息不足，也要直接回一个简短但有用的澄清或当前状态，而不是不回。
- 只有在明显与本项目无关、别人彼此闲聊、或你确实无法提供价值时，才保持安静。
