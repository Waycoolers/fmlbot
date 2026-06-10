import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import '../main.dart';
import '../views/login_screen.dart';
import 'token_storage.dart';

class ApiClient {
  late final Dio dio;
  final TokenStorage _tokenStorage = TokenStorage();

  // Обязательно подставь актуальный IP твоего компьютера в локальной сети!
  final String apiBaseUrl = 'http://192.168.0.100:8080';
  final String authBaseUrl = 'http://192.168.0.100:8081';

  ApiClient() {
    // Базовые настройки Dio
    dio = Dio(BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      contentType: 'application/json',
    ));

    // Используем QueuedInterceptorsWrapper для блокировки очереди при обновлении токена
    dio.interceptors.add(QueuedInterceptorsWrapper(

      // 1. ПЕРЕД КАЖДЫМ ЗАПРОСОМ: Достаем токен и клеим его в заголовки
      onRequest: (options, handler) async {
        final accessToken = await _tokenStorage.getAccessToken();
        if (accessToken != null) {
          options.headers['Authorization'] = 'Bearer $accessToken';
        }
        return handler.next(options);
      },

      // 2. ЕСЛИ ПРОИЗОШЛА ОШИБКА
      onError: (DioException e, handler) async {
        // Если ошибка 401 (Токен протух)
        if (e.response?.statusCode == 401) {
          final refreshToken = await _tokenStorage.getRefreshToken();

          if (refreshToken != null) {
            try {
              // ВАЖНО: Создаем новый чистый экземпляр Dio для рефреша,
              // чтобы он случайно не попал в этот же интерсептор и не вызвал бесконечный цикл.
              final refreshDio = Dio();

              // Дергаем твой Auth Service
              final refreshResponse = await refreshDio.post(
                '$authBaseUrl/auth/refresh',
                data: {'refresh_token': refreshToken},
              );

              // Достаем свежие токены
              final newAccess = refreshResponse.data['access_token'];
              final newRefresh = refreshResponse.data['refresh_token'];

              // Сохраняем их в защищенное хранилище
              await _tokenStorage.saveTokens(access: newAccess, refresh: newRefresh);

              // ПОЛНОСТЬЮ ПЕРЕЗАПИСЫВАЕМ ЗАГОЛОВОК, чтобы не было "Bearer старый, Bearer новый"
              final newHeaders = Map<String, dynamic>.from(e.requestOptions.headers);
              newHeaders['Authorization'] = 'Bearer $newAccess';

              final clonedOptions = e.requestOptions.copyWith(headers: newHeaders);

              // Повторяем упавший запрос уже с новым токеном!
              final retryResponse = await dio.fetch(clonedOptions);
              return handler.resolve(retryResponse);

            } catch (refreshError) {
              // Если обновить не вышло (Refresh Token протух)
              await _tokenStorage.clearTokens();

              // ПРИНУДИТЕЛЬНО КИДАЕМ НА ЭКРАН ЛОГИНА
              globalNavigatorKey.currentState?.pushAndRemoveUntil(
                MaterialPageRoute(builder: (_) => const LoginScreen()),
                    (route) => false,
              );

              return handler.next(e);
            }
          } else {
            // Если Refresh токена изначально нет
            await _tokenStorage.clearTokens();
            globalNavigatorKey.currentState?.pushAndRemoveUntil(
              MaterialPageRoute(builder: (_) => const LoginScreen()),
                  (route) => false,
            );
          }
        }

        // Передаем любую другую ошибку (500, 404, таймаут) дальше
        return handler.next(e);
      },
    ));
  }
}