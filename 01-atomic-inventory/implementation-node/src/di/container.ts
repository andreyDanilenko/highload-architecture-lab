import { ProductRepository } from '../repositories/product.repository';
import { TransactionRepository } from '../repositories/transaction.repository';
import { InventoryService } from '../services/inventory.service';
import { InventoryController } from '../controllers/inventory.controller';

export const productRepo = new ProductRepository();
export const transactionRepo = new TransactionRepository();

export const inventoryService = new InventoryService(productRepo, transactionRepo);

export const inventoryController = new InventoryController(inventoryService);

export const container = {
  // Repositories
  productRepo,
  transactionRepo,
  
  // Services
  inventoryService,
  
  // Controllers
  inventoryController
};
