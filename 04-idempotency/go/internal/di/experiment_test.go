package di

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Эти опыты считают бизнес-эффекты отдельно от HTTP-ответов.
// В Redis запускается реальный протокол; без LAB_REDIS_ADDR пропускаются только Redis-варианты.
func TestExperimentRepeat(t *testing.T) {
	for _, provider := range []string{"noop", "memory", "redis", "advanced"} {
		t.Run(provider, func(t *testing.T) {
			config := experimentConfig(t, provider)
			app := experimentApp(t, config)
			first := experimentPayment(app)
			second := experimentPayment(app)
			count := experimentEffects(t, app)
			want := 1
			if provider == "noop" {
				want = 2
			}
			if first.Code != http.StatusOK || second.Code != http.StatusOK || count != want {
				t.Fatalf("statuses=%d,%d effects=%d want=%d", first.Code, second.Code, count, want)
			}
			if provider != "noop" && (first.Body.String() != second.Body.String() || first.Header().Get("Content-Type") != second.Header().Get("Content-Type")) {
				t.Fatal("replay changed the response")
			}
			t.Logf("statuses=%d,%d effects=%d", first.Code, second.Code, count)
		})
	}
}

func TestExperimentTwoInstances(t *testing.T) {
	for _, provider := range []string{"noop", "memory", "redis", "advanced"} {
		t.Run(provider, func(t *testing.T) {
			config := experimentConfig(t, provider)
			one, two := experimentApp(t, config), experimentApp(t, config)
			first, second := experimentPayment(one), experimentPayment(two)
			count := experimentEffects(t, one) + experimentEffects(t, two)
			want := 1
			if provider == "noop" || provider == "memory" {
				want = 2
			}
			if first.Code != 200 || second.Code != 200 || count != want {
				t.Fatalf("statuses=%d,%d total_effects=%d want=%d", first.Code, second.Code, count, want)
			}
			if want == 1 && first.Body.String() != second.Body.String() {
				t.Fatal("second instance did not replay")
			}
			t.Logf("statuses=%d,%d total_effects=%d", first.Code, second.Code, count)
		})
	}
}

func TestExperimentLeaseExpiresBeforeCompletion(t *testing.T) {
	for _, provider := range []string{"noop", "memory", "redis", "advanced"} {
		t.Run(provider, func(t *testing.T) {
			config := experimentConfig(t, provider)
			config.LeaseTTL = 10 * time.Millisecond
			config.PaymentDelay = 100 * time.Millisecond
			app := experimentApp(t, config)
			first, second := experimentPayment(app), experimentPayment(app)
			count := experimentEffects(t, app)
			wantFirst, wantSecond, wantCount := 503, 409, 1
			if provider == "noop" {
				wantFirst, wantSecond, wantCount = 200, 200, 2
			}
			if provider == "redis" {
				wantSecond, wantCount = 503, 2
			}
			if first.Code != wantFirst || second.Code != wantSecond || count != wantCount {
				t.Fatalf("statuses=%d,%d effects=%d; want %d,%d/%d", first.Code, second.Code, count, wantFirst, wantSecond, wantCount)
			}
			t.Logf("statuses=%d,%d effects=%d", first.Code, second.Code, count)
		})
	}
}

func experimentConfig(t *testing.T, provider string) Config {
	t.Helper()
	config := DefaultConfig()
	config.Provider = provider
	config.LeaseTTL = 5 * time.Second
	config.ResultTTL = time.Minute
	config.MaxEntries = 100
	if provider != "redis" && provider != "advanced" {
		return config
	}
	config.RedisAddr = os.Getenv("LAB_REDIS_ADDR")
	if config.RedisAddr == "" {
		t.Skip("set LAB_REDIS_ADDR to run real Redis experiments")
	}
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		t.Fatal(err)
	}
	config.RedisPrefix = "lab04:experiment:" + hex.EncodeToString(token[:]) + ":"
	client := redis.NewClient(&redis.Options{Addr: config.RedisAddr, MaxRetries: -1, ContextTimeoutEnabled: true, ReadTimeout: time.Second, WriteTimeout: time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		t.Fatalf("Redis unavailable: %v", err)
	}
	t.Cleanup(func() {
		defer client.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		// Удаляем только собственные ключи опыта, не всю базу/чужое пространство имён.
		var cursor uint64
		for {
			keys, next, err := client.Scan(ctx, cursor, config.RedisPrefix+"*", 100).Result()
			if err != nil {
				t.Errorf("cleanup scan: %v", err)
				return
			}
			if len(keys) > 0 {
				if err := client.Del(ctx, keys...).Err(); err != nil {
					t.Errorf("cleanup delete: %v", err)
					return
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	})
	return config
}

func experimentApp(t *testing.T, config Config) *App {
	t.Helper()
	app, err := NewApp(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Error(err)
		}
	})
	return app
}

func experimentPayment(app *App) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", strings.NewReader(`{"from":"alice","to":"bob","amount":100}`))
	req.Header.Set("Idempotency-Key", "same-operation")
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Server.Handler().ServeHTTP(response, req)
	return response
}

func experimentEffects(t *testing.T, app *App) int {
	t.Helper()
	response := httptest.NewRecorder()
	app.Server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/effects", nil))
	var result struct {
		Count int `json:"count"`
	}
	if response.Code != 200 {
		t.Fatalf("effects status=%d", response.Code)
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result.Count
}
