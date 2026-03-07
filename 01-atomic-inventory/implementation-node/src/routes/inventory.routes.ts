import { FastifyInstance } from 'fastify';
import { inventoryController } from '../di/container';

export async function inventoryRoutes(fastify: FastifyInstance) {
  // GET /inventory/stock/:sku
  fastify.get('/inventory/stock/:sku', (request, reply) => 
    inventoryController.getStock(request, reply)
  );
  
  // POST /inventory/reserve
  fastify.post('/inventory/reserve', (request, reply) => 
    inventoryController.reserve(request, reply)
  );
}
