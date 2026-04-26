import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/constant_param.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/member_info.dart';
import 'package:flutter_mall/model/message_model.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/address/address_list.dart';
import 'package:flutter_mall/view/mine/collection/collection.dart';
import 'package:flutter_mall/view/mine/coupon/coupon_list.dart';
import 'package:flutter_mall/view/mine/focus/focus.dart';
import 'package:flutter_mall/view/mine/history/history.dart';
import 'package:flutter_mall/view/mine/login/login.dart';
import 'package:flutter_mall/view/mine/message/message.dart';
import 'package:flutter_mall/view/mine/order/order_list.dart';
import 'package:flutter_mall/view/mine/ping_jia/ping_jia.dart';
import 'package:flutter_mall/view/mine/profile/profile_edit.dart';
import 'package:flutter_mall/view/mine/setting/settings.dart';
import 'package:flutter_mall/view/digital_card/my_digital_card_page.dart';
import 'package:shared_preferences/shared_preferences.dart';

///
/// 我的页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Mine extends StatefulWidget {
  const Mine({super.key});

  @override
  State<Mine> createState() => _MineState();
}

class _MineState extends State<Mine> {
  bool _isLoggedIn = false;
  MemberInfoData? _memberInfoData;
  int _unreadMessageCount = 0;

  @override
  void initState() {
    super.initState();
    _checkLoginAndLoadData();
  }

  Future<void> _checkLoginAndLoadData() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String? savedToken = prefs.getString(token);
    final bool loggedIn = savedToken != null && savedToken.isNotEmpty;

    if (!mounted) {
      return;
    }

    setState(() {
      _isLoggedIn = loggedIn;
      if (!loggedIn) {
        _memberInfoData = null;
        _unreadMessageCount = 0;
      }
    });

    if (loggedIn) {
      await Future.wait([_queryMemberInfo(), _queryUnreadMessageCount()]);
    }
  }

  Future<void> _queryUnreadMessageCount() async {
    try {
      final Response result = await HttpUtil.get(unreadCountUrl);
      final UnreadCountModel model = UnreadCountModel.fromJson(result.data);
      if (!mounted) {
        return;
      }
      setState(() {
        _unreadMessageCount = model.unreadCount;
      });
    } catch (_) {
      // 未读数查询失败不阻塞页面展示。
    }
  }

  Future<void> _queryMemberInfo() async {
    try {
      final Response result = await HttpUtil.get(memberInfoDataUrl);
      final MemberInfoModel memberInfoModel = MemberInfoModel.fromJson(
        result.data,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _memberInfoData = memberInfoModel.data;
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: const Text('加载会员信息失败，请重试'),
          action: SnackBarAction(
            label: '重试',
            onPressed: _queryMemberInfo,
          ),
        ),
      );
    }
  }

  Future<void> _handleRefresh() async {
    await _checkLoginAndLoadData();
  }

  Future<bool> _ensureLogin() async {
    if (_isLoggedIn) {
      return true;
    }

    await Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => const Login()),
    );
    await _checkLoginAndLoadData();
    return _isLoggedIn;
  }

  Future<void> _openProfile() async {
    if (_isLoggedIn) {
      final bool? result = await Navigator.of(context).push<bool>(
        MaterialPageRoute(builder: (context) => const ProfileEdit()),
      );
      if (result == true) {
        await _queryMemberInfo();
      }
      return;
    }

    await _ensureLogin();
  }

  Future<void> _openMessages() async {
    if (!await _ensureLogin() || !mounted) {
      return;
    }

    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (context) => const Message()));
    await _queryUnreadMessageCount();
  }

  Future<void> _openSettings() async {
    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (context) => const Settings()));
    await _checkLoginAndLoadData();
  }

  Future<void> _openProtectedPage(Widget page) async {
    if (!await _ensureLogin() || !mounted) {
      return;
    }

    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (context) => page));
    await _checkLoginAndLoadData();
  }

  Future<void> _openOrderList({int initialTab = 0}) async {
    if (!await _ensureLogin() || !mounted) {
      return;
    }

    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => OrderList(initialTab: initialTab),
      ),
    );
    await _checkLoginAndLoadData();
  }

  String _displayName() {
    if (!_isLoggedIn) {
      return '欢迎登录';
    }
    final String nickname = _memberInfoData?.nickname.trim() ?? '';
    return nickname.isEmpty ? '九克城会员' : nickname;
  }

  String _displayDescription() {
    if (!_isLoggedIn) {
      return '同步订单、优惠券和收货地址';
    }

    final String signature = _memberInfoData?.signature.trim() ?? '';
    if (signature.isNotEmpty) {
      return signature;
    }
    return '订单、资产和常用服务都集中在这里';
  }

  String _maskedMobile() {
    final String mobile = _memberInfoData?.mobile.trim() ?? '';
    if (mobile.length < 7) {
      return mobile.isEmpty ? '资料待完善' : mobile;
    }
    return '${mobile.substring(0, 3)}****${mobile.substring(mobile.length - 4)}';
  }

  List<String> _heroTags() {
    if (!_isLoggedIn || _memberInfoData == null) {
      return const <String>[];
    }

    return <String>[
      'Lv.${_memberInfoData!.levelId} 会员',
      _unreadMessageCount > 0 ? '$_unreadMessageCount 条未读消息' : '消息已读',
      _maskedMobile(),
    ];
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        bottom: false,
        child: RefreshIndicator(
          color: AppColors.primary,
          onRefresh: _handleRefresh,
          child: CustomScrollView(
            physics: const BouncingScrollPhysics(
              parent: AlwaysScrollableScrollPhysics(),
            ),
            slivers: [
              _buildHeroSection(),
              if (_isLoggedIn) _buildMetricsSection(),
              _buildOrderSection(),
              _buildServiceSection(),
              const SliverToBoxAdapter(
                child: SizedBox(height: AppSpacing.xxl),
              ),
            ],
          ),
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildHeroSection() {
    final ThemeData theme = Theme.of(context);
    final List<String> heroTags = _heroTags();

    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.sm,
          AppSpacing.lg,
          0,
        ),
        child: Container(
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [AppColors.primary, AppColors.primaryDark],
            ),
            borderRadius: BorderRadius.circular(AppRadii.xxl),
            boxShadow: const [
              BoxShadow(
                color: Color(0x26101828),
                blurRadius: 28,
                offset: Offset(0, 16),
              ),
            ],
          ),
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.lg,
            AppSpacing.lg,
            AppSpacing.lg,
            AppSpacing.lg,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Text(
                    '个人中心',
                    style: theme.textTheme.titleMedium?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const Spacer(),
                  _MineTopActionButton(
                    icon: Icons.notifications_none_rounded,
                    semanticLabel: '消息中心',
                    badgeCount: _unreadMessageCount,
                    onTap: _openMessages,
                  ),
                  const SizedBox(width: AppSpacing.sm),
                  _MineTopActionButton(
                    icon: Icons.tune_rounded,
                    semanticLabel: '设置',
                    onTap: _openSettings,
                  ),
                ],
              ),
              const SizedBox(height: AppSpacing.lg),
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _MineAvatar(
                    avatarUrl: _memberInfoData?.avatar ?? '',
                    isLoggedIn: _isLoggedIn,
                  ),
                  const SizedBox(width: AppSpacing.lg),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _displayName(),
                          maxLines: _isLoggedIn ? 1 : 2,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.titleLarge?.copyWith(
                            color: Colors.white,
                            fontSize: _isLoggedIn ? 24 : 22,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        Text(
                          _displayDescription(),
                          maxLines: _isLoggedIn ? 1 : 2,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            color: Colors.white.withValues(alpha: 0.92),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (heroTags.isNotEmpty) ...[
                const SizedBox(height: AppSpacing.md),
                Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.sm,
                  children: heroTags
                      .take(2)
                      .map((tag) => _MineInfoPill(label: tag))
                      .toList(),
                ),
              ],
              const SizedBox(height: AppSpacing.lg),
              if (_isLoggedIn)
                Row(
                  children: [
                    Expanded(
                      child: _MineHeroButton(
                        label: '我的订单',
                        foregroundColor: AppColors.primaryDark,
                        backgroundColor: Colors.white,
                        onTap: () {
                          _openOrderList();
                        },
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    SizedBox(
                      width: 126,
                      child: _MineHeroButton(
                        label: '编辑资料',
                        foregroundColor: Colors.white,
                        backgroundColor: Colors.white.withValues(alpha: 0.1),
                        borderColor: Colors.white.withValues(alpha: 0.24),
                        onTap: _openProfile,
                      ),
                    ),
                  ],
                )
              else
                _MineHeroButton(
                  label: '立即登录',
                  foregroundColor: AppColors.primaryDark,
                  backgroundColor: Colors.white,
                  onTap: _openProfile,
                ),
            ],
          ),
        ),
      ),
    );
  }

  SliverToBoxAdapter _buildMetricsSection() {
    final List<_MineMetricData> metrics = [
      _MineMetricData(
        label: '积分',
        value: (_memberInfoData?.points ?? 0).toString(),
        subtitle: '会员权益',
        icon: Icons.stars_rounded,
        accentColor: AppColors.accent,
        backgroundColor: AppColors.accentSoft,
      ),
      _MineMetricData(
        label: '成长值',
        value: (_memberInfoData?.growthPoint ?? 0).toString(),
        subtitle: '等级成长',
        icon: Icons.trending_up_rounded,
        accentColor: AppColors.success,
        backgroundColor: const Color(0xFFE9F9F0),
      ),
      _MineMetricData(
        label: '优惠券',
        value: (_memberInfoData?.couponCount ?? 0).toString(),
        subtitle: '可用优惠',
        icon: Icons.confirmation_number_outlined,
        accentColor: AppColors.primaryDark,
        backgroundColor: AppColors.primarySoft,
        onTap: () {
          _openProtectedPage(const CouponList());
        },
      ),
      _MineMetricData(
        label: '数字卡片',
        value: '查看',
        subtitle: '到账进度',
        icon: Icons.style_outlined,
        accentColor: const Color(0xFF7C3AED),
        backgroundColor: const Color(0xFFF3E8FF),
        onTap: () {
          _openProtectedPage(
            const MyDigitalCardPage(intentSource: 'mine_asset_metric'),
          );
        },
      ),
      _MineMetricData(
        label: '订单数',
        value: (_memberInfoData?.orderCount ?? 0).toString(),
        subtitle: '订单进度',
        icon: Icons.receipt_long_rounded,
        accentColor: const Color(0xFF2563EB),
        backgroundColor: const Color(0xFFDBEAFE),
        onTap: () {
          _openOrderList();
        },
      ),
    ];

    return _buildSurfaceSection(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _MineSectionHeader(
            icon: Icons.dashboard_customize_rounded,
            iconColor: AppColors.primaryDark,
            iconBackground: AppColors.primarySoft,
            title: '我的资产',
            actionLabel: '查看卡片',
            onActionTap: () {
              _openProtectedPage(
                const MyDigitalCardPage(intentSource: 'mine_asset_header'),
              );
            },
          ),
          const SizedBox(height: AppSpacing.md),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: metrics.length,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              childAspectRatio: 1.68,
              crossAxisSpacing: AppSpacing.md,
              mainAxisSpacing: AppSpacing.md,
            ),
            itemBuilder: (context, index) {
              final _MineMetricData metric = metrics[index];
              return _MineMetricTile(data: metric);
            },
          ),
        ],
      ),
    );
  }

  SliverToBoxAdapter _buildOrderSection() {
    final List<_MineQuickActionData> actions = _isLoggedIn
        ? [
            _MineQuickActionData(
              label: '全部订单',
              icon: Icons.apps_rounded,
              accentColor: AppColors.primaryDark,
              backgroundColor: AppColors.primarySoft,
              onTap: () {
                _openOrderList(initialTab: 0);
              },
            ),
            _MineQuickActionData(
              label: '待支付',
              icon: Icons.payments_outlined,
              accentColor: const Color(0xFF2563EB),
              backgroundColor: const Color(0xFFDBEAFE),
              onTap: () {
                _openOrderList(initialTab: 1);
              },
            ),
            _MineQuickActionData(
              label: '待发货',
              icon: Icons.inventory_2_outlined,
              accentColor: AppColors.accent,
              backgroundColor: AppColors.accentSoft,
              onTap: () {
                _openOrderList(initialTab: 2);
              },
            ),
            _MineQuickActionData(
              label: '已完成',
              icon: Icons.task_alt_rounded,
              accentColor: AppColors.success,
              backgroundColor: const Color(0xFFE9F9F0),
              onTap: () {
                _openOrderList(initialTab: 3);
              },
            ),
          ]
        : [
            _MineQuickActionData(
              label: '订单',
              icon: Icons.receipt_long_rounded,
              accentColor: AppColors.primaryDark,
              backgroundColor: AppColors.primarySoft,
              onTap: () {
                _openOrderList(initialTab: 0);
              },
            ),
            _MineQuickActionData(
              label: '优惠券',
              icon: Icons.confirmation_number_outlined,
              accentColor: const Color(0xFF2563EB),
              backgroundColor: const Color(0xFFDBEAFE),
              onTap: () {
                _openProtectedPage(const CouponList());
              },
            ),
            _MineQuickActionData(
              label: '地址',
              icon: Icons.location_on_outlined,
              accentColor: AppColors.accent,
              backgroundColor: AppColors.accentSoft,
              onTap: () {
                _openProtectedPage(const AddressList());
              },
            ),
            _MineQuickActionData(
              label: '消息',
              icon: Icons.chat_bubble_outline_rounded,
              accentColor: AppColors.success,
              backgroundColor: const Color(0xFFE9F9F0),
              onTap: _openMessages,
            ),
          ];

    return _buildSurfaceSection(
      topPadding: _isLoggedIn ? AppSpacing.lg : AppSpacing.xl,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _MineSectionHeader(
            icon: Icons.local_shipping_outlined,
            iconColor: const Color(0xFF2563EB),
            iconBackground: const Color(0xFFDBEAFE),
            title: _isLoggedIn ? '订单入口' : '快捷入口',
            actionLabel: _isLoggedIn ? '查看全部' : null,
            onActionTap: _isLoggedIn
                ? () {
                    _openOrderList();
                  }
                : null,
          ),
          const SizedBox(height: AppSpacing.md),
          Row(
            children: actions
                .map(
                  (action) => Expanded(
                    child: Padding(
                      padding:
                          const EdgeInsets.symmetric(horizontal: AppSpacing.xs),
                      child: _MineQuickActionTile(data: action),
                    ),
                  ),
                )
                .toList(),
          ),
        ],
      ),
    );
  }

  SliverToBoxAdapter _buildServiceSection() {
    final List<_MineServiceData> services = [
      _MineServiceData(
        title: '地址管理',
        icon: Icons.location_on_outlined,
        accentColor: AppColors.primaryDark,
        backgroundColor: AppColors.primarySoft,
        onTap: () {
          _openProtectedPage(const AddressList());
        },
      ),
      _MineServiceData(
        title: '我的足迹',
        icon: Icons.history_rounded,
        accentColor: const Color(0xFF2563EB),
        backgroundColor: const Color(0xFFDBEAFE),
        onTap: () {
          _openProtectedPage(const History());
        },
      ),
      _MineServiceData(
        title: '我的关注',
        icon: Icons.visibility_outlined,
        accentColor: AppColors.accent,
        backgroundColor: AppColors.accentSoft,
        onTap: () {
          _openProtectedPage(const FocusOn());
        },
      ),
      _MineServiceData(
        title: '我的收藏',
        icon: Icons.bookmark_border_rounded,
        accentColor: const Color(0xFF7C3AED),
        backgroundColor: const Color(0xFFF3E8FF),
        onTap: () {
          _openProtectedPage(const Collection());
        },
      ),
      _MineServiceData(
        title: '我的评价',
        icon: Icons.rate_review_outlined,
        accentColor: AppColors.success,
        backgroundColor: const Color(0xFFE9F9F0),
        onTap: () {
          _openProtectedPage(const PinJia());
        },
      ),
      _MineServiceData(
        title: '售后服务',
        icon: Icons.support_agent_rounded,
        accentColor: const Color(0xFF7C3AED),
        backgroundColor: const Color(0xFFF3E8FF),
        onTap: () async {
          await _openOrderList(initialTab: 5);
        },
      ),
      _MineServiceData(
        title: '设置',
        icon: Icons.settings_outlined,
        accentColor: AppColors.textSecondary,
        backgroundColor: AppColors.surfaceMuted,
        onTap: _openSettings,
      ),
    ];

    return _buildSurfaceSection(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const _MineSectionHeader(
            icon: Icons.widgets_outlined,
            iconColor: Color(0xFF7C3AED),
            iconBackground: Color(0xFFF3E8FF),
            title: '更多服务',
          ),
          const SizedBox(height: AppSpacing.md),
          Container(
            decoration: BoxDecoration(
              color: AppColors.surfaceMuted,
              borderRadius: BorderRadius.circular(AppRadii.lg),
            ),
            child: Column(
              children: List.generate(services.length, (index) {
                final _MineServiceData service = services[index];
                return _MineServiceRow(
                  data: service,
                  showDivider: index != services.length - 1,
                );
              }),
            ),
          ),
        ],
      ),
    );
  }

  SliverToBoxAdapter _buildSurfaceSection({
    required Widget child,
    double topPadding = AppSpacing.lg,
  }) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: EdgeInsets.fromLTRB(
          AppSpacing.lg,
          topPadding,
          AppSpacing.lg,
          0,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: BorderRadius.circular(AppRadii.xl),
            border: Border.all(color: AppColors.border),
            boxShadow: const [
              BoxShadow(
                color: Color(0x12101828),
                blurRadius: 24,
                offset: Offset(0, 12),
              ),
            ],
          ),
          padding: const EdgeInsets.all(AppSpacing.md),
          child: child,
        ),
      ),
    );
  }
}

class _MineSectionHeader extends StatelessWidget {
  final IconData icon;
  final Color iconColor;
  final Color iconBackground;
  final String title;
  final String? actionLabel;
  final VoidCallback? onActionTap;

  const _MineSectionHeader({
    required this.icon,
    required this.iconColor,
    required this.iconBackground,
    required this.title,
    this.actionLabel,
    this.onActionTap,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);

    return Row(
      children: [
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: iconBackground,
            borderRadius: BorderRadius.circular(AppRadii.md),
          ),
          child: Icon(icon, color: iconColor, size: 22),
        ),
        const SizedBox(width: AppSpacing.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [Text(title, style: theme.textTheme.titleMedium)],
          ),
        ),
        if (actionLabel != null && onActionTap != null)
          TextButton(
            onPressed: onActionTap,
            child: Text(actionLabel!),
          ),
      ],
    );
  }
}

class _MineTopActionButton extends StatelessWidget {
  final IconData icon;
  final String semanticLabel;
  final int badgeCount;
  final VoidCallback onTap;

  const _MineTopActionButton({
    required this.icon,
    required this.semanticLabel,
    this.badgeCount = 0,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    Widget content = Icon(icon, color: Colors.white);
    if (badgeCount > 0) {
      content = Badge(
        label: Text(
          badgeCount > 99 ? '99+' : '$badgeCount',
          style: const TextStyle(fontSize: 9),
        ),
        child: content,
      );
    }

    return Semantics(
      button: true,
      label: semanticLabel,
      child: Material(
        color: Colors.white.withValues(alpha: 0.18),
        borderRadius: BorderRadius.circular(AppRadii.md),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.md),
          onTap: onTap,
          child: SizedBox(
            width: 40,
            height: 40,
            child: Center(child: content),
          ),
        ),
      ),
    );
  }
}

class _MineAvatar extends StatelessWidget {
  final String avatarUrl;
  final bool isLoggedIn;

  const _MineAvatar({
    required this.avatarUrl,
    required this.isLoggedIn,
  });

  @override
  Widget build(BuildContext context) {
    final Widget fallback = Container(
      width: 68,
      height: 68,
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.18),
        shape: BoxShape.circle,
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.32),
          width: 1.5,
        ),
      ),
      child: Icon(
        isLoggedIn ? Icons.person_rounded : Icons.login_rounded,
        size: 30,
        color: Colors.white,
      ),
    );

    if (!isLoggedIn || avatarUrl.trim().isEmpty) {
      return fallback;
    }

    return Container(
      width: 68,
      height: 68,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.32),
          width: 1.5,
        ),
      ),
      child: ClipOval(
        child: Image.network(
          proxyImageUrl(avatarUrl),
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) {
            return fallback;
          },
        ),
      ),
    );
  }
}

class _MineInfoPill extends StatelessWidget {
  final String label;

  const _MineInfoPill({required this.label});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.md,
        vertical: AppSpacing.xs,
      ),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.16),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: Colors.white,
              fontWeight: FontWeight.w700,
            ),
      ),
    );
  }
}

class _MineHeroButton extends StatelessWidget {
  final String label;
  final Color foregroundColor;
  final Color backgroundColor;
  final Color? borderColor;
  final VoidCallback onTap;

  const _MineHeroButton({
    required this.label,
    required this.foregroundColor,
    required this.backgroundColor,
    this.borderColor,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: backgroundColor,
      borderRadius: BorderRadius.circular(AppRadii.lg),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        onTap: onTap,
        child: Container(
          height: 44,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(AppRadii.lg),
            border:
                borderColor == null ? null : Border.all(color: borderColor!),
          ),
          alignment: Alignment.center,
          child: Text(
            label,
            style: Theme.of(context).textTheme.labelLarge?.copyWith(
                  color: foregroundColor,
                  fontWeight: FontWeight.w700,
                ),
          ),
        ),
      ),
    );
  }
}

class _MineMetricData {
  final String label;
  final String value;
  final String subtitle;
  final IconData icon;
  final Color accentColor;
  final Color backgroundColor;
  final VoidCallback? onTap;

  const _MineMetricData({
    required this.label,
    required this.value,
    required this.subtitle,
    required this.icon,
    required this.accentColor,
    required this.backgroundColor,
    this.onTap,
  });
}

class _MineMetricTile extends StatelessWidget {
  final _MineMetricData data;

  const _MineMetricTile({required this.data});

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);

    return Semantics(
      label: '${data.label}，${data.subtitle}',
      child: Material(
        color: data.backgroundColor,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.lg),
          onTap: data.onTap,
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.sm),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Container(
                      width: 32,
                      height: 32,
                      decoration: BoxDecoration(
                        color: Colors.white.withValues(alpha: 0.92),
                        borderRadius: BorderRadius.circular(AppRadii.md),
                      ),
                      child: Icon(data.icon, color: data.accentColor, size: 18),
                    ),
                    Expanded(
                      child: Align(
                        alignment: Alignment.centerRight,
                        child: Text(
                          data.label,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: theme.textTheme.labelLarge?.copyWith(
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                Text(
                  data.value,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.titleMedium?.copyWith(
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  data.subtitle,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _MineQuickActionData {
  final String label;
  final IconData icon;
  final Color accentColor;
  final Color backgroundColor;
  final VoidCallback onTap;

  const _MineQuickActionData({
    required this.label,
    required this.icon,
    required this.accentColor,
    required this.backgroundColor,
    required this.onTap,
  });
}

class _MineQuickActionTile extends StatelessWidget {
  final _MineQuickActionData data;

  const _MineQuickActionTile({required this.data});

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: data.label,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadii.lg),
          onTap: data.onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
            child: Column(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: data.backgroundColor,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Icon(data.icon, color: data.accentColor, size: 24),
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  data.label,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        color: AppColors.textSecondary,
                      ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _MineServiceData {
  final String title;
  final IconData icon;
  final Color accentColor;
  final Color backgroundColor;
  final VoidCallback onTap;

  const _MineServiceData({
    required this.title,
    required this.icon,
    required this.accentColor,
    required this.backgroundColor,
    required this.onTap,
  });
}

class _MineServiceRow extends StatelessWidget {
  final _MineServiceData data;
  final bool showDivider;

  const _MineServiceRow({
    required this.data,
    required this.showDivider,
  });

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);

    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadii.lg),
        onTap: data.onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.lg,
            vertical: AppSpacing.md,
          ),
          decoration: BoxDecoration(
            border: showDivider
                ? const Border(
                    bottom: BorderSide(color: AppColors.border),
                  )
                : null,
          ),
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: data.backgroundColor,
                  borderRadius: BorderRadius.circular(AppRadii.md),
                ),
                child: Icon(data.icon, color: data.accentColor, size: 22),
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(data.title, style: theme.textTheme.titleSmall)
                  ],
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              const Icon(
                Icons.chevron_right_rounded,
                color: AppColors.textHint,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
