import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../view_models/important_date_view_model.dart';
import '../models/important_date_model.dart';
import '../view_models/user_view_model.dart';

class AddImportantDateScreen extends StatefulWidget {
  final ImportantDateModel? dateToEdit; // Если null - создаем новую, если передали - редактируем

  const AddImportantDateScreen({super.key, this.dateToEdit});

  @override
  State<AddImportantDateScreen> createState() => _AddImportantDateScreenState();
}

class _AddImportantDateScreenState extends State<AddImportantDateScreen> {
  late final TextEditingController _titleController;
  late final TextEditingController _notifyController;
  late DateTime _selectedDate;
  late bool _isShared;
  bool _isSaving = false;

  bool get _isEditing => widget.dateToEdit != null;

  @override
  void initState() {
    super.initState();
    // Если передали дату для редактирования, заполняем поля её значениями
    _titleController = TextEditingController(text: widget.dateToEdit?.title ?? '');
    _notifyController = TextEditingController(text: widget.dateToEdit?.notifyBeforeDays.toString() ?? '7');
    _selectedDate = widget.dateToEdit?.date ?? DateTime.now();

    // По умолчанию дата не общая, если нет партнера
    final partnerExists = context.read<UserViewModel>().partner != null;
    _isShared = widget.dateToEdit?.isShared ?? (partnerExists ? true : false);
  }

  @override
  Widget build(BuildContext context) {
    final userVM = context.watch<UserViewModel>();
    final hasPartner = userVM.partner != null;

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditing ? 'Редактировать дату' : 'Добавить дату'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // 1. Название
            TextField(
              controller: _titleController,
              decoration: const InputDecoration(labelText: 'Название', border: OutlineInputBorder()),
            ),
            const SizedBox(height: 20),

            // 2. Выбор даты
            ListTile(
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8), side: BorderSide(color: Colors.grey.shade400)),
              leading: const Icon(Icons.calendar_today, color: Colors.blue),
              title: Text('Дата: ${_selectedDate.day.toString().padLeft(2, '0')}.${_selectedDate.month.toString().padLeft(2, '0')}.${_selectedDate.year}'),
              trailing: const Icon(Icons.edit),
              onTap: () async {
                final picked = await showDatePicker(
                  context: context,
                  initialDate: _selectedDate,
                  firstDate: DateTime(1900),
                  lastDate: DateTime(2100),
                );
                if (picked != null) setState(() => _selectedDate = picked);
              },
            ),
            const SizedBox(height: 20),

            // 3. Переключатель: Показываем только если есть партнер
            if (hasPartner) ...[
              SwitchListTile(
                title: const Text('Общая с партнером'),
                subtitle: const Text('Твой партнер тоже увидит эту дату'),
                value: _isShared,
                onChanged: (val) => setState(() => _isShared = val),
              ),
              const SizedBox(height: 20),
            ],

            // 4. Уведомления
            TextField(
              controller: _notifyController,
              keyboardType: TextInputType.number,
              decoration: const InputDecoration(
                  labelText: 'Уведомить за (дней)',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.notifications)
              ),
            ),
            const SizedBox(height: 40),

            // 5. Кнопка сохранения
            ElevatedButton(
              onPressed: _isSaving ? null : _saveDate,
              style: ElevatedButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 16)),
              child: _isSaving
                  ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2))
                  : Text(_isEditing ? 'Сохранить изменения' : 'Создать', style: const TextStyle(fontSize: 18)),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _saveDate() async {
    final title = _titleController.text.trim();
    final notifyDays = int.tryParse(_notifyController.text.trim());

    if (title.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Введи название даты'),
            backgroundColor: Colors.red,
          )
      );
      return;
    }

    if (notifyDays == null || notifyDays < 0 || notifyDays > 30) {
      ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Уведомление должно быть от 0 до 30 дней'),
            backgroundColor: Colors.orange,
          )
      );
      return;
    }

    setState(() => _isSaving = true);
    bool success;

    if (_isEditing) {
      success = await context.read<ImportantDateViewModel>().updateDate(
        id: widget.dateToEdit!.id,
        title: title,
        date: _selectedDate,
        isShared: _isShared,
        notifyBeforeDays: notifyDays,
      );
    } else {
      success = await context.read<ImportantDateViewModel>().addDate(
        title: title,
        date: _selectedDate,
        isShared: _isShared,
        notifyBeforeDays: notifyDays,
      );
    }

    setState(() => _isSaving = false);

    if (success && mounted) {
      Navigator.pop(context);
      ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(_isEditing ? 'Изменения сохранены!' : 'Дата добавлена!'),
            backgroundColor: Colors.green,
          )
      );
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _notifyController.dispose();
    super.dispose();
  }
}