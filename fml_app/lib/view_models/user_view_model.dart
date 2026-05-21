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

  // Добавление партнера по никнейму
  Future<bool> addPartner(String username) async {
    isLoading = true;
    notifyListeners();

    try {
      final cleanUsername = username.replaceAll('@', '').trim();

      // 1. Сначала узнаем ID партнера по его никнейму
      final userResponse = await _apiClient.dio.get('/users/by-username/$cleanUsername');
      final partnerId = userResponse.data['user_id'];

      // 2. Отправляем запрос на создание пары
      // ВНИМАНИЕ: Я предполагаю, что твой эндпоинт POST /users/pair ожидает json с partner_id.
      // Если у тебя структура другая, поправь ключи в data.
      await _apiClient.dio.post('/users/pair', data: {
        'partner_id': partnerId
      });

      // 3. Успешно! Заново скачиваем профили, чтобы обновить Главный экран
      await fetchProfiles();
      return true;

    } on DioException catch (e) {
      print('Ошибка при добавлении партнера: ${e.response?.statusCode}');
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
}