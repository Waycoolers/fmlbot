import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../view_models/compliment_view_model.dart';
import '../models/compliment_model.dart';
import '../view_models/user_view_model.dart';

class ComplimentsScreen extends StatefulWidget {
  const ComplimentsScreen({super.key});

  @override
  State<ComplimentsScreen> createState() => _ComplimentsScreenState();
}

class _ComplimentsScreenState extends State<ComplimentsScreen> {
  final TextEditingController _textController = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<ComplimentViewModel>().fetchData();
    });
  }

  @override
  Widget build(BuildContext context) {
    final vm = context.watch<ComplimentViewModel>();
    final userVM = context.watch<UserViewModel>(); // <-- ДОБАВИЛИ ЭТУ СТРОКУ

    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Комплименты'),
          bottom: const TabBar(
            tabs: [
              Tab(text: 'Мне'),
              Tab(text: 'От меня'),
            ],
          ),
        ),
        body: vm.isLoading && vm.config == null
            ? const Center(child: CircularProgressIndicator())
            : TabBarView(
          children: [
            _buildMyComplimentsTab(vm, userVM), // <-- ПЕРЕДАЕМ userVM
            _buildSentComplimentsTab(vm, userVM), // <-- ПЕРЕДАЕМ userVM
          ],
        ),
      ),
    );
  }

  // --- Вкладка "Мне" ---
  Widget _buildMyComplimentsTab(ComplimentViewModel vm, UserViewModel userVM) {
    final hasPartner = userVM.partner != null;

    return Column(
      children: [
        if (hasPartner)
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: ElevatedButton.icon(
              icon: const Icon(Icons.download),
              label: const Text(
                'Получить комплимент',
                style: TextStyle(fontSize: 18),
              ),
              style: ElevatedButton.styleFrom(
                minimumSize: const Size.fromHeight(60),
                backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                foregroundColor: Theme.of(context).colorScheme.onPrimaryContainer,
              ),
              // КНОПКА ТЕПЕРЬ ВСЕГДА АКТИВНА!
              onPressed: () async {
                // Пытаемся вытянуть комплимент и ждем результат (String или null)
                final errorMessage = await vm.receiveNextCompliment();

                // Если пришла ошибка (не null), показываем её текст
                if (errorMessage != null && mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text(errorMessage),
                      backgroundColor: Colors.orange.shade700, // Красивый оранжевый для предупреждений
                      duration: const Duration(seconds: 3),
                    ),
                  );
                } else if (mounted) {
                  // Если вернулся null, значит успех!
                  // Можно даже добавить радостный SnackBar:
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Новый комплимент получен! ❤️'),
                      backgroundColor: Colors.green,
                    ),
                  );
                }
              },
            ),
          ),

        if (!hasPartner)
          const Padding(
            padding: EdgeInsets.all(16.0),
            child: Text('Добавь партнера, чтобы получать комплименты!', style: TextStyle(color: Colors.grey)),
          ),

        Expanded(child: _buildList(vm.myCompliments, isMyTab: true)),
      ],
    );
  }

  // --- Вкладка "От меня" ---
  Widget _buildSentComplimentsTab(ComplimentViewModel vm, UserViewModel userVM) {
    final hasPartner = userVM.partner != null;

    return Column(
      children: [
        if (hasPartner)
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _textController,
                    decoration: const InputDecoration(
                      hintText: 'Написать комплимент...',
                      border: OutlineInputBorder(),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                IconButton(
                  icon: const Icon(Icons.send, color: Colors.blue),
                  onPressed: () async {
                    if (_textController.text.isNotEmpty) {
                      final success = await vm.sendCompliment(_textController.text);
                      if (success) {
                        _textController.clear();
                        FocusScope.of(context).unfocus();
                      }
                    }
                  },
                )
              ],
            ),
          ),

        if (!hasPartner)
          const Padding(
            padding: EdgeInsets.all(16.0),
            child: Text('Добавь партнера, чтобы писать комплименты!', style: TextStyle(color: Colors.grey)),
          ),

        Expanded(child: _buildList(vm.sentCompliments, isMyTab: false)),
      ],
    );
  }

  // Универсальный генератор списка
  Widget _buildList(List<ComplimentModel> list, {bool isMyTab = false}) {
    return RefreshIndicator(
      onRefresh: () async {
        await context.read<ComplimentViewModel>().fetchData();
      },
      child: list.isEmpty
          ? ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: const [
          SizedBox(height: 100),
          Center(child: Text('Пока ничего нет 😔', textAlign: TextAlign.center)),
        ],
      )
          : ListView.builder(
        physics: const AlwaysScrollableScrollPhysics(),
        itemCount: list.length,
        itemBuilder: (context, index) {
          final comp = list[index];
          final dateStr = "${comp.createdAt.day.toString().padLeft(2, '0')}.${comp.createdAt.month.toString().padLeft(2, '0')} в ${comp.createdAt.hour.toString().padLeft(2, '0')}:${comp.createdAt.minute.toString().padLeft(2, '0')}";

          final card = Card(
            margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            elevation: 2,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            child: ListTile(
              title: Text(comp.text, style: const TextStyle(fontSize: 16)),
              subtitle: Padding(
                padding: const EdgeInsets.only(top: 8.0),
                child: Text(dateStr, style: const TextStyle(color: Colors.grey)),
              ),
              trailing: isMyTab
                  ? null
                  : Icon(
                comp.isSent ? Icons.done_all : Icons.access_time,
                color: comp.isSent ? Colors.blue : Colors.grey,
              ),
            ),
          );

          if (isMyTab || comp.isSent || comp.id == null) {
            return card;
          }

          return Dismissible(
            key: ValueKey('comp_${comp.id}'),
            direction: DismissDirection.endToStart,
            background: Container(
              alignment: Alignment.centerRight,
              padding: const EdgeInsets.only(right: 20.0),
              decoration: BoxDecoration(
                color: Colors.red.shade400,
                borderRadius: BorderRadius.circular(12),
              ),
              margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: const Icon(Icons.delete, color: Colors.white),
            ),
            confirmDismiss: (direction) async {
              return await showDialog(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: const Text("Удалить комплимент?"),
                    content: const Text("Партнер еще не видел его. Удалить навсегда?"),
                    actions: [
                      TextButton(onPressed: () => Navigator.of(context).pop(false), child: const Text("Отмена")),
                      TextButton(
                        onPressed: () => Navigator.of(context).pop(true),
                        style: TextButton.styleFrom(foregroundColor: Colors.red),
                        child: const Text("Удалить"),
                      ),
                    ],
                  );
                },
              );
            },
            onDismissed: (direction) {
              context.read<ComplimentViewModel>().deleteCompliment(comp.id!);
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Комплимент удален')),
              );
            },
            child: card,
          );
        },
      ),
    );
  }
}