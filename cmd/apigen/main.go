// cmd/apigen/main.go generates typed Dart API methods from docs/swagger.json.
//
// Usage:
//
//	go run ./cmd/apigen
//
// Output: flutter_app/lib/services/api_methods.dart
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Minimal OpenAPI spec structures we care about.
type Spec struct {
	Paths map[string]PathItem `json:"paths"`
}

type PathItem map[string]Operation // "get", "post", etc.

type Operation struct {
	Summary     string         `json:"summary"`
	OperationID string         `json:"operationId"`
	Tags        []string       `json:"tags"`
	Parameters  []Parameter    `json:"parameters"`
	Responses   map[string]any `json:"responses"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"` // "query", "path"
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type Route struct {
	Method      string
	Path        string
	Summary     string
	Tag         string
	PathParams  []Parameter
	QueryParams []Parameter
	HasBody     bool // POST/PUT
}

func main() {
	raw, err := os.ReadFile("docs/swagger.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Run 'make docs' first to generate swagger.json\n")
		os.Exit(1)
	}

	var spec Spec
	if err := json.Unmarshal(raw, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Parse swagger.json: %v\n", err)
		os.Exit(1)
	}

	// Collect routes, sorted by path+method.
	var routes []Route
	for path, methods := range spec.Paths {
		for method, op := range methods {
			r := Route{
				Method:  strings.ToUpper(method),
				Path:    path,
				Summary: op.Summary,
				HasBody: method == "post" || method == "put",
			}
			if len(op.Tags) > 0 {
				r.Tag = op.Tags[0]
			}
			for _, p := range op.Parameters {
				if p.In == "path" {
					r.PathParams = append(r.PathParams, p)
				} else if p.In == "query" {
					r.QueryParams = append(r.QueryParams, p)
				}
			}
			routes = append(routes, r)
		}
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	// Generate Dart code.
	var b strings.Builder
	b.WriteString("// GENERATED — DO NOT EDIT. Run `make flutter-apigen` to regenerate.\n")
	b.WriteString("// Source: docs/swagger.json\n\n")
	b.WriteString("import 'package:flutter_app/services/api_client.dart';\n\n")
	b.WriteString("/// Typed API methods generated from the OpenAPI spec.\n")
	b.WriteString("class Api {\n")
	b.WriteString("  final ApiClient _client;\n")
	b.WriteString("  Api([ApiClient? client]) : _client = client ?? ApiClient.instance;\n\n")

	for _, r := range routes {
		// Skip system routes.
		if r.Tag == "system" || r.Path == "/docs" || strings.HasPrefix(r.Path, "/docs/") {
			continue
		}

		methodName := dartMethodName(r.Method, r.Path)
		b.WriteString(fmt.Sprintf("  /// %s\n", r.Summary))

		// Build method signature.
		var params []string
		for _, p := range r.PathParams {
			params = append(params, fmt.Sprintf("int %s", p.Name))
		}
		if r.HasBody {
			params = append(params, "Map<String, dynamic> body")
		}
		// Query params are optional named params.
		var optionals []string
		for _, p := range r.QueryParams {
			dartType := dartParamType(p.Type)
			optionals = append(optionals, fmt.Sprintf("%s? %s", dartType, p.Name))
		}

		sig := strings.Join(params, ", ")
		if len(optionals) > 0 {
			if sig != "" {
				sig += ", "
			}
			sig += "{" + strings.Join(optionals, ", ") + "}"
		}

		// Return type.
		returnType := "Future<dynamic>"
		if r.Method == "DELETE" {
			returnType = "Future<void>"
		}

		b.WriteString(fmt.Sprintf("  %s %s(%s) async {\n", returnType, methodName, sig))

		// Build query map if needed.
		if len(r.QueryParams) > 0 {
			b.WriteString("    final query = <String, String>{\n")
			for _, p := range r.QueryParams {
				b.WriteString(fmt.Sprintf("      if (%s != null) '%s': %s.toString(),\n", p.Name, p.Name, p.Name))
			}
			b.WriteString("    };\n")
		}

		// Build path with interpolation.
		dartPath := r.Path
		for _, p := range r.PathParams {
			dartPath = strings.Replace(dartPath, "{"+p.Name+"}", "$"+p.Name, 1)
		}

		// Method call.
		switch r.Method {
		case "GET":
			if len(r.QueryParams) > 0 {
				b.WriteString(fmt.Sprintf("    return _client.get('%s', query: query);\n", dartPath))
			} else {
				b.WriteString(fmt.Sprintf("    return _client.get('%s');\n", dartPath))
			}
		case "POST":
			b.WriteString(fmt.Sprintf("    return _client.post('%s', body);\n", dartPath))
		case "PUT":
			b.WriteString(fmt.Sprintf("    return _client.put('%s', body);\n", dartPath))
		case "DELETE":
			b.WriteString(fmt.Sprintf("    return _client.delete('%s');\n", dartPath))
		}

		b.WriteString("  }\n\n")
	}

	b.WriteString("}\n")

	dest := "flutter_app/lib/services/api_methods.dart"
	if err := os.WriteFile(dest, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Write %s: %v\n", dest, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Generated %s (%d routes)\n", dest, len(routes))
}

// dartMethodName converts "GET /v1/items/{id}" → "getItem".
func dartMethodName(method, path string) string {
	// Remove /v1/ prefix and {id} placeholders.
	clean := strings.TrimPrefix(path, "/v1/")
	clean = strings.ReplaceAll(clean, "/{id}", "")
	clean = strings.ReplaceAll(clean, "/", "_")

	parts := strings.Split(clean, "_")
	var result strings.Builder
	result.WriteString(strings.ToLower(method))
	for _, p := range parts {
		if p == "" {
			continue
		}
		result.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}

	// Pluralization: GET /items → listItems, GET /items/{id} → getItem
	name := result.String()
	if method == "GET" && !strings.Contains(path, "{id}") {
		name = strings.Replace(name, "get", "list", 1)
	}
	return name
}

// dartParamType maps OpenAPI types to Dart types.
func dartParamType(t string) string {
	switch t {
	case "integer":
		return "int"
	case "number":
		return "double"
	case "boolean":
		return "bool"
	default:
		return "String"
	}
}
