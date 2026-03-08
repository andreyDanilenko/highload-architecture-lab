import "./load-env";
import { pgPool, redisClient, connectRedis } from "@/config";
import { buildApp } from "@/app";

const fastify = buildApp();

const PORT = parseInt(process.env.PORT || "3000");
const HOST = process.env.HOST || "0.0.0.0";

async function start() {
	try {
		await connectRedis();
		console.log("📦 Redis connected");

		await pgPool.query("SELECT 1");
		console.log("📦 PostgreSQL connected");

		await fastify.listen({ port: PORT, host: HOST });
		console.log(`🚀 Server running on http://${HOST}:${PORT}`);
		console.log(`📊 Health check: http://${HOST}:${PORT}/health`);
	} catch (err) {
		console.error("❌ Failed to start server:", err);
		process.exit(1);
	}
}

async function gracefulShutdown(signal: string) {
	console.log(`\n🛑 Received ${signal}, shutting down...`);
	await fastify.close();
	await pgPool.end();
	if (redisClient.isReady) await redisClient.quit();
	process.exit(0);
}

process.on("SIGINT", () => gracefulShutdown("SIGINT"));
process.on("SIGTERM", () => gracefulShutdown("SIGTERM"));

start();
