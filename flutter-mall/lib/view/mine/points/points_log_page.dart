import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/member/points_log_model.dart';
import 'package:flutter_mall/utils/http_util.dart';

class PointsLogPage extends StatefulWidget {
  final int currentPoints;

  const PointsLogPage({super.key, this.currentPoints = 0});

  @override
  State<PointsLogPage> createState() => _PointsLogPageState();
}

class _PointsLogPageState extends State<PointsLogPage>
    with SingleTickerProviderStateMixin {
  static const Color _bg = Color(0xFFF7F2EA);
  static const Color _surface = Color(0xFFFFFCF7);
  static const Color _ink = Color(0xFF151821);
  static const Color _muted = Color(0xFF746D63);
  static const Color _gold = Color(0xFFC78A24);
  static const Color _goldLight = Color(0xFFFFE6A6);
  static const Color _line = Color(0xFFE8DDCC);
  static const Color _green = Color(0xFF0F766E);
  static const Color _red = Color(0xFFDC2626);

  late final TabController _tabController;

  // 每个 tab 的状态：0=全部, 1=增加, 2=减少
  final List<int> _tabs = const [0, 1, 2];
  final List<String> _tabLabels = const ['全部', '增加', '减少'];

  final List<List<PointsLogItem>> _items = [[], [], []];
  final List<int> _totals = [0, 0, 0];
  final List<int> _pages = [1, 1, 1];
  final List<bool> _loading = [false, false, false];
  final List<bool> _hasMore = [true, true, true];
  final List<ScrollController> _scrollControllers = [
    ScrollController(),
    ScrollController(),
    ScrollController(),
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: _tabs.length, vsync: this);
    _tabController.addListener(_onTabChanged);
    for (final sc in _scrollControllers) {
      sc.addListener(_onScroll);
    }
    _loadPage(0);
  }

  @override
  void dispose() {
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    for (final sc in _scrollControllers) {
      sc.dispose();
    }
    super.dispose();
  }

  void _onTabChanged() {
    if (_tabController.indexIsChanging) return;
    final idx = _tabController.index;
    if (_items[idx].isEmpty && !_loading[idx]) {
      _loadPage(idx);
    }
  }

  void _onScroll() {
    final idx = _tabController.index;
    final sc = _scrollControllers[idx];
    if (sc.position.pixels >= sc.position.maxScrollExtent - 120 &&
        !_loading[idx] &&
        _hasMore[idx]) {
      _loadPage(idx);
    }
  }

  Future<void> _loadPage(int tabIdx) async {
    if (_loading[tabIdx] || !_hasMore[tabIdx]) return;
    setState(() => _loading[tabIdx] = true);
    try {
      final Response resp = await HttpUtil.get(
        pointsLogListUrl,
        queryParameters: <String, dynamic>{
          'pageNum': _pages[tabIdx],
          'pageSize': 20,
          'changeType': _tabs[tabIdx],
        },
      );
      final model = PointsLogListResponse.fromJson(
        Map<String, dynamic>.from(resp.data as Map),
      );
      if (!mounted) return;
      setState(() {
        _totals[tabIdx] = model.total.toInt();
        _items[tabIdx].addAll(model.list);
        _pages[tabIdx]++;
        _hasMore[tabIdx] = _items[tabIdx].length < model.total;
        _loading[tabIdx] = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading[tabIdx] = false);
    }
  }

  Future<void> _refresh(int tabIdx) async {
    setState(() {
      _items[tabIdx].clear();
      _pages[tabIdx] = 1;
      _hasMore[tabIdx] = true;
    });
    await _loadPage(tabIdx);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _bg,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        scrolledUnderElevation: 0,
        foregroundColor: _ink,
        title: const Text(
          '积分明细',
          style: TextStyle(
            color: _ink,
            fontSize: 18,
            fontWeight: FontWeight.w800,
          ),
        ),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(96),
          child: Column(
            children: [
              _buildBalanceBanner(),
              TabBar(
                controller: _tabController,
                labelColor: _gold,
                unselectedLabelColor: _muted,
                indicatorColor: _gold,
                indicatorWeight: 2.5,
                labelStyle: const TextStyle(
                  fontWeight: FontWeight.w800,
                  fontSize: 14,
                ),
                tabs: _tabLabels.map((l) => Tab(text: l)).toList(),
              ),
            ],
          ),
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: List.generate(
          _tabs.length,
          (i) => _buildTabContent(i),
        ),
      ),
    );
  }

  Widget _buildBalanceBanner() {
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 0, 16, 8),
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 12),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF10131B), Color(0xFF2A2130)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(18),
      ),
      child: Row(
        children: [
          const Icon(Icons.stars_rounded, color: _gold, size: 26),
          const SizedBox(width: 10),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                '当前积分',
                style: TextStyle(
                  color: Colors.white54,
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                ),
              ),
              Text(
                widget.currentPoints.toString(),
                style: const TextStyle(
                  color: _goldLight,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                  height: 1.1,
                ),
              ),
            ],
          ),
          const Spacer(),
          Text(
            '共 ${_totals[0]} 条记录',
            style: const TextStyle(
              color: Colors.white38,
              fontSize: 12,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTabContent(int tabIdx) {
    final items = _items[tabIdx];
    final loading = _loading[tabIdx];

    if (items.isEmpty && loading) {
      return _buildSkeletonList();
    }
    if (items.isEmpty && !loading) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.stars_outlined, color: _muted.withValues(alpha: 0.4), size: 52),
            const SizedBox(height: 12),
            Text(
              '暂无积分记录',
              style: TextStyle(color: _muted.withValues(alpha: 0.7), fontSize: 14),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      color: _gold,
      onRefresh: () => _refresh(tabIdx),
      child: ListView.builder(
        controller: _scrollControllers[tabIdx],
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
        itemCount: items.length + 1,
        itemBuilder: (ctx, i) {
          if (i == items.length) {
            return _buildLoadMore(tabIdx);
          }
          return _buildLogTile(items[i]);
        },
      ),
    );
  }

  Widget _buildLogTile(PointsLogItem item) {
    final bool isAdd = item.isAdd;
    final Color amountColor = isAdd ? _green : _red;
    final String sign = isAdd ? '+' : '-';

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: _surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _line),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: isAdd
                  ? const Color(0xFFE9F9F0)
                  : const Color(0xFFFEF2F2),
              borderRadius: BorderRadius.circular(13),
            ),
            child: Icon(
              isAdd ? Icons.add_rounded : Icons.remove_rounded,
              color: amountColor,
              size: 22,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.description.trim().isEmpty
                      ? '${item.sourceLabel}积分变动'
                      : item.description,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    height: 1.35,
                  ),
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 7, vertical: 2),
                      decoration: BoxDecoration(
                        color: _goldLight.withValues(alpha: 0.6),
                        borderRadius: BorderRadius.circular(999),
                      ),
                      child: Text(
                        item.sourceLabel,
                        style: const TextStyle(
                          color: _gold,
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      item.createTime,
                      style:
                          const TextStyle(fontSize: 12, color: _muted),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Text(
            '$sign${item.changePoints}',
            style: TextStyle(
              color: amountColor,
              fontSize: 18,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSkeletonList() {
    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 32),
      itemCount: 6,
      itemBuilder: (_, __) => Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: _surface,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: _line),
        ),
        child: Row(
          children: [
            _shimmerBox(width: 40, height: 40, radius: 13),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _shimmerBox(width: double.infinity, height: 14, radius: 6),
                  const SizedBox(height: 8),
                  _shimmerBox(width: 120, height: 12, radius: 6),
                ],
              ),
            ),
            const SizedBox(width: 12),
            _shimmerBox(width: 36, height: 20, radius: 6),
          ],
        ),
      ),
    );
  }

  Widget _shimmerBox(
      {required double width, required double height, required double radius}) {
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0.3, end: 0.9),
      duration: const Duration(milliseconds: 900),
      curve: Curves.easeInOut,
      builder: (_, v, __) => Container(
        width: width,
        height: height,
        decoration: BoxDecoration(
          color: _line.withValues(alpha: v),
          borderRadius: BorderRadius.circular(radius),
        ),
      ),
      onEnd: () => setState(() {}),
    );
  }

  Widget _buildLoadMore(int tabIdx) {
    if (!_hasMore[tabIdx]) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 18),
        child: Center(
          child: Text(
            '已显示全部记录',
            style: TextStyle(color: _muted.withValues(alpha: 0.6), fontSize: 12),
          ),
        ),
      );
    }
    if (_loading[tabIdx]) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 18),
        child: Center(child: CircularProgressIndicator(color: _gold)),
      );
    }
    return const SizedBox.shrink();
  }
}
