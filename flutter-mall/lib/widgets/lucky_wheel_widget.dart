import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';

import 'package:flutter_mall/model/digital_card/draw_activity_model.dart';

const List<Color> kWheelColors = <Color>[
  Color(0xFFC9A84C),
  Color(0xFF0F766E),
  Color(0xFF6D28D9),
  Color(0xFF0369A1),
  Color(0xFFDC2626),
];

class LuckyWheelWidget extends StatefulWidget {
  final List<DrawCardPreview> slots;
  final double size;

  const LuckyWheelWidget({
    super.key,
    required this.slots,
    this.size = 280,
  });

  @override
  State<LuckyWheelWidget> createState() => LuckyWheelWidgetState();
}

class LuckyWheelWidgetState extends State<LuckyWheelWidget>
    with SingleTickerProviderStateMixin {
  Timer? _spinTimer;
  late AnimationController _decelerationController;
  double _angle = 0.0;

  @override
  void initState() {
    super.initState();
    _decelerationController = AnimationController(vsync: this);
  }

  @override
  void dispose() {
    _spinTimer?.cancel();
    _decelerationController.dispose();
    super.dispose();
  }

  void startSpin() {
    _spinTimer?.cancel();
    _spinTimer = Timer.periodic(const Duration(milliseconds: 16), (_) {
      if (mounted) setState(() => _angle += 0.12);
    });
  }

  Future<void> stopAt(int targetSlotIndex) async {
    _spinTimer?.cancel();
    _spinTimer = null;

    final double targetCenter = _slotCenterAngle(targetSlotIndex);
    final double targetMod = (2 * pi - targetCenter) % (2 * pi);
    final double current = _angle % (2 * pi);
    double delta = (targetMod - current + 2 * pi) % (2 * pi);
    if (delta < pi / 2) delta += 2 * pi;

    final double startAngle = _angle;
    final double endAngle = _angle + delta;

    final Animation<double> anim = _decelerationController.drive(
      Tween<double>(begin: startAngle, end: endAngle)
          .chain(CurveTween(curve: Curves.easeOut)),
    );

    void listener() {
      if (mounted) setState(() => _angle = anim.value);
    }

    anim.addListener(listener);
    _decelerationController.duration = const Duration(milliseconds: 2000);
    _decelerationController.reset();
    await _decelerationController.forward();
    anim.removeListener(listener);
    if (mounted) setState(() => _angle = endAngle);
  }

  double _slotCenterAngle(int slotIndex) {
    final List<DrawCardPreview> sorted = _sorted();
    final double total = sorted.fold(0.0, (double s, DrawCardPreview c) => s + c.probability);
    double angle = 0;
    for (final DrawCardPreview c in sorted) {
      final double sweep = total > 0
          ? (c.probability / total) * 2 * pi
          : (2 * pi / sorted.length);
      if (c.slotIndex == slotIndex) return angle + sweep / 2;
      angle += sweep;
    }
    return 0;
  }

  List<DrawCardPreview> _sorted() {
    final List<DrawCardPreview> list = List<DrawCardPreview>.from(widget.slots);
    list.sort((DrawCardPreview a, DrawCardPreview b) => a.slotIndex.compareTo(b.slotIndex));
    return list;
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox.square(
      dimension: widget.size,
      child: Stack(
        alignment: Alignment.center,
        children: <Widget>[
          Positioned.fill(
            top: 34,
            child: const DecoratedBox(
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                boxShadow: <BoxShadow>[
                  BoxShadow(
                    color: Color(0x47C9A84C),
                    blurRadius: 40,
                    spreadRadius: 8,
                  ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.only(top: 34),
            child: CustomPaint(
              size: Size.square(widget.size - 34),
              painter: _WheelPainter(slots: _sorted(), angle: _angle),
            ),
          ),
          const Positioned(
            top: 0,
            child: CustomPaint(
              size: Size(26, 34),
              painter: _PointerPainter(),
            ),
          ),
        ],
      ),
    );
  }
}

class _PointerPainter extends CustomPainter {
  const _PointerPainter();

  @override
  void paint(Canvas canvas, Size size) {
    final Path path = Path()
      ..moveTo(size.width / 2, size.height)
      ..lineTo(2, 2)
      ..lineTo(size.width - 2, 2)
      ..close();

    canvas.drawPath(
      path,
      Paint()
        ..color = const Color(0x66000000)
        ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 6),
    );

    canvas.drawPath(
      path,
      Paint()
        ..shader = LinearGradient(
          colors: const <Color>[Color(0xFFFFE6A6), Color(0xFFC78A24)],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
        ).createShader(Rect.fromLTWH(0, 0, size.width, size.height)),
    );

    canvas.drawPath(
      path,
      Paint()
        ..color = const Color(0xFFC78A24)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1.5,
    );
  }

  @override
  bool shouldRepaint(_PointerPainter _) => false;
}

class _WheelPainter extends CustomPainter {
  final List<DrawCardPreview> slots;
  final double angle;

  const _WheelPainter({required this.slots, required this.angle});

  @override
  void paint(Canvas canvas, Size size) {
    if (slots.isEmpty) return;
    final Offset center = Offset(size.width / 2, size.height / 2);
    final double r = size.width / 2 - 6;
    final double total = slots.fold(0.0, (double s, DrawCardPreview c) => s + c.probability);

    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.rotate(angle);
    canvas.translate(-center.dx, -center.dy);

    // Decorative outer rings
    canvas.drawCircle(center, r + 6, Paint()..color = const Color(0xFFC78A24));
    canvas.drawCircle(center, r + 3, Paint()..color = const Color(0xFF0A0A14));
    canvas.drawCircle(center, r + 1, Paint()..color = const Color(0xFFE8D494));

    // Segments
    double startAngle = -pi / 2;
    for (int i = 0; i < slots.length; i++) {
      final DrawCardPreview slot = slots[i];
      final double sweep = total > 0
          ? (slot.probability / total) * 2 * pi
          : (2 * pi / slots.length);
      final Rect segRect = Rect.fromCircle(center: center, radius: r);

      canvas.drawArc(
        segRect,
        startAngle,
        sweep,
        true,
        Paint()..color = kWheelColors[i % kWheelColors.length],
      );
      canvas.drawArc(
        segRect,
        startAngle,
        sweep,
        true,
        Paint()
          ..color = Colors.white.withValues(alpha: 0.22)
          ..style = PaintingStyle.stroke
          ..strokeWidth = 2,
      );

      if (sweep > 0.25) {
        final double mid = startAngle + sweep / 2;
        final double tr = r * 0.64;
        final double tx = center.dx + tr * cos(mid);
        final double ty = center.dy + tr * sin(mid);
        final String label = slot.rarity.isNotEmpty
            ? (slot.rarity.length > 5 ? '${slot.rarity.substring(0, 4)}…' : slot.rarity)
            : 'R${slot.slotIndex}';
        _drawText(canvas, label, Offset(tx, ty - 9), 12, bold: true);
        _drawText(
          canvas,
          '${(slot.probability * 100).toStringAsFixed(0)}%',
          Offset(tx, ty + 9),
          11,
        );
      }

      startAngle += sweep;
    }

    // Center hub
    canvas.drawCircle(center, r * 0.175, Paint()..color = const Color(0xFF0A0A14));
    canvas.drawCircle(
      center,
      r * 0.175,
      Paint()
        ..color = const Color(0xFFC78A24)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 3,
    );
    canvas.drawCircle(center, r * 0.085, Paint()..color = const Color(0xFFFFE6A6));

    canvas.restore();
  }

  void _drawText(
    Canvas canvas,
    String text,
    Offset pos,
    double fontSize, {
    bool bold = false,
  }) {
    final TextPainter tp = TextPainter(
      text: TextSpan(
        text: text,
        style: TextStyle(
          color: Colors.white,
          fontSize: fontSize,
          fontWeight: bold ? FontWeight.w900 : FontWeight.w600,
          shadows: const <Shadow>[
            Shadow(
              color: Color(0x88000000),
              blurRadius: 5,
              offset: Offset(0, 1),
            ),
          ],
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    tp.paint(canvas, pos - Offset(tp.width / 2, tp.height / 2));
  }

  @override
  bool shouldRepaint(_WheelPainter old) =>
      old.angle != angle || old.slots != slots;
}
