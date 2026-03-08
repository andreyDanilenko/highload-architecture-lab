import Fastify, { FastifyInstance } from "fastify";
import { mkdirSync } from "node:fs";
import { join } from "node:path";
import pino from "pino";
import { setupErrorHandler } from "@/plugins/error-handler";
import { healthPlugin } from "@/plugins/health";
import { registerRoutes } from "@/routes";

const LOGS_DIR = join(process.cwd(), "logs");

function createRunLogger(): pino.Logger {
	mkdirSync(LOGS_DIR, { recursive: true });
	const runId = new Date().toISOString().replace(/:/g, "-");
	const logPath = join(LOGS_DIR, `run-${runId}.log`);
	const fileStream = pino.destination({
		dest: logPath,
		append: true,
		mkdir: false,
	});
	const logger = pino(
		{ level: process.env.LOG_LEVEL || "info" },
		pino.multistream([{ stream: process.stdout }, { stream: fileStream }]),
	);
	logger.info({ logFile: logPath }, "Logging to file");
	return logger;
}

/** Use file logging only when not in test and LOG_TO_FILE is not "0". */
function getLoggerConfig():
	| { logger: true }
	| { loggerInstance: pino.Logger } {
	const logToFile =
		process.env.NODE_ENV !== "test" && process.env.LOG_TO_FILE !== "0";
	return logToFile
		? { loggerInstance: createRunLogger() }
		: { logger: true };
}

/**
 * Builds and configures the Fastify app (no listen).
 * Use for: server entry point, tests with fastify.inject().
 */
export function buildApp(): FastifyInstance {
	const fastify = Fastify({
		...getLoggerConfig(),
		trustProxy: true,
	});

	setupErrorHandler(fastify);
	fastify.register(healthPlugin);
	fastify.register(registerRoutes);

	return fastify;
}
