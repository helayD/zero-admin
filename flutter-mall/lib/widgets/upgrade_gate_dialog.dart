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
        return '建议近期升级';
      case UpgradeUpdateMode.recommended:
        return '发现更合适的新版本';
      case UpgradeUpdateMode.none:
        return '当前已是最新版本';
    }
  }

  String _summary() {
    if (policy.isUpToDate) {
      return '当前版本可继续使用。';
    }
    if (policy.blocking) {
      return '当前任务需要新版能力，升级后会自动回到刚才的位置。';
    }
    return '新版更稳定，当前页面可以稍后继续。';
  }

  String _primaryLabel() {
    if (policy.isUpToDate) {
      return '我知道了';
    }
    return launching ? '正在打开' : '立即升级';
  }

  String _versionText() {
    final current = policy.currentVersion.trim();
    final target = policy.requiredDisplayVersion.trim();
    if (current.isNotEmpty && target.isNotEmpty && current != target) {
      return '当前 $current  升至 $target';
    }
    if (target.isNotEmpty) {
      return '目标版本 $target';
    }
    if (current.isNotEmpty) {
      return '当前版本 $current';
    }
    return '版本状态已确认';
  }

  String _statusText() {
    if (policy.isUpToDate) {
      return '已满足';
    }
    return policy.blocking ? '需升级' : '可稍后';
  }

  Color _statusColor() {
    if (policy.isUpToDate) {
      return const Color(0xFF12B76A);
    }
    return policy.blocking ? const Color(0xFFE74C3C) : const Color(0xFFFA436A);
  }

  String? _deadlineText() {
    final raw = policy.deadlineAt.trim();
    if (raw.isEmpty || policy.blocking || policy.isUpToDate) {
      return null;
    }
    final parsed = DateTime.tryParse(raw);
    if (parsed == null) {
      final dateOnly = raw.split('T').first.trim();
      return dateOnly.isEmpty ? null : '建议 $dateOnly 前完成';
    }
    final local = parsed.toLocal();
    return '建议 ${local.year}年${local.month}月${local.day}日前完成';
  }

  List<_DialogPoint> _points() {
    if (policy.isUpToDate) {
      return const <_DialogPoint>[];
    }

    final deadline = _deadlineText();
    return <_DialogPoint>[
      _DialogPoint(
        icon: policy.blocking
            ? Icons.lock_open_rounded
            : Icons.verified_user_rounded,
        text: policy.blocking ? '升级后恢复当前任务' : '提升稳定性与兼容性',
      ),
      if (deadline != null)
        _DialogPoint(icon: Icons.event_available_rounded, text: deadline)
      else
        const _DialogPoint(
          icon: Icons.replay_rounded,
          text: '升级后保留恢复入口',
        ),
    ];
  }

  @override
  Widget build(BuildContext context) {
    final statusColor = _statusColor();
    final points = _points();

    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 26, vertical: 24),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 420),
        child: Container(
          padding: const EdgeInsets.fromLTRB(22, 24, 22, 18),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            boxShadow: const [
              BoxShadow(
                color: Color(0x26000000),
                blurRadius: 28,
                offset: Offset(0, 14),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFEDF2),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: const Icon(
                      Icons.system_update_alt_rounded,
                      color: Color(0xFFFA436A),
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _title(),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF303133),
                          ),
                        ),
                        const SizedBox(height: 6),
                        Text(
                          _versionText(),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF909399),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 10),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 9,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withAlpha(26),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      _statusText(),
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: statusColor,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              Text(
                _summary(),
                style: const TextStyle(
                  fontSize: 14,
                  height: 1.45,
                  color: Color(0xFF606266),
                ),
              ),
              if (points.isNotEmpty) ...[
                const SizedBox(height: 16),
                const Divider(height: 1, color: Color(0xFFEDEFF2)),
                const SizedBox(height: 14),
                for (final point in points) ...[
                  _PointRow(point: point),
                  if (point != points.last) const SizedBox(height: 10),
                ],
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
              const SizedBox(height: 20),
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
                          '正在确认升级状态...',
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
                        style: OutlinedButton.styleFrom(
                          minimumSize: const Size.fromHeight(48),
                          foregroundColor: const Color(0xFF606266),
                          side: const BorderSide(color: Color(0xFFD8DDE6)),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(14),
                          ),
                        ),
                        child: Text(policy.isUpToDate ? '关闭' : '稍后继续'),
                      ),
                    ),
                  if (onSecondaryAction != null) const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed:
                          (checking || launching) ? null : onPrimaryAction,
                      style: ElevatedButton.styleFrom(
                        minimumSize: const Size.fromHeight(48),
                        backgroundColor: const Color(0xFFFA436A),
                        foregroundColor: Colors.white,
                        elevation: 0,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14),
                        ),
                      ),
                      icon: Icon(
                        policy.isUpToDate
                            ? Icons.check_rounded
                            : Icons.open_in_new_rounded,
                        size: 18,
                      ),
                      label: Text(_primaryLabel()),
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

class _DialogPoint {
  final IconData icon;
  final String text;

  const _DialogPoint({
    required this.icon,
    required this.text,
  });
}

class _PointRow extends StatelessWidget {
  final _DialogPoint point;

  const _PointRow({required this.point});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(
          point.icon,
          size: 18,
          color: const Color(0xFFFA436A),
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Text(
            point.text,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              fontSize: 13,
              height: 1.35,
              color: Color(0xFF606266),
            ),
          ),
        ),
      ],
    );
  }
}
