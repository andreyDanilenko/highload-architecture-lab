import { createClient } from "redis";
import dotenv from "dotenv";

dotenv.config();

// Определяем правильный хост в зависимости от окружения
const redisHost = process.env.REDIS_HOST || "localhost";
const redisPort = process.env.REDIS_PORT || "6379";
const redisUrl = process.env.REDIS_URL || `redis://${redisHost}:${redisPort}`;

export const redisClient = createClient({
	url: redisUrl,
});

redisClient.on("connect", () => {
	console.log("✅ Redis connected to", redisUrl);
});

redisClient.on("error", (err) => {
	console.error("❌ Redis error:", err.message);
});

redisClient.on("end", () => {
	console.log("📦 Redis disconnected");
});

export const connectRedis = async () => {
	if (!redisClient.isOpen) {
		await redisClient.connect();
	}
	return redisClient;
};
