# Testing Scripts for Atomic Inventory Counter

## Available Scripts

### `race-test.sh`
Tests for **race conditions** by sending 100 concurrent reservation requests.

```bash
./scripts/load-test/race-test.sh
```

### `reset-db.sh`
Resets database to initial state (1000 stock, clears transactions).

```bash
./scripts/reset-db.sh
```

## What We're Testing

| Script | Purpose | Expected Result |
|--------|---------|-----------------|
| `race-test.sh` | Demonstrate race condition | Stock < 900 (lost updates) |
| `reset-db.sh` | Clean slate for next test | Stock back to 1000 |

## Quick Start

```bash
# 1. Reset database
./scripts/reset-db.sh

# 2. Run race condition test
./scripts/load-test/race-test.sh

# 3. See the problem (stock should be less than 900)
```

## Why These Tests

- **race-test.sh** - proves the naive implementation fails under concurrency
- **reset-db.sh** - ensures consistent starting point for each test run
