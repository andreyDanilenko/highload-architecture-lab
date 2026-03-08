/**
 * Root DI container. Composed from module containers.
 * Safe to use after connectRedis() and pg pool are ready (see server.ts start()).
 */
import {
	inventoryContainer,
	productRepo,
	transactionRepo,
	inventoryService,
	inventoryController,
} from "./inventory.container";

export { productRepo, transactionRepo, inventoryService, inventoryController };

export const container = {
	...inventoryContainer,
};
