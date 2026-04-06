import 'package:flutter/material.dart';
import 'package:flutter_mall/model/permission_flow_context.dart';

enum PermissionPromptAction {
  request,
  openSettings,
  alternative,
  dismiss,
}

class PermissionPromptSheet extends StatelessWidget {
  final PermissionFlowContext flow;
  final PermissionFlowStatus status;
  final String? alternativeLabel;
  final String? detailOverride;
  final PermissionPromptAction? primaryActionOverride;
  final String? primaryLabelOverride;

  const PermissionPromptSheet({
    super.key,
    required this.flow,
    required this.status,
    this.alternativeLabel,
    this.detailOverride,
    this.primaryActionOverride,
    this.primaryLabelOverride,
  });

  static Future<PermissionPromptAction?> show(
    BuildContext context, {
    required PermissionFlowContext flow,
    required PermissionFlowStatus status,
    String? alternativeLabel,
    String? detailOverride,
    PermissionPromptAction? primaryActionOverride,
    String? primaryLabelOverride,
  }) {
    return showModalBottomSheet<PermissionPromptAction>(
      context: context,
      isScrollControlled: true,
      builder: (_) => PermissionPromptSheet(
        flow: flow,
        status: status,
        alternativeLabel: alternativeLabel,
        detailOverride: detailOverride,
        primaryActionOverride: primaryActionOverride,
        primaryLabelOverride: primaryLabelOverride,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final primaryLabel = _primaryLabel();
    final title = _title();
    final detail = detailOverride ?? _detail();

    return SafeArea(
      child: Semantics(
        label: '$title。$detail',
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 28),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 12),
              Text(
                detail,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: const Color(0xFF606266),
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 20),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => Navigator.of(context).pop(_primaryAction()),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFA436A),
                    foregroundColor: Colors.white,
                    minimumSize: const Size.fromHeight(46),
                  ),
                  child: Text(primaryLabel),
                ),
              ),
              if (alternativeLabel != null && alternativeLabel!.isNotEmpty) ...[
                const SizedBox(height: 10),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    onPressed: () => Navigator.of(context)
                        .pop(PermissionPromptAction.alternative),
                    style: OutlinedButton.styleFrom(
                      minimumSize: const Size.fromHeight(46),
                    ),
                    child: Text(alternativeLabel!),
                  ),
                ),
              ],
              const SizedBox(height: 8),
              SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: () =>
                      Navigator.of(context).pop(PermissionPromptAction.dismiss),
                  child: const Text('稍后再说'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  PermissionPromptAction _primaryAction() {
    if (primaryActionOverride != null) {
      return primaryActionOverride!;
    }
    if (status == PermissionFlowStatus.permanentlyDenied) {
      return PermissionPromptAction.openSettings;
    }
    return PermissionPromptAction.request;
  }

  String _primaryLabel() {
    if (primaryLabelOverride != null && primaryLabelOverride!.isNotEmpty) {
      return primaryLabelOverride!;
    }
    if (status == PermissionFlowStatus.permanentlyDenied) {
      return '去系统设置';
    }
    return '继续授权';
  }

  String _title() {
    final sceneLabel = _sceneLabel();
    switch (status) {
      case PermissionFlowStatus.initial:
        return '为$sceneLabel开启所需权限';
      case PermissionFlowStatus.denied:
        return '$sceneLabel需要权限才能补充该能力';
      case PermissionFlowStatus.permanentlyDenied:
        return '$sceneLabel权限已被系统关闭';
      case PermissionFlowStatus.granted:
      case PermissionFlowStatus.limited:
      case PermissionFlowStatus.notRequired:
        return '$sceneLabel已可继续';
      case PermissionFlowStatus.settingsReturnPending:
        return '返回后继续$sceneLabel';
      case PermissionFlowStatus.pickerLostData:
        return '$sceneLabel正在恢复上次选择的内容';
    }
  }

  String _detail() {
    switch (flow.scene) {
      case PermissionScene.notificationSubscription:
        if (status == PermissionFlowStatus.permanentlyDenied ||
            status == PermissionFlowStatus.denied) {
          return '通知权限未开启时，你仍可继续购物，并通过站内消息中心和未读角标查看提醒。';
        }
        return '仅在你主动开启消息提醒时申请通知权限，不会在启动或浏览阶段提前打断你。';
      case PermissionScene.commentImage:
        if (status == PermissionFlowStatus.permanentlyDenied ||
            status == PermissionFlowStatus.denied) {
          return '拒绝后不会丢失评分、评价文本和已选图片，你仍可继续无图评价，或改用另一种补图方式。';
        }
        return '仅在你主动补充评价图片时请求相册或相机权限，不会影响继续提交文字评价。';
      case PermissionScene.afterSalesProof:
        if (status == PermissionFlowStatus.permanentlyDenied ||
            status == PermissionFlowStatus.denied) {
          return '拒绝后不会丢失售后类型、原因和问题描述，你仍可继续无图提交，或改用另一种凭证入口。';
        }
        return '仅在你主动补充售后凭证时请求相册或相机权限，不会阻断当前售后流程。';
    }
  }

  String _sceneLabel() {
    switch (flow.scene) {
      case PermissionScene.notificationSubscription:
        return '消息提醒';
      case PermissionScene.commentImage:
        return flow.source == PermissionFlowSource.camera ? '拍照补图' : '相册补图';
      case PermissionScene.afterSalesProof:
        return flow.source == PermissionFlowSource.camera ? '拍照上传凭证' : '相册上传凭证';
    }
  }
}
