import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/utils/http_util.dart';

import '../../../model/coupon_model.dart';

///
/// 可领取优惠券列表页面（领券中心）
///
class AvailableCouponList extends StatefulWidget {
  const AvailableCouponList({super.key});

  @override
  State<AvailableCouponList> createState() => _AvailableCouponListState();
}

class _AvailableCouponListState extends State<AvailableCouponList> {
  List<CouponData> couponListData = [];
  bool isLoading = true;

  @override
  void initState() {
    super.initState();
    queryAvailableCoupons();
  }

  void queryAvailableCoupons() async {
    setState(() {
      isLoading = true;
    });
    try {
      Response result = await HttpUtil.get(availableCouponUrl);
      CouponModel couponModel = CouponModel.fromJson(result.data);
      setState(() {
        couponListData = couponModel.data;
        isLoading = false;
      });
    } catch (e) {
      setState(() {
        isLoading = false;
      });
    }
  }

  void claimCoupon(CouponData coupon) async {
    try {
      Response result = await HttpUtil.post(addCouponUrl, data: {"couponId": coupon.id});
      Map<String, dynamic> resp = result.data;
      if (resp["code"] == 0) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(resp["message"] ?? "领取成功"), backgroundColor: Colors.green),
          );
        }
        queryAvailableCoupons();
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(resp["message"] ?? "领取失败"), backgroundColor: Colors.red),
          );
        }
      }
    } on DioException catch (e) {
      String msg = "领取失败";
      if (e.response?.data != null && e.response!.data is Map) {
        msg = e.response!.data["message"] ?? msg;
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(msg), backgroundColor: Colors.red),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('领券中心'),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: isLoading
          ? const Center(child: CircularProgressIndicator())
          : couponListData.isEmpty
              ? const Center(child: Text('暂无可领取的优惠券'))
              : ListView.builder(
                  itemCount: couponListData.length,
                  itemBuilder: (context, index) {
                    CouponData coupon = couponListData[index];
                    bool canClaim = coupon.receiveStatus == 0;
                    String btnText = canClaim ? '立即领取' : '已领取';

                    return Container(
                      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(8),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.grey.withAlpha(30),
                            blurRadius: 4,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: Row(
                        children: [
                          Container(
                            width: 100,
                            padding: const EdgeInsets.symmetric(vertical: 16),
                            decoration: BoxDecoration(
                              color: canClaim
                                  ? Color(int.parse('fa436a', radix: 16)).withAlpha(255)
                                  : Colors.grey,
                              borderRadius: const BorderRadius.only(
                                topLeft: Radius.circular(8),
                                bottomLeft: Radius.circular(8),
                              ),
                            ),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Row(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    const Text('￥',
                                        style: TextStyle(fontSize: 14, color: Colors.white)),
                                    Text('${coupon.amount}',
                                        style: const TextStyle(
                                            fontSize: 24,
                                            color: Colors.white,
                                            fontWeight: FontWeight.bold)),
                                  ],
                                ),
                                const SizedBox(height: 4),
                                Text('满${coupon.minAmount}可用',
                                    style: const TextStyle(fontSize: 11, color: Colors.white70)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Padding(
                              padding: const EdgeInsets.all(12),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(coupon.name,
                                      style: const TextStyle(
                                          fontSize: 14, fontWeight: FontWeight.w500)),
                                  const SizedBox(height: 4),
                                  Text(_getScopeText(coupon.scopeType),
                                      style: TextStyle(fontSize: 12, color: Colors.grey[600])),
                                  const SizedBox(height: 4),
                                  Text('有效期至 ${coupon.endTime}',
                                      style: TextStyle(fontSize: 11, color: Colors.grey[500])),
                                ],
                              ),
                            ),
                          ),
                          Padding(
                            padding: const EdgeInsets.only(right: 12),
                            child: ElevatedButton(
                              onPressed: canClaim ? () => claimCoupon(coupon) : null,
                              style: ElevatedButton.styleFrom(
                                backgroundColor: canClaim
                                    ? Color(int.parse('fa436a', radix: 16)).withAlpha(255)
                                    : Colors.grey[300],
                                foregroundColor: Colors.white,
                                padding:
                                    const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                                shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(16)),
                              ),
                              child: Text(btnText, style: const TextStyle(fontSize: 12)),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
    );
  }

  String _getScopeText(int scopeType) {
    switch (scopeType) {
      case 0:
        return '全场通用';
      case 1:
        return '指定分类商品可用';
      case 2:
        return '指定商品可用';
      default:
        return '全场通用';
    }
  }
}
