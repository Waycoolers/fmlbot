import 'package:dio/dio.dart';

class AuthClient {
  late final Dio dio;

  // 10.0.2.2 - это localhost компьютера для эмулятора Android.
  final String authBaseUrl = 'http://192.168.0.100:8081'; // Порт твоего Auth Service

  AuthClient() {
    // Базовые настройки Dio
    dio = Dio(BaseOptions(
      baseUrl: authBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      contentType: 'application/json',
    ));
  }
}