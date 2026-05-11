#!/usr/bin/env bash
# Story 10.11 / Task 8.4：CI 守卫脚本，防回退。
#
# 目的：
#   - 阻止任何人重新引入「免费链 mock 底座」相关的实现代码：
#       * 函数标识符 buildFreeChainReceipt / shortDigest（曾在 mock 客户端使用）
#       * 错误文案 "免费链能力未启用"
#       * chainType 返回值 "free_chain"（必须永远是 "fisco_bcos_3x" / "antchain"）
#   - 仅扫描业务源码（pkg / rpc / api / consumer / job），不扫描 _test.go，因为
#     单元测试经常需要把这些字符串当作「负面断言」来保护反向回退。
#   - 仅扫描 .go 文件，避免误报 docs/ 历史档案 / 提交信息引用。
#
# 退出码：
#   0  ✅ 干净
#   1  ❌ 检测到回退迹象
#
# 用法：
#   bash scripts/check-no-free-chain.sh
#
# 在 CI 中接入：
#   - GitHub Actions：增加 step 调用本脚本
#   - Makefile：lint 目标增加调用本脚本
#
# 相关 AC：AC2 + AC6 — 错误体系不再出现「免费链能力未启用」字串。

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# 黑名单 pattern；用 \| 拼接成单一 ERE。
PATTERN='buildFreeChainReceipt|"free_chain"|免费链能力未启用|免费链 mock|mock fiscoClient'

EXCLUDES=(
    --include='*.go'
    --exclude='*_test.go'
)

DIRS=(pkg rpc api consumer job)

# shellcheck disable=SC2068
HITS=$(grep -RnE "${EXCLUDES[@]}" "$PATTERN" ${DIRS[@]} 2>/dev/null || true)

if [[ -n "$HITS" ]]; then
    echo "❌ Story 10.11 回退守卫：检测到禁止的字符串/符号" >&2
    echo "$HITS" >&2
    echo "" >&2
    echo "如果这是合法的负面断言，请把代码移到 *_test.go；" >&2
    echo "如果这是真实业务代码，请删除并使用 fisco.NewClient 真实客户端。" >&2
    exit 1
fi

echo "✅ Story 10.11 回退守卫：业务源码未发现免费链 mock 残留"
