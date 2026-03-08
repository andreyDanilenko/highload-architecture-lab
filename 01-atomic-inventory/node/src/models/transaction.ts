type TransactionBase = {
	sku: string;
	quantity: number;
	requestId: string;
};

export interface InventoryTransaction extends TransactionBase {
	id: number;
	createdAt: Date;
}

export type CreateTransactionDTO = TransactionBase;
