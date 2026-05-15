import 'package:flutter/material.dart';
import 'package:dio/dio.dart';
import '../models/user_config_model.dart';
import '../services/api_client.dart';

class SettingsViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();

  UserConfigModel? myConfig;
  bool isLoading = false;
  String? errorMessage;

  // Получаем текущие настройки
  Future<void> fetchConfig() async {
    isLoading = true;
    errorMessage = null;
    notifyListeners();

    try {
      final response = await _apiClient.dio.get('/user_config/me');
      myConfig = UserConfigModel.fromJson(response.data);
    } on DioException catch (e) {
      errorMessage = "Ошибка загрузки: ${e.response?.statusCode}";
    } catch (e) {
      errorMessage = "Непредвиденная ошибка";
    }

    isLoading = false;
    notifyListeners();
  }

  // Обновляем лимит комплиментов
  Future<bool> updateMaxCompliments(int newMax) async {
    try {
      await _apiClient.dio.patch('/user_config/me', data: {
        'max_compliment_count': newMax,
      });
      // Обновляем локальные данные после успешного сохранения
      await fetchConfig();
      return true;
    } catch (e) {
      print('Ошибка при сохранении настроек: $e');
      return false;
    }
  }
}