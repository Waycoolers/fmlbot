import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import '../models/leisure_idea_request.dart';
import '../services/api_client.dart';

class LeisureIdeaViewModel extends ChangeNotifier {
  final ApiClient _apiClient = ApiClient();

  bool isLoading = false;
  String? lastIdea;

  Future<String?> generateIdea(LeisureIdeaRequest request) async {
    isLoading = true;
    notifyListeners();

    try {
      final response = await _apiClient.dio.post(
        '/ideas/leisure',
        data: request.toJson(),
      );

      final data = response.data;

      if (data is Map<String, dynamic>) {
        lastIdea = (data['idea'] ?? data['text'] ?? data['message'] ?? data).toString();
      } else {
        lastIdea = data.toString();
      }

      return null;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        return 'Нужна авторизация';
      }
      return e.response?.data?.toString() ?? 'Ошибка при получении идеи';
    } catch (e) {
      return 'Непредвиденная ошибка: $e';
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }
}