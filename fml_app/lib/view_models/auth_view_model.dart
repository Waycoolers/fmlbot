import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../services/api_client.dart';
import '../services/token_storage.dart';

class AuthViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();
  final TokenStorage _tokenStorage = TokenStorage();

  bool _isLoading = false;
  bool get isLoading => _isLoading;

  String? _errorMessage;
  String? get errorMessage => _errorMessage;

  Future<bool> loginWithTelegram(String username) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners(); // Даем сигнал UI показать крутилку загрузки

    try {
      // Очищаем никнейм от '@', если пользователь случайно его ввел
      final cleanUsername = username.replaceAll('@', '').trim();

      if (cleanUsername.isEmpty) {
        _errorMessage = "Пожалуйста, введите никнейм";
        _isLoading = false;
        notifyListeners();
        return false;
      }

      // 1. Получаем ID пользователя из твоего API Service
      // ВНИМАНИЕ: Убедись, что путь '/users/by-username/' совпадает с твоим API!
      final userResponse = await _apiClient.dio.get('/users/by-username/$cleanUsername');

      // Предполагаем, что бэкенд возвращает JSON, где есть поле "id" (или "user_id")
      final userId = userResponse.data['user_id'];

      // 2. Стучимся в Auth Service за токенами (используем чистый Dio, без перехватчика)
      final authResponse = await Dio().post(
        '${_apiClient.authBaseUrl}/auth/token',
        data: {'user_id': userId},
      );

      // 3. Сохраняем токены
      final accessToken = authResponse.data['access_token'];
      final refreshToken = authResponse.data['refresh_token'];
      await _tokenStorage.saveTokens(access: accessToken, refresh: refreshToken);

      _isLoading = false;
      notifyListeners();
      return true; // Успешный вход!

    } on DioException catch (e) {
      _isLoading = false;
      if (e.response?.statusCode == 404) {
        _errorMessage = "Пользователь не найден. Запустите бота!";
      } else {
        // Теперь мы увидим реальную причину (например, Connection refused)
        _errorMessage = "Ошибка соединения: ${e.response?.statusCode}. Детали: ${e.message}";
      }
      notifyListeners();
      return false;
    } catch (e) {
      _isLoading = false;
      _errorMessage = "Непредвиденная ошибка: $e";
      notifyListeners();
      return false;
    }
  }

  // Выход из аккаунта
  Future<void> logout() async {
    // В зависимости от того, как ты назвал метод в TokenStorage:
    // Если метода deleteTokens нет, создай его там (внутри просто storage.deleteAll())
    await _tokenStorage.clearTokens();
  }
}