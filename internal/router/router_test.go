package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chilljzz/gohub/internal/metrics"
	"github.com/chilljzz/gohub/internal/middleware"
	"github.com/chilljzz/gohub/internal/ws"
	"github.com/gin-gonic/gin"
)

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	manager := ws.NewManager()
	appMetrics := metrics.New(manager)

	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.PrometheusMetrics(appMetrics))

	registerSystemRoutes(r, appMetrics)

	r.GET(
		"/api/private",
		middleware.AuthMiddleware(),
		func(ctx *gin.Context) {
			ctx.Status(http.StatusNoContent)
		},
	)

	return r
}

func TestPingRoute(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(
		http.MethodGet,
		"/ping",
		nil,
	)

	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want 200",
			recorder.Code,
		)
	}

	var result struct {
		Code int `json:"code"`
		Data struct {
			Msg string `json:"msg"`
		} `json:"data"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&result,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if result.Code != 0 || result.Data.Msg != "pong" {
		t.Fatalf("unexpected response: %+v", result)
	}

	if recorder.Header().Get("X-Request-ID") == "" {
		t.Error("missing request ID")
	}

	if recorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("missing security header")
	}

}

func TestMetricsRecordsPing(t *testing.T) {
	r := newTestRouter()

	ping := httptest.NewRecorder()

	r.ServeHTTP(
		ping,
		httptest.NewRequest(
			http.MethodGet,
			"/ping",
			nil,
		),
	)

	recorder := httptest.NewRecorder()

	r.ServeHTTP(
		recorder,
		httptest.NewRequest(
			http.MethodGet,
			"/metrics",
			nil,
		),
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"metrics status = %d",
			recorder.Code,
		)
	}

	body := recorder.Body.String()

	expected := `gohub_http_requests_total{method="GET",route="/ping",status="200"} 1`

	if !strings.Contains(body, expected) {
		t.Errorf("missing expected ping metric")
	}

	if strings.Contains(body, `route="/metrics"`) {
		t.Error("/metrics should not count itself")
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	r := newTestRouter()

	recorder := httptest.NewRecorder()

	r.ServeHTTP(
		recorder,
		httptest.NewRequest(
			http.MethodGet,
			"/api/private",
			nil,
		),
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"status = %d, want 401",
			recorder.Code,
		)
	}
}
