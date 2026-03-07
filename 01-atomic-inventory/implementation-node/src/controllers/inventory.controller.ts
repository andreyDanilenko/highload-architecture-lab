import { FastifyRequest, FastifyReply } from 'fastify';
import { IInventoryService } from '@/contracts/inventory-service.contracts';
import { reserveSchema, skuParamSchema, ReserveRequest } from '@/schemas/inventory.schema';

export class InventoryController {
  constructor(private inventoryService: IInventoryService) {}

  async getStock(
    request: FastifyRequest<{ Params: { sku: string } }>,
    reply: FastifyReply
  ) {
    try {
      console.log(`📥 GET /stock called with params:`, request.params);
      
      const { sku } = skuParamSchema.parse({ sku: request.params.sku });
      
      const stock = await this.inventoryService.getBalance(sku);
      
      if (stock === null) {
        return reply.status(404).send({ 
          error: 'Product not found',
          sku 
        });
      }
      
      return reply.status(200).send({ 
        sku,
        stock,
        timestamp: new Date().toISOString()
      });
      
    } catch (error) {
      request.log.error(error);
      return reply.status(400).send({ 
        error: 'Invalid request',
        details: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  }

  async reserve(
    request: FastifyRequest<{ Body: ReserveRequest }>,
    reply: FastifyReply
  ) {
    try {
      console.log(`📥 POST /reserve called with body:`, request.body);
      
      const body = reserveSchema.parse(request.body);
      
      const result = await this.inventoryService.reserveStock({
        sku: body.sku,
        quantity: body.quantity,
        requestId: body.requestId
      });

      if (result.success) {
        return reply.status(200).send({
          success: true,
          duplicated: result.duplicated,
          newBalance: result.newBalance
        });
      } else {
        const status = result.error?.includes('not found') ? 404 :
                      result.error?.includes('Insufficient') ? 409 : 400;
        
        return reply.status(status).send({
          success: false,
          error: result.error,
          duplicated: result.duplicated
        });
      }

    } catch (error) {
      request.log.error(error);
      return reply.status(400).send({
        success: false,
        error: error instanceof Error ? error.message : 'Validation failed'
      });
    }
  }
}
