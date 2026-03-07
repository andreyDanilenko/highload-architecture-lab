type ProductFields = {
    sku: string;
    name: string;
    stockQuantity: number;
};

export interface Product extends ProductFields {
    id: number;
    version: number;
    createdAt: Date;
    updatedAt: Date;
}

export type CreateProductDTO = ProductFields;
export type UpdateStockDTO = Pick<ProductFields, 'sku'> & {
    quantity: number;
    requestId: string;
};
