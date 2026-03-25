import 'package:flutter_dotenv/flutter_dotenv.dart';

/// Centralised access to runtime configuration loaded from .env
class AppConfig {
  AppConfig._();

  static String get apiBaseUrl =>
      dotenv.env['API_BASE_URL'] ?? 'http://localhost:8080';

  static String get apiKey => dotenv.env['API_KEY'] ?? '';
}
