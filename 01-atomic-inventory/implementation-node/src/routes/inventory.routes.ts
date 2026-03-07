import { FastifyInstance } from 'fastify';
import { inventoryController } from '@/di/container';

export async function inventoryRoutes(fastify: FastifyInstance) {
  // GET /inventory/stock/:sku
  fastify.get<{ Params: { sku: string } }>(
    '/inventory/stock/:sku',
    async (request, reply) => inventoryController.getStock(request, reply)
  );
  
  // POST /inventory/reserve
  fastify.post<{ Body: { sku: string; quantity: number; requestId: string } }>(
    '/inventory/reserve',
    async (request, reply) => inventoryController.reserve(request, reply)
  );
}
