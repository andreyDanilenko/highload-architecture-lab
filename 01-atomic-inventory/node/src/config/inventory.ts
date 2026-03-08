import dotenv from "dotenv";
import { z } from "zod";

dotenv.config();

const maxRetriesSchema = z.coerce.number().int().min(1).max(100);

const raw = process.env.INVENTORY_MAX_OPTIMISTIC_RETRIES ?? 10;
const parsed = maxRetriesSchema.safeParse(raw);
const maxOptimisticRetries = parsed.success ? parsed.data : 10;

export const inventoryConfig = {
	maxOptimisticRetries,
} as const;

export type InventoryConfig = typeof inventoryConfig;
