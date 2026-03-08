import { Pool, type PoolClient } from "pg";
import {
	InventoryTransaction,
	CreateTransactionDTO,
} from "@/models/transaction";
import { ITransactionRepository } from "@/contracts/transaction-repository.contracts";
import {
	DatabaseError,
	DuplicateRequestError,
} from "@/shared/errors/app-errors";
import { isPostgresError } from "@/shared/utils/errors";

interface TransactionRow {
	id: number;
	sku: string;
	quantity: number;
	request_id: string;
	created_at: Date;
}

/**
 * Repository for inventory_transactions table.
 * Handles idempotency and audit logging of stock operations.
 */
export class TransactionRepository implements ITransactionRepository {
	constructor(private readonly pool: Pool) {}

	/**
	 * Creates transaction record. Throws DuplicateRequestError if requestId already exists.
	 */
	async create(
		transaction: CreateTransactionDTO,
	): Promise<InventoryTransaction> {
		try {
			const { rows } = await this.pool.query<TransactionRow>(
				`INSERT INTO inventory_transactions (sku, quantity, request_id) 
             VALUES ($1, $2, $3) 
             RETURNING *`,
				[transaction.sku, transaction.quantity, transaction.requestId],
			);

			return this.mapToEntity(rows[0]);
		} catch (error) {
			if (isPostgresError(error) && error.code === "23505") {
				throw new DuplicateRequestError(transaction.requestId, transaction.sku);
			}

			throw new DatabaseError("TransactionRepository.create", {
				cause: error,
				transaction,
			});
		}
	}

	/**
	 * Create transaction record using existing client (inside a transaction).
	 */
	async createWithClient(
		client: PoolClient,
		transaction: CreateTransactionDTO,
	): Promise<InventoryTransaction> {
		try {
			const { rows } = await client.query<TransactionRow>(
				`INSERT INTO inventory_transactions (sku, quantity, request_id) 
             VALUES ($1, $2, $3) 
             RETURNING *`,
				[transaction.sku, transaction.quantity, transaction.requestId],
			);
			return this.mapToEntity(rows[0]);
		} catch (error) {
			if (isPostgresError(error) && error.code === "23505") {
				throw new DuplicateRequestError(transaction.requestId, transaction.sku);
			}
			throw new DatabaseError("TransactionRepository.createWithClient", {
				cause: error,
				transaction,
			});
		}
	}

	/**
	 * Finds transaction by requestId. Returns null if not found.
	 * Used for idempotency checks.
	 */
	async findByRequestId(
		requestId: string,
	): Promise<InventoryTransaction | null> {
		try {
			const { rows } = await this.pool.query<TransactionRow>(
				"SELECT * FROM inventory_transactions WHERE request_id = $1",
				[requestId],
			);

			if (rows.length === 0) {
				return null;
			}

			return this.mapToEntity(rows[0]);
		} catch (error) {
			throw new DatabaseError("TransactionRepository.findByRequestId", {
				cause: error,
				requestId,
			});
		}
	}

	/**
	 * Returns all transactions for a SKU, newest first.
	 */
	async findBySku(sku: string): Promise<InventoryTransaction[]> {
		try {
			const { rows } = await this.pool.query<TransactionRow>(
				"SELECT * FROM inventory_transactions WHERE sku = $1 ORDER BY created_at DESC",
				[sku],
			);

			return rows.map((row) => this.mapToEntity(row));
		} catch (error) {
			throw new DatabaseError("TransactionRepository.findBySku", {
				cause: error,
				sku,
			});
		}
	}

	/**
	 * Fast existence check using SELECT 1.
	 */
	async exists(requestId: string): Promise<boolean> {
		try {
			const { rows } = await this.pool.query(
				"SELECT 1 FROM inventory_transactions WHERE request_id = $1",
				[requestId],
			);

			return rows.length > 0;
		} catch (error) {
			throw new DatabaseError("TransactionRepository.exists", {
				cause: error,
				requestId,
			});
		}
	}

	/**
	 * Returns total deducted quantity for a SKU. Returns 0 if no transactions.
	 */
	async getTotalDeducted(sku: string): Promise<number> {
		try {
			const { rows } = await this.pool.query<{ total: string }>(
				"SELECT COALESCE(SUM(quantity), 0) as total FROM inventory_transactions WHERE sku = $1",
				[sku],
			);

			return parseInt(rows[0].total);
		} catch (error) {
			throw new DatabaseError("TransactionRepository.getTotalDeducted", {
				cause: error,
				sku,
			});
		}
	}

	private mapToEntity(row: TransactionRow): InventoryTransaction {
		return {
			id: row.id,
			sku: row.sku,
			quantity: row.quantity,
			requestId: row.request_id, // snake_case -> camelCase
			createdAt: row.created_at,
		};
	}
}
