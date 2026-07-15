package http

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/acmhot100/server/internal/config"
	"gopkg.in/yaml.v3"
)

type openAPIDocument struct {
	OpenAPI string                    `yaml:"openapi"`
	Paths   map[string]map[string]any `yaml:"paths"`
	Schemas map[string]openAPISchema  `yaml:"-"`
}

type openAPISchema struct {
	Properties map[string]any `yaml:"properties"`
}

func TestOpenAPIContractCoversRegisteredRoutes(t *testing.T) {
	document := loadOpenAPIContract(t)
	if document.OpenAPI != "3.1.1" {
		t.Fatalf("openapi = %q, want 3.1.1", document.OpenAPI)
	}

	got := make([]string, 0)
	for path, operations := range document.Paths {
		for method := range operations {
			if isHTTPMethod(method) {
				got = append(got, strings.ToUpper(method)+" /api/v1"+path)
			}
		}
	}
	sort.Strings(got)
	want := []string{
		"GET /api/v1/auth/me",
		"GET /api/v1/health",
		"GET /api/v1/languages",
		"GET /api/v1/problems",
		"GET /api/v1/problems/{slug}",
		"GET /api/v1/problems/{slug}/drafts/{language_key}",
		"GET /api/v1/problems/{slug}/navigation",
		"GET /api/v1/profile/progress-by-stage",
		"GET /api/v1/profile/summary",
		"GET /api/v1/runs/{id}",
		"GET /api/v1/submissions",
		"GET /api/v1/submissions/{id}",
		"GET /api/v1/tags",
		"POST /api/v1/auth/forgot-password",
		"POST /api/v1/auth/login",
		"POST /api/v1/auth/logout",
		"POST /api/v1/auth/logout-all",
		"POST /api/v1/auth/refresh",
		"POST /api/v1/auth/register",
		"POST /api/v1/auth/resend-verification",
		"POST /api/v1/auth/reset-password",
		"POST /api/v1/auth/verify-email",
		"POST /api/v1/problems/{slug}/run",
		"POST /api/v1/problems/{slug}/submissions",
		"PUT /api/v1/problems/{slug}/drafts/{language_key}",
	}
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("contract routes differ from registered routes\ngot:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestOpenAPIContractMatchesRegisteredRoutes(t *testing.T) {
	db, _ := handlerTestDB(t)
	router := NewServer(&config.Config{AppEnv: "test", AppBaseURL: "http://localhost:5173", JudgeMode: "mock"}, db, nil)

	registered := make([]string, 0)
	for _, route := range router.Routes() {
		if strings.HasPrefix(route.Path, "/api/v1/") {
			registered = append(registered, route.Method+" "+normalizeGinPath(route.Path))
		}
	}
	sort.Strings(registered)

	document := loadOpenAPIContract(t)
	contracted := make([]string, 0)
	for path, operations := range document.Paths {
		for method := range operations {
			if isHTTPMethod(method) {
				contracted = append(contracted, strings.ToUpper(method)+" /api/v1"+path)
			}
		}
	}
	sort.Strings(contracted)
	if strings.Join(registered, "\n") != strings.Join(contracted, "\n") {
		t.Fatalf("OpenAPI paths differ from Gin routes\nregistered:\n%s\ncontracted:\n%s", strings.Join(registered, "\n"), strings.Join(contracted, "\n"))
	}
}

func TestOpenAPIContractOmitsInternalResponseFields(t *testing.T) {
	document := loadOpenAPIContract(t)
	for schemaName, schema := range document.Schemas {
		for _, forbidden := range []string{
			"dsn", "enabled", "judge0_language_id", "judge0_language_name",
			"password_hash", "reference_solution", "refresh_token", "test_cases",
		} {
			if _, exists := schema.Properties[forbidden]; exists {
				t.Errorf("schema %s exposes internal field %q", schemaName, forbidden)
			}
		}
	}
}

func loadOpenAPIContract(t *testing.T) openAPIDocument {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	contractPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "packages", "contracts", "openapi.yaml")
	contents, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("read OpenAPI contract: %v", err)
	}

	var raw struct {
		OpenAPI    string                    `yaml:"openapi"`
		Paths      map[string]map[string]any `yaml:"paths"`
		Components struct {
			Schemas map[string]openAPISchema `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(contents, &raw); err != nil {
		t.Fatalf("parse OpenAPI contract: %v", err)
	}
	return openAPIDocument{
		OpenAPI: raw.OpenAPI,
		Paths:   raw.Paths,
		Schemas: raw.Components.Schemas,
	}
}

func normalizeGinPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}
