import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_app/models/item.dart';
import 'package:flutter_app/services/api_client.dart';

/// Provider exposing the current paginated list of items.
final itemsProvider = AsyncNotifierProvider<ItemsNotifier, ListResponse<Item>>(
  ItemsNotifier.new,
);

class ItemsNotifier extends AsyncNotifier<ListResponse<Item>> {
  final _repo = ItemsRepository.instance;

  @override
  Future<ListResponse<Item>> build() => _repo.listItems();

  Future<void> refresh() async {
    state = const AsyncValue.loading();
    state = await AsyncValue.guard(() => _repo.listItems());
  }

  Future<void> create({required String name, String description = ''}) async {
    await _repo.createItem(name: name, description: description);
    await refresh();
  }

  Future<void> updateItem(
    int id, {
    required String name,
    String description = '',
  }) async {
    await _repo.updateItem(id, name: name, description: description);
    await refresh();
  }

  Future<void> delete(int id) async {
    await _repo.deleteItem(id);
    await refresh();
  }
}
