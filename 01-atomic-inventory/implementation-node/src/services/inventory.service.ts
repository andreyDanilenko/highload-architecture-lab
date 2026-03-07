import { CreateTransactionDTO } from '../models/transaction';
import { IInventoryService, ReserveResult } from '../contracts/inventory-service.contracts';
import { IProductRepository } from '../contracts/product-repository.contracts';
import { ITransactionRepository } from '../contracts/transaction-repository.contracts';

export class InventoryService implements IInventoryService {
  constructor(
    private productRepo: IProductRepository,
    private transactionRepo: ITransactionRepository
  ) {}

  async reserveStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
    console.log(`🔍 [InventoryService] reserveStock called with:`, dto);
    
    const existingTx = await this.transactionRepo.findByRequestId(dto.requestId);
    if (existingTx) {
      console.log(`⚠️ Duplicate request detected: ${dto.requestId}`);
      return {
        success: true,
        duplicated: true,
        newBalance: await this.getBalance(dto.sku) || 0
      };
    }

    const product = await this.productRepo.findBySku(dto.sku);
    if (!product) {
      console.log(`❌ Product not found: ${dto.sku}`);
      return {
        success: false,
        error: `Product not found: ${dto.sku}`
      };
    }

    if (product.stockQuantity < dto.quantity) {
      console.log(`❌ Insufficient stock: available ${product.stockQuantity}, requested ${dto.quantity}`);
      return {
        success: false,
        error: `Insufficient stock. Available: ${product.stockQuantity}`,
        newBalance: product.stockQuantity
      };
    }

    // 4. Рассчитываем новый остаток
    const newQuantity = product.stockQuantity - dto.quantity;
    
    // 5. Обновляем остаток (наивно)
    const updated = await this.productRepo.updateStockNaive(dto.sku, newQuantity);
    
    if (!updated) {
      return {
        success: false,
        error: 'Failed to update stock'
      };
    }

    const transaction = await this.transactionRepo.create(dto);

    return {
      success: true,
      duplicated: false,
      newBalance: newQuantity,
      transaction
    };
  }

  async releaseStock(dto: CreateTransactionDTO): Promise<ReserveResult> {
    console.log(`🔍 [InventoryService] releaseStock called with:`, dto);
    // TODO: Implement compensation logic
    return {
      success: true,
      newBalance: await this.getBalance(dto.sku) || 0
    };
  }

  async getBalance(sku: string): Promise<number | null> {
    console.log(`🔍 [InventoryService] getBalance called with sku: ${sku}`);
    const stock = await this.productRepo.getStock(sku);
    return stock;
  }

  async hasSufficientStock(sku: string, quantity: number): Promise<boolean> {
    console.log(`🔍 [InventoryService] hasSufficientStock called with sku: ${sku}, quantity: ${quantity}`);
    const stock = await this.getBalance(sku);
    return stock !== null && stock >= quantity;
  }
}
