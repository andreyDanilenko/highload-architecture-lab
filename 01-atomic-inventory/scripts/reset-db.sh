#!/bin/bash

echo "🔄 Resetting database to initial state..."

if ! docker info &>/dev/null; then
  echo "❌ Cannot connect to the Docker daemon. Is Docker running?"
  exit 1
fi

if ! docker exec -i inventory-postgres psql -U postgres -d inventory <<'EOF'
-- Reset product stock to 1000
UPDATE products 
SET stock_quantity = 1000, version = 0, updated_at = NOW() 
WHERE sku = 'SKU-TEST-001';

-- Clear all transactions
DELETE FROM inventory_transactions;

-- Show current state
SELECT '✅ Products:' as info, COUNT(*) FROM products;
SELECT '✅ SKU-TEST-001 stock:' as info, stock_quantity FROM products WHERE sku = 'SKU-TEST-001';
SELECT '✅ Transactions:' as info, COUNT(*) FROM inventory_transactions;
EOF
then
  echo ""
  echo "❌ Database reset failed (container 'inventory-postgres' may not be running)."
  if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -q '^inventory-postgres$'; then
    echo "   Tip: start infra from 01-atomic-inventory: make infra-up"
  fi
  exit 1
fi

echo ""
echo "✅ Database reset complete!"
echo "   Stock restored to 1000"
echo "   Transactions cleared"
