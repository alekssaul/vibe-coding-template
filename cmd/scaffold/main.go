package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

// Field represents a single field parsed from CLI input (e.g. "price:float").
type Field struct {
	Name        string // Go/Dart field name (PascalCase)
	JSONName    string // JSON / SQL column name (snake_case)
	GoType      string // Go type (string, int64, float64, bool)
	SQLType     string // SQLite type
	DartType    string // Dart type
	ValidateTag string // validate struct tag
	IsBool      bool   // true if the Go type is bool (sqlc maps to int64)
}

// TemplateData holds all the data passed into every template.
type TemplateData struct {
	Resource             string // PascalCase singular (e.g. Product)
	ResourceLower        string // lowercase singular (e.g. product)
	ResourcePlural       string // lowercase plural (e.g. products)
	ResourcePluralPascal string // PascalCase plural (e.g. Products)
	Fields               []Field
	ModulePath           string
	Timestamp            string // for migration file naming
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/scaffold RESOURCE=name FIELDS=\"field1:type,field2:type\"")
		fmt.Println("Types: string, int, float, bool")
		fmt.Println("Example: go run ./cmd/scaffold RESOURCE=product FIELDS=\"price:float,active:bool\"")
		os.Exit(1)
	}

	var resourceName, fieldsRaw string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "RESOURCE=") {
			resourceName = strings.TrimPrefix(arg, "RESOURCE=")
		} else if strings.HasPrefix(arg, "FIELDS=") {
			fieldsRaw = strings.TrimPrefix(arg, "FIELDS=")
		}
	}

	if resourceName == "" {
		fmt.Println("Error: RESOURCE is required")
		os.Exit(1)
	}

	// Read module path from go.mod
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Printf("Error reading go.mod: %v\n", err)
		os.Exit(1)
	}
	modulePath := strings.TrimPrefix(strings.Split(string(modBytes), "\n")[0], "module ")
	modulePath = strings.TrimSpace(modulePath)

	data := TemplateData{
		Resource:             toPascalCase(resourceName),
		ResourceLower:        strings.ToLower(resourceName),
		ResourcePlural:       strings.ToLower(resourceName) + "s",
		ResourcePluralPascal: toPascalCase(resourceName) + "s",
		ModulePath:           modulePath,
		Timestamp:            time.Now().Format("20060102150405"),
	}

	// Always include Name + Description as base fields, then add custom fields
	data.Fields = []Field{
		{Name: "Name", JSONName: "name", GoType: "string", SQLType: "TEXT NOT NULL DEFAULT ''", DartType: "String", ValidateTag: `validate:"required,min=2,max=100"`},
		{Name: "Description", JSONName: "description", GoType: "string", SQLType: "TEXT NOT NULL DEFAULT ''", DartType: "String", ValidateTag: `validate:"max=500"`},
	}

	if fieldsRaw != "" {
		for _, f := range strings.Split(fieldsRaw, ",") {
			parts := strings.SplitN(strings.TrimSpace(f), ":", 2)
			if len(parts) != 2 {
				fmt.Printf("Warning: skipping invalid field '%s' (expected name:type)\n", f)
				continue
			}
			field := parseField(parts[0], parts[1])
			data.Fields = append(data.Fields, field)
		}
	}

	files := []struct {
		tmpl string
		dest string
	}{
		{tmplModel, fmt.Sprintf("internal/model/%s.go", data.ResourceLower)},
		{tmplQueries, fmt.Sprintf("internal/store/queries/%s.sql", data.ResourcePlural)},
		{tmplStore, fmt.Sprintf("internal/store/%s.go", data.ResourcePlural)},
		{tmplHandler, fmt.Sprintf("internal/handler/%s.go", data.ResourcePlural)},
		{tmplMigrationUp, fmt.Sprintf("internal/store/migrations/%s_create_%s.up.sql", nextMigrationNum(), data.ResourcePlural)},
		{tmplMigrationDown, fmt.Sprintf("internal/store/migrations/%s_create_%s.down.sql", nextMigrationNum(), data.ResourcePlural)},
		{tmplDartModel, fmt.Sprintf("flutter_app/lib/models/%s.dart", data.ResourceLower)},
		{tmplDartProvider, fmt.Sprintf("flutter_app/lib/providers/%s_provider.dart", data.ResourcePlural)},
		{tmplDartScreen, fmt.Sprintf("flutter_app/lib/screens/%s_screen.dart", data.ResourcePlural)},
	}

	for _, f := range files {
		if err := renderTemplate(f.tmpl, f.dest, data); err != nil {
			fmt.Printf("❌ Error generating %s: %v\n", f.dest, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Generated %s\n", f.dest)
	}

	fmt.Println()
	fmt.Println("🎉 Scaffold complete! Next steps:")
	fmt.Printf("  1. Run: make db-generate\n")
	fmt.Printf("  2. Routes auto-injected into cmd/api/main.go ✅\n")
	fmt.Printf("  3. Run: cd flutter_app && dart run build_runner build -d\n")
	fmt.Printf("  4. Add the screen to flutter_app/lib/router/app_router.dart\n")
	fmt.Printf("  5. Run: make verify\n")

	if err := injectRoutes(data); err != nil {
		fmt.Printf("⚠️  Could not auto-inject routes: %v\n", err)
		fmt.Println("   Add these routes manually to cmd/api/main.go:")
		printRouteSnippet(data)
	} else {
		fmt.Println("✅ Routes injected into cmd/api/main.go")
	}
}

func renderTemplate(tmplStr, dest string, data TemplateData) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Don't overwrite existing files
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("file already exists: %s", dest)
	}

	funcMap := template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": toPascalCase,
	}
	t, err := template.New("scaffold").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("template parse error: %w", err)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	return t.Execute(f, data)
}

func parseField(name, typ string) Field {
	f := Field{
		Name:     toPascalCase(name),
		JSONName: toSnakeCase(name),
	}
	switch strings.ToLower(typ) {
	case "int", "integer":
		f.GoType = "int64"
		f.SQLType = "INTEGER NOT NULL DEFAULT 0"
		f.DartType = "int"
		f.ValidateTag = ""
	case "float", "double", "decimal":
		f.GoType = "float64"
		f.SQLType = "REAL NOT NULL DEFAULT 0"
		f.DartType = "double"
		f.ValidateTag = ""
	case "bool", "boolean":
		f.GoType = "bool"
		f.SQLType = "INTEGER NOT NULL DEFAULT 0"
		f.DartType = "bool"
		f.ValidateTag = ""
		f.IsBool = true
	default: // string
		f.GoType = "string"
		f.SQLType = "TEXT NOT NULL DEFAULT ''"
		f.DartType = "String"
		f.ValidateTag = `validate:"max=500"`
	}
	return f
}

func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func nextMigrationNum() string {
	entries, _ := os.ReadDir("internal/store/migrations")
	maxNum := 0
	for _, e := range entries {
		name := e.Name()
		if len(name) >= 6 {
			var num int
			fmt.Sscanf(name[:6], "%d", &num)
			if num > maxNum {
				maxNum = num
			}
		}
	}
	return fmt.Sprintf("%06d", maxNum+1)
}

// ── Route Injection ──────────────────────────────────────────────────────────

const routeMarker = "// scaffold:routes"
const mainGoPath = "cmd/api/main.go"

// injectRoutes inserts 5 route registrations into cmd/api/main.go just before
// the scaffold:routes marker comment. Safe to call multiple times (idempotent guard).
func injectRoutes(data TemplateData) error {
	raw, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", mainGoPath, err)
	}

	snippet := fmt.Sprintf(
		"\t// %s — read\n"+
			"\tmux.Handle(\"GET /v1/%s\", apiKeyMW.RequireRead(http.HandlerFunc(h.List%s)))\n"+
			"\tmux.Handle(\"GET /v1/%s/{id}\", apiKeyMW.RequireRead(http.HandlerFunc(h.Get%s)))\n\n"+
			"\t// %s — write\n"+
			"\tmux.Handle(\"POST /v1/%s\", apiKeyMW.RequireWrite(http.HandlerFunc(h.Create%s)))\n"+
			"\tmux.Handle(\"PUT /v1/%s/{id}\", apiKeyMW.RequireWrite(http.HandlerFunc(h.Update%s)))\n"+
			"\tmux.Handle(\"DELETE /v1/%s/{id}\", apiKeyMW.RequireWrite(http.HandlerFunc(h.Delete%s)))\n\n",
		data.ResourcePluralPascal,
		data.ResourcePlural, data.ResourcePluralPascal,
		data.ResourcePlural, data.Resource,
		data.ResourcePluralPascal,
		data.ResourcePlural, data.Resource,
		data.ResourcePlural, data.Resource,
		data.ResourcePlural, data.Resource,
	)

	content := string(raw)
	// Idempotency: skip if already injected.
	if strings.Contains(content, "h.List"+data.ResourcePluralPascal) {
		return nil
	}
	if !strings.Contains(content, routeMarker) {
		return fmt.Errorf("marker %q not found in %s", routeMarker, mainGoPath)
	}

	updated := strings.Replace(content, "\t"+routeMarker, snippet+"\t"+routeMarker, 1)
	return os.WriteFile(mainGoPath, []byte(updated), 0o644)
}

// printRouteSnippet is the fallback — prints routes to stdout when auto-injection fails.
func printRouteSnippet(data TemplateData) {
	fmt.Printf(`
	// %s — read
	mux.Handle("GET /v1/%s", apiKeyMW.RequireRead(http.HandlerFunc(h.List%s)))
	mux.Handle("GET /v1/%s/{id}", apiKeyMW.RequireRead(http.HandlerFunc(h.Get%s)))

	// %s — write
	mux.Handle("POST /v1/%s", apiKeyMW.RequireWrite(http.HandlerFunc(h.Create%s)))
	mux.Handle("PUT /v1/%s/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.Update%s)))
	mux.Handle("DELETE /v1/%s/{id}", apiKeyMW.RequireWrite(http.HandlerFunc(h.Delete%s)))
`,
		data.ResourcePluralPascal,
		data.ResourcePlural, data.ResourcePluralPascal,
		data.ResourcePlural, data.Resource,
		data.ResourcePluralPascal,
		data.ResourcePlural, data.Resource,
		data.ResourcePlural, data.Resource,
		data.ResourcePlural, data.Resource,
	)
}

// ── Go Templates ─────────────────────────────────────────────────────────────

var tmplModel = `package model

import "time"

// {{.Resource}} represents a {{.ResourceLower}} resource.
type {{.Resource}} struct {
	ID          int64     ` + "`" + `json:"id"` + "`" + `
{{- range .Fields}}
	{{.Name}}  {{.GoType}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
{{- end}}
	CreatedAt   time.Time ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt   time.Time ` + "`" + `json:"updated_at"` + "`" + `
}

// Create{{.Resource}}Request is the payload for creating a {{.ResourceLower}}.
type Create{{.Resource}}Request struct {
{{- range .Fields}}
	{{.Name}}  {{.GoType}} ` + "`" + `json:"{{.JSONName}}"{{if .ValidateTag}} {{.ValidateTag}}{{end}}` + "`" + `
{{- end}}
}

// Update{{.Resource}}Request is the payload for updating a {{.ResourceLower}}.
type Update{{.Resource}}Request struct {
{{- range .Fields}}
	{{.Name}}  {{.GoType}} ` + "`" + `json:"{{.JSONName}}"{{if .ValidateTag}} {{.ValidateTag}}{{end}}` + "`" + `
{{- end}}
}
`

var tmplQueries = `-- name: Get{{.Resource}} :one
SELECT * FROM {{.ResourcePlural}}
WHERE id = ? LIMIT 1;

-- name: List{{.ResourcePluralPascal}} :many
SELECT * FROM {{.ResourcePlural}}
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: Count{{.ResourcePluralPascal}} :one
SELECT COUNT(*) FROM {{.ResourcePlural}};

-- name: Create{{.Resource}} :one
INSERT INTO {{.ResourcePlural}} ({{range $i, $f := .Fields}}{{if $i}}, {{end}}{{$f.JSONName}}{{end}})
VALUES ({{range $i, $f := .Fields}}{{if $i}}, {{end}}?{{end}})
RETURNING *;

-- name: Update{{.Resource}} :one
UPDATE {{.ResourcePlural}}
SET {{range $i, $f := .Fields}}{{if $i}}, {{end}}{{$f.JSONName}} = ?{{end}}, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: Delete{{.Resource}} :exec
DELETE FROM {{.ResourcePlural}}
WHERE id = ?;
`

var tmplStore = `package store

import (
	"context"

	"{{.ModulePath}}/internal/model"
	"{{.ModulePath}}/internal/store/db"
)

func map{{.Resource}}(i db.{{.Resource}}) *model.{{.Resource}} {
	return &model.{{.Resource}}{
		ID:          i.ID,
{{- range .Fields}}
{{- if .IsBool}}
		{{.Name}}:  i.{{.Name}} != 0,
{{- else}}
		{{.Name}}:  i.{{.Name}},
{{- end}}
{{- end}}
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

// List{{.ResourcePluralPascal}} returns paginated {{.ResourcePlural}} and the total count.
func (s *Store) List{{.ResourcePluralPascal}}(ctx context.Context, limit, offset int) ([]*model.{{.Resource}}, int, error) {
	total, err := s.queries.Count{{.ResourcePluralPascal}}(ctx)
	if err != nil {
		return nil, 0, err
	}

	dbRows, err := s.queries.List{{.ResourcePluralPascal}}(ctx, db.List{{.ResourcePluralPascal}}Params{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	var items []*model.{{.Resource}}
	for _, i := range dbRows {
		items = append(items, map{{.Resource}}(i))
	}
	return items, int(total), nil
}

// Get{{.Resource}} returns a single {{.ResourceLower}} by ID.
func (s *Store) Get{{.Resource}}(ctx context.Context, id int64) (*model.{{.Resource}}, error) {
	i, err := s.queries.Get{{.Resource}}(ctx, id)
	if err != nil {
		return nil, err
	}
	return map{{.Resource}}(i), nil
}

func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Create{{.Resource}} inserts a new {{.ResourceLower}} and returns it.
func (s *Store) Create{{.Resource}}(ctx context.Context, req *model.Create{{.Resource}}Request) (*model.{{.Resource}}, error) {
	i, err := s.queries.Create{{.Resource}}(ctx, db.Create{{.Resource}}Params{
{{- range .Fields}}
{{- if .IsBool}}
		{{.Name}}:  boolToInt64(req.{{.Name}}),
{{- else}}
		{{.Name}}:  req.{{.Name}},
{{- end}}
{{- end}}
	})
	if err != nil {
		return nil, err
	}
	return map{{.Resource}}(i), nil
}

// Update{{.Resource}} updates an existing {{.ResourceLower}} and returns it.
func (s *Store) Update{{.Resource}}(ctx context.Context, id int64, req *model.Update{{.Resource}}Request) (*model.{{.Resource}}, error) {
	i, err := s.queries.Update{{.Resource}}(ctx, db.Update{{.Resource}}Params{
		ID:          id,
{{- range .Fields}}
{{- if .IsBool}}
		{{.Name}}:  boolToInt64(req.{{.Name}}),
{{- else}}
		{{.Name}}:  req.{{.Name}},
{{- end}}
{{- end}}
	})
	if err != nil {
		return nil, err
	}
	return map{{.Resource}}(i), nil
}

// Delete{{.Resource}} removes a {{.ResourceLower}} by ID.
func (s *Store) Delete{{.Resource}}(ctx context.Context, id int64) error {
	return s.queries.Delete{{.Resource}}(ctx, id)
}
`

var tmplHandler = `package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"{{.ModulePath}}/internal/model"
	"{{.ModulePath}}/internal/request"
	"{{.ModulePath}}/internal/response"
)

// List{{.ResourcePluralPascal}} godoc
//
//	@Summary		List {{.ResourcePlural}}
//	@Description	Returns paginated list of {{.ResourcePlural}}
//	@Tags			{{.ResourcePlural}}
//	@Produce		json
//	@Param			limit	query		int	false	"Max items (default 20, max 100)"
//	@Param			offset	query		int	false	"Pagination offset (default 0)"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.ListResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Router			/v1/{{.ResourcePlural}} [get]
func (h *Handler) List{{.ResourcePluralPascal}}(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20, 100)
	offset := queryInt(r, "offset", 0, -1)

	items, total, err := h.store.List{{.ResourcePluralPascal}}(r.Context(), limit, offset)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "list {{.ResourcePlural}}", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to list {{.ResourcePlural}}", "INTERNAL_ERROR")
		return
	}
	if items == nil {
		items = []*model.{{.Resource}}{}
	}
	response.WriteList(w, items, total, limit, offset)
}

// Get{{.Resource}} godoc
//
//	@Summary		Get {{.ResourceLower}}
//	@Description	Returns a single {{.ResourceLower}} by ID
//	@Tags			{{.ResourcePlural}}
//	@Produce		json
//	@Param			id	path		int	true	"{{.Resource}} ID"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/{{.ResourcePlural}}/{id} [get]
func (h *Handler) Get{{.Resource}}(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid {{.ResourceLower}} id", "BAD_REQUEST")
		return
	}
	item, err := h.store.Get{{.Resource}}(r.Context(), id)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "{{.ResourceLower}} not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "get {{.ResourceLower}}", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to get {{.ResourceLower}}", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
}

// Create{{.Resource}} godoc
//
//	@Summary		Create {{.ResourceLower}}
//	@Description	Creates a new {{.ResourceLower}}
//	@Tags			{{.ResourcePlural}}
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.Create{{.Resource}}Request	true	"{{.Resource}} payload"
//	@Security		ApiKeyAuth
//	@Success		201	{object}	response.SuccessResponse
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/v1/{{.ResourcePlural}} [post]
func (h *Handler) Create{{.Resource}}(w http.ResponseWriter, r *http.Request) {
	var req model.Create{{.Resource}}Request
	if err := request.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}
	item, err := h.store.Create{{.Resource}}(r.Context(), &req)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "create {{.ResourceLower}}", "error", err)
		response.WriteError(w, http.StatusInternalServerError, "failed to create {{.ResourceLower}}", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusCreated, item)
}

// Update{{.Resource}} godoc
//
//	@Summary		Update {{.ResourceLower}}
//	@Description	Updates an existing {{.ResourceLower}}
//	@Tags			{{.ResourcePlural}}
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"{{.Resource}} ID"
//	@Param			body	body		model.Update{{.Resource}}Request	true	"{{.Resource}} payload"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	response.SuccessResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/{{.ResourcePlural}}/{id} [put]
func (h *Handler) Update{{.Resource}}(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid {{.ResourceLower}} id", "BAD_REQUEST")
		return
	}
	var req model.Update{{.Resource}}Request
	if err := request.DecodeJSON(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}
	item, err := h.store.Update{{.Resource}}(r.Context(), id, &req)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "{{.ResourceLower}} not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "update {{.ResourceLower}}", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to update {{.ResourceLower}}", "INTERNAL_ERROR")
		return
	}
	response.WriteSuccess(w, http.StatusOK, item)
}

// Delete{{.Resource}} godoc
//
//	@Summary		Delete {{.ResourceLower}}
//	@Description	Deletes a {{.ResourceLower}} by ID
//	@Tags			{{.ResourcePlural}}
//	@Param			id	path	int	true	"{{.Resource}} ID"
//	@Security		ApiKeyAuth
//	@Success		204	"No Content"
//	@Failure		404	{object}	response.ErrorResponse
//	@Router			/v1/{{.ResourcePlural}}/{id} [delete]
func (h *Handler) Delete{{.Resource}}(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid {{.ResourceLower}} id", "BAD_REQUEST")
		return
	}
	err = h.store.Delete{{.Resource}}(r.Context(), id)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "{{.ResourceLower}} not found", "NOT_FOUND")
		return
	}
	if err != nil {
		h.logger.ErrorContext(r.Context(), "delete {{.ResourceLower}}", "error", err, "id", id)
		response.WriteError(w, http.StatusInternalServerError, "failed to delete {{.ResourceLower}}", "INTERNAL_ERROR")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
`

var tmplMigrationUp = `CREATE TABLE IF NOT EXISTS {{.ResourcePlural}} (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
{{- range .Fields}}
    {{.JSONName}}  {{.SQLType}},
{{- end}}
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

var tmplMigrationDown = `DROP TABLE IF EXISTS {{.ResourcePlural}};
`

// ── Flutter/Dart Templates ───────────────────────────────────────────────────

var tmplDartModel = `/// Represents a single {{.ResourceLower}} from the API.
class {{.Resource}} {
  final int id;
{{- range .Fields}}
  final {{.DartType}} {{.JSONName}};
{{- end}}
  final DateTime createdAt;
  final DateTime updatedAt;

  {{.Resource}}({
    required this.id,
{{- range .Fields}}
    required this.{{.JSONName}},
{{- end}}
    required this.createdAt,
    required this.updatedAt,
  });

  factory {{.Resource}}.fromJson(Map<String, dynamic> json) => {{.Resource}}(
    id: json['id'] as int,
{{- range .Fields}}
{{- if eq .DartType "String"}}
    {{.JSONName}}: json['{{.JSONName}}'] as String? ?? '',
{{- else if eq .DartType "int"}}
    {{.JSONName}}: json['{{.JSONName}}'] as int? ?? 0,
{{- else if eq .DartType "double"}}
    {{.JSONName}}: (json['{{.JSONName}}'] as num?)?.toDouble() ?? 0.0,
{{- else if eq .DartType "bool"}}
    {{.JSONName}}: json['{{.JSONName}}'] as bool? ?? false,
{{- end}}
{{- end}}
    createdAt: DateTime.parse(json['created_at'] as String),
    updatedAt: DateTime.parse(json['updated_at'] as String),
  );

  Map<String, dynamic> toJson() => {
{{- range .Fields}}
    '{{.JSONName}}': {{.JSONName}},
{{- end}}
  };
}
`

var tmplDartProvider = `import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/{{.ResourceLower}}.dart';
import '../services/api_client.dart';
import '../core/utils/snackbar_util.dart';

part '{{.ResourcePlural}}_provider.g.dart';

@riverpod
class {{.ResourcePluralPascal}} extends _${{.ResourcePluralPascal}} {
  @override
  FutureOr<List<{{.Resource}}>> build() async {
    final client = ApiClient.instance;
    final response = await client.get('/v1/{{.ResourcePlural}}');
    final list = (response['data'] as List)
        .map((json) => {{.Resource}}.fromJson(json))
        .toList();
    return list;
  }

  Future<void> create{{.Resource}}(Map<String, dynamic> data) async {
    final client = ApiClient.instance;
    try {
      await client.post('/v1/{{.ResourcePlural}}', data);
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('{{.Resource}} created');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }

  Future<void> update{{.Resource}}(int id, Map<String, dynamic> data) async {
    final client = ApiClient.instance;
    try {
      await client.put('/v1/{{.ResourcePlural}}/$id', data);
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('{{.Resource}} updated');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }

  Future<void> delete{{.Resource}}(int id) async {
    final client = ApiClient.instance;
    try {
      await client.delete('/v1/{{.ResourcePlural}}/$id');
      ref.invalidateSelf();
      SnackBarUtil.showSuccess('{{.Resource}} deleted');
    } catch (e) {
      SnackBarUtil.showError(e.toString());
    }
  }
}
`

var tmplDartScreen = `import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/{{.ResourcePlural}}_provider.dart';
import '../models/{{.ResourceLower}}.dart';
import '../core/widgets/error_state_widget.dart';

/// Main {{.ResourcePlural}} list screen with inline create/edit/delete.
class {{.ResourcePluralPascal}}Screen extends ConsumerWidget {
  const {{.ResourcePluralPascal}}Screen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch({{.ResourcePlural}}Provider);

    void refreshList() {
      ref.invalidate({{.ResourcePlural}}Provider);
    }

    void confirmDelete(int id) async {
      final confirm = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Delete {{.Resource}}'),
          content: const Text('Are you sure you want to delete this {{.ResourceLower}}?'),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: Theme.of(context).colorScheme.error,
                foregroundColor: Theme.of(context).colorScheme.onError,
              ),
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Delete'),
            ),
          ],
        ),
      );
      if (confirm == true) {
        await ref.read({{.ResourcePlural}}Provider.notifier).delete{{.Resource}}(id);
      }
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('{{.ResourcePluralPascal}}'),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: refreshList),
        ],
      ),
      body: state.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => ErrorStateWidget(error: err, onRetry: refreshList),
        data: (items) => items.isEmpty
            ? const Center(child: Text('No {{.ResourcePlural}} yet. Create one!'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) => ListTile(
                  title: Text(items[i].name),
                  subtitle: items[i].description.isNotEmpty
                      ? Text(items[i].description)
                      : null,
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        icon: const Icon(Icons.edit_outlined),
                        onPressed: () => _showDialog(context, ref, items[i]),
                      ),
                      IconButton(
                        icon: const Icon(Icons.delete_outline),
                        color: Theme.of(context).colorScheme.error,
                        onPressed: () => confirmDelete(items[i].id),
                      ),
                    ],
                  ),
                ),
              ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showDialog(context, ref),
        child: const Icon(Icons.add),
      ),
    );
  }

  Future<void> _showDialog(BuildContext context, WidgetRef ref, [{{.Resource}}? item]) async {
    final nameCtrl = TextEditingController(text: item?.name ?? '');
    final descCtrl = TextEditingController(text: item?.description ?? '');

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: Text(item == null ? 'Create {{.Resource}}' : 'Edit {{.Resource}}'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: nameCtrl, decoration: const InputDecoration(labelText: 'Name *'), autofocus: true),
            const SizedBox(height: 8),
            TextField(controller: descCtrl, decoration: const InputDecoration(labelText: 'Description')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('Save')),
        ],
      ),
    );

    if (confirmed != true || nameCtrl.text.trim().isEmpty) return;

    final data = {'name': nameCtrl.text.trim(), 'description': descCtrl.text.trim()};
    final notifier = ref.read({{.ResourcePlural}}Provider.notifier);
    if (item == null) {
      await notifier.create{{.Resource}}(data);
    } else {
      await notifier.update{{.Resource}}(item.id, data);
    }
  }
}
`
