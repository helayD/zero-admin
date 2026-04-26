import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_mall/model/digital_card/physical_fulfillment_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/physical_fulfillment_address_sheet.dart';
import 'package:flutter_mall/view/digital_card/physical_fulfillment_status_badge.dart';
import 'package:flutter_mall/view/digital_card/physical_fulfillment_timeline.dart';

typedef PhysicalFulfillmentDetailFetcher
    = Future<QueryMyPhysicalFulfillmentDetailResponse> Function(
        int assetInstanceId);

class DigitalCardPhysicalFulfillmentPage extends StatefulWidget {
  final int assetInstanceId;
  final PhysicalFulfillmentDetailFetcher? fetchDetail;

  const DigitalCardPhysicalFulfillmentPage({
    super.key,
    required this.assetInstanceId,
    this.fetchDetail,
  });

  @override
  State<DigitalCardPhysicalFulfillmentPage> createState() =>
      _DigitalCardPhysicalFulfillmentPageState();
}

class _DigitalCardPhysicalFulfillmentPageState
    extends State<DigitalCardPhysicalFulfillmentPage> {
  PhysicalFulfillmentDetailData? _detail;
  bool _isLoading = true;
  bool _isSubmitting = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadDetail();
  }

  Future<void> _loadDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final QueryMyPhysicalFulfillmentDetailResponse parsed =
          await (widget.fetchDetail?.call(widget.assetInstanceId) ??
              _fetchDetailFromApi());
      if (!mounted) return;
      setState(() {
        _detail = parsed.data;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _errorMessage = '加载实体卡进度失败，请稍后重试';
        _isLoading = false;
      });
    }
  }

  Future<QueryMyPhysicalFulfillmentDetailResponse> _fetchDetailFromApi() async {
    final Response response = await HttpUtil.get(
      queryMyPhysicalFulfillmentDetailUrl,
      queryParameters: <String, dynamic>{
        'assetInstanceId': widget.assetInstanceId,
      },
    );
    return queryMyPhysicalFulfillmentDetailResponseFromJson(
      jsonEncode(response.data),
    );
  }

  Future<void> _confirmReceipt() async {
    final PhysicalFulfillmentDetailData? detail = _detail;
    if (detail == null || _isSubmitting) return;
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: const Text('确认签收'),
        content: const Text('确认已收到实体卡吗？'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('确认'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;
    setState(() => _isSubmitting = true);
    try {
      await HttpUtil.post(
        confirmPhysicalCardReceiptUrl,
        data: <String, dynamic>{
          'fulfillmentId': detail.fulfillmentId,
          'reason': '会员确认签收',
        },
      );
      await _loadDetail();
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  Future<void> _openAddressSheet() async {
    if (_isSubmitting) return;
    setState(() => _isSubmitting = true);
    try {
      final Response response = await HttpUtil.get(addressListDataUrl);
      final AddressListModel model = AddressListModel.fromJson(response.data);
      if (!mounted) return;
      await showModalBottomSheet<void>(
        context: context,
        showDragHandle: true,
        builder: (BuildContext context) => PhysicalFulfillmentAddressSheet(
          addresses: model.data,
          selectedAddressId: null,
          onSelected: (AddressListData address) async {
            Navigator.pop(context);
            await HttpUtil.post(
              confirmPhysicalFulfillmentAddressUrl,
              data: <String, dynamic>{
                'assetInstanceId': widget.assetInstanceId,
                'addressId': address.id,
              },
            );
            await _loadDetail();
          },
        ),
      );
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(title: const Text('实体卡进度')),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading && _detail == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null && _detail == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.local_shipping_outlined, size: 52),
              const SizedBox(height: AppSpacing.md),
              Text(_errorMessage!, textAlign: TextAlign.center),
              const SizedBox(height: AppSpacing.lg),
              FilledButton(
                onPressed: _loadDetail,
                child: const Text('重新加载'),
              ),
            ],
          ),
        ),
      );
    }

    final PhysicalFulfillmentDetailData? detail = _detail;
    if (detail == null) {
      return const SizedBox.shrink();
    }

    return RefreshIndicator(
      onRefresh: _loadDetail,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: <Widget>[
          _buildSummary(detail),
          const SizedBox(height: AppSpacing.lg),
          _buildAddress(detail),
          const SizedBox(height: AppSpacing.lg),
          _buildLogistics(detail),
          const SizedBox(height: AppSpacing.lg),
          Text('履约时间线', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.md),
          PhysicalFulfillmentTimeline(timeline: detail.timeline),
        ],
      ),
    );
  }

  Widget _buildSummary(PhysicalFulfillmentDetailData detail) {
    final String statusText = detail.isBlocked
        ? detail.blockedReasonText
        : detail.fulfillmentStatusText;
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  detail.templateName.trim().isEmpty
                      ? '实体卡'
                      : detail.templateName,
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ),
              PhysicalFulfillmentStatusBadge(
                status: detail.fulfillmentStatus,
                text: statusText,
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          _buildMetaLine('卡片编号', detail.assetNo),
          _buildMetaLine('活动名称', detail.activityName),
          _buildMetaLine('获取时间', detail.obtainedAt),
          _buildMetaLine('发放状态', detail.mintStatusText),
          if (detail.complianceTipSummary.trim().isNotEmpty)
            _buildMetaLine('合规提示', detail.complianceTipSummary),
        ],
      ),
    );
  }

  Widget _buildAddress(PhysicalFulfillmentDetailData detail) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text('收货信息', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.md),
          _buildMetaLine('收件人', detail.receiverNameMasked),
          _buildMetaLine('手机号', detail.receiverPhoneMasked),
          _buildMetaLine('地址', detail.addressSummary),
          if (detail.needsAddress) ...<Widget>[
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _isSubmitting ? null : _openAddressSheet,
                icon: const Icon(Icons.location_on_outlined),
                label: const Text('确认收货地址'),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildLogistics(PhysicalFulfillmentDetailData detail) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(AppRadii.xl),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text('制作与配送', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.md),
          _buildMetaLine('制作状态', detail.productionStatusText),
          _buildMetaLine('配送状态', detail.shippingStatusText),
          _buildMetaLine('物流公司', detail.carrierName),
          _buildMetaLine('物流单号', detail.trackingNo),
          if (detail.canConfirmReceipt) ...<Widget>[
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _isSubmitting ? null : _confirmReceipt,
                icon: const Icon(Icons.check_circle_outline),
                label: const Text('确认已收到'),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildMetaLine(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 84,
            child: Text(label, style: Theme.of(context).textTheme.bodySmall),
          ),
          Expanded(
            child: Text(
              value.trim().isEmpty ? '-' : value.trim(),
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: AppColors.textPrimary,
                  ),
            ),
          ),
        ],
      ),
    );
  }
}
