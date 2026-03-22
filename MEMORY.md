# MEMORY.md

## 2026-03-22

- 用户明确要求：在 `zero-admin` 仓库里，生成同类 BMAD `create-story` 文档时，必须以项目内最近同类 story 作为风格与展开粒度基线，尤其要对齐 `1.6A` / `1.6B` 这类现有样板；不能只做到结构合规、内容可用，却在篇幅、章节密度、仓库现实细节密度上明显缩水。
- 用户明确纠正：当同仓库、同类型、同阶段的 BMAD story 输出风格不一致时，不能用“阶段不同”“后续会补厚”这类解释搪塞；应直接承认未按规则严格对齐，并优先重写或补齐到项目内统一基线。
- 执行规则：后续在审核 BMAD story 时，除了检查内容是否 `ready-for-dev`，还必须额外对比最近同类 story 的五个重点区块是否同级展开：`Tasks / Subtasks`、`Existing Repository Reality`、`Technical Requirements`、`Testing Requirements`、`Project Structure Notes`。
- 失败预防：如果新 story 与最近同类样板在风格或密度上明显不一致，不应直接给出“审核通过”结论，而应先补齐再评审。
