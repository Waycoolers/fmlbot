import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../models/user_model.dart';
import '../services/api_client.dart';

class UserViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();

  UserModel? currentUser;
  UserModel? partner;
  bool isLoading = false;

  Future<void> fetchProfiles() async {
    isLoading = true;
    notifyListeners();

    try {
      // 1. Запрашиваем свой профиль
      final myResponse = await _apiClient.dio.get('/users/me');
      currentUser = UserModel.fromJson(myResponse.data);

      // 2. Проверяем, есть ли у нас партнер (по полю partner_id из ответа)
      if (currentUser!.partnerId > 0) {
        try {
          // Запрашиваем данные партнера через специальный эндпоинт
          final partnerResponse = await _apiClient.dio.get('/users/partner');
          partner = UserModel.fromJson(partnerResponse.data);
        } on DioException catch (e) {
          // Если бэкенд возвращает 404 (Not Found), значит партнера по какой-то причине нет
          if (e.response?.statusCode == 404) {
            partner = null;
          } else {
            print('Ошибка при загрузке партнера: ${e.message}');
          }
        }
      } else {
        // Если partner_id == 0, значит мы точно не в паре
        partner = null;
      }

    } on DioException catch (e) {
      print('Ошибка сети при загрузке профиля: ${e.response?.statusCode}');
    } catch (e) {
      print('Непредвиденная ошибка: $e');
    }

    isLoading = false;
    notifyListeners();
  }

  Future<String?> addPartner(String username) async {
    isLoading = true;
    notifyListeners();

    try {
      // 1. Ищем пользователя по юзернейму
      final userResponse = await _apiClient.dio.get('/users/by-username/$username');

      final data = userResponse.data;
      // ВАЖНО: Достаем ID по обоим возможным ключам
      final partnerId = data['id'] ?? data['user_id'];

      if (partnerId == null || partnerId == 0) {
        return 'Ошибка: Сервер не вернул ID пользователя';
      }

      // ЖЕСТКАЯ защита на фронтенде от самого себя
      if (currentUser != null && partnerId == currentUser!.userId) {
        return 'Нельзя добавить самого себя! 😅';
      }

      // 2. Отправляем ID на эндпоинт создания пары
      await _apiClient.dio.post('/users/pair', data: {
        'partner_id': partnerId,
      });

      await fetchProfiles();
      return null;

    } on DioException catch (e) {
      if (e.response != null) {
        final statusCode = e.response!.statusCode;
        if (statusCode == 404) return 'Пользователь с таким ником не найден';

        final data = e.response!.data;
        if (data is Map<String, dynamic> && data['error'] != null) {
          final errStr = data['error'].toString();
          switch (errStr) {
            case 'cannot partner yourself': return 'Нельзя добавить самого себя! 😅';
            case 'already has partner': return 'У тебя уже есть пара!';
            case 'partner already has partner': return 'У этого пользователя уже есть партнер 💔';
            default: return 'Ошибка: $errStr';
          }
        }
      }
      return 'Ошибка сети. Проверь подключение';
    } catch (e) {
      return 'Произошла непредвиденная ошибка';
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  // Разрыв связи с партнером
  Future<bool> unpair() async {
    isLoading = true;
    notifyListeners();

    try {
      await _apiClient.dio.patch('/users/unpair');

      // После успешного удаления связи заново запрашиваем профили
      // Бэкенд должен вернуть нас уже без партнера (partner_id = 0)
      await fetchProfiles();
      return true;

    } on DioException catch (e) {
      print('Ошибка при разрыве связи: ${e.response?.statusCode}');
      isLoading = false;
      notifyListeners();
      return false;
    } catch (e) {
      print('Непредвиденная ошибка: $e');
      isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<bool> changePassword(String newPassword) async {
    isLoading = true;
    notifyListeners();
    try {
      // Отправляем как строку, Go json.Decoder превратит её в []byte
      await _apiClient.dio.patch('/users/me/password', data: {'password': newPassword});
      return true;
    } catch (e) {
      print('Ошибка смены пароля: $e');
      return false;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }

  Future<bool> deleteAccount() async {
    isLoading = true;
    notifyListeners();
    try {
      await _apiClient.dio.delete('/users/me');
      return true;
    } catch (e) {
      print('Ошибка удаления аккаунта: $e');
      return false;
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }
}