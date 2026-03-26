import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter_app/services/api_client.dart';

part 'auth_provider.g.dart';

const _apiKeyPrefKey = 'api_key';

/// Provides the stored API key (null if not yet set).
@riverpod
class ApiKeyNotifier extends _$ApiKeyNotifier {
  @override
  Future<String?> build() async {
    final prefs = await SharedPreferences.getInstance();
    final key = prefs.getString(_apiKeyPrefKey);
    if (key != null && key.isNotEmpty) {
      ApiClient.instance.setApiKey(key);
    }
    return key;
  }

  /// Validate and persist an API key.
  Future<void> saveKey(String key) async {
    state = const AsyncLoading();
    try {
      // Validate by hitting the health endpoint.
      ApiClient.instance.setApiKey(key);
      await ApiClient.instance.get('/health');

      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_apiKeyPrefKey, key);
      state = AsyncData(key);
    } catch (e) {
      ApiClient.instance.setApiKey('');
      state = AsyncError(
        'Invalid API key: could not connect to server',
        StackTrace.current,
      );
    }
  }

  /// Clear the stored key (logout).
  Future<void> clearKey() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_apiKeyPrefKey);
    ApiClient.instance.setApiKey('');
    state = const AsyncData(null);
  }
}
