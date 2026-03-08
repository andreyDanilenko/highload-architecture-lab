export function isPostgresError(
	error: unknown,
): error is { code: string } & Record<string, unknown> {
	return (
		error !== null &&
		typeof error === "object" &&
		"code" in error &&
		typeof (error as any).code === "string"
	);
}
