import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_app/providers/auth_provider.dart';
import 'package:flutter_app/router/app_router.dart';
import 'package:flutter_app/screens/setup_screen.dart';
import 'package:flutter_app/theme/app_theme.dart';

Future<void> main() async {
  await dotenv.load();
  runApp(const ProviderScope(child: App()));
}

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(apiKeyProvider);

    return MaterialApp(
      title: 'Template App',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      home: authState.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (_, _) => const SetupScreen(),
        data: (key) {
          if (key == null || key.isEmpty) {
            return const SetupScreen();
          }
          // Key is valid — show the main app with GoRouter.
          return _MainApp(key: const ValueKey('main'));
        },
      ),
    );
  }
}

/// Wraps the GoRouter-based app, shown only after auth.
class _MainApp extends ConsumerWidget {
  const _MainApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Template App',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      routerConfig: router,
    );
  }
}
