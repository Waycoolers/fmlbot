import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../models/important_date_model.dart';
import '../services/api_client.dart';

class ImportantDateViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();

  List<ImportantDateModel> dates = [];
  bool isLoading = false;
  String? errorMessage;

  Future<void> fetchDates() async {
    isLoading = true;
    errorMessage = null;
    notifyListeners();

    try {
      final response = await _apiClient.dio.get('/important_dates');
      final List<dynamic> data = response.data ?? [];

      dates = data.map((json) => ImportantDateModel.fromJson(json)).toList();

      // Сортируем даты так, чтобы ближайшие были сверху
      dates.sort((a, b) => a.date.compareTo(b.date));

    } on DioException catch (e) {
      errorMessage = "Ошибка загрузки дат: ${e.response?.statusCode}";
    } catch (e) {
      errorMessage = "Непредвиденная ошибка";
    }

    isLoading = false;
    notifyListeners();
  }

  // Добавление новой даты
  Future<bool> addDate({
    required String title,
    required DateTime date,
    required bool isShared,
    required int notifyBeforeDays,
  }) async {
    try {
      await _apiClient.dio.post('/important_dates', data: {
        'title': title,
        'date': date.toUtc().toIso8601String(), // Конвертируем в формат, понятный Go (RFC3339)
        'is_shared': isShared,
        'notify_before_days': notifyBeforeDays,
        'is_active': true,
        // user_id бэкенд сам достанет из твоего токена авторизации
      });

      await fetchDates(); // Сразу обновляем список после успешного добавления
      return true;
    } on DioException catch (e) {
      print('Ошибка при создании даты: ${e.response?.statusCode}');
      return false;
    } catch (e) {
      print('Непредвиденная ошибка: $e');
      return false;
    }
  }

  // Удаление даты
  Future<bool> deleteDate(int id) async {
    try {
      await _apiClient.dio.delete('/important_dates/$id');

      // Удаляем дату из локального списка, чтобы интерфейс обновился мгновенно,
      // не дожидаясь ответа от сервера с новым списком
      dates.removeWhere((d) => d.id == id);
      notifyListeners();

      return true;
    } on DioException catch (e) {
      print('Ошибка при удалении даты: ${e.response?.statusCode}');
      return false;
    } catch (e) {
      print('Непредвиденная ошибка: $e');
      return false;
    }
  }

  // Обновление даты
  Future<bool> updateDate({
    required int id,
    required String title,
    required DateTime date,
    required bool isShared,
    required int notifyBeforeDays,
  }) async {
    try {
      // 1. Обновляем основные текстовые и временные данные
      await _apiClient.dio.patch('/important_dates/$id', data: {
        'title': title,
        'date': date.toUtc().toIso8601String(),
        'notify_before_days': notifyBeforeDays,
        'is_active': true,
      });

      // 2. Отдельно обновляем статус "Общая дата с партнером"
      await _apiClient.dio.patch('/important_dates/$id/sharing', data: {
        'make_shared': isShared,
      });

      await fetchDates(); // Обновляем список после успешных запросов
      return true;
    } on DioException catch (e) {
      print('Ошибка при обновлении даты: ${e.response?.statusCode}');
      return false;
    } catch (e) {
      print('Непредвиденная ошибка: $e');
      return false;
    }
  }
}