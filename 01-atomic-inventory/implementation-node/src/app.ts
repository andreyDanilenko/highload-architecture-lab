import Fastify, { FastifyInstance } from 'fastify';
import { setupErrorHandler } from '@/plugins/error-handler';
import { healthPlugin } from '@/plugins/health';
import { registerRoutes } from '@/routes';

/**
 * Builds and configures the Fastify app (no listen).
 * Use for: server entry point, tests with fastify.inject().
 */
export function buildApp(): FastifyInstance {
  const fastify = Fastify({
    logger: true,
    trustProxy: true
  });

  setupErrorHandler(fastify);
  fastify.register(healthPlugin);
  fastify.register(registerRoutes);

  return fastify;
}
