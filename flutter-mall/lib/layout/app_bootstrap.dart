import 'package:flutter/material.dart';
import 'package:flutter_mall/layout/intent_recovery_shell.dart';

class AppBootstrap extends StatelessWidget {
  final Duration decisionTimeout;
  final RecoveryTargetBuilder? targetBuilder;
  final RecoveryFallbackBuilder? fallbackBuilder;
  final RecoveryLoginBuilder? loginBuilder;

  const AppBootstrap({
    super.key,
    this.decisionTimeout = const Duration(seconds: 3),
    this.targetBuilder,
    this.fallbackBuilder,
    this.loginBuilder,
  });

  @override
  Widget build(BuildContext context) {
    return IntentRecoveryShell(
      decisionTimeout: decisionTimeout,
      targetBuilder: targetBuilder,
      fallbackBuilder: fallbackBuilder,
      loginBuilder: loginBuilder,
      splash: Container(
        color: Colors.white,
        alignment: Alignment.center,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Image.asset(
              'images/icon_main_logo.png',
              width: 100,
              height: 100,
              fit: BoxFit.contain,
            ),
            const SizedBox(height: 20),
            const Text(
              '九克城',
              style: TextStyle(
                fontSize: 32,
                fontWeight: FontWeight.bold,
                color: Color(0xFF1a1a2e),
                letterSpacing: 4,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              '品质生活，从这里开始',
              style: TextStyle(
                fontSize: 14,
                color: Color(0xFF909399),
              ),
            ),
            const SizedBox(height: 24),
            const SizedBox(
              width: 24,
              height: 24,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          ],
        ),
      ),
    );
  }
}
