import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../view_models/auth_view_model.dart';
import '../view_models/user_view_model.dart';
import 'compliments_screen.dart';
import 'important_dates_screen.dart';
import 'login_screen.dart';
import 'settings_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<UserViewModel>().fetchProfiles();
    });
  }

  @override
  Widget build(BuildContext context) {
    final userVM = context.watch<UserViewModel>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('FML App', style: TextStyle(fontWeight: FontWeight.bold)),
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.settings),
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const SettingsScreen()),
            );
          },
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              // Спрашиваем подтверждение
              final confirm = await showDialog<bool>(
                context: context,
                builder: (ctx) => AlertDialog(
                  title: const Text('Выход'),
                  content: const Text('Ты точно хочешь выйти из аккаунта?'),
                  actions: [
                    TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Отмена')),
                    TextButton(
                        onPressed: () => Navigator.pop(ctx, true),
                        style: TextButton.styleFrom(foregroundColor: Colors.red),
                        child: const Text('Выйти')
                    ),
                  ],
                ),
              );

              if (confirm == true && context.mounted) {
                await context.read<AuthViewModel>().logout();
                // Уничтожаем историю навигации и кидаем на экран входа
                Navigator.of(context).pushAndRemoveUntil(
                  MaterialPageRoute(builder: (_) => const LoginScreen()),
                      (route) => false,
                );
              }
            },
          )
        ],
      ),
      body: userVM.isLoading && userVM.currentUser == null
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
        onRefresh: () async {
          await context.read<UserViewModel>().fetchProfiles();
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                // --- БЛОК ПРОФИЛЕЙ ---
                _buildProfileSection(userVM),
                const SizedBox(height: 24),

                // Кнопка добавления партнера (появляется только если пары нет)
                if (userVM.partner == null && userVM.currentUser != null) ...[
                  ElevatedButton.icon(
                    icon: const Icon(Icons.favorite_border),
                    label: const Text('Связать аккаунты'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.pink.shade50,
                      foregroundColor: Colors.pink,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    onPressed: () => _showAddPartnerDialog(context),
                  ),
                  const SizedBox(height: 24),
                ],

                // --- МЕНЮ АКТИВНОСТЕЙ ---
                const Text(
                  'Твои активности',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.grey),
                ),
                const SizedBox(height: 12),

                _buildMenuTile(
                  context,
                  title: 'Комплименты',
                  subtitle: 'Пиши и получай приятности',
                  icon: Icons.favorite,
                  color: Colors.redAccent,
                  onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const ComplimentsScreen())),
                ),
                const SizedBox(height: 12),

                _buildMenuTile(
                  context,
                  title: 'Важные даты',
                  subtitle: 'Важные и не очень важные события',
                  icon: Icons.calendar_month,
                  color: Colors.blueAccent,
                  onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const ImportantDatesScreen())),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // Виджет секции профилей
  Widget _buildProfileSection(UserViewModel vm) {
    if (vm.currentUser == null) return const SizedBox();

    // Если партнер есть — рисуем две карточки в ряд
    if (vm.partner != null) {
      return Row(
        children: [
          Expanded(
            child: _buildUserCard(
              username: vm.currentUser!.username,
              label: 'Это ты',
              icon: Icons.person,
              color: Colors.blue,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: _buildUserCard(
              username: vm.partner!.username,
              label: 'Партнер',
              icon: Icons.favorite,
              color: Colors.pink,
            ),
          ),
        ],
      );
    }

    // Если партнера нет — одна большая карточка
    return _buildUserCard(
      username: vm.currentUser!.username,
      label: 'Ты пока без пары 💔',
      icon: Icons.person_pin_rounded,
      color: Colors.blue,
      isFullWidth: true,
    );
  }

  // Шаблон отдельной карточки пользователя
  Widget _buildUserCard({
    required String username,
    required String label,
    required IconData icon,
    required Color color,
    bool isFullWidth = false,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 16),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.2), width: 2),
      ),
      child: Column(
        children: [
          CircleAvatar(
            radius: 28,
            backgroundColor: color.withOpacity(0.8),
            child: Icon(icon, color: Colors.white, size: 30),
          ),
          const SizedBox(height: 12),
          Text(
            '@$username',
            style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  // Красивая кнопка меню
  Widget _buildMenuTile(
      BuildContext context, {
        required String title,
        required String subtitle,
        required IconData icon,
        required Color color,
        required VoidCallback onTap,
      }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surfaceContainerHighest.withOpacity(0.5),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(color: color.withOpacity(0.1), shape: BoxShape.circle),
              child: Icon(icon, color: color),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                  Text(subtitle, style: TextStyle(fontSize: 13, color: Colors.grey.shade600)),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, color: Colors.grey),
          ],
        ),
      ),
    );
  }

  // Всплывающее окно для добавления партнера
  void _showAddPartnerDialog(BuildContext context) {
    final TextEditingController controller = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          title: const Text('Связка аккаунтов'),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(
              labelText: 'Никнейм партнера',
              hintText: 'username',
              prefixIcon: Icon(Icons.alternate_email),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Отмена'),
            ),
            ElevatedButton(
              onPressed: () async {
                // Очищаем ввод от пробелов и удаляем все символы '@'
                final cleanUsername = controller.text.trim().replaceAll('@', '');
                final currentUser = context.read<UserViewModel>().currentUser;

                if (cleanUsername.isNotEmpty) {
                  // --- ПРОВЕРКА НА САМОГО СЕБЯ ---
                  if (currentUser != null && cleanUsername.toLowerCase() == currentUser.username.toLowerCase()) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(
                        content: Text('Нельзя добавить самого себя! 😅'),
                        backgroundColor: Colors.orange,
                      ),
                    );
                    return; // Прерываем выполнение
                  }

                  Navigator.pop(ctx); // Закрываем окно
                  // Отправляем на бэкенд уже чистый юзернейм без '@'
                  final success = await context.read<UserViewModel>().addPartner(cleanUsername);

                  if (success && mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Ура! Вы теперь в паре 🎉'), backgroundColor: Colors.green),
                    );
                  } else if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('Не удалось найти партнера'), backgroundColor: Colors.red),
                    );
                  }
                }
              },
              child: const Text('Добавить'),
            ),
          ],
        );
      },
    );
  }
}