import type {
	FastifyInstance,
	FastifyPluginAsync,
	FastifyTypeProviderDefault,
	RawReplyDefaultExpression,
	RawRequestDefaultExpression,
	RawServerDefault,
} from "fastify";
import type pino from "pino";

/**
 * Fastify instance type used by this app (custom Pino logger).
 * Use this in buildApp, plugins, and route registrations so types stay consistent.
 */
export type AppFastifyInstance = FastifyInstance<
	RawServerDefault,
	RawRequestDefaultExpression<RawServerDefault>,
	RawReplyDefaultExpression<RawServerDefault>,
	pino.Logger,
	FastifyTypeProviderDefault
>;

/**
 * Plugin type that accepts our app instance (pino.Logger).
 * Use this for plugins so register() accepts them without type casts.
 */
export type AppPluginAsync = FastifyPluginAsync<
	Record<never, never>,
	RawServerDefault,
	FastifyTypeProviderDefault,
	pino.Logger
>;
