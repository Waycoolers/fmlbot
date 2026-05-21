import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../services/utils.dart';
import '../view_models/important_date_view_model.dart';
import '../models/important_date_model.dart';
import 'add_important_date_screen.dart';

class ImportantDatesScreen extends StatefulWidget {
  const ImportantDatesScreen({super.key});

  @override
  State<ImportantDatesScreen> createState() => _ImportantDatesScreenState();
}

class _ImportantDatesScreenState extends State<ImportantDatesScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<ImportantDateViewModel>().fetchDates();
    });
  }

  @override
  Widget build(BuildContext context) {
    final vm = context.watch<ImportantDateViewModel>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Важные даты'),
      ),
      body: vm.isLoading && vm.dates.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
        onRefresh: () async {
          await context.read<ImportantDateViewModel>().fetchDates();
        },
        child: _buildBody(vm),
      ),
      // Кнопка для добавления новой даты (плавающая внизу)
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          // Открываем экран создания даты
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const AddImportantDateScreen()),
          );
        },
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildBody(ImportantDateViewModel vm) {
    if (vm.errorMessage != null) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          const SizedBox(height: 100),
          Center(child: Text(vm.errorMessage!, style: const TextStyle(color: Colors.red))),
        ],
      );
    }

    if (vm.dates.isEmpty) {
      return ListView(
        physics: const AlwaysScrollableScrollPhysics(),
        children: const [
          SizedBox(height: 100),
          Center(child: Text('Важных дат пока нет 😔')),
        ],
      );
    }

    return ListView.builder(
      physics: const AlwaysScrollableScrollPhysics(), // Обязательно добавляем физику!
      padding: const EdgeInsets.all(16),
      itemCount: vm.dates.length,
      itemBuilder: (context, index) {
        final dateItem = vm.dates[index];
        return _DateCard(date: dateItem);
      },
    );
  }
}

// Виджет отдельной карточки даты
class _DateCard extends StatelessWidget {
  final ImportantDateModel date;

  const _DateCard({required this.date});

  @override
  Widget build(BuildContext context) {
    final dateString = "${date.date.day.toString().padLeft(2, '0')}.${date.date.month.toString().padLeft(2, '0')}.${date.date.year}";

    return Dismissible(
      key: ValueKey(date.id),
      direction: DismissDirection.endToStart,
      background: Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 20.0),
        decoration: BoxDecoration(
          color: Colors.red.shade400,
          borderRadius: BorderRadius.circular(12),
        ),
        margin: const EdgeInsets.only(bottom: 12),
        child: const Icon(Icons.delete, color: Colors.white),
      ),
      // --- ДОБАВИЛИ ПОДТВЕРЖДЕНИЕ УДАЛЕНИЯ ---
      confirmDismiss: (direction) async {
        return await showDialog(
          context: context,
          builder: (BuildContext context) {
            return AlertDialog(
              title: const Text("Подтверждение"),
              content: const Text("Ты уверен, что хочешь удалить эту дату?"),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(false), // Возвращает false (отмена)
                  child: const Text("Отмена"),
                ),
                TextButton(
                  onPressed: () => Navigator.of(context).pop(true), // Возвращает true (удаляем)
                  style: TextButton.styleFrom(foregroundColor: Colors.red),
                  child: const Text("Удалить"),
                ),
              ],
            );
          },
        );
      },
      // ----------------------------------------
      onDismissed: (direction) {
        context.read<ImportantDateViewModel>().deleteDate(date.id);
        showFmlSnackBar(context, 'Дата "${date.title}" удалена', backgroundColor: Colors.green);
      },
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              // Передаем текущую дату в форму
              builder: (context) => AddImportantDateScreen(dateToEdit: date),
            ),
          );
        },
        borderRadius: BorderRadius.circular(12),
        child: Card(
          elevation: 2,
          margin: const EdgeInsets.only(bottom: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // --- Обернули Text в Expanded, чтобы он правильно обрезался! ---
                    Expanded(
                      child: Text(
                        date.title,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                      ),
                    ),
                    const SizedBox(width: 8), // Отступ между названием и датой
                    Text(
                      dateString,
                      style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Icon(
                      date.isShared ? Icons.people : Icons.person,
                      size: 16,
                      color: date.isShared ? Colors.blue : Colors.grey,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      date.isShared ? 'Общая с партнером' : 'Для меня',
                      style: TextStyle(
                        color: date.isShared ? Colors.blue : Colors.grey[700],
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  'Напоминание: за ${date.notifyBeforeDays} дней',
                  style: TextStyle(color: Colors.grey[600], fontSize: 13),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}