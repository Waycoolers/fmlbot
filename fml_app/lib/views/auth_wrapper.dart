import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../view_models/auth_view_model.dart';
import 'home_screen.dart';
import 'login_screen.dart';

class AuthWrapper extends StatefulWidget {
  const AuthWrapper({super.key});

  @override
  State<AuthWrapper> createState() => _AuthWrapperState();
}

class _AuthWrapperState extends State<AuthWrapper> {
  bool _isChecking = true;
  bool _isAuthenticated = false;

  @override
  void initState() {
    super.initState();
    _initAuthCheck();
  }

  Future<void> _initAuthCheck() async {
    // Вызываем созданный нами метод из AuthViewModel
    final hasToken = await context.read<AuthViewModel>().checkAutoLogin();

    if (mounted) {
      setState(() {
        _isAuthenticated = hasToken;
        _isChecking = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    // 1. Пока идет чтение из хранилища — показываем красивый экран загрузки (Splash Screen)
    if (_isChecking) {
      return Scaffold(
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Логотип приложения (тот же, что и на экране логина)
              const Icon(Icons.favorite, size: 100, color: Colors.pinkAccent),
              const SizedBox(height: 24),
              const Text(
                'FML App',
                style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold, letterSpacing: 1.2),
              ),
              const SizedBox(height: 48),
              // Небольшой и аккуратный индикатор загрузки снизу
              SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(
                  strokeWidth: 2.5,
                  color: Theme.of(context).colorScheme.primary,
                ),
              ),
            ],
          ),
        ),
      );
    }

    // 2. Если токен найден — перенаправляем на Главный экран, если нет — на Логин
    return _isAuthenticated ? const HomeScreen() : const LoginScreen();
  }
}