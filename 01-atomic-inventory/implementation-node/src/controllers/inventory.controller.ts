import { FastifyRequest } from "fastify";
import { IInventoryService } from "@/contracts/inventory-service.contracts";
import {
	reserveSchema,
	skuParamSchema,
	ReserveRequest,
} from "@/schemas/inventory.schema";

export class InventoryController {
	constructor(private inventoryService: IInventoryService) {}

	async getStock(request: FastifyRequest<{ Params: { sku: string } }>) {
		const { sku } = skuParamSchema.parse({ sku: request.params.sku });
		request.log.debug({ sku }, "Getting stock");

		const stock = await this.inventoryService.getBalance(sku);

		return {
			sku,
			stock,
			timestamp: new Date().toISOString(),
		};
	}

	/** Naive reserve endpoint. Demo/load-test only — race condition under concurrency. */
	async reserve(request: FastifyRequest<{ Body: ReserveRequest }>) {
		const body = reserveSchema.parse(request.body);
		request.log.debug({ body }, "Processing reservation (naive)");

		const result = await this.inventoryService.reserveStock({
			sku: body.sku,
			quantity: body.quantity,
			requestId: body.requestId,
		});

		return {
			success: true,
			duplicated: result.duplicated,
			newBalance: result.newBalance,
		};
	}

	async reservePessimistic(request: FastifyRequest<{ Body: ReserveRequest }>) {
		const body = reserveSchema.parse(request.body);
		request.log.debug({ body }, "Processing reservation (pessimistic)");

		const result = await this.inventoryService.reserveStockPessimistic({
			sku: body.sku,
			quantity: body.quantity,
			requestId: body.requestId,
		});

		return {
			success: true,
			duplicated: result.duplicated,
			newBalance: result.newBalance,
		};
	}
}
