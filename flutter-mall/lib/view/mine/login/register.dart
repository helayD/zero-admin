import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_mall/config/service_url.dart';
import 'package:flutter_mall/layout/main_tab.dart';
import 'package:flutter_mall/model/app_recent_context.dart';
import 'package:flutter_mall/theme/app_theme.dart';
import 'package:flutter_mall/utils/app_recovery_store.dart';
import 'package:flutter_mall/utils/http_util.dart';

import '../../../model/login_model.dart';

///
/// 注册页面
///
/// 日期：2026/03/24
///
class Register extends StatefulWidget {
  final String? redirectRoute;
  final AppRecentContext? recoveryIntent;

  const Register({
    super.key,
    this.redirectRoute,
    this.recoveryIntent,
  });

  @override
  State<Register> createState() => _RegisterState();
}

class _RegisterState extends State<Register> {
  static final RegExp _mobileRegExp = RegExp(r'^1[3-9]\d{9}$');

  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  final TextEditingController _mobileController = TextEditingController();
  final TextEditingController _nicknameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _confirmPasswordController =
      TextEditingController();
  final FocusNode _mobileFocusNode = FocusNode();
  final FocusNode _nicknameFocusNode = FocusNode();
  final FocusNode _passwordFocusNode = FocusNode();
  final FocusNode _confirmPasswordFocusNode = FocusNode();

  bool _isLoading = false;
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  String? _submitError;

  @override
  void dispose() {
    _mobileController.dispose();
    _nicknameController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _mobileFocusNode.dispose();
    _nicknameFocusNode.dispose();
    _passwordFocusNode.dispose();
    _confirmPasswordFocusNode.dispose();
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

  String? _validateNickname(String? value) {
    final nickname = value?.trim() ?? '';
    if (nickname.isEmpty) {
      return '请输入昵称';
    }
    return null;
  }

  String? _validatePassword(String? value) {
    final password = value ?? '';
    if (password.isEmpty) {
      return '请输入密码';
    }
    if (password.length < 6) {
      return '密码长度不能少于 6 位';
    }
    return null;
  }

  String? _validateConfirmPassword(String? value) {
    final confirmPassword = value ?? '';
    if (confirmPassword.isEmpty) {
      return '请再次输入密码';
    }
    if (confirmPassword != _passwordController.value.text) {
      return '两次输入的密码不一致';
    }
    return null;
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
        return recoveryHint.startsWith('登录后')
            ? recoveryHint.replaceFirst('登录后', '注册完成后')
            : recoveryHint;
      }
      switch (intent.targetType) {
        case AppRecentTargetType.orderDetail:
          return '注册完成后将继续查看订单详情。';
        case AppRecentTargetType.orderList:
          return '注册完成后将继续回到订单列表。';
        case AppRecentTargetType.cart:
          return '注册完成后将继续访问购物车。';
        case AppRecentTargetType.productDetail:
          return '注册完成后将继续查看商品详情。';
        case AppRecentTargetType.couponList:
          return '注册完成后将继续查看优惠券资产。';
        case AppRecentTargetType.couponCenter:
          return '注册完成后将继续进入领券中心。';
        case AppRecentTargetType.afterSalesApply:
          return '注册完成后将继续填写售后申请。';
        case AppRecentTargetType.commentCompose:
          return '注册完成后将继续完成评价内容。';
        case AppRecentTargetType.settings:
          return '注册完成后将继续访问账户设置。';
        case AppRecentTargetType.digitalCardAssetList:
          return '注册完成后将继续查看我的数字卡片。';
        case AppRecentTargetType.digitalCardAssetDetail:
          return '注册完成后将继续查看数字卡片详情。';
        case AppRecentTargetType.home:
        case AppRecentTargetType.activity:
        case AppRecentTargetType.subject:
        case AppRecentTargetType.preferredArea:
          return '注册完成后将继续返回你刚才的浏览场景。';
      }
    }

    if ((widget.redirectRoute ?? '').trim().isNotEmpty) {
      return '注册成功后会自动跳回你刚才的目标页面。';
    }

    return null;
  }

  Future<void> _submitRegisterData() async {
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
      final Map<String, dynamic> registerMap = <String, dynamic>{
        'mobile': _mobileController.value.text.trim(),
        'nickname': _nicknameController.value.text.trim(),
        'password': _passwordController.value.text,
        'confirmPassword': _confirmPasswordController.value.text,
        'source': 1,
      };

      final Response result =
          await HttpUtil.post(registerDataUrl, data: registerMap);
      final LoginModel loginModel = LoginModel.fromJson(result.data);

      if (loginModel.code == 0) {
        final authToken =
            '${loginModel.data.tokenHead} ${loginModel.data.token}';
        await AppRecoveryStore.persistAuthToken(authToken);

        if (!mounted) {
          return;
        }
        Navigator.of(context).pop(true);
        return;
      }

      if (!mounted) {
        return;
      }
      setState(() {
        _submitError = loginModel.message.trim().isNotEmpty
            ? loginModel.message
            : '注册失败，请核对信息后重试';
      });
    } on DioException catch (e) {
      if (!mounted) {
        return;
      }
      String msg = '注册失败，请稍后重试';
      if (e.response?.data is Map) {
        msg = (e.response?.data as Map)['message']?.toString() ?? msg;
      }
      setState(() => _submitError = msg);
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() => _submitError = '注册失败，请检查网络连接后重试');
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

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
              Icons.person_add_alt_1_rounded,
              color: AppColors.primary,
            ),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '注册后继续当前操作',
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
                            ),
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
                                      '创建账号',
                                      style: textTheme.titleLarge,
                                    ),
                                    const SizedBox(height: AppSpacing.xs),
                                    Text(
                                      '注册后会自动登录，并继续当前操作。',
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
                                  '新会员',
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
                                '填写注册信息',
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
                                        controller: _mobileController,
                                        focusNode: _mobileFocusNode,
                                        keyboardType: TextInputType.phone,
                                        textInputAction: TextInputAction.next,
                                        maxLength: 11,
                                        autofillHints: const [
                                          AutofillHints.telephoneNumber,
                                        ],
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) {
                                          _nicknameFocusNode.requestFocus();
                                        },
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '手机号',
                                          hint: '请输入常用手机号',
                                          icon: Icons.phone_iphone_rounded,
                                        ).copyWith(counterText: ''),
                                        validator: _validateMobile,
                                      ),
                                      const SizedBox(height: AppSpacing.lg),
                                      TextFormField(
                                        controller: _nicknameController,
                                        focusNode: _nicknameFocusNode,
                                        keyboardType: TextInputType.name,
                                        textInputAction: TextInputAction.next,
                                        maxLength: 20,
                                        autofillHints: const [
                                          AutofillHints.name,
                                        ],
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) {
                                          _passwordFocusNode.requestFocus();
                                        },
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '昵称',
                                          hint: '用于订单与评论展示',
                                          icon: Icons.badge_outlined,
                                        ).copyWith(counterText: ''),
                                        validator: _validateNickname,
                                      ),
                                      const SizedBox(height: AppSpacing.lg),
                                      TextFormField(
                                        controller: _passwordController,
                                        focusNode: _passwordFocusNode,
                                        obscureText: _obscurePassword,
                                        enableSuggestions: false,
                                        autocorrect: false,
                                        textInputAction: TextInputAction.next,
                                        autofillHints: const [
                                          AutofillHints.newPassword,
                                        ],
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) {
                                          _confirmPasswordFocusNode
                                              .requestFocus();
                                        },
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '密码',
                                          hint: '至少 6 位，建议包含字母和数字',
                                          icon: Icons.lock_outline_rounded,
                                          suffixIcon: IconButton(
                                            splashRadius: 20,
                                            onPressed: () {
                                              setState(() {
                                                _obscurePassword =
                                                    !_obscurePassword;
                                              });
                                            },
                                            icon: Icon(
                                              _obscurePassword
                                                  ? Icons
                                                      .visibility_off_outlined
                                                  : Icons.visibility_outlined,
                                              color: AppColors.textHint,
                                            ),
                                          ),
                                        ),
                                        validator: _validatePassword,
                                      ),
                                      const SizedBox(height: AppSpacing.lg),
                                      TextFormField(
                                        controller: _confirmPasswordController,
                                        focusNode: _confirmPasswordFocusNode,
                                        obscureText: _obscureConfirmPassword,
                                        enableSuggestions: false,
                                        autocorrect: false,
                                        textInputAction: TextInputAction.done,
                                        onChanged: (_) => _clearSubmitError(),
                                        onFieldSubmitted: (_) =>
                                            _submitRegisterData(),
                                        decoration: _buildInputDecoration(
                                          context: context,
                                          label: '确认密码',
                                          hint: '请再次输入密码',
                                          icon: Icons.verified_user_outlined,
                                          suffixIcon: IconButton(
                                            splashRadius: 20,
                                            onPressed: () {
                                              setState(() {
                                                _obscureConfirmPassword =
                                                    !_obscureConfirmPassword;
                                              });
                                            },
                                            icon: Icon(
                                              _obscureConfirmPassword
                                                  ? Icons
                                                      .visibility_off_outlined
                                                  : Icons.visibility_outlined,
                                              color: AppColors.textHint,
                                            ),
                                          ),
                                        ),
                                        validator: _validateConfirmPassword,
                                      ),
                                      const SizedBox(height: AppSpacing.md),
                                      Row(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          const Padding(
                                            padding: EdgeInsets.only(top: 1),
                                            child: Icon(
                                              Icons.info_outline_rounded,
                                              size: 16,
                                              color: AppColors.textHint,
                                            ),
                                          ),
                                          const SizedBox(width: AppSpacing.sm),
                                          Expanded(
                                            child: Text(
                                              '昵称用于前台展示，密码不会明文展示。注册完成后将自动登录，无需重复操作。',
                                              style:
                                                  textTheme.bodySmall?.copyWith(
                                                color: AppColors.textHint,
                                                height: 1.5,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                      if (_submitError != null) ...[
                                        const SizedBox(height: AppSpacing.lg),
                                        Container(
                                          padding: const EdgeInsets.all(
                                            AppSpacing.md,
                                          ),
                                          decoration: BoxDecoration(
                                            color: theme
                                                .colorScheme.errorContainer,
                                            borderRadius: BorderRadius.circular(
                                              AppRadii.lg,
                                            ),
                                          ),
                                          child: Row(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              Icon(
                                                Icons.error_outline_rounded,
                                                size: 18,
                                                color: theme.colorScheme
                                                    .onErrorContainer,
                                              ),
                                              const SizedBox(
                                                width: AppSpacing.sm,
                                              ),
                                              Expanded(
                                                child: Text(
                                                  _submitError!,
                                                  style: textTheme.bodySmall
                                                      ?.copyWith(
                                                    color: theme.colorScheme
                                                        .onErrorContainer,
                                                    height: 1.5,
                                                  ),
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ],
                                      const SizedBox(height: AppSpacing.xl),
                                      SizedBox(
                                        height: 52,
                                        child: ElevatedButton(
                                          onPressed: _isLoading
                                              ? null
                                              : _submitRegisterData,
                                          style: ElevatedButton.styleFrom(
                                            backgroundColor: AppColors.primary,
                                            foregroundColor: Colors.white,
                                            disabledBackgroundColor:
                                                AppColors.primary.withValues(
                                              alpha: 0.45,
                                            ),
                                            elevation: 0,
                                            shape: RoundedRectangleBorder(
                                              borderRadius:
                                                  BorderRadius.circular(
                                                AppRadii.lg,
                                              ),
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
                                              : Text(
                                                  '创建账号并登录',
                                                  style: textTheme.labelLarge
                                                      ?.copyWith(
                                                    color: Colors.white,
                                                  ),
                                                ),
                                        ),
                                      ),
                                      const SizedBox(height: AppSpacing.lg),
                                      OutlinedButton(
                                        onPressed: _isLoading
                                            ? null
                                            : () => Navigator.of(context)
                                                .maybePop(),
                                        style: OutlinedButton.styleFrom(
                                          minimumSize: const Size.fromHeight(
                                            48,
                                          ),
                                          side: const BorderSide(
                                            color: AppColors.border,
                                          ),
                                          foregroundColor:
                                              AppColors.textPrimary,
                                          shape: RoundedRectangleBorder(
                                            borderRadius: BorderRadius.circular(
                                              AppRadii.lg,
                                            ),
                                          ),
                                        ),
                                        child: const Text('已有账号，返回登录'),
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
