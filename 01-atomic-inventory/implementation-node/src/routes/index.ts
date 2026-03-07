import { FastifyInstance } from 'fastify';
import { inventoryRoutes } from '@/routes//inventory.routes';

export async function registerRoutes(fastify: FastifyInstance) {
  fastify.register(inventoryRoutes);
}
