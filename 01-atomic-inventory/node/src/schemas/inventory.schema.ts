import { z } from "zod";

export const reserveSchema = z.object({
	sku: z.string().min(1).max(50),
	quantity: z.number().int().positive().max(100),
	requestId: z.uuid(),
});

export type ReserveRequest = z.infer<typeof reserveSchema>;

export const skuParamSchema = z.object({
	sku: z.string().min(1).max(50),
});

export const reserveResponseSchema = z.object({
	success: z.boolean(),
	duplicated: z.boolean().optional(),
	newBalance: z.number().int().optional(),
});

export const stockResponseSchema = z.object({
	sku: z.string(),
	stock: z.number().int(),
	timestamp: z.string(),
});

export const reserveSchemaJSON = {
	body: reserveSchema,
	response: {
		200: reserveResponseSchema,
	},
};

export const stockSchemaJSON = {
	params: skuParamSchema,
	response: {
		200: stockResponseSchema,
	},
};
