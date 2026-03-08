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
	 * Reserve stock for a product
	 * @param dto - transaction data with sku, quantity, requestId
	 * @returns ReserveResult with status and details
	 */
	reserveStock(dto: CreateTransactionDTO): Promise<ReserveResult>;

	/**
	 * Release previously reserved stock (compensation)
	 * @param dto - transaction data
	 * @returns ReserveResult with status
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
