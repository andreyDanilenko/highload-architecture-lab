import { IRedisStockStore } from "@/contracts/redis-stock.contracts";

/** Minimal Redis client for stock: get, set, and either eval(options) or sendCommand(['EVAL', ...]). */
export interface RedisClientLike {
	get(key: string): Promise<string | null>;
	set(key: string, value: string): Promise<unknown>;
	eval?(
		script: string,
		options: { keys?: string[]; arguments?: (string | number)[] },
	): Promise<unknown>;
	sendCommand?(args: string[]): Promise<unknown>;
}

const KEY_PREFIX = "inventory:stock:";

function key(sku: string): string {
	return `${KEY_PREFIX}${sku}`;
}

/**
 * Lua: if current >= quantity then DECRBY and return new balance; else return -1.
 * KEYS[1] = key, ARGV[1] = quantity (number to subtract).
 */
const DECR_IF_SUFFICIENT_SCRIPT = `
local k = KEYS[1]
local qty = tonumber(ARGV[1])
local cur = redis.call('GET', k)
if cur == false then return -1 end
cur = tonumber(cur)
if cur >= qty then
  redis.call('DECRBY', k, qty)
  return cur - qty
else
  return -1
end
`;

/**
 * Lua: if key missing, set to initialValue; then if current >= quantity, DECRBY and return new balance; else return -1.
 * Atomic seed + decrement — no race on cold start.
 * KEYS[1] = key, ARGV[1] = initialValue (when key missing), ARGV[2] = quantity to subtract.
 */
const INIT_AND_DECR_IF_SUFFICIENT_SCRIPT = `
local k = KEYS[1]
local init = tonumber(ARGV[1])
local qty = tonumber(ARGV[2])
local cur = redis.call('GET', k)
if cur == false then
  redis.call('SET', k, init)
  cur = init
else
  cur = tonumber(cur)
end
if cur >= qty then
  redis.call('DECRBY', k, qty)
  return cur - qty
else
  return -1
end
`;

export class RedisStockRepository implements IRedisStockStore {
	constructor(private readonly client: RedisClientLike) {}

	async get(sku: string): Promise<number | null> {
		const val = await this.client.get(key(sku));
		if (val === null || val === undefined) return null;
		const n = Number.parseInt(val, 10);
		return Number.isNaN(n) ? null : n;
	}

	async set(sku: string, quantity: number): Promise<void> {
		await this.client.set(key(sku), String(quantity));
	}

	async decrementIfSufficient(sku: string, quantity: number): Promise<number | null> {
		const k = key(sku);
		const q = String(quantity);
		let raw: unknown;
		if (typeof this.client.eval === "function") {
			raw = await this.client.eval(DECR_IF_SUFFICIENT_SCRIPT, {
				keys: [k],
				arguments: [q],
			});
		} else if (typeof this.client.sendCommand === "function") {
			// Fallback: raw EVAL script numkeys key [key ...] arg [arg ...]
			raw = await this.client.sendCommand([
				"EVAL",
				DECR_IF_SUFFICIENT_SCRIPT,
				"1",
				k,
				q,
			]);
		} else {
			throw new Error("Redis client has neither eval nor sendCommand");
		}
		const result = Number(raw);
		if (Number.isNaN(result) || result < 0) return null;
		return result;
	}

	async decrementIfSufficientOrInit(
		sku: string,
		initialValue: number,
		quantity: number,
	): Promise<number | null> {
		const k = key(sku);
		const init = String(initialValue);
		const q = String(quantity);
		let raw: unknown;
		if (typeof this.client.eval === "function") {
			raw = await this.client.eval(INIT_AND_DECR_IF_SUFFICIENT_SCRIPT, {
				keys: [k],
				arguments: [init, q],
			});
		} else if (typeof this.client.sendCommand === "function") {
			raw = await this.client.sendCommand([
				"EVAL",
				INIT_AND_DECR_IF_SUFFICIENT_SCRIPT,
				"1",
				k,
				init,
				q,
			]);
		} else {
			throw new Error("Redis client has neither eval nor sendCommand");
		}
		const result = Number(raw);
		if (Number.isNaN(result) || result < 0) return null;
		return result;
	}
}
