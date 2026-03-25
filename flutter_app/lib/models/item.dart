import 'package:flutter_app/services/api_client.dart';

/// Represents a single item from the API.
class Item {
  final int id;
  final String name;
  final String description;
  final DateTime createdAt;
  final DateTime updatedAt;

  Item({
    required this.id,
    required this.name,
    required this.description,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Item.fromJson(Map<String, dynamic> json) => Item(
    id: json['id'] as int,
    name: json['name'] as String,
    description: json['description'] as String? ?? '',
    createdAt: DateTime.parse(json['created_at'] as String),
    updatedAt: DateTime.parse(json['updated_at'] as String),
  );

  Map<String, dynamic> toJson() => {'name': name, 'description': description};
}

/// Data layer for items — called directly by Riverpod providers.
class ItemsRepository {
  ItemsRepository._();
  static final ItemsRepository instance = ItemsRepository._();

  final _client = ApiClient.instance;

  Future<ListResponse<Item>> listItems({int limit = 20, int offset = 0}) async {
    final data =
        await _client.get(
              '/v1/items',
              query: {'limit': '$limit', 'offset': '$offset'},
            )
            as Map<String, dynamic>;
    return ListResponse(
      data: (data['data'] as List)
          .map((e) => Item.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: data['total'] as int,
      limit: data['limit'] as int,
      offset: data['offset'] as int,
    );
  }

  Future<Item> getItem(int id) async {
    final data = await _client.get('/v1/items/$id') as Map<String, dynamic>;
    return Item.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<Item> createItem({
    required String name,
    String description = '',
  }) async {
    final data =
        await _client.post('/v1/items', {
              'name': name,
              'description': description,
            })
            as Map<String, dynamic>;
    return Item.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<Item> updateItem(
    int id, {
    required String name,
    String description = '',
  }) async {
    final data =
        await _client.put('/v1/items/$id', {
              'name': name,
              'description': description,
            })
            as Map<String, dynamic>;
    return Item.fromJson(data['data'] as Map<String, dynamic>);
  }

  Future<void> deleteItem(int id) => _client.delete('/v1/items/$id');
}
