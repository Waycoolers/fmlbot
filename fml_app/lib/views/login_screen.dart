import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../services/utils.dart';
import '../view_models/auth_view_model.dart';
import 'home_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _confirmPasswordController = TextEditingController();

  bool _isPasswordObscured = true;
  bool _isLoginMode = true; // Переключатель: Вход или Регистрация

  @override
  Widget build(BuildContext context) {
    final authVM = context.watch<AuthViewModel>();

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Icon(Icons.favorite, size: 80, color: Colors.pinkAccent),
                const SizedBox(height: 24),

                Text(
                  _isLoginMode ? 'Добро пожаловать в FML' : 'Создание аккаунта',
                  style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  _isLoginMode ? 'Войди, чтобы продолжить' : 'Придумай логин и пароль',
                  style: const TextStyle(fontSize: 16, color: Colors.grey),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),

                TextField(
                  controller: _usernameController,
                  decoration: const InputDecoration(
                    labelText: 'Логин',
                    hintText: 'Введите ваш логин',
                    prefixIcon: Icon(Icons.person),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.all(Radius.circular(12)),
                    ),
                  ),
                ),
                const SizedBox(height: 16),

                TextField(
                  controller: _passwordController,
                  obscureText: _isPasswordObscured,
                  decoration: InputDecoration(
                    labelText: 'Пароль',
                    prefixIcon: const Icon(Icons.lock_outline),
                    border: const OutlineInputBorder(
                      borderRadius: BorderRadius.all(Radius.circular(12)),
                    ),
                    suffixIcon: IconButton(
                      icon: Icon(
                        _isPasswordObscured ? Icons.visibility_off : Icons.visibility,
                      ),
                      onPressed: () {
                        setState(() {
                          _isPasswordObscured = !_isPasswordObscured;
                        });
                      },
                    ),
                  ),
                ),

                // --- ЖЕСТКАЯ ФИКСАЦИЯ МЕСТА ДЛЯ 3-ГО ПОЛЯ ---
                IgnorePointer(
                  ignoring: _isLoginMode, // Блокируем нажатия, когда мы в режиме "Входа"
                  child: AnimatedOpacity(
                    opacity: _isLoginMode ? 0.0 : 1.0, // Плавное исчезновение/появление
                    duration: const Duration(milliseconds: 300),
                    child: Padding(
                      padding: const EdgeInsets.only(top: 16.0),
                      child: TextField(
                        controller: _confirmPasswordController,
                        obscureText: _isPasswordObscured,
                        decoration: const InputDecoration(
                          labelText: 'Повторите пароль',
                          prefixIcon: Icon(Icons.lock_reset),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.all(Radius.circular(12)),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),

                // --- ЖЕСТКАЯ ФИКСАЦИЯ МЕСТА ДЛЯ ТЕКСТА ОШИБКИ ---
                SizedBox(
                  height: 40, // Выделяем ровно 40 пикселей, чтобы кнопки не съезжали
                  child: Center(
                    child: AnimatedOpacity(
                      opacity: authVM.errorMessage != null ? 1.0 : 0.0,
                      duration: const Duration(milliseconds: 300),
                      child: Text(
                        authVM.errorMessage ?? '',
                        style: const TextStyle(color: Colors.red, fontWeight: FontWeight.bold),
                        textAlign: TextAlign.center,
                      ),
                    ),
                  ),
                ),

                // Главная кнопка
                ElevatedButton(
                  onPressed: authVM.isLoading
                      ? null
                      : (_isLoginMode ? _handleLogin : _handleRegister),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: authVM.isLoading
                      ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                      : Text(_isLoginMode ? 'Войти' : 'Зарегистрироваться', style: const TextStyle(fontSize: 18)),
                ),

                // Переключатель между экранами
                TextButton(
                  onPressed: () {
                    setState(() {
                      _isLoginMode = !_isLoginMode;
                      authVM.clearError(); // Сбрасываем ошибку при переключении
                    });
                  },
                  child: Text(
                    _isLoginMode ? 'Нет аккаунта? Создать' : 'Уже есть аккаунт? Войти',
                    style: const TextStyle(color: Colors.blue),
                  ),
                ),

                const SizedBox(height: 8),
                const Row(
                  children: [
                    Expanded(child: Divider()),
                    Padding(
                      padding: EdgeInsets.symmetric(horizontal: 16.0),
                      child: Text('ИЛИ', style: TextStyle(color: Colors.grey)),
                    ),
                    Expanded(child: Divider()),
                  ],
                ),
                const SizedBox(height: 24),

                // Кнопка быстрого входа через Telegram
                OutlinedButton.icon(
                  onPressed: _openTelegramBot,
                  icon: const Icon(Icons.telegram, color: Colors.blue, size: 28),
                  label: const Text(
                    'Быстрый вход через Telegram',
                    style: TextStyle(color: Colors.blue, fontSize: 16),
                  ),
                  style: OutlinedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    side: BorderSide(color: Colors.blue.shade200, width: 1.5),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    backgroundColor: Colors.blue.withOpacity(0.05),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _openTelegramBot() async {
    final Uri url = Uri.parse('https://t.me/fml_tg_bot');
    try {
      await launchUrl(url, mode: LaunchMode.externalApplication);
    } catch (e) {
      if (mounted) {
        showFmlSnackBar(context, 'Не удалось открыть Telegram', backgroundColor: Colors.red);
      }
    }
  }

  Future<void> _handleLogin() async {
    final username = _usernameController.text;
    final password = _passwordController.text;

    if (username.isEmpty || password.isEmpty) {
      showFmlSnackBar(context, 'Пожалуйста, заполни оба поля', backgroundColor: Colors.orange);
      return;
    }

    FocusScope.of(context).unfocus();

    final success = await context.read<AuthViewModel>().login(username, password);

    if (success && mounted) {
      Navigator.of(context).pushAndRemoveUntil(
        MaterialPageRoute(builder: (_) => const HomeScreen()),
            (route) => false,
      );
    }
  }

  Future<void> _handleRegister() async {
    final username = _usernameController.text;
    final password = _passwordController.text;
    final confirmPassword = _confirmPasswordController.text;

    if (username.isEmpty || password.isEmpty || confirmPassword.isEmpty) {
      showFmlSnackBar(context, 'Пожалуйста, заполни все поля', backgroundColor: Colors.orange);
      return;
    }

    if (password != confirmPassword) {
      showFmlSnackBar(context, 'Пароли не совпадают', backgroundColor: Colors.red);
      return;
    }

    FocusScope.of(context).unfocus();

    final success = await context.read<AuthViewModel>().register(username, password);

    if (success && mounted) {
      Navigator.of(context).pushAndRemoveUntil(
        MaterialPageRoute(builder: (_) => const HomeScreen()),
            (route) => false,
      );
    }
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }
}