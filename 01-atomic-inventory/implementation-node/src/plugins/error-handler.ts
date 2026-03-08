import { FastifyInstance, FastifyRequest } from "fastify";
import { z, ZodError } from "zod";
import { AppError } from "@/shared/errors/app-errors";

const isDev = process.env.NODE_ENV !== "production";

function withMeta<T extends Record<string, unknown>>(
	request: FastifyRequest,
	payload: T,
): T & { timestamp: string; requestId: string | undefined } {
	return {
		...payload,
		timestamp: new Date().toISOString(),
		requestId: request.id,
	};
}

function hasStatusCode(
	err: unknown,
): err is { statusCode: number; code?: string; message?: string } {
	return (
		err !== null &&
		typeof err === "object" &&
		"statusCode" in err &&
		typeof (err as { statusCode: unknown }).statusCode === "number"
	);
}

function buildErrorPayload(
	request: FastifyRequest,
	error: unknown,
): { status: number; body: Record<string, unknown> } {
	if (error instanceof ZodError) {
		return {
			status: 400,
			body: withMeta(request, {
				error: "Validation Error",
				code: "VALIDATION_FAILED",
				details: z.treeifyError(error),
			}),
		};
	}

	if (error instanceof AppError) {
		return {
			status: error.statusCode,
			body: withMeta(request, {
				code: error.code,
				error: error.message,
				...(error.details !== undefined && { details: error.details }),
			}),
		};
	}

	// Fastify errors (FST_ERR_*) and any error that already has statusCode
	if (hasStatusCode(error)) {
		const status =
			error.statusCode >= 400 && error.statusCode < 600
				? error.statusCode
				: 500;
		return {
			status,
			body: withMeta(request, {
				error: error.message ?? "Request failed",
				code: error.code ?? "UNKNOWN_ERROR",
			}),
		};
	}

	return {
		status: 500,
		body: withMeta(request, {
			error: "Internal Server Error",
			code: "INTERNAL_SERVER_ERROR",
		}),
	};
}

export function setupErrorHandler(fastify: FastifyInstance) {
	fastify.setErrorHandler((error, request, reply) => {
		request.log.error(
			{
				err: error,
				req: request.id,
				url: request.url,
				method: request.method,
				...(isDev && {
					body: request.body,
					params: request.params,
					query: request.query,
				}),
			},
			"Request failed",
		);

		const { status, body } = buildErrorPayload(request, error);
		return reply.status(status).send(body);
	});
}
