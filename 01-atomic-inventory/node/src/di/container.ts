import { ProductRepository } from "@/repositories/product.repository";
import {
	RedisStockRepository,
	type RedisClientLike,
} from "@/repositories/redis-stock.repository";
import { TransactionRepository } from "@/repositories/transaction.repository";
import { InventoryService } from "@/services/inventory.service";
import { InventoryController } from "@/controllers/inventory.controller";
import { pgPool, inventoryConfig, redisClient } from "@/config";

export const productRepo = new ProductRepository(pgPool);
export const transactionRepo = new TransactionRepository(pgPool);
// redisClient has get/set; eval() for Lua is available at runtime (node-redis EVAL)
export const redisStockRepo = new RedisStockRepository(
	redisClient as unknown as RedisClientLike,
);

export const inventoryService = new InventoryService(
	productRepo,
	transactionRepo,
	pgPool,
	inventoryConfig,
	redisStockRepo,
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
