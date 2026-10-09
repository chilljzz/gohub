package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chilljzz/gohub/internal/ws"
)

func TestMetricsHandler(t *testing.T) {
	manager := ws.NewManager()
	m := New(manager)

	m.ObserveHTTPRequest(
		http.MethodGet,
		"/ping",
		http.StatusOK,
		20*time.Millisecond,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	recorder := httptest.NewRecorder()

	m.Handler().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want 200",
			recorder.Code,
		)
	}

	body := recorder.Body.String()

	expected := `gohub_http_requests_total{method="GET",route="/ping",status="200"} 1`

	if !strings.Contains(body, expected) {
		t.Errorf(
			"metrics output missing expected counter",
		)
	}

	if !strings.Contains(
		body,
		"gohub_http_request_duration_seconds_bucket",
	) {
		t.Error("missing HTTP duration histogram")
	}
}
