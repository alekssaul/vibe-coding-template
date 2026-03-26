// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'auth_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning
/// Provides the stored API key (null if not yet set).

@ProviderFor(ApiKeyNotifier)
final apiKeyProvider = ApiKeyNotifierProvider._();

/// Provides the stored API key (null if not yet set).
final class ApiKeyNotifierProvider
    extends $AsyncNotifierProvider<ApiKeyNotifier, String?> {
  /// Provides the stored API key (null if not yet set).
  ApiKeyNotifierProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'apiKeyProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$apiKeyNotifierHash();

  @$internal
  @override
  ApiKeyNotifier create() => ApiKeyNotifier();
}

String _$apiKeyNotifierHash() => r'2817670aa6fff7c99d005e949337c3153a4d6e44';

/// Provides the stored API key (null if not yet set).

abstract class _$ApiKeyNotifier extends $AsyncNotifier<String?> {
  FutureOr<String?> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<String?>, String?>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<String?>, String?>,
              AsyncValue<String?>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
