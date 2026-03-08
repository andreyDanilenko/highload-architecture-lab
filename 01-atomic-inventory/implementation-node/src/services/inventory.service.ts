import type { Pool } from "pg";
import { CreateTransactionDTO } from "@/models/transaction";
import {
	IInventoryService,
	ReserveResult,
} from "@/contracts/inventory-service.contracts";
import { IProductRepository } from "@/contracts/product-repository.contracts";
import { ITransactionRepository } from "@/contracts/transaction-repository.contracts";
import {
	NotFoundError,
	InsufficientStockError,
	BusinessError,
} from "@/shared/errors/app-errors";
import { withTransaction } from "@/shared/db/transaction";

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
	) {}

	async reserveStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
		const existingTx = await this.transactionRepo.findByRequestId(
			dto.requestId,
		);
		if (existingTx) {
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

		// 2. Check product exists
		const product = await this.productRepo.findBySku(dto.sku);
		if (!product) {
			throw new NotFoundError("Product", { sku: dto.sku });
		}

		// 3. Check stock
		if (product.stockQuantity < dto.quantity) {
			throw new InsufficientStockError(
				dto.sku,
				dto.quantity,
				product.stockQuantity,
			);
		}

		// 4. Compute new balance
		const newQuantity = product.stockQuantity - dto.quantity;

		// 5. Update (naive, no locking)
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

		// 6. Create transaction record
		const transaction = await this.transactionRepo.create(dto);

		return {
			success: true,
			duplicated: false,
			newBalance: newQuantity,
			transaction,
		};
	}

	async reserveStockPessimistic(
		dto: CreateTransactionDTO,
	): Promise<ReserveResult> {
		const existingTx = await this.transactionRepo.findByRequestId(
			dto.requestId,
		);
		if (existingTx) {
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

		return withTransaction(this.pool, async (client) => {
			const product = await this.productRepo.findBySkuWithLock(client, dto.sku);
			if (!product) throw new NotFoundError("Product", { sku: dto.sku });

			if (product.stockQuantity < dto.quantity) {
				throw new InsufficientStockError(
					dto.sku,
					dto.quantity,
					product.stockQuantity,
				);
			}

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
			return {
				success: true,
				duplicated: false,
				newBalance: newQuantity,
				transaction,
			};
		});
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

	async releaseStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
		// TODO: compensation / rollback implementation
		throw new Error("Method not implemented");
	}
}
