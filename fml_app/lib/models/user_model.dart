class UserModel {
  final int userId;
  final String username;
  final int partnerId;

  UserModel({
    required this.userId,
    required this.username,
    required this.partnerId,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      userId: json['user_id'] ?? 0,
      username: json['username'] ?? '',
      partnerId: json['partner_id'] ?? 0,
    );
  }
}