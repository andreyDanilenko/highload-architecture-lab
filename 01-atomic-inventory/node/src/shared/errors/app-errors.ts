export class AppError extends Error {
	constructor(
		public statusCode: number,
		public message: string,
		public code: string,
		public details?: unknown,
	) {
		super(message);
		this.name = "AppError";
		Error.captureStackTrace(this, this.constructor);
	}
}

// 400 - Bad Request
export class ValidationError extends AppError {
	constructor(message: string, details?: unknown) {
		super(400, message, "VALIDATION_ERROR", details);
	}
}

export class BusinessError extends AppError {
	constructor(message: string, code: string, details?: unknown) {
		super(400, message, code, details);
	}
}

// 404 - Not Found
export class NotFoundError extends AppError {
	constructor(resource: string, identifier?: string | Record<string, unknown>) {
		const details =
			typeof identifier === "string" ? { id: identifier } : identifier;

		super(404, `${resource} not found`, "RESOURCE_NOT_FOUND", details);
	}
}

// 409 - Conflict
export class InsufficientStockError extends AppError {
	constructor(sku: string, requested: number, available?: number) {
		super(409, `Insufficient stock for SKU: ${sku}`, "INSUFFICIENT_STOCK", {
			sku,
			requested,
			available,
		});
	}
}

export class DuplicateRequestError extends AppError {
	constructor(requestId: string, sku: string) {
		super(
			409,
			`Duplicate reservation request: ${requestId}`,
			"DUPLICATE_REQUEST",
			{ requestId, sku },
		);
	}
}

export class VersionConflictError extends AppError {
	constructor(sku: string, currentVersion: number, expectedVersion: number) {
		super(409, `Version conflict for SKU ${sku}`, "VERSION_CONFLICT", {
			sku,
			currentVersion,
			expectedVersion,
		});
	}
}

// 500 - Server Error
export class DatabaseError extends AppError {
	constructor(operation: string, error: unknown) {
		super(500, `Database error during ${operation}`, "DATABASE_ERROR", {
			operation,
			error: error instanceof Error ? error.message : String(error),
		});
	}
}
