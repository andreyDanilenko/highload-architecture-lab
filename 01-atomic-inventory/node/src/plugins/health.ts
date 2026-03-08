import { FastifyInstance } from "fastify";
import { pgPool, redisClient } from "@/config";

export async function healthPlugin(fastify: FastifyInstance) {
	fastify.get("/health", async () => {
		const dbStatus = await pgPool
			.query("SELECT 1 as connected")
			.then(() => "ok" as const)
			.catch(() => "error" as const);

		const redisStatus = redisClient.isReady
			? ("ok" as const)
			: ("disconnected" as const);

		return {
			status: "ok" as const,
			timestamp: new Date().toISOString(),
			uptime: process.uptime(),
			databases: {
				postgres: dbStatus,
				redis: redisStatus,
			},
		};
	});
}
