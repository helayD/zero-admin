import 'package:flutter/material.dart';
import 'package:flutter_mall/model/address_list.dart';
import 'package:flutter_mall/theme/app_theme.dart';

class PhysicalFulfillmentAddressSheet extends StatelessWidget {
  final List<AddressListData> addresses;
  final int? selectedAddressId;
  final ValueChanged<AddressListData> onSelected;

  const PhysicalFulfillmentAddressSheet({
    super.key,
    required this.addresses,
    required this.selectedAddressId,
    required this.onSelected,
  });

  @override
  Widget build(BuildContext context) {
    if (addresses.isEmpty) {
      return Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Text(
          '暂无可用收货地址，请先在地址簿中新增地址。',
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      );
    }

    return SafeArea(
      child: ListView.separated(
        shrinkWrap: true,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
        itemCount: addresses.length,
        separatorBuilder: (_, __) => const SizedBox(height: AppSpacing.sm),
        itemBuilder: (BuildContext context, int index) {
          final AddressListData item = addresses[index];
          final bool selected = selectedAddressId == item.id;
          return InkWell(
            borderRadius: BorderRadius.circular(AppRadii.lg),
            onTap: () => onSelected(item),
            child: Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: AppColors.surface,
                borderRadius: BorderRadius.circular(AppRadii.lg),
                border: Border.all(
                  color: selected ? const Color(0xFF2563EB) : AppColors.border,
                ),
              ),
              child: Row(
                children: <Widget>[
                  Icon(
                    selected
                        ? Icons.radio_button_checked_rounded
                        : Icons.radio_button_unchecked_rounded,
                    color: selected
                        ? const Color(0xFF2563EB)
                        : AppColors.textSecondary,
                  ),
                  const SizedBox(width: AppSpacing.md),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          item.contactText,
                          style:
                              Theme.of(context).textTheme.titleSmall?.copyWith(
                                    color: AppColors.textPrimary,
                                  ),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        Text(
                          item.fullAddress,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}
