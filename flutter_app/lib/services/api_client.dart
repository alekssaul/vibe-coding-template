import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

/// API response wrapper for paginated lists.
class ListResponse<T> {
  final List<T> data;
  final int total;
  final int limit;
  final int offset;

  ListResponse({
    required this.data,
    required this.total,
    required this.limit,
    required this.offset,
  });
}

/// Typed exception surfacing API error responses.
class ApiException implements Exception {
  final int statusCode;
  final String message;
  final String code;

  ApiException({
    required this.statusCode,
    required this.message,
    required this.code,
  });

  @override
  String toString() => 'ApiException($statusCode): $message';
}

/// Central HTTP client for the Go API.
class ApiClient {
  ApiClient._();
  static final ApiClient instance = ApiClient._();

  String? _overrideKey;

  String get _baseUrl => dotenv.env['API_BASE_URL'] ?? 'http://localhost:8080';
  String get _apiKey => _overrideKey ?? dotenv.env['API_KEY'] ?? '';

  /// Override the API key (used by setup screen).
  void setApiKey(String key) => _overrideKey = key.isEmpty ? null : key;

  Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    'X-API-Key': _apiKey,
  };

  Uri _uri(String path, [Map<String, String>? query]) =>
      Uri.parse('$_baseUrl$path').replace(queryParameters: query);

  Future<dynamic> _checkResponse(http.Response res) async {
    if (res.statusCode >= 200 && res.statusCode < 300) {
      if (res.body.isEmpty) return null;
      return json.decode(res.body);
    }
    final body = json.decode(res.body) as Map<String, dynamic>;
    throw ApiException(
      statusCode: res.statusCode,
      message: body['error'] as String? ?? 'Unknown error',
      code: body['code'] as String? ?? 'UNKNOWN',
    );
  }

  Future<dynamic> get(String path, {Map<String, String>? query}) async {
    final res = await http.get(_uri(path, query), headers: _headers);
    return _checkResponse(res);
  }

  Future<dynamic> post(String path, Map<String, dynamic> body) async {
    final res = await http.post(
      _uri(path),
      headers: _headers,
      body: json.encode(body),
    );
    return _checkResponse(res);
  }

  Future<dynamic> put(String path, Map<String, dynamic> body) async {
    final res = await http.put(
      _uri(path),
      headers: _headers,
      body: json.encode(body),
    );
    return _checkResponse(res);
  }

  Future<void> delete(String path) async {
    final res = await http.delete(_uri(path), headers: _headers);
    await _checkResponse(res);
  }
}
