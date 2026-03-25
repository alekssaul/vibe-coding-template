// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'items_provider.dart';

// **************************************************************************
// RiverpodGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, type=warning

@ProviderFor(Items)
final itemsProvider = ItemsProvider._();

final class ItemsProvider extends $AsyncNotifierProvider<Items, List<Item>> {
  ItemsProvider._()
    : super(
        from: null,
        argument: null,
        retry: null,
        name: r'itemsProvider',
        isAutoDispose: true,
        dependencies: null,
        $allTransitiveDependencies: null,
      );

  @override
  String debugGetCreateSourceHash() => _$itemsHash();

  @$internal
  @override
  Items create() => Items();
}

String _$itemsHash() => r'c2fb6b7c0f2bcaeedcd92aa250b77bc9c8e94a80';

abstract class _$Items extends $AsyncNotifier<List<Item>> {
  FutureOr<List<Item>> build();
  @$mustCallSuper
  @override
  void runBuild() {
    final ref = this.ref as $Ref<AsyncValue<List<Item>>, List<Item>>;
    final element =
        ref.element
            as $ClassProviderElement<
              AnyNotifier<AsyncValue<List<Item>>, List<Item>>,
              AsyncValue<List<Item>>,
              Object?,
              Object?
            >;
    element.handleCreate(ref, build);
  }
}
