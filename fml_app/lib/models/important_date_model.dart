class ImportantDateModel {
  final int id;
  final String title;
  final DateTime date;
  final bool isActive;
  final int notifyBeforeDays;
  final int? userId; // Чья это дата
  final int? partnerId; // Если не null, значит дата общая

  ImportantDateModel({
    required this.id,
    required this.title,
    required this.date,
    required this.isActive,
    required this.notifyBeforeDays,
    this.userId,
    this.partnerId,
  });

  factory ImportantDateModel.fromJson(Map<String, dynamic> json) {
    return ImportantDateModel(
      id: json['id'] ?? 0,
      title: json['title'] ?? '',
      date: json['date'] != null ? DateTime.parse(json['date']) : DateTime.now(),
      isActive: json['is_active'] ?? true,
      notifyBeforeDays: json['notify_before_days'] ?? 0,
      // В Go sql.NullInt64 обычно отдается как объект { "Int64": 123, "Valid": true } или просто число,
      // если ты настроил кастомный маршалинг. Достанем безопасно:
      userId: json['user_id'] is Map ? json['user_id']['Int64'] : json['user_id'],
      partnerId: json['partner_id'] is Map ? json['partner_id']['Int64'] : json['partner_id'],
    );
  }

  // Удобный геттер: общая ли это дата?
  bool get isShared => partnerId != null && partnerId! > 0;
}