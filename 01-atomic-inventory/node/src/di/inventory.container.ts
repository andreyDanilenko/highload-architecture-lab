import { ProductRepository } from "@/repositories/product.repository";
import { RedisStockRepository } from "@/repositories/redis-stock.repository";
import { TransactionRepository } from "@/repositories/transaction.repository";
import { InventoryService } from "@/services/inventory.service";
import { InventoryController } from "@/controllers/inventory.controller";
import { pgPool, inventoryConfig, redisClient } from "@/config";

const productRepo = new ProductRepository(pgPool);
const transactionRepo = new TransactionRepository(pgPool);
const redisStockRepo = new RedisStockRepository(redisClient);

const inventoryService = new InventoryService(
	productRepo,
	transactionRepo,
	pgPool,
	inventoryConfig,
	redisStockRepo,
);

const inventoryController = new InventoryController(inventoryService);

export const inventoryContainer = {
	productRepo,
	transactionRepo,
	redisStockRepo,
	inventoryService,
	inventoryController,
};

export {
	productRepo,
	transactionRepo,
	redisStockRepo,
	inventoryService,
	inventoryController,
};
