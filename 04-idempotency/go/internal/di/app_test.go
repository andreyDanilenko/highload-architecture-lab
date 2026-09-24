package di_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"idempotency/internal/di"
)

func TestInvalidConfigurationFailsBeforeStarting(t *testing.T) {
	for _, change := range []func(*di.Config){
		func(c *di.Config) { c.Provider = "typo" },
		func(c *di.Config) { c.LeaseTTL = 0 },
		func(c *di.Config) { c.ResultTTL = -time.Second },
		func(c *di.Config) { c.MaxEntries = 0 },
		func(c *di.Config) { c.MaxBodyBytes = 0 },
		func(c *di.Config) { c.PaymentDelay = -time.Second },
	} {
		c := di.DefaultConfig()
		change(&c)
		if app, err := di.NewApp(c); err == nil {
			_ = app.Close()
			t.Fatalf("invalid config accepted: %+v", c)
		}
	}
}

func TestDefaultAppReadinessAndDiagnostics(t *testing.T) {
	app, err := di.NewApp(di.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	for _, path := range []string{"/healthz", "/metrics", "/api/v1/effects"} {
		w := httptest.NewRecorder()
		app.Server.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Errorf("path=%s status=%d body=%s", path, w.Code, w.Body.String())
		}
	}
}
