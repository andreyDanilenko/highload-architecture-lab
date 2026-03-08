import { ProductRepository } from "@/repositories/product.repository";
import { TransactionRepository } from "@/repositories/transaction.repository";
import { InventoryService } from "@/services/inventory.service";
import { InventoryController } from "@/controllers/inventory.controller";
import { pgPool, inventoryConfig } from "@/config";

export const productRepo = new ProductRepository(pgPool);
export const transactionRepo = new TransactionRepository(pgPool);

export const inventoryService = new InventoryService(
	productRepo,
	transactionRepo,
	pgPool,
	inventoryConfig,
);

export const inventoryController = new InventoryController(inventoryService);

export const container = {
	// Repositories
	productRepo,
	transactionRepo,

	// Services
	inventoryService,

	// Controllers
	inventoryController,
};
