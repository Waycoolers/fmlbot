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
      final cleanUsername = username.trim().replaceAll('@', '');
      final cleanPassword = password.trim();

      final response = await _authClient.dio.post('/auth/token', data: {
        'username': cleanUsername,
        'password': cleanPassword,
      });

      final accessToken = response.data['access_token'];
      final refreshToken = response.data['refresh_token'];

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
            _errorMessage = 'Пользователь не найден. Вы уже зарегистрировались?';
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

  // НОВЫЙ МЕТОД: Настоящая регистрация
  Future<bool> register(String username, String password) async {
    _isLoading = true;
    _errorMessage = null;
    notifyListeners();

    try {
      final cleanUsername = username.trim().replaceAll('@', '');
      final cleanPassword = password.trim();

      final response = await _authClient.dio.post('/register', data: {
        'username': cleanUsername,
        'password': cleanPassword,
      });

      // Go-бэкенд возвращает 200 OK и сразу отдает готовую пару токенов!
      if (response.statusCode == 200 || response.statusCode == 201) {
        final accessToken = response.data['access_token'];
        final refreshToken = response.data['refresh_token'];

        // Сразу сохраняем токены, не вызывая метод login повторно
        await _tokenStorage.saveTokens(
          access: accessToken,
          refresh: refreshToken,
        );

        _isLoading = false;
        notifyListeners();
        return true;
      } else {
        _errorMessage = 'Не удалось зарегистрироваться';
        _isLoading = false;
        notifyListeners();
        return false;
      }
    } on DioException catch (e) {
      _isLoading = false;
      if (e.response != null) {
        final statusCode = e.response!.statusCode;
        if (statusCode == 409) {
          _errorMessage = 'Этот логин уже занят';
        } else if (statusCode == 400) {
          final data = e.response!.data;
          if (data is Map<String, dynamic> && data['error'] != null) {
            _errorMessage = data['error'].toString();
          } else {
            _errorMessage = 'Логин или пароль не соответствуют требованиям';
          }
        } else {
          _errorMessage = 'Ошибка сервера при регистрации';
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

  // Метод очистки ошибки (чтобы сбрасывать красную надпись при переключении Вход/Регистрация)
  void clearError() {
    if (_errorMessage != null) {
      _errorMessage = null;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    await _tokenStorage.clearTokens();
  }

  Future<bool> checkAutoLogin() async {
    final accessToken = await _tokenStorage.getAccessToken();
    return accessToken != null;
  }
}