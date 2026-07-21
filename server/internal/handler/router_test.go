package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"star/server/internal/config"
)

func TestHealthzDoesNotDependOnDatabase(t *testing.T) {
	routes := &Routes{Cfg: &config.Config{CORSOrigins: "*"}}
	engine := routes.Build()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want 200", recorder.Code)
	}
}
