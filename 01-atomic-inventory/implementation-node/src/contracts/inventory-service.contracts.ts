import { Product } from "@/models/product";
import {
	InventoryTransaction,
	CreateTransactionDTO,
} from "@/models/transaction";

export interface ReserveResult {
	success: boolean;
	product?: Product;
	transaction?: InventoryTransaction;
	error?: string;
	duplicated?: boolean;
	newBalance?: number;
}

export interface IInventoryService {
	/**
	 * Reserve stock (naive, no locking). Race condition under concurrency — lost updates.
	 * @deprecated For demo/load-test only. Use reserveStockPessimistic (or optimistic/redis) in production.
	 */
	reserveStock(dto: CreateTransactionDTO): Promise<ReserveResult>;

	/**
	 * Reserve stock using pessimistic locking (SELECT FOR UPDATE in a single transaction).
	 */
	reserveStockPessimistic(dto: CreateTransactionDTO): Promise<ReserveResult>;

	/**
	 * Release previously reserved stock (compensation).
	 * @deprecated Not implemented. Do not use in production until implemented.
	 */
	releaseStock(dto: CreateTransactionDTO): Promise<ReserveResult>;

	/**
	 * Get current balance for a product
	 * @param sku - product SKU
	 * @returns current stock
	 * @throws NotFoundError if product does not exist
	 */
	getBalance(sku: string): Promise<number>;

	/**
	 * Check if product has sufficient stock
	 * @param sku - product SKU
	 * @param quantity - required quantity
	 * @returns true if enough stock
	 */
	hasSufficientStock(sku: string, quantity: number): Promise<boolean>;
}
