# Task 04: Idempotency contract

This describes the current Go implementation. Commands, experiments, and validation evidence belong in the [task README](../../README.md) and [experiment notebook](../experiment-notes.md). The scenarios below are expectations to test, not recorded results.

## Purpose and boundaries

Compare four strategies behind one HTTP middleware: `noop`, `memory`, `redis`, and `advanced`. The experiment separates admission of an attempt, a simulated business effect, and storage of its response. No provider makes these three events one transaction.

`POST /api/v1/payments` accepts one JSON object:

```json
{"from":"account-a","to":"account-b","amount":100}
```

`amount` is a positive `int64` in minor units. Accounts must be nonempty, distinct, and at most 128 bytes. Unknown fields, trailing JSON, fractional amounts, and invalid payments are rejected. The service records a simulated effect; it does not move real funds or maintain account balances.

## Request identity

- Exactly one `Idempotency-Key`: 1–128 visible ASCII bytes, excluding spaces.
- Optional `X-Demo-Tenant`: 1–64 visible ASCII bytes, default `lab`. This demonstrates scope separation and does not authenticate a tenant.
- Storage key: SHA-256 of method, matched route pattern, tenant, and client key, separated unambiguously.
- Fingerprint: SHA-256 of the concrete escaped path, raw query string, and raw request body, without normalization. Changing a path parameter, query string, JSON whitespace, or field order causes a mismatch under the same scope/key while the record exists.
- Request bodies are bounded before admission. Authentication, canonical JSON, query normalization, and other header parameters are outside this endpoint's identity contract. Wildcard path values affect the fingerprint, not the scope: a different value under the same matched route/key conflicts rather than receiving an independent key.

## Provider API

[Provider](../../go/pkg/idempotency/provider.go) exposes:

```go
Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (Claim, error)
Complete(ctx context.Context, key, owner string, response Response, retention time.Duration) error
Name() string
```

A claim contains either an owner token or a stored response. `Complete` checks the attempt's owner in stateful providers. The token protects the record from a stale writer; it is not fencing enforced by the business effect.

| Strategy | Admission and replay | Expired unfinished attempt |
|----------|----------------------|----------------------------|
| `noop` | Every request reaches the handler; no fingerprint check or replay | No record exists |
| `memory` | Mutex-protected map, bounded entry count, response copies | Record remains; returns unknown, no takeover |
| `redis` | `SET NX` with pending TTL; `WATCH`/transaction for completion | Record disappears; a later call can execute again |
| `advanced` | Lua checks state and owner; Redis time defines lease | Pending has no TTL; returns unknown, no takeover |

All stateful providers retain completed responses for `ResultTTL`, starting at completion. After retention expires, the same key may execute again. Memory is process-local; Redis strategies coordinate only replicas sharing the same Redis and namespace. Do not mix strategies in one namespace: use a distinct `-redis-prefix` for a different strategy or independent experiment.

## HTTP flow and responses

1. Validate key, demo tenant, and body size; calculate scope and fingerprint.
2. Acquire admission, replay a stored response, or return a provider error.
3. Run the handler into a bounded response buffer.
4. Pass status, body bytes, and only `Content-Type`/`Location` to Complete using a context detached from client cancellation but limited by `CompleteTimeout`; noop deliberately stores nothing.
5. Send the buffered response to the client. A failed completion does not send the buffered business response as if persistence succeeded.

Stateful providers store completed handler responses, including ordinary 4xx/5xx responses. The same body/key replays that error until retention expires; an edited body conflicts while the record exists. Streaming, cookies, arbitrary headers, and unbounded bodies are not supported. A panic or response overflow leaves the attempt unfinished.

| Situation | HTTP result |
|-----------|-------------|
| Missing/invalid key or demo tenant | 400 |
| Request body above limit | 413 |
| Valid active attempt already exists | 409 `in_progress`, `Retry-After: 1` |
| Same scope/key with different path, query, or body | 409 `key_mismatch` |
| Provider reports expired/unknown ownership | 409 `outcome_unknown` |
| Store unavailable or memory capacity exhausted | 503 `idempotency_unavailable` |
| Completion fails or response exceeds buffer limit | 503 `outcome_unknown` |
| Handler panics | 500 `outcome_unknown` |

Unknown outcome is not proof that no effect occurred. A new client key bypasses this record and is not a recovery protocol.

## Observation and configuration

`GET /api/v1/effects` returns `{count, payments}` from a bounded, process-local in-memory journal. It deliberately does not deduplicate. Query every process in a multi-instance experiment. Restart loses that process's journal, and transaction IDs are local; this journal cannot prove durable crash recovery. Redis may retain a response after the corresponding process journal has been lost.

`GET /metrics` reports fixed outcome labels and aggregate duration sum/count, not latency quantiles. `GET /healthz` checks Redis for Redis strategies. These endpoints are educational observation tools, not authenticated operational APIs.

CLI selects `-provider noop|memory|redis|advanced` (default `memory`). Defaults: `-lease-ttl 30s`, `-result-ttl 24h`, `-complete-timeout 2s`, request/response limits 65536 bytes, `-max-entries 10000`, `-payment-delay 0`. The delay accepts 0–10s; the capacity bounds both the memory provider and each process's effect journal. See `go run ./cmd/server -help` from `04-idempotency/go` for all flags.

## Required questions for an experiment

- Does identical replay preserve the response without adding an effect?
- Do changed path, query, or body parameters conflict while the key is retained?
- What happens when the operation lasts longer than its lease?
- Can a stale owner overwrite another attempt's response?
- What happens after completed retention expires or stored data is lost?

For `advanced`, an expired pending record stays in quarantine until its outcome is resolved. No reconciliation, lease renewal, or administrative resolution API is implemented. Redis eviction, deletion, or data loss can remove the protection. Atomic Lua does not make the external effect atomic with response storage.
