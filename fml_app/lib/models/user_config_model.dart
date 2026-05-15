class UserConfigModel {
  final int maxComplimentCount;
  final int complimentTokenBucket;
  final DateTime? dailyMessageTime;

  UserConfigModel({
    required this.maxComplimentCount,
    required this.complimentTokenBucket,
    this.dailyMessageTime,
  });

  factory UserConfigModel.fromJson(Map<String, dynamic> json) {
    return UserConfigModel(
      maxComplimentCount: json['max_compliment_count'] ?? 0,
      complimentTokenBucket: json['compliment_token_bucket'] ?? 0,
      dailyMessageTime: json['daily_message_time'] != null
          ? DateTime.tryParse(json['daily_message_time'])
          : null,
    );
  }
}