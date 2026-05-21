import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ThemeViewModel extends ChangeNotifier {
  static const String _themePrefKey = 'theme_preference';

  ThemeMode _themeMode = ThemeMode.system;
  ThemeMode get themeMode => _themeMode;

  ThemeViewModel() {
    _loadTheme();
  }

  // Загружаем сохраненную тему
  Future<void> _loadTheme() async {
    final prefs = await SharedPreferences.getInstance();
    final String? themeStr = prefs.getString(_themePrefKey);

    if (themeStr != null) {
      _themeMode = ThemeMode.values.firstWhere(
              (e) => e.toString() == themeStr,
          orElse: () => ThemeMode.system
      );
      notifyListeners();
    }
  }

  // Сохраняем новую тему
  Future<void> setTheme(ThemeMode mode) async {
    if (_themeMode == mode) return;

    _themeMode = mode;
    notifyListeners();

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_themePrefKey, mode.toString());
  }
}