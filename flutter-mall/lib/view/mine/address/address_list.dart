import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/view/mine/address/address_edit.dart';

///
/// 收货地址列表页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class AddressList extends StatefulWidget {
  const AddressList({super.key});

  @override
  State<AddressList> createState() => _AddressListState();
}

class _AddressListState extends State<AddressList> {
  static const Color _themeColor = Color(0xFFFA436A);
  static const Color _pageBg = Color(0xFFF7F7F7);
  static const Color _textPrimary = Color(0xFF303133);
  static const Color _textHint = Color(0xFF909399);

  List<AddressListData> addressListData = <AddressListData>[];
  bool _isLoading = true;
  bool _isRefreshing = false;
  String? _errorMessage;
  int? _actionAddressId;

  @override
  void initState() {
    super.initState();
    queryAddressList();
  }

  Future<void> queryAddressList({bool showLoading = true}) async {
    if (showLoading && mounted) {
      setState(() {
        _isLoading = true;
        _errorMessage = null;
      });
    }

    try {
      final Response result = await HttpUtil.get(addressListDataUrl);
      final addressListModel = AddressListModel.fromJson(result.data);
      if (!mounted) return;

      if (addressListModel.code != 0) {
        setState(() {
          _isLoading = false;
          _isRefreshing = false;
          _errorMessage = addressListModel.message.trim().isEmpty
              ? '加载失败，请重试'
              : addressListModel.message;
        });
        return;
      }

      final sortedData = List<AddressListData>.from(addressListModel.data)
        ..sort((a, b) {
          final defaultCompare = b.isDefault.compareTo(a.isDefault);
          if (defaultCompare != 0) return defaultCompare;
          return b.id.compareTo(a.id);
        });

      setState(() {
        addressListData = sortedData;
        _isLoading = false;
        _isRefreshing = false;
        _errorMessage = null;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _isRefreshing = false;
        _errorMessage = '加载失败，请重试';
      });
    }
  }

  Future<void> _refreshAddressList() async {
    setState(() => _isRefreshing = true);
    await queryAddressList(showLoading: false);
  }

  Future<void> _deleteAddress(AddressListData data) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除地址'),
        content: Text('确定删除 ${data.receiverName} 的收货地址吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('删除', style: TextStyle(color: _themeColor)),
          ),
        ],
      ),
    );

    if (confirmed != true || _actionAddressId != null) return;

    setState(() => _actionAddressId = data.id);
    try {
      final Response result = await HttpUtil.get(
        deleteAddressDataUrl,
        queryParameters: <String, dynamic>{'ids': data.id},
      );
      if (!mounted) return;
      if (result.data is Map && result.data['code'] == 0) {
        _showSnackBar('已删除收货地址');
        await queryAddressList(showLoading: false);
      } else {
        _showSnackBar(result.data?['message']?.toString() ?? '删除失败');
      }
    } catch (_) {
      if (!mounted) return;
      _showSnackBar('网络请求失败，请稍后重试');
    } finally {
      if (mounted) {
        setState(() => _actionAddressId = null);
      }
    }
  }

  Future<void> _setDefault(AddressListData data) async {
    if (data.isDefault == 1 || _actionAddressId != null) return;

    setState(() => _actionAddressId = data.id);
    try {
      final Response result = await HttpUtil.post(
        updateAddressStatusDataUrl,
        data: <String, dynamic>{'id': data.id, 'isDefault': 1},
      );
      if (!mounted) return;
      if (result.data is Map && result.data['code'] == 0) {
        _showSnackBar('已设为默认地址');
        await queryAddressList(showLoading: false);
      } else {
        _showSnackBar(result.data?['message']?.toString() ?? '设置失败');
      }
    } catch (_) {
      if (!mounted) return;
      _showSnackBar('网络请求失败，请稍后重试');
    } finally {
      if (mounted) {
        setState(() => _actionAddressId = null);
      }
    }
  }

  Future<void> _navigateToEdit({AddressListData? addressData}) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (context) => AddressEdit(
          addressData: addressData,
          initialDefault: addressData == null && addressListData.isEmpty,
        ),
      ),
    );
    if (result == true) {
      await queryAddressList(showLoading: false);
    }
  }

  void _showSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _pageBg,
      appBar: AppBar(
        backgroundColor: Colors.white,
        foregroundColor: Colors.black,
        elevation: 0,
        title: const Text(
          '收货地址',
          style: TextStyle(
            fontSize: 18,
            color: Colors.black,
            fontWeight: FontWeight.w700,
          ),
        ),
        centerTitle: true,
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(child: _buildBody()),
            _buildBottomAction(),
          ],
        ),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null) {
      return _buildScrollableState(
        icon: Icons.wifi_off_rounded,
        title: _errorMessage!,
        subtitle: '请检查网络或稍后重试',
        actionLabel: '重试',
        onAction: queryAddressList,
      );
    }

    if (addressListData.isEmpty) {
      return _buildScrollableState(
        icon: Icons.location_on_outlined,
        title: '暂无收货地址',
        subtitle: '添加地址后，下单时会自动带入收货信息',
        actionLabel: '新增地址',
        onAction: () => _navigateToEdit(),
      );
    }

    return RefreshIndicator(
      onRefresh: _refreshAddressList,
      color: _themeColor,
      child: ListView.separated(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.fromLTRB(14, 12, 14, 16),
        itemCount: addressListData.length + (_isRefreshing ? 1 : 0),
        separatorBuilder: (_, __) => const SizedBox(height: 10),
        itemBuilder: (context, index) {
          if (_isRefreshing && index == 0) {
            return const SizedBox(height: 1);
          }
          final data = addressListData[index - (_isRefreshing ? 1 : 0)];
          return _AddressCard(
            data: data,
            busy: _actionAddressId == data.id,
            onEdit: () => _navigateToEdit(addressData: data),
            onDelete: () => _deleteAddress(data),
            onSetDefault: () => _setDefault(data),
          );
        },
      ),
    );
  }

  Widget _buildScrollableState({
    required IconData icon,
    required String title,
    required String subtitle,
    required String actionLabel,
    required VoidCallback onAction,
  }) {
    return RefreshIndicator(
      onRefresh: _refreshAddressList,
      color: _themeColor,
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.symmetric(horizontal: 32),
        children: [
          SizedBox(height: MediaQuery.of(context).size.height * 0.22),
          Icon(icon, size: 58, color: const Color(0xFFD0D5DD)),
          const SizedBox(height: 16),
          Text(
            title,
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontSize: 17,
              fontWeight: FontWeight.w700,
              color: _textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            subtitle,
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontSize: 14,
              height: 1.5,
              color: _textHint,
            ),
          ),
          const SizedBox(height: 18),
          Center(
            child: OutlinedButton(
              onPressed: onAction,
              style: OutlinedButton.styleFrom(
                foregroundColor: _themeColor,
                side: const BorderSide(color: Color(0xFFFFB3C5)),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(22),
                ),
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 10),
              ),
              child: Text(actionLabel),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBottomAction() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
      child: SizedBox(
        width: double.infinity,
        height: 52,
        child: ElevatedButton.icon(
          onPressed: () => _navigateToEdit(),
          style: ElevatedButton.styleFrom(
            backgroundColor: _themeColor,
            foregroundColor: Colors.white,
            elevation: 0,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
            textStyle: const TextStyle(
              fontSize: 17,
              fontWeight: FontWeight.w700,
            ),
          ),
          icon: const Icon(Icons.add_location_alt_rounded),
          label: const Text('新增地址'),
        ),
      ),
    );
  }
}

class _AddressCard extends StatelessWidget {
  final AddressListData data;
  final bool busy;
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  final VoidCallback onSetDefault;

  const _AddressCard({
    required this.data,
    required this.busy,
    required this.onEdit,
    required this.onDelete,
    required this.onSetDefault,
  });

  static const Color _themeColor = Color(0xFFFA436A);
  static const Color _textPrimary = Color(0xFF303133);
  static const Color _textSecondary = Color(0xFF606266);
  static const Color _textHint = Color(0xFF909399);

  @override
  Widget build(BuildContext context) {
    final isDefault = data.isDefault == 1;

    return Container(
      padding: const EdgeInsets.fromLTRB(16, 16, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: isDefault ? const Color(0xFFFFD2DC) : const Color(0xFFEFEFEF),
        ),
        boxShadow: const [
          BoxShadow(
            color: Color(0x08000000),
            blurRadius: 10,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(
                  data.contactText,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontSize: 17,
                    fontWeight: FontWeight.w700,
                    color: _textPrimary,
                  ),
                ),
              ),
              if (busy)
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              else
                PopupMenuButton<String>(
                  padding: EdgeInsets.zero,
                  icon: const Icon(Icons.more_horiz_rounded),
                  onSelected: (value) {
                    switch (value) {
                      case 'edit':
                        onEdit();
                        break;
                      case 'delete':
                        onDelete();
                        break;
                    }
                  },
                  itemBuilder: (context) => const [
                    PopupMenuItem(
                      value: 'edit',
                      child: ListTile(
                        dense: true,
                        leading: Icon(Icons.edit_location_alt_outlined),
                        title: Text('编辑'),
                      ),
                    ),
                    PopupMenuItem(
                      value: 'delete',
                      child: ListTile(
                        dense: true,
                        leading: Icon(Icons.delete_outline_rounded),
                        title: Text('删除'),
                      ),
                    ),
                  ],
                ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            data.fullAddress,
            style: const TextStyle(
              fontSize: 15,
              height: 1.45,
              color: _textSecondary,
            ),
          ),
          if (data.tag.trim().isNotEmpty || isDefault) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                if (isDefault)
                  const _AddressBadge(
                    text: '默认',
                    backgroundColor: Color(0xFFFFEDF2),
                    textColor: _themeColor,
                  ),
                if (data.tag.trim().isNotEmpty)
                  _AddressBadge(
                    text: data.tag.trim(),
                    backgroundColor: const Color(0xFFF2F4F7),
                    textColor: _textHint,
                  ),
              ],
            ),
          ],
          const SizedBox(height: 12),
          const Divider(height: 1, color: Color(0xFFEDEDED)),
          const SizedBox(height: 4),
          Row(
            children: [
              TextButton.icon(
                onPressed: busy ? null : onSetDefault,
                style: TextButton.styleFrom(
                  foregroundColor: isDefault ? _themeColor : _textHint,
                  padding: const EdgeInsets.only(right: 10),
                ),
                icon: Icon(
                  isDefault
                      ? Icons.check_circle_rounded
                      : Icons.radio_button_unchecked_rounded,
                  size: 19,
                ),
                label: Text(isDefault ? '默认地址' : '设为默认'),
              ),
              const Spacer(),
              TextButton.icon(
                onPressed: busy ? null : onEdit,
                style: TextButton.styleFrom(foregroundColor: _textSecondary),
                icon: const Icon(Icons.edit_outlined, size: 18),
                label: const Text('编辑'),
              ),
              TextButton.icon(
                onPressed: busy ? null : onDelete,
                style: TextButton.styleFrom(foregroundColor: _textSecondary),
                icon: const Icon(Icons.delete_outline_rounded, size: 18),
                label: const Text('删除'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _AddressBadge extends StatelessWidget {
  final String text;
  final Color backgroundColor;
  final Color textColor;

  const _AddressBadge({
    required this.text,
    required this.backgroundColor,
    required this.textColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w700,
          color: textColor,
        ),
      ),
    );
  }
}
