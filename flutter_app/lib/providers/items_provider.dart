import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/item.dart';
import '../services/api_client.dart';
import '../core/utils/snackbar_util.dart';

part 'items_provider.g.dart';

@riverpod
class Items extends _$Items {
  @override
  FutureOr<List<Item>> build() async {
    final client = ApiClient.instance;
    final response = await client.get('/v1/items');
    final itemsList = (response['data'] as List)
        .map((json) => Item.fromJson(json))
        .toList();
    return itemsList;
  }

  Future<void> createItem(String name, String description) async {
    final client = ApiClient.instance;
    try {
      await client.post('/v1/items', {
        'name': name,
        'description': description,
      });
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('Item created');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }

  Future<void> updateItem(int id, String name, String description) async {
    final client = ApiClient.instance;
    try {
      await client.put('/v1/items/$id', {
        'name': name,
        'description': description,
      });
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('Item updated');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }

  Future<void> deleteItem(int id) async {
    final client = ApiClient.instance;
    try {
      await client.delete('/v1/items/$id');
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('Item deleted');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }
}
