package optimistic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"anti-bruteforce/internal/domain"

	"github.com/redis/go-redis/v9"
)

// errOverLimit — специальная ошибка-маркер для случая "превышен лимит"
// Нужна, чтобы отличать "429" (лимит) от "надо повторить" и "реальной ошибки"
var errOverLimit = errors.New("over limit")

// OptimisticLimiter реализует rate limiting через оптимистичные транзакции в Redis
// Смысл подхода: "не блокирую, но слежу, не изменилось ли ничего за время моей работы"
type OptimisticLimiter struct {
	client *redis.Client
}

func NewOptimisticLimiter(client *redis.Client) *OptimisticLimiter {
	return &OptimisticLimiter{client: client}
}

// Allow проверяет, можно ли пропустить запрос от данного IP
// Возвращает:
//   - true, nil — запрос разрешен
//   - false, nil — запрос отклонен (лимит исчерпан)
//   - false, error — техническая ошибка (Redis недоступен и т.д.)
func (o *OptimisticLimiter) Allow(ip string, limit int, windowSec int64) (bool, error) {
	// Контекст для операций с Redis. В этом примере просто пустой,
	// но в реальном проекте сюда можно добавить таймаут
	ctx := context.Background()

	// ---------- ПОДГОТОВКА ДАННЫХ ----------
	// В оптимистичном подходе у нас ТОЛЬКО ОДИН КЛЮЧ в Redis — сами данные
	// Никакого отдельного ключа-замка нет!
	key := fmt.Sprintf("rate:%s", ip)

	// Текущее время — нужно для расчета окна и для метки запроса
	now := time.Now()

	// Граница окна: сейчас минус windowSec секунд
	// Например, сейчас 10:30:00, windowSec=60 → cutoff = 10:29:00
	cutoff := now.Add(-time.Duration(windowSec) * time.Second).Unix()

	// cutoffStr нужен для удаления старых записей
	// cutoff-1 потому что cutoff включительно, а нам нужны строго меньше cutoff
	cutoffStr := fmt.Sprintf("%d", cutoff-1)

	// Для нового запроса готовим:
	// score — время запроса в секундах (для сортировки в Sorted Set)
	// member — уникальный ID запроса (наносекунды, чтобы различать запросы с одинаковым временем)
	score := float64(now.Unix())
	member := fmt.Sprintf("%d", now.UnixNano())

	// Сколько раз пробовать при конфликте
	const maxRetries = 3
	var lastErr error

	// ---------- ЦИКЛ ПОВТОРНЫХ ПОПЫТОК ----------
	// В оптимистичном подходе мы повторяем не при "не могу взять замок",
	// а при "данные изменились, пока я работал"
	for attempt := 0; attempt < maxRetries; attempt++ {
		// ---------- WATCH: НАЧИНАЕМ НАБЛЮДЕНИЕ ----------
		// REAL REDIS COMMAND: WATCH rate:192.168.1.1
		//
		// WATCH говорит Redis: "следи за этим ключом. Если кто-то его изменит,
		// пока я выполняю транзакцию — отмени её"
		//
		// Это АНАЛОГ version из первой задачи! Вместо явной версии мы следим
		// за состоянием всего ключа целиком
		err := o.client.Watch(ctx, func(tx *redis.Tx) error {
			// ---------- ШАГ 1: ЧИТАЕМ ТЕКУЩИЕ ДАННЫЕ ----------
			// REAL REDIS COMMAND: ZRANGE rate:192.168.1.1 0 -1 WITHSCORES
			//
			// ВНИМАНИЕ! В отличие от pessimistic, где мы делали ZCOUNT прямо в Redis,
			// здесь мы читаем ВСЕ записи и фильтруем на стороне Go
			// Это может быть проблемой при миллионе запросов от одного IP
			entries, err := tx.ZRangeWithScores(ctx, key, 0, -1).Result()
			if err != nil && err != redis.Nil {
				return err
			}

			// ---------- ШАГ 2: ФИЛЬТРУЕМ И СЧИТАЕМ (НА СТОРОНЕ КЛИЕНТА) ----------
			// Проходим по всем записям и считаем только те, что попадают в окно
			var count int
			for _, z := range entries {
				if z.Score >= float64(cutoff) {
					count++
				}
			}

			// ---------- ШАГ 3: ПРОВЕРЯЕМ ЛИМИТ ----------
			if count >= limit {
				// Специальная ошибка-маркер для случая "превышен лимит"
				return errOverLimit
			}

			// ---------- ШАГ 4: ПОДГОТАВЛИВАЕМ ИЗМЕНЕНИЯ (В ТРАНЗАКЦИИ) ----------
			// REAL REDIS COMMANDS:
			// MULTI
			// ZREMRANGEBYSCORE rate:192.168.1.1 -inf 1678901234
			// ZADD rate:192.168.1.1 1678901295 1678901295123456789
			// EXEC
			//
			// TxPipelined автоматически оборачивает команды в MULTI/EXEC
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				// Удаляем старые записи (строго меньше cutoff)
				pipe.ZRemRangeByScore(ctx, key, "-inf", cutoffStr)
				// Добавляем новый запрос
				pipe.ZAdd(ctx, key, redis.Z{
					Score:  score,
					Member: member,
				})
				return nil
			})
			return err
		}, key) // WATCH применяется к этому ключу

		// ---------- ШАГ 5: АНАЛИЗИРУЕМ РЕЗУЛЬТАТ ----------

		// 5.1: Ошибки нет — значит EXEC выполнился успешно, данные записаны
		if err == nil {
			return true, nil
		}

		// 5.2: Специальная ошибка "превышен лимит" — возвращаем false без retry
		if errors.Is(err, errOverLimit) {
			return false, nil
		}

		// 5.3: TxFailedErr — это значит, что WATCH сработал!
		// Кто-то изменил ключ между нашим чтением и EXEC
		// Это АНАЛОГ version mismatch из первой задачи
		// Просто запоминаем ошибку и идем на следующий виток цикла
		if errors.Is(err, redis.TxFailedErr) {
			lastErr = err
			continue
		}

		// 5.4: Любая другая ошибка — реальная проблема с Redis
		return false, err
	}

	// Если вышли из цикла после всех попыток — значит, так и не смогли
	// выполнить транзакцию из-за постоянных конфликтов
	return false, lastErr
}

// Проверка реализации интерфейса на этапе компиляции
var _ domain.RateLimiter = (*OptimisticLimiter)(nil)
