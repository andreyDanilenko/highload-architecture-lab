import type { Pool } from "pg";
import { Product } from "@/models/product";
import { CreateTransactionDTO } from "@/models/transaction";
import {
	IInventoryService,
	ReserveResult,
} from "@/contracts/inventory-service.contracts";
import { IProductRepository } from "@/contracts/product-repository.contracts";
import { IRedisStockStore } from "@/contracts/redis-stock.contracts";
import { ITransactionRepository } from "@/contracts/transaction-repository.contracts";
import {
	NotFoundError,
	InsufficientStockError,
	BusinessError,
} from "@/shared/errors/app-errors";
import { withTransaction } from "@/shared/db/transaction";
import type { InventoryConfig } from "@/config/inventory";

/**
 * Domain layer: business rules and domain errors only.
 * - Decides "can this operation happen?" and throws AppError (NotFound, InsufficientStock, etc.) when not.
 * - Does not know about HTTP; controller + error-handler map errors to status codes.
 */
export class InventoryService implements IInventoryService {
	constructor(
		private productRepo: IProductRepository,
		private transactionRepo: ITransactionRepository,
		private pool: Pool,
		private config: InventoryConfig,
		private redisStore: IRedisStockStore,
	) {}

	/** Resolve idempotent reserve: return result if duplicate request, otherwise null. */
	private async resolveIdempotentReserve(
		dto: CreateTransactionDTO,
	): Promise<ReserveResult | null> {
		const existingTx = await this.transactionRepo.findByRequestId(
			dto.requestId,
		);
		if (!existingTx) return null;
		if (existingTx.sku !== dto.sku || existingTx.quantity !== dto.quantity) {
			throw new BusinessError(
				"Request payload mismatch",
				"PAYLOAD_MISMATCH",
				{ requestId: dto.requestId, existing: existingTx, new: dto },
			);
		}
		return {
			success: true,
			duplicated: true,
			newBalance: await this.getBalance(dto.sku),
		};
	}

	/** Throw NotFoundError or InsufficientStockError if product missing or stock too low. */
	private validateReserveQuantity(
		product: Product | null,
		dto: CreateTransactionDTO,
	): asserts product is Product {
		if (!product) {
			throw new NotFoundError("Product", { sku: dto.sku });
		}
		if (product.stockQuantity < dto.quantity) {
			throw new InsufficientStockError(
				dto.sku,
				dto.quantity,
				product.stockQuantity,
			);
		}
	}

	/** Build success result for a new (non-duplicate) reservation. */
	private successResult(
		newBalance: number,
		transaction: Awaited<
			ReturnType<ITransactionRepository["create"]>
		>,
	): ReserveResult {
		return {
			success: true,
			duplicated: false,
			newBalance,
			transaction,
		};
	}

	/**
	 * Naive reserve: read-modify-write without lock. Race condition under concurrency.
	 * Kept for demo and load-test only; use reserveStockPessimistic or reserveStockOptimistic in production.
	 */
	async reserveStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
		const idempotent = await this.resolveIdempotentReserve(dto);
		if (idempotent) return idempotent;

		const product = await this.productRepo.findBySku(dto.sku);
		this.validateReserveQuantity(product, dto);

		const newQuantity = product.stockQuantity - dto.quantity;
		const updated = await this.productRepo.updateStockNaive(
			dto.sku,
			newQuantity,
		);
		if (!updated) {
			throw new BusinessError("Failed to update stock", "UPDATE_FAILED", {
				sku: dto.sku,
				newQuantity,
			});
		}

		const transaction = await this.transactionRepo.create(dto);
		return this.successResult(newQuantity, transaction);
	}

	async reserveStockPessimistic(
		dto: CreateTransactionDTO,
	): Promise<ReserveResult> {
		const idempotent = await this.resolveIdempotentReserve(dto);
		if (idempotent) return idempotent;

		return withTransaction(this.pool, async (client) => {
			const product = await this.productRepo.findBySkuWithLock(client, dto.sku);
			this.validateReserveQuantity(product, dto);

			const newQuantity = product.stockQuantity - dto.quantity;
			const updated = await this.productRepo.updateStockWithClient(
				client,
				dto.sku,
				newQuantity,
			);

			if (!updated) {
				throw new BusinessError("Failed to update stock", "UPDATE_FAILED", {
					sku: dto.sku,
					newQuantity,
				});
			}

			const transaction = await this.transactionRepo.createWithClient(
				client,
				dto,
			);
			return this.successResult(newQuantity, transaction);
		});
	}

	/**
	 * Reserve stock using optimistic locking: read version, update with version check, retry on conflict.
	 */
	async reserveStockOptimistic(
		dto: CreateTransactionDTO,
	): Promise<ReserveResult> {
		const idempotent = await this.resolveIdempotentReserve(dto);
		if (idempotent) return idempotent;

		for (let attempt = 0; attempt < this.config.maxOptimisticRetries; attempt++) {
			const product = await this.productRepo.findBySku(dto.sku);
			this.validateReserveQuantity(product, dto);

			const newQuantity = product.stockQuantity - dto.quantity;
			const updated = await this.productRepo.updateStock(
				dto.sku,
				newQuantity,
				product.version,
			);

			if (updated) {
				const transaction = await this.transactionRepo.create(dto);
				return this.successResult(newQuantity, transaction);
			}
			// Version conflict — retry (re-read and try again)
		}

		throw new BusinessError(
			"Optimistic lock: too many retries",
			"OPTIMISTIC_RETRY_EXHAUSTED",
			{ sku: dto.sku, requestId: dto.requestId },
		);
	}

	/**
	 * Reserve stock using Redis atomic counter; persist to PostgreSQL.
	 */
	async reserveStockRedis(dto: CreateTransactionDTO): Promise<ReserveResult> {
		const idempotent = await this.resolveIdempotentReserve(dto);
		if (idempotent) return idempotent;

		const product = await this.productRepo.findBySku(dto.sku);
		if (!product) {
			throw new NotFoundError("Product", { sku: dto.sku });
		}

		// Atomic: init from PG if key missing, then decrement (no race on cold start)
		const newBalance = await this.redisStore.decrementIfSufficientOrInit(
			dto.sku,
			product.stockQuantity,
			dto.quantity,
		);
		if (newBalance === null) {
			const current = await this.redisStore.get(dto.sku);
			throw new InsufficientStockError(
				dto.sku,
				dto.quantity,
				current ?? 0,
			);
		}

		const transaction = await this.transactionRepo.create(dto);
		// Sync PG so getBalance and other strategies stay consistent
		await this.productRepo.updateStockNaive(dto.sku, newBalance);

		return this.successResult(newBalance, transaction);
	}

	async getBalance(sku: string): Promise<number> {
		const stock = await this.productRepo.getStock(sku);
		if (stock === null) {
			throw new NotFoundError("Product", { sku });
		}
		return stock;
	}

	async hasSufficientStock(sku: string, quantity: number): Promise<boolean> {
		const stock = await this.getBalance(sku);
		return stock >= quantity;
	}

	/**
	 * @deprecated Not implemented. Do not use in production until compensation/rollback is implemented.
	 */
	async releaseStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
		// TODO: compensation / rollback implementation
		throw new Error("Method not implemented");
	}
}
