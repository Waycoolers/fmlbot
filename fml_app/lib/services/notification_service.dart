import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:dio/dio.dart';
import 'api_client.dart';

class NotificationService {
  final FirebaseMessaging _fcm = FirebaseMessaging.instance;
  final ApiClient _apiClient = ApiClient();

  // Создаём экземпляр для локальных уведомлений
  final FlutterLocalNotificationsPlugin _localNotifications =
  FlutterLocalNotificationsPlugin();

  Future<void> init() async {
    // 1. Запрос разрешения на push-уведомления
    final settings = await _fcm.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );

    if (settings.authorizationStatus == AuthorizationStatus.authorized ||
        settings.authorizationStatus == AuthorizationStatus.provisional) {

      // 2. Настройка канала уведомлений (обязательно для Android)
      const androidChannel = AndroidNotificationChannel(
        'compliments_channel',
        'Комплименты',
        description: 'Уведомления о новых комплиментах',
        importance: Importance.high,
      );

      final androidPlugin =
      _localNotifications.resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin>();
      await androidPlugin?.createNotificationChannel(androidChannel);

      // 3. Инициализация локальных уведомлений (для iOS и Android)
      const initSettings = InitializationSettings(
        android: AndroidInitializationSettings('ic_notification'),
        iOS: DarwinInitializationSettings(),
      );
      await _localNotifications.initialize(initSettings);

      // 4. Получение и отправка FCM-токена
      final token = await _fcm.getToken();
      if (token != null) {
        await _sendTokenToBackend(token);
      }

      // 5. Слежение за обновлением токена
      _fcm.onTokenRefresh.listen((newToken) async {
        await _sendTokenToBackend(newToken);
      });

      // 6. Обработка входящих уведомлений
      FirebaseMessaging.onMessage.listen(_handleForegroundMessage);
      FirebaseMessaging.onMessageOpenedApp.listen(_handleMessageOpened);

      // 7. Проверка, не запущено ли приложение по уведомлению (из убитого состояния)
      final initialMessage = await _fcm.getInitialMessage();
      if (initialMessage != null) {
        _handleMessageOpened(initialMessage);
      }
    }
  }

  // Когда приложение на переднем плане
  void _handleForegroundMessage(RemoteMessage message) {
    // Извлекаем данные
    final notification = message.notification;
    if (notification != null) {
      // Показываем локальное уведомление
      _showLocalNotification(notification.title ?? '', notification.body ?? '');
    }
  }

  // Когда приложение было в фоне/закрыто и пользователь тапнул по уведомлению
  void _handleMessageOpened(RemoteMessage message) {
    // Здесь можно реализовать переход на конкретный экран
    print('Переход по уведомлению: ${message.data}');
  }

  // Показ локального уведомления
  Future<void> _showLocalNotification(String title, String body) async {
    const androidDetails = AndroidNotificationDetails(
      'compliments_channel',
      'Комплименты',
      channelDescription: 'Уведомления о новых комплиментах',
      importance: Importance.high,
      priority: Priority.high,
      icon: 'ic_notification',
    );
    const iosDetails = DarwinNotificationDetails();
    const details = NotificationDetails(android: androidDetails, iOS: iosDetails);

    await _localNotifications.show(
      0, // ID уведомления (можно генерировать динамически)
      title,
      body,
      details,
    );
  }

  // Отправка токена на бэкенд
  Future<void> _sendTokenToBackend(String token) async {
    try {
      await _apiClient.dio.post('/users/fcm-token', data: {
        'fcm_token': token,
      });
      print('FCM Token успешно отправлен на бэкенд');
    } catch (e) {
      print('Ошибка при отправке FCM Token: $e');
    }
  }
}