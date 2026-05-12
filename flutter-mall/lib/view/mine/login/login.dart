import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_router.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';

import '../../../model/login_model.dart';

///
/// 登录页面
///
/// 作者：David
/// 日期：2023/11/21 17:17
///
class Login extends StatefulWidget {
  final String? redirectRoute;
  final AppRecentContext? recoveryIntent;

  const Login({super.key, this.redirectRoute, this.recoveryIntent});

  @override
  State<Login> createState() => _LoginState();
}

class _LoginState extends State<Login> {
  // Story 3.1.1: 验证码合并登录注册
  //   - 手机号正则与后端 ums-rpc 保持一致：^1[3-9]\d{9}$
  //   - 验证码 6 位纯数字（AC-12 硬约束；mock provider 固定下发 "123456"）
  //   - 60 秒发送冷却窗口，前后端双向 enforce
  static final RegExp _mobileRegExp = RegExp(r'^1[3-9]\d{9}$');
  static final RegExp _smsCodeRegExp = RegExp(r'^\d{6}$');
  static const int _smsCooldownSeconds = 60;

  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _smsCodeController = TextEditingController();
  final FocusNode _mobileFocusNode = FocusNode();
  final FocusNode _smsCodeFocusNode = FocusNode();

  bool _isLoading = false;
  bool _isSendingCode = false;
  int _cooldownRemaining = 0;
  Timer? _cooldownTimer;
  String? _submitError;

  @override
  void dispose() {
    _cooldownTimer?.cancel();
    _usernameController.dispose();
    _smsCodeController.dispose();
    _mobileFocusNode.dispose();
    _smsCodeFocusNode.dispose();
    super.dispose();
  }

  String? _validateMobile(String? value) {
    final mobile = value?.trim() ?? '';
    if (mobile.isEmpty) {
      return '请输入手机号';
    }
    if (!_mobileRegExp.hasMatch(mobile)) {
      return '请输入正确的 11 位手机号';
    }
    return null;
  }

  String? _validateSmsCode(String? value) {
    final code = value?.trim() ?? '';
    if (code.isEmpty) {
      return '请输入验证码';
    }
    if (!_smsCodeRegExp.hasMatch(code)) {
      return '请输入 6 位验证码';
    }
    return null;
  }

  void _startCooldown() {
    _cooldownTimer?.cancel();
    setState(() => _cooldownRemaining = _smsCooldownSeconds);
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      if (_cooldownRemaining <= 1) {
        timer.cancel();
        setState(() => _cooldownRemaining = 0);
        return;
      }
      setState(() => _cooldownRemaining -= 1);
    });
  }

  Future<void> _sendSmsCode() async {
    if (_isSendingCode || _cooldownRemaining > 0) {
      return;
    }
    final mobile = _usernameController.text.trim();
    final mobileError = _validateMobile(mobile);
    if (mobileError != null) {
      setState(() => _submitError = mobileError);
      _mobileFocusNode.requestFocus();
      return;
    }

    setState(() {
      _isSendingCode = true;
      _submitError = null;
    });

    try {
      final Response result = await HttpUtil.post(
        sendSmsCodeUrl,
        data: {'mobile': mobile, 'scene': 1},
      );
      final data = result.data;
      final int code = (data is Map && data['code'] is int) ? data['code'] : -1;
      if (code == 0) {
        // UX: 发送成功给予轻量触觉反馈 + 自动聚焦验证码输入框
        HapticFeedback.lightImpact();
        _startCooldown();
        _smsCodeFocusNode.requestFocus();
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('验证码已发送，请注意查收'),
              duration: Duration(seconds: 2),
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      } else {
        final String msg = (data is Map ? data['message']?.toString() : null) ??
            '验证码发送失败，请稍后重试';
        if (mounted) setState(() => _submitError = msg);
      }
    } on DioException catch (e) {
      String msg = '验证码发送失败，请稍后重试';
      if (e.response?.data is Map) {
        msg = (e.response?.data as Map)['message']?.toString() ?? msg;
      }
      if (mounted) setState(() => _submitError = msg);
    } catch (_) {
      if (mounted) setState(() => _submitError = '验证码发送失败，请检查网络');
    } finally {
      if (mounted) setState(() => _isSendingCode = false);
    }
  }

  void _clearSubmitError() {
    if (_submitError == null) {
      return;
    }
    setState(() => _submitError = null);
  }

  String? _resolveRecoveryHint() {
    final intent = widget.recoveryIntent;
    if (intent != null) {
      final recoveryHint = intent.recoveryHint.trim();
      if (recoveryHint.isNotEmpty) {
        return recoveryHint;
      }
      switch (intent.targetType) {
        case AppRecentTargetType.orderDetail:
          return '登录后将继续查看订单详情。';
        case AppRecentTargetType.orderList:
          return '登录后将继续回到订单列表。';
        case AppRecentTargetType.cart:
          return '登录后将继续访问购物车。';
        case AppRecentTargetType.productDetail:
          return '登录后将继续查看商品详情。';
        case AppRecentTargetType.couponList:
          return '登录后将继续查看优惠券资产。';
        case AppRecentTargetType.couponCenter:
          return '登录后将继续进入领券中心。';
        case AppRecentTargetType.afterSalesApply:
          return '登录后将继续填写售后申请。';
        case AppRecentTargetType.commentCompose:
          return '登录后将继续完成评价内容。';
        case AppRecentTargetType.settings:
          return '登录后将继续访问账户设置。';
        case AppRecentTargetType.digitalCardAssetList:
          return '登录后将继续查看我的提货卡。';
        case AppRecentTargetType.digitalCardAssetDetail:
          return '登录后将继续查看提货卡详情。';
        case AppRecentTargetType.digitalCardClaim:
          return '登录后将继续领取分享的提货卡。';
        case AppRecentTargetType.home:
        case AppRecentTargetType.activity:
        case AppRecentTargetType.subject:
        case AppRecentTargetType.preferredArea:
          return '登录后将继续返回你刚才的浏览场景。';
      }
    }

    if ((widget.redirectRoute ?? '').trim().isNotEmpty) {
      return '登录成功后会自动跳转回上一目标页面。';
    }

    return null;
  }

  Future<void> _submitLoginData() async {
    FocusScope.of(context).unfocus();

    if (_isLoading) {
      return;
    }

    final formState = _formKey.currentState;
    if (formState == null || !formState.validate()) {
      return;
    }

    setState(() {
      _isLoading = true;
      _submitError = null;
    });

    try {
      final Map<String, dynamic> loginMap = <String, dynamic>{
        'mobile': _usernameController.value.text.trim(),
        'code': _smsCodeController.value.text.trim(),
      };

      final Response result = await HttpUtil.post(smsLoginUrl, data: loginMap);
      final LoginModel loginModel = LoginModel.fromJson(result.data);

      if (loginModel.code == 0) {
        // UX: 登录成功中等强度触觉反馈，与安全操作一致
        HapticFeedback.mediumImpact();

        // 提交 AutofillContext，提示系统记住成功凭据
        TextInput.finishAutofillContext();

        final authToken =
            "${loginModel.data.tokenHead} ${loginModel.data.token}";
        await AppRecoveryStore.persistAuthToken(authToken);

        if (!mounted) {
          return;
        }

        // 新用户首次登录时给出差异化提示（仅 SnackBar，不阻断路由恢复）
        if (loginModel.data.isNewUser) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('账号创建成功，欢迎加入九克城'),
              duration: Duration(seconds: 2),
              behavior: SnackBarBehavior.floating,
            ),
          );
          // 给 SnackBar 一个短暂的可见窗口，避免被后续 Navigator.pushAndRemoveUntil 立即覆盖。
          await Future<void>.delayed(const Duration(milliseconds: 300));
          if (!mounted) return;
        }

        final pendingIntent = await AppRecoveryStore.consumePendingIntent();
        if (!mounted) {
          return;
        }
        final restoreIntent = pendingIntent ?? widget.recoveryIntent;
        await AppRecoveryRouter.restoreAfterLogin(
          context,
          recoveryIntent: restoreIntent,
          redirectRoute: widget.redirectRoute,
        );
        return;
      }

      if (!mounted) {
        return;
      }
      setState(() {
        _submitError = loginModel.message.trim().isNotEmpty
            ? loginModel.message
            : '登录失败，请检查手机号与验证码';
      });
    } on DioException catch (e) {
      if (!mounted) {
        return;
      }
      String msg = '登录失败，请稍后重试';
      if (e.response?.data is Map) {
        msg = (e.response?.data as Map)['message']?.toString() ?? msg;
      }
      setState(() => _submitError = msg);
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() => _submitError = '登录失败，请检查网络连接后重试');
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  InputDecoration _buildInputDecoration({
    required BuildContext context,
    required String label,
    required String hint,
    required IconData icon,
    Widget? suffixIcon,
  }) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final baseBorder = OutlineInputBorder(
      borderRadius: BorderRadius.circular(AppRadii.lg),
      borderSide: const BorderSide(color: AppColors.border),
    );

    return InputDecoration(
      labelText: label,
      hintText: hint,
      prefixIcon: Icon(icon, color: AppColors.textHint),
      suffixIcon: suffixIcon,
      filled: true,
      fillColor: AppColors.surfaceMuted,
      contentPadding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.lg,
        vertical: 18,
      ),
      border: baseBorder,
      enabledBorder: baseBorder,
      focusedBorder: baseBorder.copyWith(
        borderSide: const BorderSide(
          color: AppColors.primary,
          width: 1.5,
        ),
      ),
      errorBorder: baseBorder.copyWith(
        borderSide: BorderSide(color: colorScheme.error),
      ),
      focusedErrorBorder: baseBorder.copyWith(
        borderSide: BorderSide(
          color: colorScheme.error,
          width: 1.5,
        ),
      ),
      floatingLabelStyle: theme.textTheme.labelLarge?.copyWith(
        color: AppColors.primary,
      ),
      hintStyle: theme.textTheme.bodyMedium?.copyWith(
        color: AppColors.textHint,
      ),
    );
  }

  Widget _buildRecoveryBanner(BuildContext context, String message) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: AppColors.primarySoft,
        borderRadius: BorderRadius.circular(AppRadii.lg),
        border: Border.all(
          color: AppColors.primary.withValues(alpha: 0.14),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(AppRadii.md),
            ),
            child: const Icon(
              Icons.lock_person_rounded,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '继续刚才的操作',
                  style: theme.textTheme.titleSmall,
                ),
                const SizedBox(height: AppSpacing.xs),
                Text(
                  message,
                  style: theme.textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                    height: 1.5,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // Story 3.1.1: 旧的「跳转独立注册页」入口已下线，验证码合并接口在 _submitLoginData
  // 中自动建号，无需单独的 Register 页面。

  Future<void> _goToHome() async {
    await AppRecoveryStore.clearPendingIntent();
    await AppRecoveryStore.clearActiveIntentCandidate();
    await AppRecoveryStore.clearPendingUpgradeContext();
    if (!mounted) {
      return;
    }
    await Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(
        builder: (_) => MainTab(intentSource: widget.recoveryIntent?.source),
      ),
      (route) => false,
    );
  }

  @override
  Widget build(BuildContext context) {
    final ThemeData theme = Theme.of(context);
    final TextTheme textTheme = theme.textTheme;
    final String? recoveryHint = _resolveRecoveryHint();
    final double bottomInset = MediaQuery.viewInsetsOf(context).bottom;

    return Scaffold(
      backgroundColor: AppColors.background,
      body: GestureDetector(
        onTap: () => FocusScope.of(context).unfocus(),
        child: DecoratedBox(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Color(0xFFFFF4F6),
                AppColors.background,
                Colors.white,
              ],
              stops: [0.0, 0.48, 1.0],
            ),
          ),
          child: SafeArea(
            child: LayoutBuilder(
              builder: (context, constraints) {
                return SingleChildScrollView(
                  keyboardDismissBehavior:
                      ScrollViewKeyboardDismissBehavior.onDrag,
                  padding: EdgeInsets.fromLTRB(
                    AppSpacing.xl,
                    AppSpacing.lg,
                    AppSpacing.xl,
                    bottomInset + AppSpacing.xxl,
                  ),
                  child: ConstrainedBox(
                    constraints: BoxConstraints(
                      minHeight: constraints.maxHeight - AppSpacing.lg,
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            if (Navigator.of(context).canPop())
                              Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  borderRadius:
                                      BorderRadius.circular(AppRadii.md),
                                  onTap: () => Navigator.of(context).maybePop(),
                                  child: Ink(
                                    width: 44,
                                    height: 44,
                                    decoration: BoxDecoration(
                                      color: Colors.white,
                                      borderRadius:
                                          BorderRadius.circular(AppRadii.md),
                                      border: Border.all(
                                        color: AppColors.border,
                                      ),
                                    ),
                                    child: const Icon(
                                      Icons.arrow_back_ios_new_rounded,
                                      size: 18,
                                      color: AppColors.textPrimary,
                                    ),
                                  ),
                                ),
                              )
                            else
                              const SizedBox(width: 44, height: 44),
                            const Spacer(),
                            TextButton.icon(
                              onPressed: _goToHome,
                              style: TextButton.styleFrom(
                                minimumSize: const Size(0, 44),
                                padding: const EdgeInsets.symmetric(
                                  horizontal: AppSpacing.md,
                                  vertical: AppSpacing.sm,
                                ),
                                foregroundColor: AppColors.textPrimary,
                                backgroundColor: Colors.white,
                                shape: RoundedRectangleBorder(
                                  borderRadius:
                                      BorderRadius.circular(AppRadii.md),
                                  side: const BorderSide(
                                    color: AppColors.border,
                                  ),
                                ),
                              ),
                              icon: const Icon(
                                Icons.home_outlined,
                                size: 18,
                              ),
                              label: Text(
                                '返回首页',
                                style: textTheme.labelLarge?.copyWith(
                                  color: AppColors.textPrimary,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: AppSpacing.md),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: AppSpacing.lg,
                            vertical: AppSpacing.lg,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.92),
                            borderRadius: BorderRadius.circular(AppRadii.xl),
                            border: Border.all(
                              color: AppColors.border.withValues(alpha: 0.88),
                            ),
                          ),
                          child: Row(
                            children: [
                              Container(
                                width: 42,
                                height: 42,
                                padding: const EdgeInsets.all(9),
                                decoration: BoxDecoration(
                                  color: AppColors.primarySoft,
                                  borderRadius:
                                      BorderRadius.circular(AppRadii.md),
                                ),
                                child: Image.asset(
                                  'images/icon_main_logo.png',
                                  fit: BoxFit.contain,
                                ),
                              ),
                              const SizedBox(width: AppSpacing.md),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '手机号登录',
                                      style: textTheme.titleLarge,
                                    ),
                                    const SizedBox(height: AppSpacing.xs),
                                    Text(
                                      '输入手机号即可登录，未注册手机号将自动创建账号。',
                                      style: textTheme.bodySmall?.copyWith(
                                        color: AppColors.textSecondary,
                                        height: 1.45,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: AppSpacing.sm,
                                  vertical: 6,
                                ),
                                decoration: BoxDecoration(
                                  color: AppColors.primarySoft,
                                  borderRadius:
                                      BorderRadius.circular(AppRadii.sm),
                                ),
                                child: Text(
                                  '会员中心',
                                  style: textTheme.labelMedium?.copyWith(
                                    color: AppColors.primaryDark,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: AppSpacing.md),
                        Container(
                          padding: const EdgeInsets.all(AppSpacing.xl),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(AppRadii.xxl),
                            border: Border.all(
                              color: AppColors.border.withValues(alpha: 0.82),
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: AppColors.textPrimary.withValues(
                                  alpha: 0.06,
                                ),
                                blurRadius: 24,
                                offset: const Offset(0, 12),
                              ),
                            ],
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                '手机号验证码登录',
                                style: textTheme.titleLarge,
                              ),
                              if (recoveryHint != null) ...[
                                const SizedBox(height: AppSpacing.md),
                                _buildRecoveryBanner(context, recoveryHint),
                              ],
                              const SizedBox(height: AppSpacing.lg),
                              AutofillGroup(
                                child: Form(
                                  key: _formKey,
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.stretch,
                                    children: [
                                      TextFormField(
                                        controller: _usernameController,
                                        focusNode: _mobileFocusNode,
                                        keyboardType: TextInputType.phone,
                                        textInputAction: TextInputAction.next,
                                        maxLength: 11,
                                        inputFormatters: [
                                          FilteringTextInputFormatter
                                              .digitsOnly,
                                          LengthLimitingTextInputFormatter(11),
                                        ],
                                        autofillHints: const [
                                          AutofillHints.telephoneNumber,
                                        ],
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) {
                                          _smsCodeFocusNode.requestFocus();
                                        },
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '手机号',
                                          hint: '请输入 11 位手机号',
                                          icon: Icons.phone_iphone_rounded,
                                        ).copyWith(counterText: ''),
                                        validator: _validateMobile,
                                      ),
                                      const SizedBox(height: AppSpacing.lg),
                                      // Story 3.1.1 + UX Review: 验证码输入 + 「获取验证码」按钮
                                      //   - autofillHints.oneTimeCode: 支持 iOS/Android 系统自动从 SMS 填充验证码
                                      //   - 按钮高度 ≥ 44px（UX touch target 规范）
                                      //   - disabled 状态显式设色，倒计时期间视觉明显
                                      TextFormField(
                                        controller: _smsCodeController,
                                        focusNode: _smsCodeFocusNode,
                                        keyboardType: TextInputType.number,
                                        textInputAction: TextInputAction.done,
                                        maxLength: 6,
                                        inputFormatters: [
                                          FilteringTextInputFormatter
                                              .digitsOnly,
                                          LengthLimitingTextInputFormatter(6),
                                        ],
                                        autofillHints: const [
                                          AutofillHints.oneTimeCode,
                                        ],
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) =>
                                            _submitLoginData(),
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '验证码',
                                          hint: '请输入短信验证码',
                                          icon: Icons.message_outlined,
                                          suffixIcon: Padding(
                                            padding: const EdgeInsets.only(
                                              right: AppSpacing.xs,
                                            ),
                                            child: TextButton(
                                              onPressed: (_isSendingCode ||
                                                      _cooldownRemaining > 0)
                                                  ? null
                                                  : _sendSmsCode,
                                              style: TextButton.styleFrom(
                                                minimumSize:
                                                    const Size(108, 44),
                                                padding: const EdgeInsets
                                                    .symmetric(
                                                  horizontal: AppSpacing.sm,
                                                ),
                                                foregroundColor:
                                                    AppColors.primary,
                                                disabledForegroundColor:
                                                    AppColors.textHint,
                                                shape: RoundedRectangleBorder(
                                                  borderRadius:
                                                      BorderRadius.circular(
                                                    AppRadii.md,
                                                  ),
                                                ),
                                              ),
                                              child: _isSendingCode
                                                  ? const SizedBox(
                                                      width: 18,
                                                      height: 18,
                                                      child:
                                                          CircularProgressIndicator(
                                                        strokeWidth: 2,
                                                      ),
                                                    )
                                                  : Text(
                                                      _cooldownRemaining > 0
                                                          ? '${_cooldownRemaining}s 后重发'
                                                          : '获取验证码',
                                                      style: const TextStyle(
                                                        fontWeight:
                                                            FontWeight.w600,
                                                      ),
                                                    ),
                                            ),
                                          ),
                                        ).copyWith(counterText: ''),
                                        validator: _validateSmsCode,
                                      ),
                                      const SizedBox(height: AppSpacing.md),
                                      Row(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          const Padding(
                                            padding: EdgeInsets.only(
                                              top: 1,
                                            ),
                                            child: Icon(
                                              Icons.info_outline_rounded,
                                              size: 16,
                                              color: AppColors.textHint,
                                            ),
                                          ),
                                          const SizedBox(width: AppSpacing.sm),
                                          Expanded(
                                            child: Text(
                                              '未注册手机号将自动创建账号并登录。验证码 5 分钟内有效，60 秒内不能重复获取。',
                                              style:
                                                  textTheme.bodySmall?.copyWith(
                                                color: AppColors.textHint,
                                                height: 1.45,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                      AnimatedSwitcher(
                                        duration:
                                            const Duration(milliseconds: 180),
                                        child: _submitError == null
                                            ? const SizedBox(
                                                key: ValueKey('submit-empty'),
                                                height: 0,
                                              )
                                            : Padding(
                                                key: ValueKey(_submitError),
                                                padding: const EdgeInsets.only(
                                                  top: AppSpacing.lg,
                                                ),
                                                child: Semantics(
                                                  liveRegion: true,
                                                  child: Container(
                                                    padding:
                                                        const EdgeInsets.all(
                                                      AppSpacing.lg,
                                                    ),
                                                    decoration: BoxDecoration(
                                                      color: theme.colorScheme
                                                          .errorContainer,
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                        AppRadii.lg,
                                                      ),
                                                      border: Border.all(
                                                        color: theme
                                                            .colorScheme.error
                                                            .withValues(
                                                          alpha: 0.18,
                                                        ),
                                                      ),
                                                    ),
                                                    child: Row(
                                                      crossAxisAlignment:
                                                          CrossAxisAlignment
                                                              .start,
                                                      children: [
                                                        Icon(
                                                          Icons.error_outline,
                                                          size: 18,
                                                          color: theme
                                                              .colorScheme
                                                              .onErrorContainer,
                                                        ),
                                                        const SizedBox(
                                                          width: AppSpacing.sm,
                                                        ),
                                                        Expanded(
                                                          child: Text(
                                                            _submitError!,
                                                            style: textTheme
                                                                .bodySmall
                                                                ?.copyWith(
                                                              color: theme
                                                                  .colorScheme
                                                                  .onErrorContainer,
                                                              height: 1.5,
                                                            ),
                                                          ),
                                                        ),
                                                      ],
                                                    ),
                                                  ),
                                                ),
                                              ),
                                      ),
                                      const SizedBox(height: AppSpacing.xl),
                                      FilledButton(
                                        onPressed: _isLoading
                                            ? null
                                            : _submitLoginData,
                                        style: FilledButton.styleFrom(
                                          minimumSize:
                                              const Size(double.infinity, 56),
                                          backgroundColor: AppColors.primary,
                                          disabledBackgroundColor:
                                              AppColors.textHint.withValues(
                                            alpha: 0.35,
                                          ),
                                          foregroundColor: Colors.white,
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              AppRadii.lg,
                                            ),
                                          ),
                                          textStyle:
                                              textTheme.titleSmall?.copyWith(
                                            color: Colors.white,
                                          ),
                                        ),
                                        child: _isLoading
                                            ? const SizedBox(
                                                width: 22,
                                                height: 22,
                                                child:
                                                    CircularProgressIndicator(
                                                  strokeWidth: 2.2,
                                                  color: Colors.white,
                                                ),
                                              )
                                            : const Text('登录并继续'),
                                      ),
                                      const SizedBox(height: AppSpacing.md),
                                      // Story 3.1.1: 旧的「新用户注册」按钮下线—
                                      // 验证码接口自动处理未注册手机号。
                                      Center(
                                        child: Text(
                                          '登录后可同步订单、优惠券、收货地址与售后进度',
                                          textAlign: TextAlign.center,
                                          style: textTheme.bodySmall?.copyWith(
                                            color: AppColors.textHint,
                                            height: 1.45,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
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
          ),
        ),
      ),
    );
  }
}
