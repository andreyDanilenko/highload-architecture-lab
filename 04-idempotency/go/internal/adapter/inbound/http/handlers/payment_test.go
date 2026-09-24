package handlers_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"idempotency/internal/adapter/inbound/http/handlers"
	"idempotency/internal/usecase"
	"idempotency/pkg/idempotency"
)

func TestPaymentValidatesBeforeRecordingEffect(t *testing.T) {
	service := usecase.NewPaymentService(0, 10)
	handler := handlers.NewPaymentHandler(service)
	for _, body := range []string{
		`{"from":"a","to":"b","amount":1.5}`,
		`{"from":"a","to":"b","amount":0}`,
		`{"from":"a","to":"a","amount":1}`,
		`{"from":" ","to":"b","amount":1}`,
		`{"from":"a","to":"b","amount":1,"unknown":true}`,
		`{"from":"a","to":"b","amount":1} {}`,
		`null`,
		`{"from":"a","to":"b","amount":9223372036854775808}`,
	} {
		w := httptest.NewRecorder()
		handler.CreatePayment(w, httptest.NewRequest("POST", "/payments", strings.NewReader(body)))
		if w.Code != 400 {
			t.Errorf("body=%s status=%d", body, w.Code)
		}
	}
	if count := len(service.Effects()); count != 0 {
		t.Fatalf("invalid requests produced %d effects", count)
	}
}

func TestPaymentCachesRejectionAndCountsEffectsIndependently(t *testing.T) {
	service := usecase.NewPaymentService(0, 1)
	payment := handlers.NewPaymentHandler(service)
	middleware, err := idempotency.NewHTTPMiddleware(idempotency.NewMemoryProvider(10), idempotency.HTTPOptions{LeaseTTL: time.Minute, ResultTTL: time.Hour, CompleteTimeout: time.Second, MaxBodyBytes: 1024, MaxResponseBytes: 1024, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	if err != nil {
		t.Fatal(err)
	}
	h := middleware.Wrap(http.HandlerFunc(payment.CreatePayment))
	send := func(key, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/payments", strings.NewReader(body))
		r.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	valid := `{"from":"a","to":"b","amount":100}`
	first := send("first", valid)
	replay := send("first", valid)
	if first.Code != 200 || replay.Body.String() != first.Body.String() || len(service.Effects()) != 1 {
		t.Fatalf("replay=%d effects=%d", replay.Code, len(service.Effects()))
	}
	full := send("next", valid)
	fullReplay := send("next", valid)
	if full.Code != 503 || full.Body.String() != fullReplay.Body.String() || len(service.Effects()) != 1 {
		t.Fatalf("capacity result=%d effects=%d", full.Code, len(service.Effects()))
	}
	invalid := send("invalid", `{"from":"a","to":"b","amount":-1}`)
	invalidReplay := send("invalid", `{"from":"a","to":"b","amount":-1}`)
	if invalid.Code != 400 || invalidReplay.Body.String() != invalid.Body.String() {
		t.Fatal("handler rejection was not replayed")
	}
	corrected := send("invalid", valid)
	if corrected.Code != 409 || !strings.Contains(corrected.Body.String(), "key_mismatch") {
		t.Fatalf("corrected payload reused key: %d %s", corrected.Code, corrected.Body.String())
	}
	ledger := httptest.NewRecorder()
	payment.ListEffects(ledger, httptest.NewRequest("GET", "/effects", nil))
	if !strings.Contains(ledger.Body.String(), `"count":1`) || !strings.Contains(ledger.Body.String(), `"transaction_id":"tx_1"`) {
		t.Fatalf("ledger=%s", ledger.Body.String())
	}
}
