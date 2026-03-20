---
name: build-tag
description: Create and push a new incremented version tag to remote. Use when releasing a new version, creating a tag, or when the user says "release tag" or "new tag".
allowed-tools: Bash, Read, Grep, LS
disable-model-invocation: true
argument-hint: "[tag_prefix]"
---

# Build Tag

检查代码状态、获取最新 tag、自动递增版本号并推送到远程。

## Scripts

| Script | Purpose |
|--------|---------|
| `check-clean.py` | 检查工作区干净且所有代码已推送到远程 |
| `get-latest-tag.py` | 获取远程最新 tag 并解析版本号 |
| `bump-tag.py` | 创建新 tag 并推送到远程（失败自动回滚） |

---

## 3-Step Workflow

### Step 1: 检查代码状态

确保所有代码已提交并推送到远程：

```bash
python3 .Codex/skills/build-tag/scripts/check-clean.py
```

**检查项**:
- 无未暂存的更改
- 无已暂存未提交的更改
- 无未跟踪的文件
- 本地与远程分支同步（无领先/落后提交）

如果检查失败，按提示操作后重试。

---

### Step 2: 获取最新 Tag

获取远程最新版本号并计算下一个版本：

```bash
python3 .Codex/skills/build-tag/scripts/get-latest-tag.py "${ARGUMENTS:-v}"
```

- 默认 tag 前缀为 `v`（如 `v1.2.3`）
- 可通过 `$ARGUMENTS` 指定其他前缀
- 自动解析 `major.minor.patch` 并计算 `NEXT_TAG`（patch +1）
- 如果无任何 tag，从 `v0.0.1` 开始

**输出确认**:
```
最新 Tag: v1.2.3
NEXT_TAG=v1.2.4
```

向用户确认即将创建的 tag 版本号，等待确认后继续。

---

### Step 3: 创建并推送 Tag

创建 annotated tag 并推送到远程：

```bash
python3 .Codex/skills/build-tag/scripts/bump-tag.py <next_tag> "Release <next_tag>"
```

- 创建 annotated tag（包含消息）
- 自动推送到远程
- 推送失败时自动回滚（删除本地 tag）

---

## Output Summary

```
✅ Build Tag 完成

分支: {current_branch}
提交: {commit_hash}
最新 Tag: {previous_tag} → 新 Tag: {new_tag}
消息: Release {new_tag}

远程已同步: origin/{new_tag}
```

---

## Error Handling

| 错误 | 解决方案 |
|------|----------|
| 工作区不干净 | `git add -A && git commit -m 'message'` |
| 本地领先远程 | `git push origin <branch>` |
| 本地落后远程 | `git pull origin <branch>` |
| Tag 已存在 | 检查版本号，使用不同版本 |
| 推送失败 | 检查网络或运行 `gh auth login` |

---

## Language

- **输出**: 中文
- **Tag 格式**: `v{major}.{minor}.{patch}`（默认）
