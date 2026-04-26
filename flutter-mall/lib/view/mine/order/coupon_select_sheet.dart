import 'package:flutter/material.dart';
import 'package:flutter_mall/model/confirm_order.dart';
import 'package:flutter_mall/theme/app_theme.dart';

/// 优惠券选择底部弹窗（Story 5-3 Task 6）
class CouponSelectSheet extends StatefulWidget {
  final List<ConfirmCouponData> enableList;
  final List<ConfirmCouponData> disableList;
  final ConfirmCouponData? selectedCoupon;

  const CouponSelectSheet({
    super.key,
    required this.enableList,
    required this.disableList,
    this.selectedCoupon,
  });

  @override
  State<CouponSelectSheet> createState() => _CouponSelectSheetState();
}

class _CouponSelectSheetState extends State<CouponSelectSheet>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  ConfirmCouponData? _selected;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _selected = widget.selectedCoupon;
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _onConfirm() {
    Navigator.of(context).pop(_selected);
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: MediaQuery.of(context).size.height * 0.72,
      decoration: const BoxDecoration(
        color: AppColors.background,
        borderRadius: BorderRadius.vertical(top: Radius.circular(AppRadii.xxl)),
      ),
      child: Column(
        children: [
          SizedBox(
            width: double.infinity,
            height: 64,
            child: Stack(
              alignment: Alignment.center,
              children: [
                const Text(
                  "选择优惠券",
                  style: TextStyle(
                    fontSize: 16,
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                Positioned(
                  top: AppSpacing.sm,
                  right: AppSpacing.sm,
                  child: IconButton(
                    icon: const Icon(Icons.close),
                    color: AppColors.textPrimary,
                    iconSize: 24,
                    onPressed: () => Navigator.of(context).pop(),
                    tooltip: "关闭",
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          Container(
            color: AppColors.surface,
            child: TabBar(
              controller: _tabController,
              indicatorColor: AppColors.price,
              indicatorSize: TabBarIndicatorSize.label,
              indicatorWeight: 3,
              labelColor: AppColors.price,
              labelStyle: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w700,
              ),
              unselectedLabelColor: AppColors.textHint,
              unselectedLabelStyle: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
              ),
              tabs: [
                Tab(text: "可用优惠券（${widget.enableList.length}）"),
                Tab(text: "不可用（${widget.disableList.length}）"),
              ],
            ),
          ),
          const Divider(height: 1),
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildCouponList(widget.enableList, enabled: true),
                _buildCouponList(widget.disableList, enabled: false),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.lg,
              AppSpacing.md,
              AppSpacing.lg,
              AppSpacing.md,
            ),
            decoration: const BoxDecoration(
              color: AppColors.surface,
              border: Border(
                top: BorderSide(color: AppColors.border, width: 1),
              ),
            ),
            child: SafeArea(
              top: false,
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  if (_selected != null)
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            _selected!.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              fontSize: 14,
                              color: AppColors.textPrimary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            "-￥${_selected!.amount.toStringAsFixed(2)}",
                            style: const TextStyle(
                              fontSize: 16,
                              color: AppColors.price,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    const Expanded(
                      child: Text(
                        "暂不选择优惠券",
                        style: TextStyle(
                          fontSize: 14,
                          color: AppColors.textHint,
                        ),
                      ),
                    ),
                  const SizedBox(width: AppSpacing.md),
                  SizedBox(
                    width: 132,
                    height: 48,
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        foregroundColor: Colors.white,
                        elevation: 0,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(24),
                        ),
                      ),
                      onPressed: _onConfirm,
                      child: const Text(
                        "确定",
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCouponList(
    List<ConfirmCouponData> coupons, {
    required bool enabled,
  }) {
    if (coupons.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.local_offer_outlined,
              size: 48,
              color: AppColors.border,
            ),
            const SizedBox(height: AppSpacing.md),
            Text(
              enabled ? "暂无可用优惠券" : "暂无不可用优惠券",
              style: const TextStyle(
                fontSize: 14,
                color: AppColors.textHint,
              ),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.md,
        AppSpacing.md,
        AppSpacing.md,
        AppSpacing.lg,
      ),
      itemCount: coupons.length,
      itemBuilder: (context, index) {
        final coupon = coupons[index];
        final isSelected = _selected?.id == coupon.id;
        return _CouponCard(
          coupon: coupon,
          enabled: enabled,
          isSelected: isSelected,
          onTap: enabled
              ? () {
                  setState(() {
                    _selected = isSelected ? null : coupon;
                  });
                }
              : null,
        );
      },
    );
  }
}

class _CouponCard extends StatelessWidget {
  final ConfirmCouponData coupon;
  final bool enabled;
  final bool isSelected;
  final VoidCallback? onTap;

  const _CouponCard({
    required this.coupon,
    required this.enabled,
    required this.isSelected,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final amountColor = enabled ? AppColors.price : AppColors.textHint;
    final titleColor = enabled ? AppColors.textPrimary : AppColors.textHint;
    final subColor = enabled ? AppColors.textHint : AppColors.border;
    final borderColor = isSelected ? AppColors.primary : AppColors.border;

    return Opacity(
      opacity: enabled ? 1.0 : 0.72,
      child: Container(
        margin: const EdgeInsets.only(bottom: AppSpacing.md),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(AppRadii.lg),
          border: Border.all(
            color: borderColor,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        clipBehavior: Clip.antiAlias,
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            child: IntrinsicHeight(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Container(
                    width: 112,
                    padding: const EdgeInsets.symmetric(
                      horizontal: AppSpacing.sm,
                      vertical: AppSpacing.md,
                    ),
                    color: enabled
                        ? AppColors.price.withValues(alpha: 0.08)
                        : AppColors.surfaceMuted,
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          "￥${coupon.amount.toStringAsFixed(2)}",
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontSize: 19,
                            height: 1.1,
                            fontWeight: FontWeight.w800,
                            color: amountColor,
                          ),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        Text(
                          "满${coupon.minAmount.toStringAsFixed(0)}可用",
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            fontSize: 11,
                            color: subColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: AppSpacing.lg,
                        vertical: AppSpacing.md,
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            coupon.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              fontSize: 14,
                              height: 1.3,
                              fontWeight: FontWeight.w600,
                              color: titleColor,
                            ),
                          ),
                          const SizedBox(height: AppSpacing.xs),
                          Text(
                            "有效期至 ${coupon.endTime}",
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              fontSize: 12,
                              color: subColor,
                            ),
                          ),
                          if (!enabled && coupon.disableReason.isNotEmpty) ...[
                            const SizedBox(height: AppSpacing.xs),
                            Text(
                              coupon.disableReason,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontSize: 12,
                                color: AppColors.accent,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                  SizedBox(
                    width: 52,
                    child: Center(
                      child: enabled
                          ? Icon(
                              isSelected
                                  ? Icons.check_circle
                                  : Icons.circle_outlined,
                              color: isSelected
                                  ? AppColors.primary
                                  : AppColors.border,
                              size: 24,
                            )
                          : const SizedBox.shrink(),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
