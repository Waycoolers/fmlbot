import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/notification_service.dart';
import '../services/utils.dart';
import '../view_models/auth_view_model.dart';
import '../view_models/user_view_model.dart';
import '../view_models/important_date_view_model.dart'; // <-- Добавили импорт
import 'compliments_screen.dart';
import 'important_dates_screen.dart';
import 'login_screen.dart';
import 'settings_screen.dart';
import '../view_models/leisure_idea_view_model.dart';
import '../models/leisure_idea_request.dart';

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
      // <-- Теперь при открытии главного экрана грузим еще и даты
      context.read<ImportantDateViewModel>().fetchDates();

      NotificationService().init();
    });
  }

  @override
  Widget build(BuildContext context) {
    final userVM = context.watch<UserViewModel>();
    final dateVM = context.watch<ImportantDateViewModel>();

    // --- НОВАЯ КРАСИВАЯ ЛОГИКА ДЛЯ ДАТЫ ---
    Widget? datesSubtitleWidget;

    if (dateVM.isLoading) {
      datesSubtitleWidget = Text('Загрузка...', style: TextStyle(fontSize: 13, color: Colors.grey.shade600));
    } else if (dateVM.dates.isNotEmpty && dateVM.upcomingDate != null) {
      final closest = dateVM.upcomingDate!;

      final now = DateTime.now();
      final today = DateTime(now.year, now.month, now.day);

      // Считаем когда она будет в этом (или следующем) году
      DateTime nextOccurrence = DateTime(today.year, closest.date.month, closest.date.day);
      if (nextOccurrence.isBefore(today)) {
        nextOccurrence = DateTime(today.year + 1, closest.date.month, closest.date.day);
      }

      final daysLeft = nextOccurrence.difference(today).inDays;

      String daysText;
      if (daysLeft == 0) daysText = 'Сегодня!';
      else if (daysLeft == 1) daysText = 'Завтра';
      else daysText = 'Через $daysLeft дн.';

      datesSubtitleWidget = Row(
        children: [
          Expanded(
            child: Text(
              closest.title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(fontSize: 13, color: Colors.blue.shade700, fontWeight: FontWeight.w600),
            ),
          ),
          const SizedBox(width: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: Colors.blue.shade100,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              daysText,
              style: TextStyle(fontSize: 11, color: Colors.blue.shade900, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      );
    } else {
      datesSubtitleWidget = Text('Важных дат нет!', style: TextStyle(fontSize: 13, color: Colors.grey.shade600));
    }

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
          : userVM.currentUser == null // <-- ПРОВЕРКА НА ОШИБКУ ЗАГРУЗКИ / НЕТ СЕТИ
          ? Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.wifi_off_rounded, size: 80, color: Colors.grey.shade400),
              const SizedBox(height: 16),
              const Text(
                'Нет связи с сервером',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                'Не удалось загрузить профиль.\nПроверь подключение к интернету и попробуй снова.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey, fontSize: 16),
              ),
              const SizedBox(height: 32),
              FilledButton.icon(
                onPressed: () {
                  context.read<UserViewModel>().fetchProfiles();
                  context.read<ImportantDateViewModel>().fetchDates();
                },
                icon: const Icon(Icons.refresh),
                label: const Text('Обновить'),
                style: FilledButton.styleFrom(
                  padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                ),
              ),
            ],
          ),
        ),
      )
          : RefreshIndicator(
        onRefresh: () async {
          await context.read<UserViewModel>().fetchProfiles();
          await context.read<ImportantDateViewModel>().fetchDates();
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _buildProfileSection(userVM),
                const SizedBox(height: 24),

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
                  subtitleWidget: datesSubtitleWidget,
                  icon: Icons.calendar_month,
                  color: Colors.blueAccent,
                  onTap: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const ImportantDatesScreen())),
                ),

                const SizedBox(height: 12),

                _buildMenuTile(
                  context,
                  title: 'Идея для досуга',
                  subtitle: 'Подберём занятие по настроению',
                  icon: Icons.lightbulb,
                  color: Colors.amber,
                  onTap: () => _showLeisureIdeaDialog(context),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

// ... остальной код (методы _buildProfileSection, _buildUserCard, _buildMenuTile, _showAddPartnerDialog) остается без изменений!

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
        String? subtitle,
        Widget? subtitleWidget, // <-- Поддержка красивого подзаголовка
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
                  Text(
                    title,
                    style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                  // Если передали обычный текст
                  if (subtitle != null)
                    Text(
                      subtitle,
                      style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
                      maxLines: 1,                      // <-- Жесткое ограничение в 1 строку
                      overflow: TextOverflow.ellipsis,  // <-- Троеточие для длинных текстов
                    ),
                  // Если передали кастомный виджет (как для нашей даты)
                  if (subtitleWidget != null) subtitleWidget,
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
                // Очищаем ввод
                final cleanUsername = controller.text.trim().replaceAll('@', '');

                if (cleanUsername.isNotEmpty) {
                  Navigator.pop(ctx); // Закрываем диалоговое окно

                  // Ждем ответ от ViewModel. errorMessage будет null, если пара создана.
                  final errorMessage = await context.read<UserViewModel>().addPartner(cleanUsername);

                  if (mounted) {
                    if (errorMessage == null) {
                      // Успех!
                      showFmlSnackBar(context, 'Ура! Вы теперь в паре 🎉', backgroundColor: Colors.green);
                    } else {
                      // Показываем конкретную ошибку пользователю
                      showFmlSnackBar(context, errorMessage, backgroundColor: Colors.red);
                    }
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

  void _showLeisureIdeaDialog(BuildContext context) {
    final locationController = TextEditingController();
    final extraContextController = TextEditingController();

    String activityLevel = 'medium';
    String budget = 'medium';

    showDialog(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (ctx, setState) {
            return AlertDialog(
              title: const Text('Идея для досуга'),
              content: SingleChildScrollView(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    TextField(
                      controller: locationController,
                      decoration: const InputDecoration(
                        labelText: 'Локация',
                        hintText: 'например: дом, парк, город',
                      ),
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      value: activityLevel,
                      decoration: const InputDecoration(labelText: 'Уровень активности'),
                      items: const [
                        DropdownMenuItem(value: 'low', child: Text('Низкий')),
                        DropdownMenuItem(value: 'medium', child: Text('Средний')),
                        DropdownMenuItem(value: 'high', child: Text('Высокий')),
                      ],
                      onChanged: (value) {
                        if (value != null) setState(() => activityLevel = value);
                      },
                    ),
                    const SizedBox(height: 12),
                    DropdownButtonFormField<String>(
                      value: budget,
                      decoration: const InputDecoration(labelText: 'Бюджет'),
                      items: const [
                        DropdownMenuItem(value: 'free', child: Text('Бесплатно')),
                        DropdownMenuItem(value: 'low', child: Text('Низкий')),
                        DropdownMenuItem(value: 'medium', child: Text('Средний')),
                        DropdownMenuItem(value: 'high', child: Text('Высокий')),
                      ],
                      onChanged: (value) {
                        if (value != null) setState(() => budget = value);
                      },
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: extraContextController,
                      maxLines: 3,
                      decoration: const InputDecoration(
                        labelText: 'Доп. контекст',
                        hintText: 'например: вдвоём, на вечер, без машины',
                      ),
                    ),
                  ],
                ),
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(ctx),
                  child: const Text('Отмена'),
                ),
                ElevatedButton(
                  onPressed: () async {
                    final location = locationController.text.trim();
                    final extraContext = extraContextController.text.trim();

                    Navigator.pop(ctx);

                    final vm = context.read<LeisureIdeaViewModel>();
                    final error = await vm.generateIdea(
                      LeisureIdeaRequest(
                        location: location,
                        activityLevel: activityLevel,
                        budget: budget,
                        extraContext: extraContext,
                      ),
                    );

                    if (!context.mounted) return;

                    if (error != null) {
                      showFmlSnackBar(context, error, backgroundColor: Colors.red);
                      return;
                    }

                    showDialog(
                      context: context,
                      builder: (_) => AlertDialog(
                        title: const Text('Вот идея'),
                        content: Text(vm.lastIdea ?? 'Идея не пришла'),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(context),
                            child: const Text('Ок'),
                          ),
                        ],
                      ),
                    );
                  },
                  child: const Text('Получить'),
                ),
              ],
            );
          },
        );
      },
    );
  }
}