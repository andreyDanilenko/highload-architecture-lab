import { z } from 'zod';

export const reserveSchema = z.object({
  sku: z.string().min(1).max(50),
  quantity: z.number().int().positive().max(100),
  requestId: z.uuid()
});

export type ReserveRequest = z.infer<typeof reserveSchema>;

export const skuParamSchema = z.object({
  sku: z.string().min(1).max(50)
});

export type SkuParam = z.infer<typeof skuParamSchema>;
