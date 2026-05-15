import 'package:flutter/material.dart';
import 'package:fml_app/view_models/settings_view_model.dart';
import 'view_models/important_date_view_model.dart';
import 'package:provider/provider.dart';
import 'view_models/auth_view_model.dart';
import 'view_models/user_view_model.dart';
import 'view_models/compliment_view_model.dart';
import 'views/login_screen.dart';

void main() {
  runApp(
    // Оборачиваем приложение в MultiProvider
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthViewModel()),
        ChangeNotifierProvider(create: (_) => UserViewModel()),
        ChangeNotifierProvider(create: (_) => ComplimentViewModel()),
        ChangeNotifierProvider(create: (_) => ImportantDateViewModel()),
        ChangeNotifierProvider(create: (_) => SettingsViewModel()),
      ],
      child: const FmlApp(),
    ),
  );
}

class FmlApp extends StatelessWidget {
  const FmlApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FML App',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepPurple),
        useMaterial3: true,
      ),
      home: const LoginScreen(),
    );
  }
}