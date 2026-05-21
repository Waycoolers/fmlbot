import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/material.dart';
import 'package:fml_app/services/notification_service.dart';
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

@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
  print("Фоновое уведомление получено: ${message.messageId}");
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );
  await NotificationService().init();

  runApp(
    // Оборачиваем приложение в MultiProvider
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthViewModel()),
        ChangeNotifierProvider(create: (_) => UserViewModel()),
        ChangeNotifierProvider(create: (_) => ComplimentViewModel()),
        ChangeNotifierProvider(create: (_) => ImportantDateViewModel()),
        ChangeNotifierProvider(create: (_) => SettingsViewModel()),
        ChangeNotifierProvider(create: (_) => ThemeViewModel()),
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