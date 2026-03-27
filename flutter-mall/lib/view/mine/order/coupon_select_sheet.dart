import 'package:flutter/material.dart';
import 'package:flutter_mall/model/confirm_order.dart';

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
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);
    return Container(
      height: MediaQuery.of(context).size.height * 0.7,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
      ),
      child: Column(
        children: [
          // 顶部栏
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            child: Row(
              children: [
                const Expanded(
                  child: Text(
                    "选择优惠券",
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(context).pop(),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          // Tab 切换
          Container(
            color: Colors.white,
            child: TabBar(
              controller: _tabController,
              indicatorColor: themeColor,
              labelColor: themeColor,
              unselectedLabelColor: Colors.grey,
              tabs: [
                Tab(text: "可用优惠券（${widget.enableList.length}）"),
                Tab(text: "不可用（${widget.disableList.length}）"),
              ],
            ),
          ),
          const Divider(height: 1),
          // 内容区
          Expanded(
            child: TabBarView(
              controller: _tabController,
              children: [
                _buildCouponList(widget.enableList, enabled: true),
                _buildCouponList(widget.disableList, enabled: false),
              ],
            ),
          ),
          // 底部按钮
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.grey.withAlpha(51),
                  blurRadius: 8,
                  offset: const Offset(0, -2),
                ),
              ],
            ),
            child: SafeArea(
              child: Row(
                children: [
                  if (_selected != null)
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            _selected!.name,
                            style: TextStyle(
                              fontSize: 14,
                              color: themeColor,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          Text(
                            "-￥${_selected!.amount.toStringAsFixed(2)}",
                            style: TextStyle(
                              fontSize: 16,
                              color: themeColor,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    const Expanded(
                      child: Text(
                        "暂不选择优惠券",
                        style: TextStyle(fontSize: 14, color: Colors.grey),
                      ),
                    ),
                  const SizedBox(width: 12),
                  ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: themeColor,
                      foregroundColor: Colors.white,
                      minimumSize: const Size(100, 44),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(22),
                      ),
                    ),
                    onPressed: _onConfirm,
                    child: const Text("确定", style: TextStyle(fontSize: 15)),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCouponList(List<ConfirmCouponData> coupons, {required bool enabled}) {
    if (coupons.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.local_offer_outlined,
              size: 48,
              color: Colors.grey.shade300,
            ),
            const SizedBox(height: 12),
            Text(
              enabled ? "暂无可用优惠券" : "暂无可用优惠券",
              style: TextStyle(fontSize: 14, color: Colors.grey.shade400),
            ),
          ],
        ),
      );
    }
    return ListView.builder(
      padding: const EdgeInsets.all(12),
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
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);
    final greyColor = Color(int.parse('909399', radix: 16)).withAlpha(255);
    final disabledColor = Colors.grey.shade300;

    return Opacity(
      opacity: enabled ? 1.0 : 0.5,
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          margin: const EdgeInsets.only(bottom: 10),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: isSelected ? themeColor : (enabled ? Colors.grey.shade200 : Colors.grey.shade300),
              width: isSelected ? 2 : 1,
            ),
          ),
          child: Row(
            children: [
              // 金额区
              Container(
                width: 90,
                padding: const EdgeInsets.symmetric(vertical: 14),
                decoration: BoxDecoration(
                  color: enabled
                      ? themeColor.withAlpha(25)
                      : Colors.grey.shade100,
                  borderRadius: const BorderRadius.only(
                    topLeft: Radius.circular(7),
                    bottomLeft: Radius.circular(7),
                  ),
                ),
                child: Column(
                  children: [
                    Text(
                      "￥${coupon.amount.toStringAsFixed(2)}",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: enabled ? themeColor : disabledColor,
                      ),
                    ),
                    Text(
                      "满${coupon.minAmount.toStringAsFixed(0)}可用",
                      style: TextStyle(
                        fontSize: 10,
                        color: enabled ? greyColor : disabledColor,
                      ),
                    ),
                  ],
                ),
              ),
              // 信息区
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.all(10),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        coupon.name,
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: enabled ? Colors.black87 : disabledColor,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 3),
                      Text(
                        "有效期至 ${coupon.endTime}",
                        style: TextStyle(
                          fontSize: 11,
                          color: enabled ? greyColor : disabledColor,
                        ),
                      ),
                      if (!enabled && coupon.disableReason.isNotEmpty) ...[
                        const SizedBox(height: 3),
                        Text(
                          coupon.disableReason,
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.orange.shade700,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
              // 选中标识
              if (enabled)
                Padding(
                  padding: const EdgeInsets.only(right: 12),
                  child: Icon(
                    isSelected ? Icons.check_circle : Icons.circle_outlined,
                    color: isSelected ? themeColor : Colors.grey.shade300,
                    size: 22,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
