class ComplimentModel {
  final int? id;
  final String text;
  final bool isSent;
  final DateTime createdAt;

  ComplimentModel({
    this.id,
    required this.text,
    required this.isSent,
    required this.createdAt,
  });

  factory ComplimentModel.fromJson(Map<String, dynamic> json) {
    return ComplimentModel(
      id: json['id'],
      text: json['text'] ?? '',
      isSent: json['is_sent'] ?? false,
      // Парсим дату (если вдруг придет null, ставим текущую)
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }
}