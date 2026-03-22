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
