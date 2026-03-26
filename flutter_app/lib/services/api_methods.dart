// GENERATED — DO NOT EDIT. Run `make flutter-apigen` to regenerate.
// Source: docs/swagger.json

import 'package:flutter_app/services/api_client.dart';

/// Typed API methods generated from the OpenAPI spec.
class Api {
  final ApiClient _client;
  Api([ApiClient? client]) : _client = client ?? ApiClient.instance;

  /// List items
  Future<dynamic> listItems({int? limit, int? offset, String? search}) async {
    final query = <String, String>{
      if (limit != null) 'limit': limit.toString(),
      if (offset != null) 'offset': offset.toString(),
      if (search != null) 'search': search.toString(),
    };
    return _client.get('/v1/items', query: query);
  }

  /// Create item
  Future<dynamic> postItems(Map<String, dynamic> body) async {
    return _client.post('/v1/items', body);
  }

  /// Delete item
  Future<void> deleteItems(int id) async {
    return _client.delete('/v1/items/$id');
  }

  /// Get item
  Future<dynamic> getItems(int id) async {
    return _client.get('/v1/items/$id');
  }

  /// Update item
  Future<dynamic> putItems(int id, Map<String, dynamic> body) async {
    return _client.put('/v1/items/$id', body);
  }

  /// List API keys
  Future<dynamic> listKeys() async {
    return _client.get('/v1/keys');
  }

  /// Create API key
  Future<dynamic> postKeys(Map<String, dynamic> body) async {
    return _client.post('/v1/keys', body);
  }

  /// Delete API key
  Future<void> deleteKeys(int id) async {
    return _client.delete('/v1/keys/$id');
  }

}
