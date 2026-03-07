import { InventoryTransaction, CreateTransactionDTO } from '@/models/transaction';

export interface ITransactionRepository {
  /**
   * Create new inventory transaction
   * @param transaction - transaction data with requestId
   * @returns created transaction
   */
  create(transaction: CreateTransactionDTO): Promise<InventoryTransaction>;

  /**
   * Find transaction by requestId (for idempotency)
   * @param requestId - unique request identifier
   * @returns transaction or null if not found
   */
  findByRequestId(requestId: string): Promise<InventoryTransaction | null>;

  /**
   * Find all transactions for a product
   * @param sku - product SKU
   * @returns array of transactions
   */
  findBySku(sku: string): Promise<InventoryTransaction[]>;

  /**
   * Check if transaction exists for requestId
   * @param requestId - unique request identifier
   * @returns true if exists
   */
  exists(requestId: string): Promise<boolean>;

  /**
   * Get total quantity deducted for a product
   * @param sku - product SKU
   * @returns total deducted quantity
   */
  getTotalDeducted(sku: string): Promise<number>;
}
