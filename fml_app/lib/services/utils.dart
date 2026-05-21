import 'package:flutter/material.dart';

void showFmlSnackBar(BuildContext context, String message, {Color backgroundColor = Colors.green}) {
  final messenger = ScaffoldMessenger.of(context);

  // 1. Мгновенно скрываем текущий, если он есть
  messenger.removeCurrentSnackBar();

  // 2. Показываем новый
  messenger.showSnackBar(
    SnackBar(
      content: Text(message),
      backgroundColor: backgroundColor,
      duration: const Duration(seconds: 2), // Время показа
      behavior: SnackBarBehavior.floating,   // Чтобы он "летал" над кнопками
      margin: const EdgeInsets.all(16),     // Отступы
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    ),
  );
}