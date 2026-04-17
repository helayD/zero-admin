import 'package:flutter/material.dart';
import 'package:flutter_mall/model/app_version_policy.dart';

class UpgradeGateDialog extends StatelessWidget {
  final AppVersionPolicy policy;
  final bool launching;
  final bool checking;
  final String? errorMessage;
  final VoidCallback onPrimaryAction;
  final VoidCallback? onSecondaryAction;

  const UpgradeGateDialog({
    super.key,
    required this.policy,
    required this.launching,
    required this.checking,
    required this.errorMessage,
    required this.onPrimaryAction,
    required this.onSecondaryAction,
  });

  String _title() {
    switch (policy.updateMode) {
      case UpgradeUpdateMode.force:
        return '升级后继续当前任务';
      case UpgradeUpdateMode.deadline:
        return '请尽快升级客户端';
      case UpgradeUpdateMode.recommended:
        return '发现更合适的新版本';
      case UpgradeUpdateMode.none:
        return '当前已是最新版本';
    }
  }

  String _summary() {
    if (policy.isUpToDate) {
      return '当前版本 ${policy.currentVersion} 已满足系统要求。';
    }
    if (policy.blocking) {
      return '当前版本 ${policy.currentVersion} 不满足目标能力要求，需升级到 ${policy.requiredDisplayVersion} 后才能继续。';
    }
    return '当前版本 ${policy.currentVersion} 建议升级到 ${policy.requiredDisplayVersion}，升级后可获得更完整的恢复与兼容能力。';
  }

  String _primaryLabel() {
    if (policy.isUpToDate) {
      return '我知道了';
    }
    return launching ? '正在打开升级入口...' : '立即升级';
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 520),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                _title(),
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF303133),
                ),
              ),
              const SizedBox(height: 12),
              Text(
                _summary(),
                style: const TextStyle(
                  fontSize: 14,
                  height: 1.5,
                  color: Color(0xFF606266),
                ),
              ),
              const SizedBox(height: 16),
              _InfoRow(
                  label: '受影响能力',
                  value: policy.affectedCapabilities.isEmpty
                      ? '未提供'
                      : policy.affectedCapabilities.join(' / ')),
              _InfoRow(
                  label: '阻断级别', value: policy.blocking ? '强制阻断' : '可稍后继续'),
              if (policy.effectiveAt.trim().isNotEmpty)
                _InfoRow(label: '生效时间', value: policy.effectiveAt),
              if (policy.deadlineAt.trim().isNotEmpty)
                _InfoRow(label: '最迟升级', value: policy.deadlineAt),
              if (policy.storeTarget.trim().isNotEmpty)
                _InfoRow(label: '升级入口', value: policy.storeTarget),
              if (policy.recoveryHint.trim().isNotEmpty) ...[
                const SizedBox(height: 12),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF4F8FF),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '恢复说明：${policy.recoveryHint}',
                    style: const TextStyle(
                      fontSize: 13,
                      height: 1.5,
                      color: Color(0xFF4E5969),
                    ),
                  ),
                ),
              ],
              if (policy.releaseNotesSummary.trim().isNotEmpty) ...[
                const SizedBox(height: 12),
                Text(
                  policy.releaseNotesSummary,
                  style: const TextStyle(
                    fontSize: 13,
                    height: 1.5,
                    color: Color(0xFF909399),
                  ),
                ),
              ],
              if (errorMessage != null && errorMessage!.trim().isNotEmpty) ...[
                const SizedBox(height: 12),
                Text(
                  errorMessage!,
                  style: const TextStyle(
                    fontSize: 13,
                    color: Color(0xFFE74C3C),
                  ),
                ),
              ],
              const SizedBox(height: 18),
              if (checking)
                const Padding(
                  padding: EdgeInsets.only(bottom: 12),
                  child: Row(
                    children: [
                      SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                      SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          '正在校验升级状态并准备恢复任务...',
                          style: TextStyle(
                            fontSize: 13,
                            color: Color(0xFF909399),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              Row(
                children: [
                  if (onSecondaryAction != null)
                    Expanded(
                      child: OutlinedButton(
                        onPressed: checking ? null : onSecondaryAction,
                        child: Text(policy.isUpToDate ? '关闭' : '稍后继续'),
                      ),
                    ),
                  if (onSecondaryAction != null) const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed:
                          (checking || launching) ? null : onPrimaryAction,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFA436A),
                        foregroundColor: Colors.white,
                      ),
                      child: Text(_primaryLabel()),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;

  const _InfoRow({
    required this.label,
    required this.value,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: RichText(
        text: TextSpan(
          style: const TextStyle(
            fontSize: 13,
            height: 1.4,
            color: Color(0xFF606266),
          ),
          children: [
            TextSpan(
              text: '$label：',
              style: const TextStyle(
                fontWeight: FontWeight.w600,
                color: Color(0xFF303133),
              ),
            ),
            TextSpan(text: value),
          ],
        ),
      ),
    );
  }
}
