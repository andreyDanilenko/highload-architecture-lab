// Package idempotency сравнивает хранение ключей повторных HTTP-операций.
// Атомарность записи провайдера не включает эффект во внешней системе.
package idempotency

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

// Response хранит транспортный результат, а не указатель на доменную структуру.
// Поэтому memory и Redis возвращают одинаковые байты, статус и выбранные заголовки.
type Response struct {
	Status int         `json:"status"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
}

// Claim допускает ровно одно из двух состояний: владелец попытки или готовый ответ.
// Ошибки обработки и неизвестный исход возвращаются отдельно через error.
type Claim struct {
	Owner    string
	Response *Response
}

// Provider — порт хранения. Key уже содержит scope операции; fingerprint проверяет
// соответствие параметров. Lease ограничивает владение, retention — хранение ответа.
// Complete обязан отклонять чужого/устаревшего владельца. Noop намеренно не хранит
// состояние, а базовый Redis удаляет pending по TTL: это изучаемые ограничения.
type Provider interface {
	Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (Claim, error)
	Complete(ctx context.Context, key, owner string, response Response, retention time.Duration) error
	Name() string
}

// newOwner не связывает владение с instance ID: у каждой попытки свой token.
func newOwner() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(token[:]), nil
}

func cloneResponse(response Response) Response {
	response.Header = response.Header.Clone()
	response.Body = append([]byte(nil), response.Body...)
	return response
}
