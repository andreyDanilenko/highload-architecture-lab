package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisFactory func(*redis.Client, string) Provider

var redisFactories = []struct {
	name string
	new  redisFactory
}{
	{"redis", func(c *redis.Client, prefix string) Provider { return NewRedisProvider(c, prefix) }},
	{"advanced", func(c *redis.Client, prefix string) Provider { return NewAdvancedProvider(c, prefix) }},
}

// Тесты используют настоящий Redis только при явном LAB_REDIS_ADDR.
// Каждый тест удаляет ровно свой ключ; чужие данные и FLUSHDB не затрагиваются.
func redisTestStore(t *testing.T) (context.Context, *redis.Client, string) {
	t.Helper()
	addr := os.Getenv("LAB_REDIS_ADDR")
	if addr == "" {
		t.Skip("set LAB_REDIS_ADDR to run Redis integration tests")
	}
	client := redis.NewClient(&redis.Options{
		Addr: addr, MaxRetries: -1, ContextTimeoutEnabled: true,
		DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	owner, err := newOwner()
	if err != nil {
		t.Fatal(err)
	}
	prefix := "lab-idempotency-test:" + owner + ":"
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()
		if err := client.Del(cleanupCtx, prefix+"request").Err(); err != nil {
			t.Errorf("cleanup owned key: %v", err)
		}
		_ = client.Close()
	})
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("LAB_REDIS_ADDR was set but Redis is unavailable: %v", err)
	}
	return ctx, client, prefix
}

func mustRedisClaim(t *testing.T, ctx context.Context, p Provider, lease time.Duration) Claim {
	t.Helper()
	claim, err := p.Acquire(ctx, "request", "fingerprint", lease)
	if err != nil || claim.Owner == "" || claim.Response != nil {
		t.Fatalf("Acquire: claim=%+v err=%v", claim, err)
	}
	return claim
}

func redisEventually(t *testing.T, ctx context.Context, check func(context.Context) (bool, error)) {
	t.Helper()
	pollCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		ok, err := check(pollCtx)
		if err != nil {
			t.Fatalf("poll Redis state: %v", err)
		}
		if ok {
			return
		}
		select {
		case <-pollCtx.Done():
			t.Fatal("Redis state did not reach the expected boundary before deadline")
		case <-ticker.C:
		}
	}
}

func waitRedisKeyGone(t *testing.T, ctx context.Context, client *redis.Client, key string) {
	t.Helper()
	redisEventually(t, ctx, func(ctx context.Context) (bool, error) {
		count, err := client.Exists(ctx, key).Result()
		return count == 0, err
	})
}

func waitAdvancedLease(t *testing.T, ctx context.Context, client *redis.Client, key string) {
	t.Helper()
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var record redisRecord
	if err := json.Unmarshal(data, &record); err != nil || record.LeaseUntilMS == 0 {
		t.Fatalf("pending record has no lease: %v", err)
	}
	redisEventually(t, ctx, func(ctx context.Context) (bool, error) {
		now, err := client.Time(ctx).Result()
		return now.UnixMilli() >= record.LeaseUntilMS, err
	})
}

func TestRedisProvidersConcurrentAcquire(t *testing.T) {
	for _, factory := range redisFactories {
		t.Run(factory.name, func(t *testing.T) {
			ctx, client, prefix := redisTestStore(t)
			providers := []Provider{factory.new(client, prefix), factory.new(client, prefix)}
			const attempts = 24
			results := make(chan error, attempts)
			start := make(chan struct{})
			var wg sync.WaitGroup
			for i := 0; i < attempts; i++ {
				wg.Add(1)
				go func(p Provider) {
					defer wg.Done()
					<-start
					claim, err := p.Acquire(ctx, "request", "fingerprint", 10*time.Second)
					if err == nil && (claim.Owner == "" || claim.Response != nil) {
						err = fmt.Errorf("successful acquire returned no new owner")
					}
					results <- err
				}(providers[i%len(providers)])
			}
			close(start)
			wg.Wait()
			close(results)
			owners := 0
			for err := range results {
				if err == nil {
					owners++
				} else if !errors.Is(err, ErrInProgress) {
					t.Errorf("contended acquire: %v", err)
				}
			}
			if owners != 1 {
				t.Fatalf("owners=%d, want one across provider instances", owners)
			}
		})
	}
}

func TestRedisProvidersReplayAndMismatch(t *testing.T) {
	for _, factory := range redisFactories {
		t.Run(factory.name, func(t *testing.T) {
			ctx, client, prefix := redisTestStore(t)
			first, second := factory.new(client, prefix), factory.new(client, prefix)
			claim := mustRedisClaim(t, ctx, first, 5*time.Second)
			if _, err := second.Acquire(ctx, "request", "different", time.Second); !errors.Is(err, ErrMismatch) {
				t.Fatalf("pending mismatch: %v", err)
			}
			response := Response{
				Status: http.StatusCreated,
				Header: http.Header{"Content-Type": {"application/octet-stream"}, "X-Values": {"first", "second"}},
				Body:   []byte{0, 255, '{', '}', '\n'},
			}
			if err := first.Complete(ctx, "request", claim.Owner, response, 5*time.Second); err != nil {
				t.Fatal(err)
			}
			replayed, err := second.Acquire(ctx, "request", "fingerprint", time.Second)
			if err != nil || replayed.Owner != "" || !reflect.DeepEqual(replayed.Response, &response) {
				t.Fatalf("replay changed response: claim=%+v err=%v", replayed, err)
			}
			if _, err := second.Acquire(ctx, "request", "different", time.Second); !errors.Is(err, ErrMismatch) {
				t.Fatalf("completed mismatch: %v", err)
			}
			if err := first.Complete(ctx, "request", claim.Owner, Response{Status: 500}, time.Second); !errors.Is(err, ErrNotOwner) {
				t.Fatalf("completed response was mutable: %v", err)
			}
			again, err := second.Acquire(ctx, "request", "fingerprint", time.Second)
			if err != nil || !reflect.DeepEqual(again.Response, &response) {
				t.Fatalf("repeat completion changed stored response: %v", err)
			}
		})
	}
}

func TestRedisBasicLeaseExpiryPermitsRepeatedEffect(t *testing.T) {
	ctx, client, prefix := redisTestStore(t)
	p := NewRedisProvider(client, prefix)
	first := mustRedisClaim(t, ctx, p, 40*time.Millisecond)
	effects := 1 // Внешний эффект уже произошёл, но процесс не сохранил ответ.
	waitRedisKeyGone(t, ctx, client, prefix+"request")
	if err := p.Complete(ctx, "request", first.Owner, Response{Status: 200}, time.Second); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("completion after key expiry: %v", err)
	}
	second := mustRedisClaim(t, ctx, NewRedisProvider(client, prefix), 5*time.Second)
	effects++ // TTL удалил сведения о первой попытке и разрешил повтор.
	if first.Owner == second.Owner || effects != 2 {
		t.Fatal("expected a new owner and a demonstrably repeated effect")
	}
}

func TestAdvancedLeaseExpiryQuarantinesUnknownOutcome(t *testing.T) {
	ctx, client, prefix := redisTestStore(t)
	p := NewAdvancedProvider(client, prefix)
	claim := mustRedisClaim(t, ctx, p, 40*time.Millisecond)
	waitAdvancedLease(t, ctx, client, prefix+"request")
	for i := 0; i < 3; i++ {
		other := NewAdvancedProvider(client, prefix)
		if _, err := other.Acquire(ctx, "request", "fingerprint", time.Second); !errors.Is(err, ErrOutcomeUnknown) {
			t.Fatalf("expired pending allowed a retry: %v", err)
		}
	}
	if err := p.Complete(ctx, "request", claim.Owner, Response{Status: 200}, time.Second); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("completion after lease expiry: %v", err)
	}
	if _, err := p.Acquire(ctx, "request", "different", time.Second); !errors.Is(err, ErrMismatch) {
		t.Fatalf("fingerprint mismatch after expiry: %v", err)
	}
	ttl, err := client.PTTL(ctx, prefix+"request").Result()
	if err != nil || ttl != -1 {
		t.Fatalf("quarantined pending must have no expiry: ttl=%v err=%v", ttl, err)
	}
}

func TestRedisProvidersStaleOwnerCannotComplete(t *testing.T) {
	for _, factory := range redisFactories {
		t.Run(factory.name, func(t *testing.T) {
			ctx, client, prefix := redisTestStore(t)
			p := factory.new(client, prefix)
			stale := mustRedisClaim(t, ctx, p, 40*time.Millisecond)
			if factory.name == "advanced" {
				waitAdvancedLease(t, ctx, client, prefix+"request")
				// Моделируем явную внешнюю сверку и очистку только ключа этого теста.
				// Сам AdvancedProvider истёкший pending никогда не удаляет.
				if err := client.Del(ctx, prefix+"request").Err(); err != nil {
					t.Fatal(err)
				}
			} else {
				waitRedisKeyGone(t, ctx, client, prefix+"request")
			}
			current := mustRedisClaim(t, ctx, factory.new(client, prefix), 5*time.Second)
			if err := p.Complete(ctx, "request", stale.Owner, Response{Status: 500}, time.Second); !errors.Is(err, ErrNotOwner) {
				t.Fatalf("stale owner completion: %v", err)
			}
			if _, err := p.Acquire(ctx, "request", "fingerprint", time.Second); !errors.Is(err, ErrInProgress) {
				t.Fatalf("stale owner changed current pending: %v", err)
			}
			want := Response{Status: 201, Body: []byte("current owner")}
			if err := p.Complete(ctx, "request", current.Owner, want, time.Second); err != nil {
				t.Fatal(err)
			}
			got, err := p.Acquire(ctx, "request", "fingerprint", time.Second)
			if err != nil || !reflect.DeepEqual(got.Response, &want) {
				t.Fatalf("current owner response: %+v err=%v", got, err)
			}
		})
	}
}

func TestRedisProvidersCompletedRetentionExpires(t *testing.T) {
	for _, factory := range redisFactories {
		t.Run(factory.name, func(t *testing.T) {
			ctx, client, prefix := redisTestStore(t)
			p := factory.new(client, prefix)
			first := mustRedisClaim(t, ctx, p, 5*time.Second)
			if err := p.Complete(ctx, "request", first.Owner, Response{Status: 200}, 40*time.Millisecond); err != nil {
				t.Fatal(err)
			}
			waitRedisKeyGone(t, ctx, client, prefix+"request")
			second := mustRedisClaim(t, ctx, p, time.Second)
			if first.Owner == second.Owner {
				t.Fatal("completed TTL must bound deduplication retention")
			}
		})
	}
}

func TestRedisProvidersValidateBeforeStorage(t *testing.T) {
	for _, factory := range redisFactories {
		t.Run(factory.name, func(t *testing.T) {
			p := factory.new(nil, "test:")
			ctx := context.Background()
			for _, input := range []struct {
				key, fingerprint string
				lease            time.Duration
			}{
				{"", "fingerprint", time.Second},
				{"request", "", time.Second},
				{"request", "fingerprint", 0},
				{"request", "fingerprint", -time.Second},
			} {
				if _, err := p.Acquire(ctx, input.key, input.fingerprint, input.lease); !errors.Is(err, ErrInvalidArgument) {
					t.Fatalf("invalid acquire reached storage: %v", err)
				}
			}
			if err := p.Complete(ctx, "request", "owner", Response{Status: 200}, 0); !errors.Is(err, ErrInvalidArgument) {
				t.Fatalf("invalid retention reached storage: %v", err)
			}
			cancelled, cancel := context.WithCancel(ctx)
			cancel()
			if _, err := p.Acquire(cancelled, "request", "fingerprint", time.Second); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancelled acquire reached storage: %v", err)
			}
			if err := p.Complete(cancelled, "request", "owner", Response{Status: 200}, time.Second); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancelled completion reached storage: %v", err)
			}
		})
	}
}

func TestRedisTTLSubmillisecondAndOverflow(t *testing.T) {
	// Нельзя округлением превратить короткий lease в бессрочный SET или переполнить duration.
	for _, test := range []struct {
		input time.Duration
		want  int64
	}{
		{time.Nanosecond, 1},
		{time.Millisecond, 1},
		{time.Millisecond + time.Nanosecond, 2},
		{time.Duration(1<<63-1) / time.Millisecond * time.Millisecond, int64(time.Duration(1<<63-1) / time.Millisecond)},
	} {
		got, err := redisTTL(test.input)
		if err != nil || got != test.want || time.Duration(got)*time.Millisecond <= 0 {
			t.Fatalf("duration=%v: ms=%d err=%v", test.input, got, err)
		}
	}
	if _, err := redisTTL(time.Duration(1<<63 - 1)); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("unrepresentable millisecond ceiling: %v", err)
	}
}

// net/http допускает трёхзначные финальные статусы до 999; JSON replay должен
// сохранять тот же диапазон, что и responseBuffer, включая нестандартные коды.
func TestRedisReplayStatusMatchesHTTPBuffer(t *testing.T) {
	for _, status := range []int{200, 599, 700, 999} {
		data, err := json.Marshal(Response{Status: status, Body: []byte("response")})
		if err != nil {
			t.Fatal(err)
		}
		claim, err := replayFromJSON(string(data))
		if err != nil || claim.Response.Status != status {
			t.Fatalf("status %d: claim=%+v err=%v", status, claim, err)
		}
	}
	for _, status := range []int{0, 100, 199, 1000} {
		data, err := json.Marshal(Response{Status: status})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := replayFromJSON(string(data)); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("status %d: %v", status, err)
		}
	}
}
