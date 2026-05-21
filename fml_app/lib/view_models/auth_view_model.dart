import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../services/auth_client.dart';
import '../services/token_storage.dart';

class AuthViewModel extends ChangeNotifier {
  final AuthClient _authClient = AuthClient();
  final TokenStorage _tokenStorage = TokenStorage();

  bool _isLoading = false;
  bool get isLoading => _isLoading;

  String? _errorMessage;
  String? get errorMessage => _errorMessage;

  // Метод входа по логину и паролю
  Future<bool> login(String username, String password) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      // Очищаем юзернейм от '@' и лишних пробелов
      final cleanUsername = username.trim().replaceAll('@', '');
      final cleanPassword = password.trim();

      // Отправляем запрос на твой микросервис авторизации
      final response = await _authClient.dio.post('/auth/token', data: {
        'username': cleanUsername,
        'password': cleanPassword,
      });

      // Достаем токены
      final accessToken = response.data['access_token'];
      final refreshToken = response.data['refresh_token'];

      // Сохраняем в защищенное хранилище
      await _tokenStorage.saveTokens(
        access: accessToken,
        refresh: refreshToken,
      );

      _isLoading = false;
      notifyListeners();
      return true;

    } on DioException catch (e) {
      _isLoading = false;

      if (e.response != null) {
        final statusCode = e.response!.statusCode;
        switch (statusCode) {
          case 404:
            _errorMessage = 'Пользователь не найден. Ты уже зарегистрировался в боте?';
            break;
          case 401:
            _errorMessage = 'Неверный пароль';
            break;
          case 400:
            _errorMessage = 'Ошибка запроса: неверный формат данных';
            break;
          case 500:
            _errorMessage = 'Ошибка на сервере. Попробуй позже';
            break;
          default:
            _errorMessage = 'Ошибка ($statusCode): ${e.response!.statusMessage}';
        }
      } else {
        _errorMessage = 'Ошибка сети. Проверь подключение';
      }

      notifyListeners();
      return false;

    } catch (e) {
      _isLoading = false;
      _errorMessage = 'Произошла непредвиденная ошибка';
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    await _tokenStorage.clearTokens();
  }

  // Проверка наличия токена для автоматического входа
  Future<bool> checkAutoLogin() async {
    final accessToken = await _tokenStorage.getAccessToken();
    // Если токен есть — возвращаем true, если пусто — false
    return accessToken != null;
  }
}