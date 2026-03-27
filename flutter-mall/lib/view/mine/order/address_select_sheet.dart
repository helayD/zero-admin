import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/address/address_edit.dart';
import 'package:flutter_mall/model/confirm_order.dart';

/// 地址选择底部弹窗（Story 5-3 Task 5）
class AddressSelectSheet extends StatefulWidget {
  final List<ConfirmAddress> addresses;
  final int selectedAddressId;
  final VoidCallback? onAddressesChanged;  // LOW-2: 刷新后通知父组件

  const AddressSelectSheet({
    super.key,
    required this.addresses,
    required this.selectedAddressId,
    this.onAddressesChanged,
  });

  @override
  State<AddressSelectSheet> createState() => _AddressSelectSheetState();
}

class _AddressSelectSheetState extends State<AddressSelectSheet> {
  late int _selectedId;

  @override
  void initState() {
    super.initState();
    _selectedId = widget.selectedAddressId;
  }

  void _onConfirm() {
    Navigator.of(context).pop(_selectedId);
  }

  @override
  Widget build(BuildContext context) {
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);
    return Container(
      height: MediaQuery.of(context).size.height * 0.65,
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
                    "选择收货地址",
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.close),
                  onPressed: () => Navigator.of(context).pop(_selectedId),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          // 地址列表
          Expanded(
            child: widget.addresses.isEmpty
                ? _buildEmpty()
                : ListView.builder(
                    itemCount: widget.addresses.length,
                    itemBuilder: (context, index) {
                      final addr = widget.addresses[index];
                      final isSelected = addr.id == _selectedId;
                      return _AddressItem(
                        address: addr,
                        isSelected: isSelected,
                        onTap: () {
                          setState(() {
                            _selectedId = addr.id;
                          });
                        },
                        onSetDefault: () => _setDefault(addr),
                      );
                    },
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
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _navigateToAddAddress,
                      icon: const Icon(Icons.add, size: 18),
                      label: const Text("新增地址"),
                      style: OutlinedButton.styleFrom(
                        foregroundColor: themeColor,
                        side: BorderSide(color: themeColor),
                        minimumSize: const Size(0, 44),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(22),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: themeColor,
                        foregroundColor: Colors.white,
                        minimumSize: const Size(0, 44),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(22),
                        ),
                      ),
                      onPressed: _onConfirm,
                      child: const Text("确定", style: TextStyle(fontSize: 15)),
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

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.location_off, size: 48, color: Colors.grey.shade300),
          const SizedBox(height: 12),
          Text(
            "暂无收货地址",
            style: TextStyle(fontSize: 14, color: Colors.grey.shade400),
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: _navigateToAddAddress,
            style: ElevatedButton.styleFrom(
              backgroundColor:
                  Color(int.parse('fa436a', radix: 16)).withAlpha(255),
              foregroundColor: Colors.white,
            ),
            child: const Text("新增地址"),
          ),
        ],
      ),
    );
  }

  void _navigateToAddAddress() async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (context) => const AddressEdit(),
      ),
    );
    if (result == true && mounted) {
      widget.onAddressesChanged?.call();  // LOW-2: 通知父组件刷新地址列表
    }
  }

  Future<void> _setDefault(ConfirmAddress addr) async {
    try {
      await HttpUtil.post(
        updateAddressStatusDataUrl,
        data: {"id": addr.id, "isDefault": 1},
      );
      if (mounted) {
        setState(() {
          _selectedId = addr.id;
        });
      }
    } catch (_) {}
  }
}

class _AddressItem extends StatelessWidget {
  final ConfirmAddress address;
  final bool isSelected;
  final VoidCallback onTap;
  final VoidCallback onSetDefault;

  const _AddressItem({
    required this.address,
    required this.isSelected,
    required this.onTap,
    required this.onSetDefault,
  });

  @override
  Widget build(BuildContext context) {
    final themeColor =
        Color(int.parse('fa436a', radix: 16)).withAlpha(255);
    final greyColor = Color(int.parse('909399', radix: 16)).withAlpha(255);

    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        decoration: BoxDecoration(
          border: Border(
            bottom: BorderSide(
              width: 1,
              color: Color(int.parse('f5f5f5', radix: 16)).withAlpha(255),
            ),
          ),
          color: isSelected ? themeColor.withAlpha(13) : Colors.transparent,
        ),
        child: Row(
          children: [
            Icon(
              isSelected ? Icons.check_circle : Icons.circle_outlined,
              color: isSelected ? themeColor : Colors.grey.shade300,
              size: 22,
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        "${address.receiverName} ${address.receiverPhone}",
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      if (address.isDefault == 1) ...[
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 6, vertical: 1),
                          decoration: BoxDecoration(
                            border: Border.all(color: themeColor, width: 1),
                            borderRadius: BorderRadius.circular(2),
                          ),
                          child: Text(
                            "默认",
                            style: TextStyle(fontSize: 11, color: themeColor),
                          ),
                        ),
                      ],
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    "${address.province} ${address.city} ${address.district} ${address.detailAddress}",
                    style: TextStyle(fontSize: 13, color: greyColor),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            if (address.isDefault != 1)
              TextButton(
                onPressed: onSetDefault,
                style: TextButton.styleFrom(
                  foregroundColor: greyColor,
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  minimumSize: Size.zero,
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                ),
                child: const Text("设为默认", style: TextStyle(fontSize: 12)),
              ),
          ],
        ),
      ),
    );
  }
}
