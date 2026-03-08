import { Pool } from "pg";
import { Product, CreateProductDTO } from "@/models/product";
import { IProductRepository } from "@/contracts/product-repository.contracts";
import { DatabaseError, BusinessError } from "@/shared/errors/app-errors";

interface ProductRow {
	id: number;
	sku: string;
	name: string;
	stock_quantity: number;
	version: number;
	created_at: Date;
	updated_at: Date;
}

/**
 * Repository for products table.
 * Handles product CRUD with optimistic locking support.
 */
export class ProductRepository implements IProductRepository {
	constructor(private readonly pool: Pool) {}

	async findBySku(sku: string): Promise<Product | null> {
		try {
			const { rows } = await this.pool.query<ProductRow>(
				"SELECT * FROM products WHERE sku = $1",
				[sku],
			);

			return rows.length === 0 ? null : this.mapToEntity(rows[0]);
		} catch (error) {
			throw new DatabaseError("ProductRepository.findBySku", {
				cause: error,
				sku,
			});
		}
	}

	async getStock(sku: string): Promise<number | null> {
		try {
			const { rows } = await this.pool.query<{ stock_quantity: number }>(
				"SELECT stock_quantity FROM products WHERE sku = $1",
				[sku],
			);

			return rows[0]?.stock_quantity ?? null;
		} catch (error) {
			throw new DatabaseError("ProductRepository.getStock", {
				cause: error,
				sku,
			});
		}
	}

	async updateStockNaive(sku: string, newQuantity: number): Promise<boolean> {
		try {
			const { rowCount } = await this.pool.query(
				`UPDATE products 
         SET stock_quantity = $1, updated_at = NOW() 
         WHERE sku = $2`,
				[newQuantity, sku],
			);

			return (rowCount ?? 0) > 0;
		} catch (error) {
			throw new DatabaseError("ProductRepository.updateStockNaive", {
				cause: error,
				sku,
				newQuantity,
			});
		}
	}

	/**
	 * Optimistic locking update. Returns false if version mismatch.
	 */
	async updateStock(
		sku: string,
		newQuantity: number,
		version: number,
	): Promise<boolean> {
		try {
			const { rowCount } = await this.pool.query(
				`UPDATE products 
         SET stock_quantity = $1, 
             version = version + 1, 
             updated_at = NOW() 
         WHERE sku = $2 AND version = $3`,
				[newQuantity, sku, version],
			);

			return (rowCount ?? 0) > 0;
		} catch (error) {
			throw new DatabaseError("ProductRepository.updateStock", {
				cause: error,
				sku,
				newQuantity,
				version,
			});
		}
	}

	async exists(sku: string): Promise<boolean> {
		try {
			const { rows } = await this.pool.query(
				"SELECT 1 FROM products WHERE sku = $1",
				[sku],
			);
			return rows.length > 0;
		} catch (error) {
			throw new DatabaseError("ProductRepository.exists", {
				cause: error,
				sku,
			});
		}
	}

	async create(product: CreateProductDTO): Promise<Product> {
		try {
			const exists = await this.exists(product.sku);
			if (exists) {
				throw new BusinessError(
					`SKU ${product.sku} already exists`,
					"DUPLICATE_SKU",
					{ sku: product.sku },
				);
			}

			const { rows } = await this.pool.query<ProductRow>(
				`INSERT INTO products (sku, name, stock_quantity) 
         VALUES ($1, $2, $3) 
         RETURNING *`,
				[product.sku, product.name, product.stockQuantity],
			);

			return this.mapToEntity(rows[0]);
		} catch (error) {
			if (error instanceof BusinessError) throw error;
			throw new DatabaseError("ProductRepository.create", {
				cause: error,
				product,
			});
		}
	}

	private mapToEntity(row: ProductRow): Product {
		return {
			id: row.id,
			sku: row.sku,
			name: row.name,
			stockQuantity: row.stock_quantity,
			version: row.version,
			createdAt: row.created_at,
			updatedAt: row.updated_at,
		};
	}
}
