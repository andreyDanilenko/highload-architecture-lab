// Product errors
export class ProductNotFoundError extends Error {
    constructor(public readonly sku: string) {
      super(`Product not found: ${sku}`);
      this.name = 'ProductNotFoundError';
    }
  }
  
  export class DuplicateSkuError extends Error {
    constructor(public readonly sku: string) {
      super(`SKU already exists: ${sku}`);
      this.name = 'DuplicateSkuError';
    }
  }
  
  // Stock errors
  export class InsufficientStockError extends Error {
    constructor(
      public readonly sku: string, 
      public readonly available: number,
      public readonly requested: number
    ) {
      super(`Insufficient stock for SKU ${sku}: available ${available}, requested ${requested}`);
      this.name = 'InsufficientStockError';
    }
  }
  
  // Concurrency errors
  export class VersionConflictError extends Error {
    constructor(
      public readonly sku: string, 
      public readonly currentVersion: number,
      public readonly expectedVersion: number
    ) {
      super(`Version conflict for SKU ${sku}: current ${currentVersion}, expected ${expectedVersion}`);
      this.name = 'VersionConflictError';
    }
  }
  
  // Idempotency errors
  export class DuplicateRequestError extends Error {
    constructor(
      public readonly requestId: string,
      public readonly existingTransaction: any
    ) {
      super(`Duplicate request: ${requestId}`);
      this.name = 'DuplicateRequestError';
    }
  }
  
  export class RequestPayloadMismatchError extends Error {
    constructor(
      public readonly requestId: string,
      public readonly expected: any,
      public readonly received: any
    ) {
      super(`Request payload mismatch for ${requestId}`);
      this.name = 'RequestPayloadMismatchError';
    }
  }
  
  // Database errors
  export class DatabaseError extends Error {
    constructor(
      public readonly operation: string,
      public readonly cause: any
    ) {
      super(`Database error during ${operation}: ${cause.message}`);
      this.name = 'DatabaseError';
    }
  }
  
  // Validation errors
  export class ValidationError extends Error {
    constructor(
      public readonly field: string,
      public readonly message: string
    ) {
      super(`Validation error on ${field}: ${message}`);
      this.name = 'ValidationError';
    }
  }
