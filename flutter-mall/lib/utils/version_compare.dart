List<int> splitVersion(String value) {
  return value.trim().split('.').where((part) => part.isNotEmpty).map((part) {
    final match = RegExp(r'^(\d+)').firstMatch(part.trim());
    if (match == null) {
      return 0;
    }
    return int.tryParse(match.group(1) ?? '0') ?? 0;
  }).toList();
}

int compareVersion(String current, String required) {
  final currentParts = splitVersion(current);
  final requiredParts = splitVersion(required);
  final maxLength = currentParts.length > requiredParts.length
      ? currentParts.length
      : requiredParts.length;

  for (var i = 0; i < maxLength; i++) {
    final currentValue = i < currentParts.length ? currentParts[i] : 0;
    final requiredValue = i < requiredParts.length ? requiredParts[i] : 0;
    if (currentValue > requiredValue) {
      return 1;
    }
    if (currentValue < requiredValue) {
      return -1;
    }
  }
  return 0;
}
