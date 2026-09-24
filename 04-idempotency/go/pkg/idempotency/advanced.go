package idempotency

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdvancedProvider оставляет истёкший pending в карантине до явной сверки исхода.
// Защита действует пока запись есть в Redis: eviction, потеря данных и истечение
// completed TTL выходят за её границы. Внешний эффект не входит в транзакцию Lua.
type AdvancedProvider struct {
	client *redis.Client
	prefix string
}

func NewAdvancedProvider(client *redis.Client, prefix string) *AdvancedProvider {
	return &AdvancedProvider{client: client, prefix: prefix}
}

func (p *AdvancedProvider) Name() string { return "advanced" }

// Шаг 1: часы Redis задают общий lease; pending не получает TTL, чтобы его
// исчезновение не разрешило молча повторить операцию с неизвестным исходом.
const advancedAcquireLua = `
local now = redis.call('TIME')
local now_ms = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
local raw = redis.call('GET', KEYS[1])
if not raw then
    local record = {fingerprint=ARGV[1], owner=ARGV[2], status='pending', lease_until_ms=now_ms + tonumber(ARGV[3])}
    redis.call('SET', KEYS[1], cjson.encode(record))
    return {1, ''}
end
local record = cjson.decode(raw)
if record.fingerprint ~= ARGV[1] then return {3, ''} end
if record.status == 'completed' then return {2, record.response_json} end
if record.status ~= 'pending' or not record.lease_until_ms then return {6, ''} end
if now_ms >= record.lease_until_ms then return {5, ''} end
return {4, ''}
`

// Шаг 2: результат сохраняет только действующий lease и тот же owner.
// После окончания lease даже прежний owner должен пройти сверку исхода.
const advancedCompleteLua = `
local raw = redis.call('GET', KEYS[1])
if not raw then return 3 end
local record = cjson.decode(raw)
if record.status ~= 'pending' or record.owner ~= ARGV[1] or ARGV[1] == '' then return 3 end
if not record.lease_until_ms then return 6 end
local now = redis.call('TIME')
local now_ms = tonumber(now[1]) * 1000 + math.floor(tonumber(now[2]) / 1000)
if now_ms >= record.lease_until_ms then return 5 end
record.status = 'completed'
record.response_json = ARGV[2]
redis.call('SET', KEYS[1], cjson.encode(record), 'PX', ARGV[3])
return 1
`

func (p *AdvancedProvider) Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (Claim, error) {
	if err := ctx.Err(); err != nil {
		return Claim{}, err
	}
	if key == "" || fingerprint == "" || lease <= 0 {
		return Claim{}, ErrInvalidArgument
	}
	if p.client == nil {
		return Claim{}, ErrUnavailable
	}
	leaseMS, err := redisTTL(lease)
	if err != nil {
		return Claim{}, err
	}
	owner, err := newOwner()
	if err != nil {
		return Claim{}, ErrUnavailable
	}
	// EVAL — один запрос; повтор после транспортной ошибки здесь не запускается.
	result, err := p.client.Eval(ctx, advancedAcquireLua, []string{p.prefix + key}, fingerprint, owner, leaseMS).Slice()
	if err != nil {
		return Claim{}, redisFailure(err)
	}
	if len(result) != 2 {
		return Claim{}, ErrUnavailable
	}
	code, ok := result[0].(int64)
	if !ok {
		return Claim{}, ErrUnavailable
	}
	switch code {
	case 1:
		return Claim{Owner: owner}, nil
	case 2:
		data, ok := result[1].(string)
		if !ok {
			return Claim{}, ErrUnavailable
		}
		return replayFromJSON(data)
	case 3:
		return Claim{}, ErrMismatch
	case 4:
		return Claim{}, ErrInProgress
	case 5:
		return Claim{}, ErrOutcomeUnknown
	default:
		return Claim{}, ErrUnavailable
	}
}

func (p *AdvancedProvider) Complete(ctx context.Context, key, owner string, response Response, retention time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if retention <= 0 {
		return ErrInvalidArgument
	}
	if p.client == nil {
		return ErrUnavailable
	}
	retentionMS, err := redisTTL(retention)
	if err != nil {
		return err
	}
	data, err := json.Marshal(response)
	if err != nil {
		return ErrUnavailable
	}
	code, err := p.client.Eval(ctx, advancedCompleteLua, []string{p.prefix + key}, owner, string(data), retentionMS).Int64()
	if err != nil {
		return redisFailure(err)
	}
	switch code {
	case 1:
		return nil
	case 3:
		return ErrNotOwner
	case 5:
		return ErrOutcomeUnknown
	default:
		return ErrUnavailable
	}
}

var _ Provider = (*AdvancedProvider)(nil)
