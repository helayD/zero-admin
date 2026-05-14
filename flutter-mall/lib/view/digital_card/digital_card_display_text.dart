import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';

class DigitalCardStatusCopy {
  final String label;
  final String description;
  final String actionHint;

  const DigitalCardStatusCopy({
    required this.label,
    required this.description,
    this.actionHint = '',
  });
}

String digitalCardUserFacingText(String text) {
  String value = text.trim();
  if (value.isEmpty) {
    return '';
  }

  final Map<String, String> replacements = <String, String>{
    'FISCO BCOS': '服务',
    'AntChain': '服务',
    'FISCO': '服务',
    '蚂蚁链': '服务',
    '区块链': '服务',
    '链上处理中': '到账中',
    '链上回执': '处理结果',
    '链上状态': '发放状态',
    '链路任务': '处理任务',
    '上链': '发放',
    '链上': '',
    '链路': '流程',
    'token': '编号',
    'Token': '编号',
    'TOKEN': '编号',
  };

  replacements.forEach((String source, String target) {
    value = value.replaceAll(source, target);
  });

  return value
      .replaceAll(RegExp(r'\s+'), ' ')
      .replaceAll('，，', '，')
      .replaceAll('。。', '。')
      .replaceAll('，。', '。')
      .trim();
}

/// 将时间线 operationType / statusText 等 snake_case 状态码转为用户可读中文。
/// 优先精确匹配，未命中时走 [digitalCardUserFacingText] 做通用脱敏。
String digitalCardOperationLabel(String operationType) {
  const Map<String, String> _opLabels = <String, String>{
    // 资产创建
    'asset_created': '卡片已创建',
    'asset_created_from_order': '下单获得卡片',
    // 发放流程
    'mint_requested': '发放申请已提交',
    'mint_dispatching': '发放处理中',
    'mint_succeeded': '卡片已到账',
    'mint_failed': '发放失败',
    'mint_frozen': '发放暂停',
    'mint_manual_review': '人工复核中',
    'mint_retry_requested': '重新发放中',
    'mint_compensating': '发放补偿中',
    // 合规 / 展示
    'asset_compliance_review': '合规复核中',
    'asset_offlined': '卡片已下线',
    // 转移
    'asset_transferred': '卡片已转移',
    'holder_transferred': '持有人已变更',
  };

  final String key = operationType.trim().toLowerCase();
  if (_opLabels.containsKey(key)) {
    return _opLabels[key]!;
  }
  // 兜底：通用脱敏替换
  return digitalCardUserFacingText(operationType);
}

/// 将时间线 statusText（to_status 等）转为用户可读中文。
String digitalCardStatusLabel(String statusText) {
  const Map<String, String> _statusLabels = <String, String>{
    'asset_created': '已创建',
    'asset_processing': '处理中',
    'asset_success': '已到账',
    'asset_pending': '待到账',
    'asset_failed': '发放失败',
    'mint_pending': '待发放',
    'mint_processing': '发放中',
    'mint_success': '已到账',
    'mint_failed': '发放失败',
    'mint_frozen': '已暂停',
    'mint_compensating': '补偿中',
    'mint_manual_review': '人工复核',
    'compliance_clear': '合规通过',
    'compliance_review': '合规复核中',
    'compliance_restricted': '受限展示',
    'compliance_recycled': '已回收',
    'display_visible': '可展示',
    'display_hidden': '受限展示',
    'display_offlined': '已下线',
    'display_recycled': '已回收',
  };

  final String key = statusText.trim().toLowerCase();
  if (_statusLabels.containsKey(key)) {
    return _statusLabels[key]!;
  }
  return digitalCardUserFacingText(statusText);
}

String digitalCardMintStatusText(String status, String statusText) {
  switch (status) {
    case 'mint_success':
      return '已到账';
    case 'mint_processing':
      return '到账中';
    case 'mint_failed':
      return '发放失败';
    case 'mint_pending':
      return '待发放';
    default:
      final String label = digitalCardUserFacingText(statusText);
      return label.isEmpty ? digitalCardUserFacingText(status) : label;
  }
}

String digitalCardDisplayStatusText(String status, String statusText) {
  switch (status) {
    case 'display_visible':
      return '可展示';
    case 'display_hidden':
      return '受限展示';
    case 'display_offlined':
      return '已下线';
    case 'display_recycled':
      return '已回收';
    default:
      final String label = digitalCardUserFacingText(statusText);
      return label.isEmpty ? digitalCardUserFacingText(status) : label;
  }
}

String digitalCardAssetStatusText(String status, String statusText) {
  switch (status) {
    case 'asset_created':
    case 'asset_processing':
      return '到账中';
    case 'asset_success':
      return '已到账';
    case 'asset_pending':
      return '待到账';
    case 'asset_failed':
      return '发放失败';
    default:
      final String label = digitalCardUserFacingText(statusText);
      return label.isEmpty ? digitalCardUserFacingText(status) : label;
  }
}

DigitalCardStatusCopy digitalCardAssetPrimaryCopy(
  DigitalCardAssetItem item, {
  String latestStatusSummary = '',
}) {
  final String summary = digitalCardUserFacingText(
    latestStatusSummary.trim().isNotEmpty
        ? latestStatusSummary
        : item.complianceRuleSummary,
  );
  String usefulSummary(String label) {
    if (summary.isEmpty || summary == label || summary == '人工复核中') {
      return '';
    }
    return summary;
  }

  if (item.complianceStatus == 'compliance_recycled' ||
      item.displayStatus == 'display_recycled') {
    return const DigitalCardStatusCopy(
      label: '已回收',
      description: '这张卡片已完成回收，当前不可继续展示。',
      actionHint: '如有疑问，请联系平台客服。',
    );
  }

  if (item.displayStatus == 'display_offlined') {
    return const DigitalCardStatusCopy(
      label: '已下线',
      description: '这张卡片当前已停止展示，详情以平台通知为准。',
      actionHint: '后续状态变化会自动同步。',
    );
  }

  if (item.complianceStatus == 'compliance_review') {
    final String description = usefulSummary('合规复核中');
    return DigitalCardStatusCopy(
      label: '合规复核中',
      description: description.isEmpty ? '卡片正在复核，结果更新后会自动同步。' : description,
      actionHint: '无需重复操作，稍后刷新查看结果。',
    );
  }

  if (item.complianceStatus == 'compliance_restricted' ||
      item.displayStatus == 'display_hidden') {
    return DigitalCardStatusCopy(
      label: '受限展示',
      description: usefulSummary('受限展示').isEmpty
          ? '当前卡片暂不公开展示，详情以合规说明为准。'
          : usefulSummary('受限展示'),
      actionHint: '可进入详情页查看原因与最新进展。',
    );
  }

  if (item.mintStatus == 'mint_failed') {
    return const DigitalCardStatusCopy(
      label: '发放失败',
      description: '卡片暂未到账，系统已记录本次失败结果。',
      actionHint: '请稍后刷新，仍未恢复时联系平台客服。',
    );
  }

  if (item.mintStatus == 'mint_processing') {
    return const DigitalCardStatusCopy(
      label: '到账中',
      description: '卡片正在到账，通常稍后会自动更新。',
      actionHint: '无需操作，稍后刷新即可。',
    );
  }

  if (item.mintStatus == 'mint_pending') {
    return const DigitalCardStatusCopy(
      label: '待到账',
      description: '卡片已进入发放队列，到账后会自动出现在卡包中。',
      actionHint: '请稍后刷新查看。',
    );
  }

  if (item.mintStatus == 'mint_success') {
    return const DigitalCardStatusCopy(
      label: '已到账',
      description: '卡片已到账，可以在卡包中查看和展示。',
    );
  }

  final String label = digitalCardMintStatusText(
    item.mintStatus,
    item.mintStatusText,
  );
  return DigitalCardStatusCopy(
    label: label.isEmpty ? '状态待同步' : label,
    description: summary.isEmpty ? '当前状态以服务端确认为准。' : summary,
    actionHint: '下拉刷新可获取最新状态。',
  );
}

DigitalCardStatusCopy digitalCardDrawResultCopy(DrawMemberRecord record) {
  final String label = digitalCardAssetStatusText(
    record.assetStatus,
    record.assetStatusText,
  );

  switch (record.assetStatus) {
    case 'asset_created':
    case 'asset_processing':
      return const DigitalCardStatusCopy(
        label: '到账中',
        description: '提货卡正在到账，稍后会同步到我的提货卡。',
        actionHint: '可以先关闭弹窗，稍后在卡包查看。',
      );
    case 'asset_success':
      return const DigitalCardStatusCopy(
        label: '已到账',
        description: '提货卡已到账，可以前往我的提货卡查看。',
      );
    case 'asset_failed':
      return const DigitalCardStatusCopy(
        label: '发放失败',
        description: '提货卡暂未到账，系统已记录本次失败结果。',
        actionHint: '请稍后刷新，仍未恢复时联系平台客服。',
      );
    default:
      return DigitalCardStatusCopy(
        label: label.isEmpty ? '已中奖待到账' : label,
        description: label.isEmpty
            ? '提货卡正在准备到账，稍后可在我的提货卡查看。'
            : '当前状态为 $label，稍后可在我的提货卡查看。',
      );
  }
}
