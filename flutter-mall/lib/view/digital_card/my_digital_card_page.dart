import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/digital_card/digital_card_asset_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_detail_page.dart';
import 'package:flutter_mall/view/digital_card/digital_card_asset_tile.dart';

typedef DigitalCardAssetListFetcher
    = Future<QueryMyDigitalCardAssetListResponse> Function();

class MyDigitalCardPage extends StatefulWidget {
  final String? intentSource;
  final DigitalCardAssetListFetcher? fetchList;

  const MyDigitalCardPage({
    super.key,
    this.intentSource,
    this.fetchList,
  });

  @override
  State<MyDigitalCardPage> createState() => _MyDigitalCardPageState();
}

class _MyDigitalCardPageState extends State<MyDigitalCardPage> {
  List<DigitalCardAssetItem> _assets = <DigitalCardAssetItem>[];
  bool _isLoading = true;
  String? _errorMessage;

  AppRecentContext _buildRecoveryContext([String? source]) {
    return AppRecentContext.create(
      targetType: AppRecentTargetType.digitalCardAssetList,
      source: source ?? widget.intentSource ?? 'digital_card_asset_list',
      requiresAuth: true,
      fallbackType: AppRecentTargetType.home,
      fallbackTabIndex: 0,
    );
  }

  @override
  void initState() {
    super.initState();
    AppRecoveryStore.saveActiveIntentCandidate(_buildRecoveryContext());
    _loadAssets();
  }

  @override
  void dispose() {
    AppRecoveryStore.clearActiveIntentCandidateIfMatches(
      AppRecentTargetType.digitalCardAssetList,
    );
    super.dispose();
  }

  Future<void> _loadAssets() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final QueryMyDigitalCardAssetListResponse parsed =
          await (widget.fetchList?.call() ?? _fetchListFromApi());

      if (!mounted) {
        return;
      }

      setState(() {
        _assets = parsed.data.list;
        _isLoading = false;
      });
      await AppRecoveryStore.saveRecentContext(
        _buildRecoveryContext('digital_card_asset_view'),
      );
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = '加载我的数字卡片失败，请稍后重试';
        _assets = <DigitalCardAssetItem>[];
        _isLoading = false;
      });
    }
  }

  Future<QueryMyDigitalCardAssetListResponse> _fetchListFromApi() async {
    final Response response = await HttpUtil.get(
      queryMyDigitalCardAssetListUrl,
      queryParameters: <String, dynamic>{'pageNum': 1, 'pageSize': 50},
    );
    return queryMyDigitalCardAssetListResponseFromJson(
        jsonEncode(response.data));
  }

  Future<void> _openDetail(DigitalCardAssetItem item) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => DigitalCardAssetDetailPage(
          assetInstanceId: item.assetInstanceId,
          initialItem: item,
          intentSource: widget.intentSource ?? 'digital_card_asset_list',
        ),
      ),
    );
    if (!mounted) {
      return;
    }
    await _loadAssets();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('我的数字卡片'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return _buildStateCard(
        icon: Icons.wifi_tethering_error_rounded,
        title: '加载失败',
        description: _errorMessage!,
        actionLabel: '重新加载',
        onTap: _loadAssets,
      );
    }
    if (_assets.isEmpty) {
      return _buildStateCard(
        icon: Icons.style_outlined,
        title: '还没有数字卡片',
        description: '参与抽卡成功后，到账进度、发放状态和受限说明都会集中展示在这里。',
        actionLabel: '下拉刷新',
        onTap: _loadAssets,
      );
    }

    final int restrictedCount = _assets
        .where((DigitalCardAssetItem item) => item.hasRestriction)
        .length;
    return RefreshIndicator(
      onRefresh: _loadAssets,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
        children: <Widget>[
          _buildSummaryCard(restrictedCount),
          const SizedBox(height: AppSpacing.lg),
          ..._assets.map(
            (DigitalCardAssetItem item) => Padding(
              padding: const EdgeInsets.only(bottom: AppSpacing.md),
              child: DigitalCardAssetTile(
                item: item,
                onTap: () {
                  _openDetail(item);
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryCard(int restrictedCount) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: <Color>[Color(0xFF1F2937), Color(0xFFB4234B)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(AppRadii.xxl),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            '已收录 ${_assets.length} 张数字卡片',
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: Colors.white,
                  fontSize: 24,
                ),
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(
            restrictedCount > 0
                ? '$restrictedCount 张处于受限展示或合规复核中，详情页会告诉你当前原因与下一步。'
                : '到账进度、发放状态和合规提示都以服务端确认为准，不会在本地自行推断。',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: Colors.white.withValues(alpha: 0.82),
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildStateCard({
    required IconData icon,
    required String title,
    required String description,
    required String actionLabel,
    required Future<void> Function() onTap,
  }) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 52, color: AppColors.primaryDark),
            const SizedBox(height: AppSpacing.md),
            Text(
              title,
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              description,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: AppSpacing.lg),
            FilledButton(
              onPressed: () {
                onTap();
              },
              child: Text(actionLabel),
            ),
          ],
        ),
      ),
    );
  }
}
