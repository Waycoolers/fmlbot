import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenStorage {
  // Инициализируем безопасное хранилище устройства
  final _storage = const FlutterSecureStorage();

  // Сохраняем оба токена
  Future<void> saveTokens({required String access, required String refresh}) async {
    await _storage.write(key: 'access_token', value: access);
    await _storage.write(key: 'refresh_token', value: refresh);
  }

  // Достаем access токен
  Future<String?> getAccessToken() async {
    return await _storage.read(key: 'access_token');
  }

  // Достаем refresh токен
  Future<String?> getRefreshToken() async {
    return await _storage.read(key: 'refresh_token');
  }

  // Удаляем токены (при выходе из аккаунта)
  Future<void> clearTokens() async {
    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
  }
}