import 'package:dio/dio.dart';
import 'token_storage.dart';

class ApiClient {
  late final Dio dio;
  final TokenStorage _tokenStorage = TokenStorage();

  // 10.0.2.2 - это localhost компьютера для эмулятора Android.
  final String apiBaseUrl = 'http://10.0.2.2:8080'; // Порт твоего API Service
  final String authBaseUrl = 'http://10.0.2.2:8081'; // Порт твоего Auth Service

  ApiClient() {
    // Базовые настройки Dio
    dio = Dio(BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 5),
      receiveTimeout: const Duration(seconds: 3),
      contentType: 'application/json',
    ));

    // Добавляем перехватчик (Interceptor)
    dio.interceptors.add(InterceptorsWrapper(

      // 1. ПЕРЕД КАЖДЫМ ЗАПРОСОМ: Достаем токен и клеим его в заголовки
      onRequest: (options, handler) async {
        final accessToken = await _tokenStorage.getAccessToken();
        if (accessToken != null) {
          options.headers['Authorization'] = 'Bearer $accessToken';
        }
        return handler.next(options); // Продолжаем запрос
      },

      // 2. ЕСЛИ ПРОИЗОШЛА ОШИБКА (например, токен истек)
      onError: (DioException e, handler) async {
        if (e.response?.statusCode == 401) { // 401 Unauthorized
          final refreshToken = await _tokenStorage.getRefreshToken();

          if (refreshToken != null) {
            try {
              // Пытаемся получить новую пару токенов у Auth Service
              final refreshResponse = await Dio().post(
                '$authBaseUrl/auth/refresh',
                data: {'refresh_token': refreshToken},
              );

              final newAccess = refreshResponse.data['access_token'];
              final newRefresh = refreshResponse.data['refresh_token'];

              // Сохраняем новые токены
              await _tokenStorage.saveTokens(access: newAccess, refresh: newRefresh);

              // Повторяем тот запрос, который упал с ошибкой 401, но уже с новым токеном!
              e.requestOptions.headers['Authorization'] = 'Bearer $newAccess';
              final retryResponse = await Dio().fetch(e.requestOptions);
              return handler.resolve(retryResponse); // Успешно возвращаем результат

            } catch (refreshError) {
              // Если обновить не вышло (refresh_token тоже протух) - чистим всё
              await _tokenStorage.clearTokens();
              // TODO: Сделать перенаправление на экран логина
            }
          }
        }
        return handler.next(e); // Передаем ошибку дальше, если это не 401
      },
    ));
  }
}