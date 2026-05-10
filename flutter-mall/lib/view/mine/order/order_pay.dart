import 'dart:async';
import 'dart:io' show Platform;

import 'package:flutter/material.dart';

// alipay_kit: ^6.0.0
import 'package:alipay_kit/alipay_kit.dart';
// fluwx: ^5.0.0
import 'package:fluwx/fluwx.dart' as fluwx;

import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/upgrade_gate_page.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/model/upgrade_gate_context.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';
import 'package:flutter_mall/utils/upgrade_gate_service.dart';
import 'package:flutter_mall/view/digital_card/my_digital_card_page.dart';

///
/// 订单支付页面（Story 5.4 重构 + Story 5.5 支付发起）
///
/// 重构内容:
/// - 增加 orderId / orderSn / payType 参数，支持幂等命中后跳转已有订单
/// - 展示订单摘要（订单号 / 金额 / 支付方式）
///
/// Story 5.5 新增:
/// - 发起真实支付（支付宝 / 微信 SDK）
/// - 支付结果轮询（最长 5 分钟）
/// - 支付结果页（成功 / 失败 / 待确认）
/// - 待支付恢复入口（onResume 刷新）
///
/// 作者：LiuFeiHua
/// 日期：2023/11/21 17:17
/// Story 5.4: Task 8 重构
/// Story 5.5: Tasks 7-10 实现
///
class OrderPay extends StatefulWidget {
  /// 订单ID（Story 5.4 新增）
  final int? orderId;

  /// 订单编号（Story 5.4 新增）
  final String? orderSn;

  /// 支付方式：1=支付宝，2=微信（Story 5.4 新增）
  final int? payType;

  /// 支付金额（元）
  final double amount;

  /// Story 10.7 Review Fix HIGH-1: 订单是否包含提货卡类履约商品。
  /// 只有为 true 时支付成功页才展示"提货卡已发放"提示和"查看我的提货卡"按钮。
  final bool hasDigitalCard;

  const OrderPay({
    super.key,
    this.orderId,
    this.orderSn,
    this.payType,
    required this.amount,
    this.hasDigitalCard = false,
  });

  @override
  State<OrderPay> createState() => _OrderPayState();
}

// ==================== 页面状态枚举 ====================
enum PayPageState {
  /// 初始态：支付选择
  initial,

  /// 加载中：正在发起支付
  loading,

  /// 轮询中：等待支付结果
  polling,

  /// 支付成功
  success,

  /// 支付失败
  failed,

  /// 订单已取消 / 超时
  cancelled,
}

// ==================== 轮询结果 ====================
class PayQueryResult {
  final int orderStatus; // 0=待支付 1=已支付 2=已取消
  final int payStatus; // 0=未支付 1=已支付
  final String message;
  final int expireTime; // 剩余秒数

  PayQueryResult({
    required this.orderStatus,
    required this.payStatus,
    required this.message,
    required this.expireTime,
  });

  factory PayQueryResult.fromResp(Map<String, dynamic> resp) {
    return PayQueryResult(
      orderStatus: resp['orderStatus'] ?? 0,
      payStatus: resp['payStatus'] ?? 0,
      message: resp['message'] ?? '',
      expireTime: resp['expireTime'] ?? 0,
    );
  }
}

class _OrderPayState extends State<OrderPay> with WidgetsBindingObserver {
  // ==================== 状态 ====================
  late int _selectedPayType;
  PayPageState _pageState = PayPageState.initial;
  String? _payErrorMessage;
  Timer? _pollingTimer;
  Timer? _countdownTimer;
  int _expireTime = 0; // 剩余支付秒数
  int _pollCount = 0;
  bool _isPaying = false;

  // AlipayKit 支付结果监听
  StreamSubscription? _alipaySubscription;

  // ==================== 生命周期 ====================
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _selectedPayType = widget.payType ?? 2; // 默认微信
    _initAlipayListener();
    // Task 9.1: 进入页面时自动查询当前状态（待支付恢复入口）
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _queryPayStatus();
    });
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _pollingTimer?.cancel();
    _countdownTimer?.cancel();
    _alipaySubscription?.cancel();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // Task 9.5: onResume 时重新查询状态
    if (state == AppLifecycleState.resumed) {
      _queryPayStatus();
    }
  }

  // ==================== Alipay 监听 ====================
  void _initAlipayListener() {
    _alipaySubscription = AlipayKitPlatform.instance.payResp().listen((resp) {
      if (!mounted) return;
      if (resp.isSuccessful) {
        _onPaymentReturned(isSuccess: true);
      } else if (resp.isCancelled) {
        // canceled
        setState(() {
          _pageState = PayPageState.initial;
        });
      } else {
        setState(() {
          _pageState = PayPageState.failed;
          _payErrorMessage = resp.memo ?? "支付宝支付失败";
        });
      }
    });
  }

  // ==================== 核心：发起支付 ====================
  /// Task 7: 发起真实支付
  Future<void> _doPay() async {
    if (!mounted || _isPaying) return;
    if (widget.orderId == null || widget.orderId == 0) {
      _showToast("订单信息异常，请返回重新下单");
      return;
    }
    final allowPay = await _ensureUpgradeReady();
    if (!allowPay || !mounted) {
      return;
    }

    setState(() {
      _pageState = PayPageState.loading;
      _isPaying = true;
      _payErrorMessage = null;
    });

    try {
      // Task 7.1: 调用 orderPayUrl 发起预下单
      final resp = await HttpUtil.post(
        orderPayUrl,
        data: {
          'orderId': widget.orderId,
          'payType': _selectedPayType,
        },
      );

      final data = resp.data;
      if (!mounted) return;

      if (data['code'] != 0) {
        setState(() {
          _pageState = PayPageState.failed;
          _payErrorMessage = data['message'] ?? '支付发起失败';
          _isPaying = false;
        });
        return;
      }

      final payParams = data['data'] as String? ?? '';

      // 模拟支付（payType=99）直接返回成功，无需调用 SDK
      if (_selectedPayType == 99) {
        if (!mounted) return;
        setState(() {
          _pageState = PayPageState.success;
          _isPaying = false;
        });
        return;
      }

      if (payParams.isEmpty) {
        setState(() {
          _pageState = PayPageState.failed;
          _payErrorMessage = '支付参数为空，请检查支付配置';
          _isPaying = false;
        });
        return;
      }

      // Task 7.2/7.3: 根据 PayType 调起对应 SDK
      await _invokePaymentSDK(payParams);
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _pageState = PayPageState.failed;
        _payErrorMessage = '系统异常，请稍后重试';
        _isPaying = false;
      });
    }
  }

  Future<bool> _ensureUpgradeReady() async {
    final policy = await UpgradeGateService.queryPolicy(
      scene: 'order_pay',
      targetType: appRecentTargetTypeToValue(AppRecentTargetType.orderDetail),
      targetId: widget.orderId,
    );
    if (!policy.hasUpgradeGate) {
      return true;
    }
    if (!mounted) {
      return true;
    }
    final navigator = Navigator.of(context);
    final pendingUpgrade = PendingUpgradeContext.forOrderPay(
      policy: policy,
      orderId: widget.orderId ?? 0,
      payAmount: widget.amount,
      payType: _selectedPayType,
      orderSn: widget.orderSn ?? '',
      fallbackContext: AppRecentContext.create(
        targetType: AppRecentTargetType.orderDetail,
        targetId: widget.orderId,
        source: 'order_pay',
        requiresAuth: true,
        fallbackType: AppRecentTargetType.orderList,
        fallbackTabIndex: 1,
      ),
    );
    await AppRecoveryStore.savePendingUpgradeContext(pendingUpgrade);
    if (!mounted) {
      return true;
    }
    final result = await navigator.push<bool>(
      MaterialPageRoute(
        builder: (_) => UpgradeGatePage(
          pendingContext: pendingUpgrade,
          initialPolicy: policy,
          onResolved: (gateContext, _) async {
            Navigator.of(gateContext).pop(true);
          },
          onContinueLater: policy.canContinueLater
              ? (gateContext, _) async {
                  Navigator.of(gateContext).pop(true);
                }
              : null,
        ),
      ),
    );
    return result == true;
  }

  /// Task 7.2/7.3: 调起支付 SDK
  Future<void> _invokePaymentSDK(String payParams) async {
    if (_selectedPayType == 1) {
      // 支付宝 App 支付
      try {
        await AlipayKitPlatform.instance.pay(orderInfo: payParams);
      } catch (e) {
        if (!mounted) return;
        setState(() {
          _pageState = PayPageState.failed;
          _payErrorMessage = '支付宝 SDK 调用失败';
          _isPaying = false;
        });
      }
    } else {
      // 微信支付（H5 支付方案）
      // payParams 可能是 mweb_url（微信 H5 支付）
      await _handleWechatPay(payParams);
    }
  }

  /// Task 7.3: 微信支付处理
  Future<void> _handleWechatPay(String mwebUrl) async {
    if (Platform.isAndroid) {
      // Android: 检查微信是否安装并调起
      final isInstalled = await fluwx.Fluwx().isWeChatInstalled;
      if (isInstalled == true) {
        // Android 微信 APP 支付: payParams 应为微信支付 SDK 返回的调起参数
        // 由于 Story 5.5 微信 SDK 未完全接入，显示引导页
        if (!mounted) return;
        await _showWXGuideDialog(mwebUrl);
      } else {
        // 微信未安装，显示引导
        if (!mounted) return;
        await _showWXNotInstalledDialog();
      }
    } else {
      // iOS: 微信 H5 支付需引导用户在 Safari 中打开 mweb_url
      if (!mounted) return;
      await _showWXGuideDialog(mwebUrl);
    }

    // SDK 返回后（用户从微信切回 App）→ 开始轮询
    _onPaymentReturned(isSuccess: true);
  }

  /// 微信引导 Dialog（H5 mweb_url）
  Future<void> _showWXGuideDialog(String mwebUrl) async {
    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        title: const Text('微信支付'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              '即将调起微信支付...',
              style: TextStyle(fontSize: 14),
            ),
            const SizedBox(height: 8),
            Text(
              '请在微信中完成支付后返回本 App',
              style: TextStyle(fontSize: 13, color: Colors.grey[600]),
            ),
            const SizedBox(height: 8),
            Text(
              'mweb_url: ${mwebUrl.length > 30 ? '${mwebUrl.substring(0, 30)}...' : mwebUrl}',
              style: TextStyle(fontSize: 11, color: Colors.grey[400]),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('已支付，返回查询'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('取消'),
          ),
        ],
      ),
    );
  }

  Future<void> _showWXNotInstalledDialog() async {
    await showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('提示'),
        content: const Text('微信未安装，请安装后重试'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('确定'),
          ),
        ],
      ),
    );
    if (!mounted) return;
    setState(() {
      _pageState = PayPageState.initial;
      _isPaying = false;
    });
  }

  // ==================== 支付返回 → 轮询 ====================
  void _onPaymentReturned({required bool isSuccess}) {
    if (!mounted) return;
    if (isSuccess) {
      // Task 7.4: 开始轮询支付结果
      _startPolling();
    }
  }

  /// Task 7.4: 轮询支付结果（每 3s，最长 5 分钟）
  void _startPolling() {
    setState(() {
      _pageState = PayPageState.polling;
      _isPaying = false;
      _pollCount = 0;
    });

    // 立即查询一次
    _pollPayStatus();

    // 每 3s 轮询一次
    _pollingTimer?.cancel();
    _pollingTimer = Timer.periodic(const Duration(seconds: 3), (timer) {
      _pollCount++;
      if (_pollCount >= 100) {
        // 5 分钟超时（100 * 3s = 300s）
        timer.cancel();
        if (!mounted) return;
        setState(() {
          _pageState = PayPageState.polling;
          _payErrorMessage = '支付结果确认中，请稍后在订单详情查看';
        });
        return;
      }
      _pollPayStatus();
    });
  }

  /// 执行一次轮询
  Future<void> _pollPayStatus() async {
    if (widget.orderId == null) return;

    try {
      final resp = await HttpUtil.get(
        '$orderPayQueryUrl?orderId=${widget.orderId}',
      );
      if (!mounted) return;

      final data = resp.data;
      final result = PayQueryResult.fromResp(data);

      _stopPollingIfDone(result);
    } catch (e) {
      // 网络错误，继续轮询
    }
  }

  /// 根据查询结果决定是否停止轮询
  void _stopPollingIfDone(PayQueryResult result) {
    if (!mounted) return;

    switch (result.orderStatus) {
      case 1: // 已支付
        _pollingTimer?.cancel();
        setState(() {
          _pageState = PayPageState.success;
        });
        break;
      case 2: // 已取消 / 超时
        _pollingTimer?.cancel();
        setState(() {
          _pageState = PayPageState.cancelled;
          _payErrorMessage =
              result.message.isNotEmpty ? result.message : '订单已取消';
        });
        break;
      default:
        // 仍待支付（orderStatus=0）：继续轮询
        setState(() {
          _expireTime = result.expireTime;
        });
        // 启动倒计时
        _startCountdown(result.expireTime);
    }
  }

  /// Task 9.1/9.5: 进入页面 / onResume 时查询当前状态
  Future<void> _queryPayStatus() async {
    if (widget.orderId == null) return;
    if (_pageState != PayPageState.initial) return; // 非初始态不覆盖

    try {
      final resp = await HttpUtil.get(
        '$orderPayQueryUrl?orderId=${widget.orderId}',
      );
      if (!mounted) return;

      final data = resp.data;
      final result = PayQueryResult.fromResp(data);

      switch (result.orderStatus) {
        case 1: // 已支付 → Task 9.2
          setState(() => _pageState = PayPageState.success);
          break;
        case 2: // 已取消 / 超时 → Task 9.4
          setState(() {
            _pageState = PayPageState.cancelled;
            _payErrorMessage =
                result.message.isNotEmpty ? result.message : '订单已超时取消';
          });
          break;
        default:
          // 待支付 → Task 9.3：显示剩余时间 + 继续支付按钮
          setState(() {
            _expireTime = result.expireTime;
          });
          _startCountdown(result.expireTime);
      }
    } catch (e) {
      // 查询失败，保持初始态
    }
  }

  void _startCountdown(int seconds) {
    _countdownTimer?.cancel();
    setState(() => _expireTime = seconds);
    if (seconds <= 0) {
      if (!mounted) return;
      setState(() => _pageState = PayPageState.cancelled);
      return;
    }
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {
        _expireTime--;
        if (_expireTime <= 0) {
          timer.cancel();
          _pageState = PayPageState.cancelled;
          _payErrorMessage = '订单已超时取消';
        }
      });
    });
  }

  String _formatCountdown(int seconds) {
    if (seconds <= 0) return '00:00';
    final mins = seconds ~/ 60;
    final secs = seconds % 60;
    return '${mins.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  void _showToast(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), duration: const Duration(seconds: 2)),
    );
  }

  // ==================== 工具方法 ====================
  String _getPayTypeLabel(int payType) {
    switch (payType) {
      case 1:
        return '支付宝支付';
      case 2:
        return '微信支付';
      case 99:
        return '模拟支付（测试）';
      default:
        return '未知支付方式';
    }
  }

  // ==================== 构建 ====================
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: Colors.white,
        title: const Text('支付'),
        titleTextStyle: const TextStyle(fontSize: 16, color: Colors.black),
        centerTitle: true,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    switch (_pageState) {
      case PayPageState.initial:
        return _buildInitialState();
      case PayPageState.loading:
        return _buildLoadingState();
      case PayPageState.polling:
        return _buildPollingState();
      case PayPageState.success:
        return _buildSuccessState();
      case PayPageState.failed:
        return _buildFailedState();
      case PayPageState.cancelled:
        return _buildCancelledState();
    }
  }

  // ==================== Task 8/9/10: 各状态 UI ====================

  /// 初始态：支付选择
  Widget _buildInitialState() {
    final border = BorderSide(width: 1, color: const Color(0xFFF5F5F5));
    final textPrimary = const Color(0xFF303133);
    final textSecondary = const Color(0xFF909399);
    final themeColor = const Color(0xFFFA436A);

    return Container(
      color: Colors.white,
      width: MediaQuery.of(context).size.width,
      padding: const EdgeInsets.only(top: 44),
      child: Column(
        children: [
          // Task 9.3: 待支付状态 - 剩余时间提示
          if (_expireTime > 0) _buildExpireTimeBanner(),

          // 订单摘要卡片
          if (widget.orderSn != null && widget.orderSn!.isNotEmpty)
            _buildOrderSummaryCard(textPrimary, textSecondary, themeColor),

          const SizedBox(height: 30),

          // 支付金额
          _buildAmountDisplay(textPrimary, textSecondary),

          const SizedBox(height: 30),

          // 支付方式选择
          _buildPayTypeSelector(border, textPrimary, textSecondary, themeColor),

          // 确认支付按钮
          _buildPayButton(themeColor),

          if (widget.orderId == null || widget.orderId == 0)
            Padding(
              padding: const EdgeInsets.only(top: 16),
              child: Text(
                '如支付遇到问题，请在订单详情页重新发起支付',
                style: TextStyle(fontSize: 12, color: textSecondary),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildExpireTimeBanner() {
    final isUrgent = _expireTime <= 300; // ≤5分钟红色警告
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(vertical: 8),
      color: isUrgent ? Colors.red[50] : Colors.orange[50],
      child: Text(
        '剩余支付时间：${_formatCountdown(_expireTime)}',
        textAlign: TextAlign.center,
        style: TextStyle(
          fontSize: 13,
          color: isUrgent ? Colors.red : Colors.orange[800],
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }

  Widget _buildOrderSummaryCard(
      Color textPrimary, Color textSecondary, Color themeColor) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(horizontal: 16),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: themeColor.withAlpha(25),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: themeColor.withAlpha(51)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.receipt_long, size: 16, color: themeColor),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  '订单号：${widget.orderSn}',
                  style: TextStyle(
                      fontSize: 13,
                      color: textPrimary,
                      fontWeight: FontWeight.w500),
                ),
              ),
            ],
          ),
          if (widget.orderId != null && widget.orderId! > 0) ...[
            const SizedBox(height: 4),
            Text(
              '订单ID：${widget.orderId}',
              style: TextStyle(fontSize: 12, color: textSecondary),
            ),
          ],
          const SizedBox(height: 4),
          Row(
            children: [
              Icon(Icons.payment, size: 14, color: textSecondary),
              const SizedBox(width: 4),
              Text(
                '支付方式：${_getPayTypeLabel(widget.payType ?? 2)}',
                style: TextStyle(fontSize: 12, color: textSecondary),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAmountDisplay(Color textPrimary, Color textSecondary) {
    return Column(
      children: [
        Text('支付金额', style: TextStyle(fontSize: 12, color: textSecondary)),
        const SizedBox(height: 5),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text('￥', style: TextStyle(fontSize: 20, color: textPrimary)),
            Text(
              widget.amount.toString(),
              style: TextStyle(fontSize: 25, color: textPrimary),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildPayTypeSelector(BorderSide border, Color textPrimary,
      Color textSecondary, Color themeColor) {
    return Column(
      children: [
        // 微信
        Container(
          height: 60,
          margin: const EdgeInsets.only(left: 30),
          padding: const EdgeInsets.only(right: 18),
          decoration: BoxDecoration(border: Border(bottom: border)),
          child: Row(
            children: [
              Image.asset('images/wx_pay.png', height: 27, width: 26),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text('微信支付',
                        style: TextStyle(fontSize: 16, color: textPrimary)),
                    const SizedBox(height: 2),
                    Text('推荐使用微信支付',
                        style: TextStyle(fontSize: 12, color: textSecondary)),
                  ],
                ),
              ),
              Transform.scale(
                scale: 1.3,
                child: Checkbox(
                  value: _selectedPayType == 2,
                  activeColor: themeColor,
                  shape: const CircleBorder(),
                  side: const BorderSide(width: 1, color: Colors.black26),
                  onChanged: (value) {
                    if (value == true) setState(() => _selectedPayType = 2);
                  },
                ),
              ),
            ],
          ),
        ),
        // 支付宝
        Container(
          height: 60,
          margin: const EdgeInsets.only(left: 30),
          padding: const EdgeInsets.only(right: 18),
          decoration: BoxDecoration(border: Border(bottom: border)),
          child: Row(
            children: [
              Image.asset('images/ali_pay.png', height: 27, width: 26),
              const SizedBox(width: 16),
              Expanded(
                child: Text('支付宝支付',
                    style: TextStyle(fontSize: 16, color: textPrimary)),
              ),
              Transform.scale(
                scale: 1.3,
                child: Checkbox(
                  value: _selectedPayType == 1,
                  activeColor: themeColor,
                  shape: const CircleBorder(),
                  side: const BorderSide(width: 1, color: Colors.black26),
                  onChanged: (value) {
                    if (value == true) setState(() => _selectedPayType = 1);
                  },
                ),
              ),
            ],
          ),
        ),
        // 模拟支付（测试专用）
        Container(
          height: 60,
          margin: const EdgeInsets.only(left: 30, bottom: 30),
          padding: const EdgeInsets.only(right: 18),
          decoration: BoxDecoration(border: Border(bottom: border)),
          child: Row(
            children: [
              Icon(Icons.science, size: 26, color: Colors.orange[700]),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text('模拟支付（测试）',
                        style: TextStyle(fontSize: 16, color: textPrimary)),
                    const SizedBox(height: 2),
                    Text('跳过真实支付，直接完成订单',
                        style: TextStyle(fontSize: 12, color: textSecondary)),
                  ],
                ),
              ),
              Transform.scale(
                scale: 1.3,
                child: Checkbox(
                  value: _selectedPayType == 99,
                  activeColor: Colors.orange[700],
                  shape: const CircleBorder(),
                  side: const BorderSide(width: 1, color: Colors.black26),
                  onChanged: (value) {
                    if (value == true) setState(() => _selectedPayType = 99);
                  },
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildPayButton(Color themeColor) {
    return GestureDetector(
      onTap: _doPay,
      child: Container(
        alignment: Alignment.center,
        width: MediaQuery.of(context).size.width,
        height: 40,
        margin: const EdgeInsets.only(top: 15, left: 30, right: 30),
        decoration: BoxDecoration(
          color: themeColor,
          borderRadius: BorderRadius.circular(5),
        ),
        child: const Text(
          '确认支付',
          style: TextStyle(color: Colors.white, fontSize: 16),
        ),
      ),
    );
  }

  /// Task 8.3/Task 7.4: 轮询中态
  Widget _buildPollingState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(
            width: 48,
            height: 48,
            child: CircularProgressIndicator(strokeWidth: 3),
          ),
          const SizedBox(height: 24),
          const Text(
            '支付确认中...',
            style: TextStyle(fontSize: 16, color: Color(0xFF303133)),
          ),
          const SizedBox(height: 8),
          Text(
            '$_pollCount / 100 次查询',
            style: const TextStyle(fontSize: 13, color: Color(0xFF909399)),
          ),
          if (_expireTime > 0) ...[
            const SizedBox(height: 16),
            Text(
              '剩余支付时间：${_formatCountdown(_expireTime)}',
              style: const TextStyle(fontSize: 13, color: Color(0xFFFA436A)),
            ),
          ],
          const SizedBox(height: 8),
          const Text(
            '请在支付 App 中完成付款后返回',
            style: TextStyle(fontSize: 12, color: Color(0xFF909399)),
          ),
        ],
      ),
    );
  }

  /// Task 8.1: 支付成功态
  /// Story 10.6: 增加提货卡履约提示
  /// UX 优化: 添加动画效果、改进颜色对比度、优化布局
  Widget _buildSuccessState() {
    return Container(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [
            Color(0xFFF0FFF4), // 浅绿色渐变开始
            Color(0xFFFFFFFF), // 白色渐变结束
          ],
        ),
      ),
      child: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // 成功图标 - 带缩放动画
              TweenAnimationBuilder<double>(
                tween: Tween(begin: 0.0, end: 1.0),
                duration: const Duration(milliseconds: 600),
                curve: Curves.elasticOut,
                builder: (context, value, child) {
                  return Transform.scale(
                    scale: value,
                    child: child,
                  );
                },
                child: Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    color: const Color(0xFF52C41A),
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF52C41A).withValues(alpha: 0.3),
                        blurRadius: 20,
                        offset: const Offset(0, 8),
                      ),
                    ],
                  ),
                  child: const Icon(
                    Icons.check_rounded,
                    color: Colors.white,
                    size: 56,
                  ),
                ),
              ),
              const SizedBox(height: 32),
              
              // 支付成功标题
              const Text(
                '支付成功',
                style: TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w700,
                  color: Color(0xFF1E293B), // 深色，对比度更高
                ),
              ),
              const SizedBox(height: 12),
              
              // 订单号
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  '订单号：${widget.orderSn ?? '-'}',
                  style: const TextStyle(
                    fontSize: 14,
                    color: Color(0xFF475569),
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              const SizedBox(height: 32),

              // Story 10.7 Review Fix HIGH-1: 只有订单包含提货卡履约时才展示"提货卡已发放"
              if (widget.hasDigitalCard) ...[
                Container(
                  margin: const EdgeInsets.symmetric(horizontal: 8),
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        Color(0xFFEFF6FF),
                        Color(0xFFDBEAFE),
                      ],
                    ),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: const Color(0xFFBFDBFE),
                      width: 1.5,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF3B82F6).withValues(alpha: 0.1),
                        blurRadius: 12,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: Column(
                    children: [
                      Container(
                        width: 56,
                        height: 56,
                        decoration: BoxDecoration(
                          color: const Color(0xFF3B82F6).withValues(alpha: 0.1),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.card_giftcard_rounded,
                          size: 28,
                          color: Color(0xFF2563EB),
                        ),
                      ),
                      const SizedBox(height: 16),
                      const Text(
                        '提货卡已发放',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF1E40AF),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '您的提货卡已自动发放至卡包\n无需手动操作，可随时查看',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 14,
                          color: const Color(0xFF1E40AF).withValues(alpha: 0.8),
                          height: 1.5,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 32),

                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: OutlinedButton.icon(
                    onPressed: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => const MyDigitalCardPage(),
                        ),
                      );
                    },
                    icon: const Icon(
                      Icons.card_giftcard_rounded,
                      size: 22,
                    ),
                    label: const Text(
                      '查看我的提货卡',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: const Color(0xFF2563EB),
                      backgroundColor: Colors.white,
                      side: const BorderSide(
                        color: Color(0xFF93C5FD),
                        width: 1.5,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
                const SizedBox(height: 16),
              ],
              
              // 查看订单详情按钮 - 优化样式
              SizedBox(
                width: double.infinity,
                height: 52,
                child: ElevatedButton(
                  onPressed: () {
                    // 跳转订单详情
                    Navigator.of(context).pop();
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF52C41A),
                    foregroundColor: Colors.white,
                    elevation: 4,
                    shadowColor: const Color(0xFF52C41A).withValues(alpha: 0.4),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: const Text(
                    '查看订单详情',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              
              // 底部提示
              Text(
                '如有疑问，请联系客服',
                style: TextStyle(
                  fontSize: 13,
                  color: const Color(0xFF94A3B8),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Task 8.2: 支付失败态
  Widget _buildFailedState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              color: Colors.red[50],
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.close, color: Colors.red[400], size: 48),
          ),
          const SizedBox(height: 24),
          const Text(
            '支付失败',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: Color(0xFFFF4D4F),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            _payErrorMessage ?? '支付失败，请重试',
            style: const TextStyle(fontSize: 13, color: Color(0xFF909399)),
          ),
          const SizedBox(height: 32),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              OutlinedButton(
                onPressed: () {
                  Navigator.of(context).pop(); // 返回订单详情
                },
                child: const Text('返回订单详情'),
              ),
              const SizedBox(width: 16),
              ElevatedButton(
                onPressed: () {
                  setState(() => _pageState = PayPageState.initial);
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFA436A),
                  foregroundColor: Colors.white,
                ),
                child: const Text('重新支付'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// Task 9.4: 已取消 / 超时态
  Widget _buildCancelledState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              color: Colors.grey[100],
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.access_time, color: Colors.grey[400], size: 48),
          ),
          const SizedBox(height: 24),
          const Text(
            '订单已超时取消',
            style: TextStyle(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: Color(0xFF909399),
            ),
          ),
          const SizedBox(height: 8),
          Text(
            _payErrorMessage ?? '支付时间已超时，订单已自动关闭',
            style: const TextStyle(fontSize: 13, color: Color(0xFF909399)),
          ),
          const SizedBox(height: 32),
          ElevatedButton(
            onPressed: () {
              Navigator.of(context).pop(); // 返回订单列表
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF909399),
              foregroundColor: Colors.white,
              minimumSize: const Size(200, 40),
            ),
            child: const Text('返回订单列表'),
          ),
        ],
      ),
    );
  }

  /// Task 10.1: 加载中遮罩
  Widget _buildLoadingState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(
            width: 48,
            height: 48,
            child: CircularProgressIndicator(strokeWidth: 3),
          ),
          const SizedBox(height: 24),
          Text(
            '正在发起${_getPayTypeLabel(_selectedPayType)}...',
            style: const TextStyle(fontSize: 15, color: Color(0xFF303133)),
          ),
        ],
      ),
    );
  }
}
