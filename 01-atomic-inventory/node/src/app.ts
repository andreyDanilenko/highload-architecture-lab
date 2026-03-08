import Fastify from "fastify";
import { getLoggerConfig } from "@/logger";
import { setupErrorHandler } from "@/plugins/error-handler";
import { healthPlugin } from "@/plugins/health";
import { registerRoutes } from "@/routes";
import type { AppFastifyInstance } from "@/types/app-instance";

/**
 * Builds and configures the Fastify app (no listen).
 * Use for: server entry point, tests with fastify.inject().
 */
export function buildApp(): AppFastifyInstance {
	const fastify = Fastify({
		...getLoggerConfig(),
		trustProxy: true,
	}) as AppFastifyInstance;

	setupErrorHandler(fastify);
	fastify.register(healthPlugin);
	fastify.register(registerRoutes);

	return fastify;
}
