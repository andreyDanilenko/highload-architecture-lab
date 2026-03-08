import { mkdirSync } from "node:fs";
import { join } from "node:path";
import pino from "pino";

export const LOGS_DIR = join(process.cwd(), "logs");

export function createRunLogger(): pino.Logger {
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

/**
 * Config for Fastify logger: either default (logger: true) or custom Pino instance (file + stdout).
 * Use file only when not in test and LOG_TO_FILE is not "0".
 */
export function getLoggerConfig():
	| { logger: true }
	| { loggerInstance: pino.Logger } {
	const logToFile =
		process.env.NODE_ENV !== "test" && process.env.LOG_TO_FILE !== "0";
	return logToFile ? { loggerInstance: createRunLogger() } : { logger: true };
}
