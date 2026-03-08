/**
 * Redis-backed stock counter for atomic reserve (Strategy 4).
 * Key pattern: inventory:stock:{sku} → number.
 */
export interface IRedisStockStore {
	/** Get current stock for SKU, or null if not set. */
	get(sku: string): Promise<number | null>;

	/** Set stock (e.g. seed from PostgreSQL). */
	set(sku: string, quantity: number): Promise<void>;

	/**
	 * Atomically decrement by quantity if current >= quantity.
	 * @returns new balance after decrement, or null if insufficient stock.
	 */
	decrementIfSufficient(sku: string, quantity: number): Promise<number | null>;

	/**
	 * Atomically: if key missing, set to initialValue; then decrement by quantity if current >= quantity.
	 * Avoids race on cold start (many requests seeing null and all doing set + decrement).
	 * @returns new balance after decrement, or null if insufficient stock.
	 */
	decrementIfSufficientOrInit(
		sku: string,
		initialValue: number,
		quantity: number,
	): Promise<number | null>;

	/**
	 * Atomically add quantity to stock (Redis INCRBY). Used for compensating transaction:
	 * if PG write fails after Redis decrement, call increment to restore Redis counter.
	 * @returns new balance after increment.
	 */
	increment(sku: string, quantity: number): Promise<number>;
}
