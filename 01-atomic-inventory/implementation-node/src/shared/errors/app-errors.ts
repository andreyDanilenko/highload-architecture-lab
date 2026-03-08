export class AppError extends Error {
    constructor(
      public statusCode: number,
      public message: string,
      public code: string,
      public details?: unknown
    ) {
      super(message);
      this.name = 'AppError';
      Error.captureStackTrace(this, this.constructor);
    }
  }
  
  export class BusinessError extends AppError {
    constructor(message: string, code: string, details?: unknown) {
      super(400, message, code, details);
    }
  }
  
  export class NotFoundError extends AppError {
    constructor(resource: string, identifier?: string | Record<string, unknown>) {
      const details = typeof identifier === 'string' 
        ? { id: identifier } 
        : identifier;
      
      super(
        404, 
        `${resource} not found`, 
        'RESOURCE_NOT_FOUND',
        details
      );
    }
  }
  
  export class InsufficientStockError extends AppError {
    constructor(sku: string, requested: number, available?: number) {
      super(
        409,
        `Insufficient stock for SKU: ${sku}`,
        'INSUFFICIENT_STOCK',
        { sku, requested, available }
      );
    }
  }
  
  export class DuplicateRequestError extends AppError {
    constructor(requestId: string, sku: string) {
      super(
        409,
        `Duplicate reservation request: ${requestId}`,
        'DUPLICATE_REQUEST',
        { requestId, sku }
      );
    }
  }
  
  export class ValidationError extends AppError {
    constructor(message: string, details?: unknown) {
      super(400, message, 'VALIDATION_ERROR', details);
    }
  }
