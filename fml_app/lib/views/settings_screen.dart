import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../view_models/important_date_view_model.dart';
import '../view_models/settings_view_model.dart';
import '../view_models/user_view_model.dart'; // Добавили импорт

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
    final userVM = context.watch<UserViewModel>(); // Подключаем юзера, чтобы знать про партнера

    return Scaffold(
      appBar: AppBar(
        title: const Text('Настройки'),
      ),
      body: vm.isLoading && vm.myConfig == null
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Лимиты комплиментов',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),

            Card(
              margin: EdgeInsets.zero,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              child: SwitchListTile(
                title: const Text('Без ограничений'),
                subtitle: const Text('Партнер сможет присылать сколько угодно комплиментов'),
                secondary: const Icon(Icons.all_inclusive, color: Colors.amber),
                value: _isUnlimited,
                onChanged: (val) {
                  setState(() {
                    _isUnlimited = val;
                    if (val) _limitController.clear();
                  });
                },
              ),
            ),

            const SizedBox(height: 16),

            AnimatedOpacity(
              duration: const Duration(milliseconds: 300),
              opacity: _isUnlimited ? 0.5 : 1.0,
              child: IgnorePointer(
                ignoring: _isUnlimited,
                child: TextField(
                  controller: _limitController,
                  readOnly: _isUnlimited,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: 'Максимум в день',
                    hintText: _isUnlimited ? '∞' : 'Например: 10',
                    border: const OutlineInputBorder(),
                    prefixIcon: const Icon(Icons.favorite_border),
                  ),
                ),
              ),
            ),

            const SizedBox(height: 32),

            ElevatedButton(
              onPressed: _isSaving ? null : () => _saveSettings(vm),
              style: ElevatedButton.styleFrom(
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              ),
              child: _isSaving
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('Сохранить', style: TextStyle(fontSize: 18)),
            ),

            // --- ОПАСНАЯ ЗОНА (показываем только если есть партнер) ---
            if (userVM.partner != null) ...[
              const SizedBox(height: 48), // Большой отступ, чтобы отделить от обычных настроек
              const Text(
                'Опасная зона',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.red),
              ),
              const SizedBox(height: 8),
              ListTile(
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                tileColor: Colors.red.withOpacity(0.1),
                leading: const Icon(Icons.link_off, color: Colors.red),
                title: const Text(
                    'Разорвать связь с партнером',
                    style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold)
                ),
                onTap: () => _showUnpairDialog(context),
              ),
            ]
          ],
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
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Введи число от 1 до 100'), backgroundColor: Colors.red),
        );
        return;
      }
      finalValue = input;
    }

    setState(() => _isSaving = true);
    final success = await vm.updateMaxCompliments(finalValue);
    setState(() => _isSaving = false);

    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Настройки успешно обновлены! ✨'), backgroundColor: Colors.green),
      );
      FocusScope.of(context).unfocus();
    }
  }

  // Диалог подтверждения разрыва связи
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
              Navigator.pop(ctx); // Закрываем диалог

              // 1. Ищем и удаляем все общие даты
              final dateVM = context.read<ImportantDateViewModel>();
              await dateVM.fetchDates(); // На всякий случай обновляем список

              // Фильтруем только общие
              final sharedDates = dateVM.dates.where((d) => d.isShared).toList();

              // Удаляем их по одной через API
              for (var date in sharedDates) {
                await dateVM.deleteDate(date.id);
              }

              // 2. Разрываем связь с партнером
              final success = await context.read<UserViewModel>().unpair();

              if (success && mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Связь разорвана 💔'),
                    backgroundColor: Colors.orange,
                  ),
                );
                Navigator.pop(context);
              } else if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Не удалось разорвать связь'),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Разорвать'),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _limitController.dispose();
    super.dispose();
  }
}