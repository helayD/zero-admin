# BMAD 自动推进规则 v1（zero-admin）

## 目标

为 `zero-admin` 项目建立一套稳定的自动推进机制：

> 定时获取真实状态 → 计算当前最优解 → 通过 codex-autopilot 执行 → 每完成一个 BMAD 节点就切换到一个新会话。

本规则明确：
- 状态从哪里读取
- 如何判断“当前最优下一步”
- 何时自动执行，何时必须人工确认
- 如何保证 BMAD 流程与 codex-autopilot 协同，而不是互相打架

---

## 一、适用范围

本规则当前仅默认适用于：
- 项目：`zero-admin`
- 工作方式：`BMAD + codex-autopilot`
- 执行介质：tmux + watchdog + Codex CLI

不适用于：
- ACP 路由
- 普通长对话持续编码
- 无 story / 无 sprint-status 的临时任务

---

## 二、核心原则

### 1. BMAD 是流程真相源
项目推进以 BMAD 工件为准，而不是聊天上下文为准。

优先读取：
- `_opcos/implementation-artifacts/sprint-status.yaml`
- 当前 story 文件
- `_opcos/planning-artifacts/` 下的 epic / prd / architecture / ux-design
- 必要时 project-context / AGENTS / repo 内 BMAD skill 工件

### 2. codex-autopilot 是执行器，不是流程定义者
codex-autopilot 负责：
- 执行任务
- 维持窗口
- 恢复会话
- 自动继续
- 收集状态

但 **下一步做什么**，由 BMAD 状态机决定。

### 3. 每个 BMAD 节点独立会话
必须遵守：
- `create-story` 一个会话
- `story-review / refine-story` 一个会话
- `dev-story` 一个会话
- `code-review` 一个会话
- `fix-review-findings` 一个会话
- `correct-course` 一个会话

禁止把多个 BMAD 节点塞进同一个长会话里持续滚动。

### 4. 每次先算最优解，再执行
自动化不允许“看到未完成就继续硬做”。

每次触发时必须先判断：
- 当前在哪个 BMAD 节点
- 上一步是否真的完成
- 当前最优下一步是什么
- 是否存在阻塞或需要人工决策

---

## 三、状态源定义

每次自动触发时，至少读取以下状态源。

### A. BMAD 工件状态
1. `sprint-status.yaml`
   - 判断当前 epic / story 顺序
   - 判断第一条 `ready-for-dev`
   - 判断是否已经 `in-progress` / `review` / `done`

2. 当前 story 文件
   - `Status`
   - `Tasks / Subtasks`
   - `Dev Agent Record`
   - `Completion Notes`
   - `File List`

3. planning artifacts
   - `epic.md`
   - `prd.md`
   - `architecture.md`
   - `ux-design.md`

### B. 仓库实施状态
1. `git status --short`
2. 最近提交记录
3. 是否存在未提交改动
4. 是否存在代码已改但 story 未收口的情况

### C. autopilot 运行状态
1. `~/.autopilot/state/zero-admin.json`
2. `~/.autopilot/logs/watchdog.log`
3. tmux 对应窗口是否活跃
4. Codex 是否卡在：
   - 登录选择页
   - 权限确认
   - shell 态
   - 空闲无任务

---

## 四、BMAD 自动推进状态机

### 状态 1：没有可开发 story
判定条件：
- `sprint-status.yaml` 中不存在 `ready-for-dev` story
- 且下一条故事仍为 `backlog`

最优动作：
- 运行 `create-story`

输出：
- 新建 story 文件
- 更新 `sprint-status` 到 `ready-for-dev`
- 创建新的独立会话

---

### 状态 2：story 已 ready-for-dev，但尚未进入实现
判定条件：
- 存在 `ready-for-dev` story
- 仓库里没有与该 story 对应的明确实现推进
- 没有大量未提交改动指向该 story

最优动作：
- 运行 `dev-story`

输出：
- 新建独立实现会话
- 开始针对该 story 编码

---

### 状态 3：代码已推进，但 story 工件未同步
判定条件：
- `git status` 有明显该 story 相关改动
- 但 story 仍是 `ready-for-dev` 或未更新

最优动作：
- 运行“阶段性验收分析”
- 判断是否：
  1. 继续开发
  2. 补测试
  3. 收口 story
  4. 进入 review

注意：
- 该状态下不能盲目继续写代码
- 必须先做一次现状诊断

---

### 状态 4：实现基本完成，待 review
判定条件：
- Acceptance Criteria 基本满足
- 关键测试通过
- story 文件可收口

最优动作：
- 更新 story 文档
- 更新 `sprint-status`
- 新开会话进入 `code-review`

---

### 状态 5：review 未通过
判定条件：
- review 返回问题
- 存在待修复项

最优动作：
- 新建一个 `fix-review-findings` 会话
- 不复用旧 dev-story 会话

---

### 状态 6：BMAD 节点完成
判定条件：
- 当前节点目标达成
- 工件状态与代码状态一致

最优动作：
- 关闭当前节点语义
- 下一次调度时再从状态机选择下一节点

---

## 五、最优解判定优先级

每次自动触发时，按以下优先级决定动作。

### 优先级 P0：阻塞处理
优先处理以下问题：
- Codex 卡在登录页 / 权限页 / shell
- 测试失败
- 关键工件缺失
- 当前 story 与代码改动不一致
- 存在明显越界开发风险

### 优先级 P1：流程闭环
如果没有阻塞，则优先保证 BMAD 流程闭环：
- create-story
- dev-story
- story 收口
- code-review
- 修复 review 问题

### 优先级 P2：局部优化
只有在 P0/P1 都稳定时，才做：
- 额外清理
- 非关键重构
- 次要体验优化

---

## 六、定时触发规则

### 工作时段
建议：
- 每 10~15 分钟触发一次状态检查

### 非工作时段
建议：
- 每 30~60 分钟触发一次
- 如无新状态变化，不执行实际任务

### 每次触发做什么
每次定时任务只做两件事：
1. 读取状态
2. 判断是否触发下一步

不允许定时任务本身承担整条开发流程。
它只是：
- 状态检查器
- 任务派发器

真正编码由 codex-autopilot 窗口执行。

---

## 七、codex-autopilot 对接方式

### 对接原则
BMAD dispatcher 不直接“实现代码”，只向 autopilot 派发明确任务。

### 任务格式建议
每次只派发一个清晰任务，例如：
- `create-story 2-2`
- `review-story 2-2`
- `dev-story 2-2`
- `analyze-current-state 2-2`
- `code-review 2-2`
- `fix-review-findings 2-2`

### 禁止事项
禁止把这类模糊任务直接喂给 autopilot：
- “继续做”
- “你看着办”
- “顺着往下推进”

自动化里必须始终是**单节点、单目标、单边界**任务。

---

## 八、自动执行边界

### 可自动执行
以下动作默认允许自动执行：
- `create-story`
- story 复核 / 修订
- `dev-story`
- `code-review`
- review 问题修复
- 测试 / 校验 / 文档更新
- story / sprint-status 收口

### 必须人工确认
以下动作默认必须先问用户：
- 跨 epic 调整范围
- 大规模 schema 重构
- 生产发布
- 删除大量代码 / 数据
- 改 BMAD 主计划 / sprint 顺序
- 修改项目级流程规则本身

---

## 九、完成判定

### create-story 完成条件
- story 文件已生成
- `sprint-status` 更新为 `ready-for-dev`
- story 边界清晰

### dev-story 完成条件
- 关键 AC 满足
- 测试通过
- story 文件已更新：
  - 任务勾选
  - Completion Notes
  - File List
  - Status
- `sprint-status` 同步推进

### review-ready 完成条件
- 代码状态与 story 状态一致
- 没有明显伪完成
- review 所需上下文完整

---

## 十、推荐的 v1 落地顺序

### 第一步：先做规则，不急着全自动
先按本规则人工/半自动执行，验证：
- 状态源是否够用
- 决策是否稳定
- 会话边界是否清楚

### 第二步：加 dispatcher
增加一个轻量 dispatcher：
- 定时检查状态
- 输出唯一最优动作
- 派发给 codex-autopilot

### 第三步：加守护与回退
当遇到以下情况时自动停止推进：
- 登录阻断
- 权限阻断
- 测试红线
- story 与代码状态冲突

---

## 十一、当前默认执行约定（立即生效）

对 `zero-admin`，从现在开始默认采用以下规则：

1. 按 BMAD 流程推进项目
2. 每个 BMAD 节点使用一个新会话语义
3. 每次先读取真实状态，再决定最优下一步
4. codex-autopilot 作为执行器，不作为流程定义者
5. 未完成不伪装为完成
6. 代码状态与 BMAD 工件状态必须最终一致

---

## 十二、下一步实施建议

建议紧接着做两件事：

1. 为 `zero-admin` 增加一个 BMAD dispatcher 规则文件
2. 将 watchdog / task queue 与该规则对接，形成：
   - 定时检查
   - 计算最优解
   - 派发单节点任务

这样才能真正实现：

> 自动拿状态 → 自动判断 → 自动推进 → 自动切换 BMAD 节点会话
