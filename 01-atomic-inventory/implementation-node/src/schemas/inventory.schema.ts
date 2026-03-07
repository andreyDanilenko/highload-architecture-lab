import { z } from 'zod';

export const reserveSchema = z.object({
  sku: z.string().min(1).max(50),
  quantity: z.number().int().positive().max(100),
  requestId: z.string().uuid()
});

export type ReserveRequest = z.infer<typeof reserveSchema>;

export const skuParamSchema = z.object({
  sku: z.string().min(1).max(50)
});

export const reserveResponseSchema = z.object({
  success: z.boolean(),
  duplicated: z.boolean().optional(),
  newBalance: z.number().int().optional(),
  error: z.string().optional()
});

export type ReserveResponse = z.infer<typeof reserveResponseSchema>;
