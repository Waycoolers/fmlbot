import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/utils.dart';
import '../view_models/auth_view_model.dart';
import '../view_models/important_date_view_model.dart';
import '../view_models/settings_view_model.dart';
import '../view_models/theme_view_model.dart';
import '../view_models/user_view_model.dart';
import 'login_screen.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  final TextEditingController _limitController = TextEditingController();
  bool _isUnlimited = false;
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    _loadCurrentConfig();
  }

  void _loadCurrentConfig() {
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final vm = context.read<SettingsViewModel>();
      await vm.fetchConfig();
      if (vm.myConfig != null) {
        setState(() {
          if (vm.myConfig!.maxComplimentCount == -1) {
            _isUnlimited = true;
            _limitController.text = "";
          } else {
            _isUnlimited = false;
            _limitController.text = vm.myConfig!.maxComplimentCount.toString();
          }
        });
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final vm = context.watch<SettingsViewModel>();
    final userVM = context.watch<UserViewModel>();
    final themeVM = context.watch<ThemeViewModel>();

    // Определяем, включена ли темная тема прямо сейчас
    bool isDarkMode = themeVM.themeMode == ThemeMode.dark;
    if (themeVM.themeMode == ThemeMode.system) {
      isDarkMode = MediaQuery.of(context).platformBrightness == Brightness.dark;
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Настройки'),
        centerTitle: true,
      ),
      // Убрали Center(CircularProgressIndicator) из корня,
      // чтобы экран всегда загружался, даже если нет сети
      body: ListView(
        padding: const EdgeInsets.symmetric(vertical: 16.0),
        children: [

          // --- СЕКЦИЯ 1: ОФОРМЛЕНИЕ (Всегда доступно) ---
          _buildSectionHeader(context, 'Экран'),
          SwitchListTile(
            secondary: Icon(isDarkMode ? Icons.dark_mode : Icons.light_mode),
            title: const Text('Тёмная тема'),
            value: isDarkMode,
            onChanged: (bool value) {
              context.read<ThemeViewModel>().setTheme(
                value ? ThemeMode.dark : ThemeMode.light,
              );
            },
          ),

          const Padding(
            padding: EdgeInsets.symmetric(vertical: 8.0),
            child: Divider(indent: 16, endIndent: 16),
          ),

          // --- РАЗВИЛКА: ПРОВЕРКА ИНТЕРНЕТА ---

          if (vm.isLoading && vm.myConfig == null) ...[
            // 1. Состояние: Идет загрузка настроек
            const SizedBox(height: 32),
            const Center(child: CircularProgressIndicator()),

          ] else if (vm.myConfig != null) ...[
            // 2. Состояние: Есть интернет, настройки загружены

            // --- СЕКЦИЯ 2: ЛИМИТЫ ---
            _buildSectionHeader(context, 'Комплименты'),
            SwitchListTile(
              secondary: const Icon(Icons.all_inclusive),
              title: const Text('Без ограничений'),
              subtitle: const Text('Партнер сможет писать сколько угодно'),
              value: _isUnlimited,
              onChanged: (val) {
                setState(() {
                  _isUnlimited = val;
                  if (val) _limitController.clear();
                });
              },
            ),

            AnimatedSize(
              duration: const Duration(milliseconds: 300),
              curve: Curves.easeInOut,
              child: _isUnlimited
                  ? const SizedBox.shrink()
                  : Padding(
                padding: const EdgeInsets.fromLTRB(72, 8, 16, 16),
                child: TextField(
                  controller: _limitController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: 'Максимум в день',
                    filled: true,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: BorderSide.none,
                    ),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 16),
                  ),
                ),
              ),
            ),

            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
              child: FilledButton(
                onPressed: _isSaving ? null : () => _saveSettings(vm),
                style: FilledButton.styleFrom(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: _isSaving
                    ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : const Text('Сохранить лимиты', style: TextStyle(fontSize: 16)),
              ),
            ),

            // --- СЕКЦИЯ 3: ОПАСНАЯ ЗОНА ---
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16.0),
              child: Divider(indent: 16, endIndent: 16),
            ),
            _buildSectionHeader(context, 'Опасная зона', color: Colors.red),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
              child: Card(
                elevation: 0,
                margin: EdgeInsets.zero,
                color: Colors.red.withOpacity(0.1),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                child: Column(
                  children: [
                    // 1. Смена пароля
                    ListTile(
                      leading: const Icon(Icons.password, color: Colors.red),
                      title: const Text('Сменить пароль', style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold)),
                      onTap: () => _showChangePasswordDialog(context),
                    ),
                    const Divider(height: 1, indent: 16, endIndent: 16),
                    // 2. Разрыв связи (твой старый функционал)
                    if (userVM.partner != null) ...{
                      ListTile(
                        leading: const Icon(Icons.link_off, color: Colors.red),
                        title: const Text('Разорвать связь с партнером', style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold)),
                        onTap: () => _showUnpairDialog(context),
                      ),
                    },
                    const Divider(height: 1, indent: 16, endIndent: 16),
                    // 3. Удаление аккаунта
                    ListTile(
                      leading: const Icon(Icons.delete_forever, color: Colors.red),
                      title: const Text('Удалить аккаунт', style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold)),
                      onTap: () => _showDeleteAccountDialog(context),
                    ),
                  ],
                ),
              ),
            ),

          ] else ...[
            // 3. Состояние: НЕТ ИНТЕРНЕТА
            const SizedBox(height: 32),
            Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 32.0),
                child: Column(
                  children: [
                    Icon(Icons.cloud_off, size: 48, color: Colors.grey.shade400),
                    const SizedBox(height: 16),
                    Text(
                      'Нет подключения к сети',
                      style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey.shade700),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Настройка лимитов и управление парой станут доступны, когда появится интернет.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey.shade600, fontSize: 14),
                    ),
                  ],
                ),
              ),
            ),
          ]
        ],
      ),
    );
  }

  // Вспомогательный метод для отрисовки красивых заголовков секций как в Pixel
  Widget _buildSectionHeader(BuildContext context, String title, {Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
      child: Text(
        title,
        style: Theme.of(context).textTheme.titleSmall?.copyWith(
          color: color ?? Theme.of(context).colorScheme.primary,
          fontWeight: FontWeight.bold,
          letterSpacing: 0.5,
        ),
      ),
    );
  }

  Future<void> _saveSettings(SettingsViewModel vm) async {
    int finalValue;

    if (_isUnlimited) {
      finalValue = -1;
    } else {
      final input = int.tryParse(_limitController.text.trim());
      if (input == null || input < 1 || input > 100) {
        showFmlSnackBar(context, 'Введи число от 1 до 100', backgroundColor: Colors.red);
        return;
      }
      finalValue = input;
    }

    setState(() => _isSaving = true);
    final success = await vm.updateMaxCompliments(finalValue);
    setState(() => _isSaving = false);

    if (success && mounted) {
      showFmlSnackBar(context, 'Настройки сохранены ✨', backgroundColor: Colors.green);
      FocusScope.of(context).unfocus();
    }
  }

  void _showUnpairDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Разорвать связь?'),
        content: const Text(
          'Ты действительно хочешь удалить партнера?\nВся ваша общая история в боте будет потеряна.',
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Отмена')
          ),
          TextButton(
            onPressed: () async {
              Navigator.pop(ctx);

              final dateVM = context.read<ImportantDateViewModel>();
              await dateVM.fetchDates();

              final sharedDates = dateVM.dates.where((d) => d.isShared).toList();

              for (var date in sharedDates) {
                await dateVM.deleteDate(date.id);
              }

              final success = await context.read<UserViewModel>().unpair();

              if (success && mounted) {
                showFmlSnackBar(context, 'Связь разорвана 💔', backgroundColor: Colors.orange);
                Navigator.pop(context);
              } else if (mounted) {
                showFmlSnackBar(context, 'Не удалось разорвать связь', backgroundColor: Colors.red);
              }
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Разорвать'),
          ),
        ],
      ),
    );
  }

  void _showChangePasswordDialog(BuildContext context) {
    final TextEditingController passController = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Смена пароля'),
        content: TextField(
          controller: passController,
          obscureText: true,
          decoration: const InputDecoration(
            labelText: 'Новый пароль',
            helperText: 'От 8 до 32 символов, цифры и буквы',
          ),
        ),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Отмена')
          ),
          FilledButton(
            onPressed: () async {
              // ВЫЗОВ ВАЛИДАЦИИ
              final error = validatePassword(passController.text);

              if (error != null) {
                // Если есть ошибка, показываем её пользователю
                showFmlSnackBar(context, error, backgroundColor: Colors.red);
                return;
              }

              // Если всё ок, идем на сервер
              final success = await context.read<UserViewModel>().changePassword(passController.text);
              if (success && mounted) {
                showFmlSnackBar(context, 'Пароль успешно обновлен', backgroundColor: Colors.green);
                Navigator.pop(ctx);
              } else if (mounted) {
                showFmlSnackBar(context, 'Ошибка при смене пароля', backgroundColor: Colors.red);
              }
            },
            child: const Text('Сохранить'),
          ),
        ],
      ),
    );
  }

  void _showDeleteAccountDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Удаление аккаунта', style: TextStyle(color: Colors.red)),
        content: const Text('Это действие нельзя отменить. Все твои данные будут удалены навсегда.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Отмена')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red),
            onPressed: () async {
              final success = await context.read<UserViewModel>().deleteAccount();
              if (success && mounted) {
                await context.read<AuthViewModel>().logout();
                Navigator.of(context).pushAndRemoveUntil(
                  MaterialPageRoute(builder: (_) => const LoginScreen()),
                      (route) => false,
                );
              }
            },
            child: const Text('Удалить навсегда'),
          ),
        ],
      ),
    );
  }

  String? validatePassword(String password) {
    if (password.length < 8) return 'Минимум 8 символов';
    if (password.length > 32) return 'Максимум 32 символа';

    bool hasUpper = false;
    bool hasLower = false;
    bool hasNumber = false;
    bool hasLetter = false;

    for (int i = 0; i < password.length; i++) {
      int code = password.codeUnitAt(i);

      // Проверка на допустимые символы (ASCII 33 '!' - 126 '~')
      if (code < 33 || code > 126) return 'Недопустимый символ';

      if (code >= 65 && code <= 90) { // A-Z
        hasUpper = true;
        hasLetter = true;
      } else if (code >= 97 && code <= 122) { // a-z
        hasLower = true;
        hasLetter = true;
      } else if (code >= 48 && code <= 57) { // 0-9
        hasNumber = true;
      }
    }

    if (!hasLetter) return 'Пароль должен содержать буквы';
    if (!hasUpper) return 'Нужна заглавная буква';
    if (!hasLower) return 'Нужна строчная буква';
    if (!hasNumber) return 'Нужна хотя бы одна цифра';

    return null; // Пароль валиден
  }

  @override
  void dispose() {
    _limitController.dispose();
    super.dispose();
  }
}