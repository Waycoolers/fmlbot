import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../models/compliment_model.dart';
import '../models/user_config_model.dart';
import '../services/api_client.dart';

class ComplimentViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();

  List<ComplimentModel> myCompliments = []; // Мне
  List<ComplimentModel> sentCompliments = []; // От меня
  UserConfigModel? config;

  bool isLoading = false;
  String? errorMessage;

  // Загружаем и комплименты, и лимиты
  Future<void> fetchData() async {
    isLoading = true;
    errorMessage = null;
    notifyListeners();

    try {
      // 1. Мои комплименты (Вкладка "От меня")
      final sentResponse = await _apiClient.dio.get('/compliments');
      final List<dynamic> sentData = sentResponse.data ?? [];

      sentCompliments = sentData.map((json) => ComplimentModel.fromJson(json)).toList();
      sentCompliments.sort((a, b) => b.createdAt.compareTo(a.createdAt));

      // 2. Полученные комплименты (Вкладка "Мне") - ЖДЕТ ТВОЕГО НОВОГО ЭНДПОИНТА
      try {
        final receivedResponse = await _apiClient.dio.get('/compliments/received');
        final List<dynamic> receivedData = receivedResponse.data ?? [];

        myCompliments = receivedData.map((json) => ComplimentModel.fromJson(json)).toList();
        myCompliments.sort((a, b) => b.createdAt.compareTo(a.createdAt));
      } on DioException catch (e) {
        // Если ты еще не сделал эндпоинт, бэкенд вернет 404.
        // Мы просто игнорируем ошибку, чтобы экран не падал, и оставляем список пустым.
        if (e.response?.statusCode == 404) {
          print('Эндпоинт /compliments/received еще не готов');
          myCompliments = [];
        } else {
          rethrow;
        }
      }

      // 3. Получаем конфиг (лимиты)
      final configResponse = await _apiClient.dio.get('/user_config/me');
      config = UserConfigModel.fromJson(configResponse.data);

    } on DioException catch (e) {
      errorMessage = "Ошибка загрузки: ${e.response?.statusCode}";
    } catch (e) {
      errorMessage = "Непредвиденная ошибка";
    }

    isLoading = false;
    notifyListeners();
  }

  // Отправка комплимента
  Future<bool> sendCompliment(String text) async {
    try {
      await _apiClient.dio.post('/compliments', data: {
        'text': text,
        'is_sent': true,
      });
      await fetchData(); // Обновляем списки после отправки
      return true;
    } catch (e) {
      return false;
    }
  }

  // Запрос нового комплимента от партнера
  // Запрос нового комплимента от партнера
  Future<String?> receiveNextCompliment() async {
    try {
      await _apiClient.dio.post('/compliments/next');
      await fetchData(); // Успех! Обновляем списки
      return null;
    } on DioException catch (e) {
      // Если бэкенд ответил с ошибкой (не 2xx код)
      if (e.response != null) {
        final statusCode = e.response!.statusCode;
        final data = e.response!.data;

        // 1. StatusGone (410) - Нет комплиментов
        if (statusCode == 410) {
          return '📭 Пока для тебя нет новых комплиментов.\nНамекни своему партнеру!';
        }

        // 2. StatusTooManyRequests (429) - Лимиты
        if (statusCode == 429) {
          // Проверяем, есть ли поле minutes (значит ведро пустое)
          if (data is Map<String, dynamic> && data.containsKey('minutes')) {
            final minutes = data['minutes'];
            return '⏳ Немного терпения. Следующий комплимент будет доступен через $minutes мин.';
          } else {
            // Если поля minutes нет — значит исчерпан дневной лимит
            return '🌙 На сегодня лимит исчерпан. Завтра будет продолжение 💛';
          }
        }
      }
      return 'Не удалось получить комплимент. Ошибка сети.';
    } catch (e) {
      return 'Произошла непредвиденная ошибка.';
    }
  }

  // Удаление заготовленного комплимента
  Future<bool> deleteCompliment(int id) async {
    try {
      await _apiClient.dio.delete('/compliments/$id');

      // Удаляем из локального списка, чтобы UI обновился мгновенно
      sentCompliments.removeWhere((c) => c.id == id);
      notifyListeners();

      return true;
    } catch (e) {
      print('Ошибка при удалении комплимента: $e');
      return false;
    }
  }
}