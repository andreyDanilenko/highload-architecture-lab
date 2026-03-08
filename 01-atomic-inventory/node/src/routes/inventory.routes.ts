import { FastifyInstance } from "fastify";
import { inventoryController } from "@/di/container";

export async function inventoryRoutes(fastify: FastifyInstance) {
	// GET /inventory/stock/:sku
	fastify.get<{ Params: { sku: string } }>("/stock/:sku", async (request) =>
		inventoryController.getStock(request),
	);

	// POST /inventory/reserve — naive (no locking). Demo/load-test only. Do not use in production.
	fastify.post<{ Body: { sku: string; quantity: number; requestId: string } }>(
		"/reserve",
		async (request) => inventoryController.reserve(request),
	);

	// POST /inventory/reserve/pessimistic
	fastify.post<{ Body: { sku: string; quantity: number; requestId: string } }>(
		"/reserve/pessimistic",
		async (request) => inventoryController.reservePessimistic(request),
	);

	// POST /inventory/reserve/optimistic
	fastify.post<{ Body: { sku: string; quantity: number; requestId: string } }>(
		"/reserve/optimistic",
		async (request) => inventoryController.reserveOptimistic(request),
	);

	// POST /inventory/reserve/redis
	fastify.post<{ Body: { sku: string; quantity: number; requestId: string } }>(
		"/reserve/redis",
		async (request) => inventoryController.reserveRedis(request),
	);
}
