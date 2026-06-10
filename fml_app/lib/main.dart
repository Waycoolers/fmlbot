import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:fml_app/view_models/settings_view_model.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:fml_app/views/auth_wrapper.dart';
import 'firebase_options.dart';
import 'view_models/important_date_view_model.dart';
import 'package:provider/provider.dart';
import 'view_models/auth_view_model.dart';
import 'view_models/user_view_model.dart';
import 'view_models/compliment_view_model.dart';
import 'view_models/theme_view_model.dart';
import 'view_models/leisure_idea_view_model.dart';

@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
  print("Фоновое уведомление получено: ${message.messageId}");
}

final GlobalKey<NavigatorState> globalNavigatorKey = GlobalKey<NavigatorState>();

Future<void> main() async {
  // Обязательная инициализация движка Flutter
  WidgetsFlutterBinding.ensureInitialized();

  // Инициализируем ядро Firebase (это быстро)
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  // Регистрируем фоновый обработчик (он должен быть тут)
  FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);

  // СРАЗУ запускаем UI! Никаких долгих await перед runApp.
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthViewModel()),
        ChangeNotifierProvider(create: (_) => UserViewModel()),
        ChangeNotifierProvider(create: (_) => ComplimentViewModel()),
        ChangeNotifierProvider(create: (_) => ImportantDateViewModel()),
        ChangeNotifierProvider(create: (_) => SettingsViewModel()),
        ChangeNotifierProvider(create: (_) => ThemeViewModel()),
        ChangeNotifierProvider(create: (_) => LeisureIdeaViewModel()),
      ],
      child: const FmlApp(),
    ),
  );
}

class FmlApp extends StatelessWidget {
  const FmlApp({super.key});

  @override
  Widget build(BuildContext context) {
    final themeVM = context.watch<ThemeViewModel>();
    return MaterialApp(
      title: 'FML App',
      navigatorKey: globalNavigatorKey,
      themeMode: themeVM.themeMode,
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: Colors.pink,
        brightness: Brightness.light,
        useMaterial3: true,
      ),

      // Темная тема
      darkTheme: ThemeData(
        colorSchemeSeed: Colors.pink,
        brightness: Brightness.dark,
        useMaterial3: true,
      ),
      home: const AuthWrapper(),
    );
  }
}