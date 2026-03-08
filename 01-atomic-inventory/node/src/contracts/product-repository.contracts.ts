import type { PoolClient } from "pg";
import { Product, CreateProductDTO } from "@/models/product";

export interface IProductRepository {
	/**
	 * Find product by SKU
	 * @param sku - product SKU
	 * @returns Product or null if not found
	 */
	findBySku(sku: string): Promise<Product | null>;

	/**
	 * Get current stock quantity
	 * @param sku - product SKU
	 * @returns stock quantity or null if product not found
	 */
	getStock(sku: string): Promise<number | null>;

	/**
	 * Update stock with optimistic locking (using version)
	 * @param sku - product SKU
	 * @param newQuantity - new stock quantity
	 * @param version - current version for optimistic lock
	 * @returns true if updated, false if version mismatch
	 */
	updateStock(
		sku: string,
		newQuantity: number,
		version: number,
	): Promise<boolean>;

	/**
	 * Simple UPDATE of stock (no lock/version). Safe when writing a known value (e.g. syncing PG after Redis reserve).
	 * Do not use for read-modify-write under concurrency — use updateStockWithClient or updateStock (optimistic) instead.
	 */
	updateStockNaive(sku: string, newQuantity: number): Promise<boolean>;

	/**
	 * Create new product
	 * @param product - product data
	 * @returns created product
	 * @throws DuplicateSkuError if SKU already exists
	 */
	create(product: CreateProductDTO): Promise<Product>;

	/**
	 * Check if product exists
	 * @param sku - product SKU
	 * @returns true if exists
	 */
	exists(sku: string): Promise<boolean>;

	/**
	 * Find product by SKU and lock row for update (pessimistic locking).
	 * Must be called inside a transaction (same client).
	 */
	findBySkuWithLock(client: PoolClient, sku: string): Promise<Product | null>;

	/**
	 * Update stock using existing transaction client.
	 * Must be called inside a transaction after findBySkuWithLock.
	 */
	updateStockWithClient(
		client: PoolClient,
		sku: string,
		newQuantity: number,
	): Promise<boolean>;

	/**
	 * Atomically decrement stock by quantity. Use inside a transaction (e.g. after Redis reserve).
	 * @returns true if row was updated (sufficient stock), false if stock_quantity < quantity
	 */
	decrementStockWithClient(
		client: PoolClient,
		sku: string,
		quantity: number,
	): Promise<boolean>;
}
