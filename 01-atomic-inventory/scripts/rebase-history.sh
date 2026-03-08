#!/usr/bin/env bash
#
# Rebase 01-atomic-inventory branch into a structured commit history.
# Squashes related commits into logical groups (see comments below).
#
# Usage:
#   From repo root: ./01-atomic-inventory/scripts/rebase-history.sh
#   Or from 01-atomic-inventory: ./scripts/rebase-history.sh
#
# Prerequisites:
#   - Clean working tree (commit or stash changes first).
#   - Branch 01-atomic-inventory (or adjust BASE/UPTO below).
#
# After rebase:
#   git push --force-with-lease origin 01-atomic-inventory
#
# If you had uncommitted changes (e.g. logger + AppFastifyInstance): stash before
# running, then after rebase run git stash pop and commit as a separate commit.

set -e

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if ! git diff-index --quiet HEAD -- 2>/dev/null; then
  echo "Working tree has uncommitted changes. Commit or stash them first."
  exit 1
fi

# Last commit before lab work (README cleanup)
BASE="${BASE:-c0a22ec}"
# Optional: end at a specific commit (default: HEAD)
UPTO="${UPTO:-HEAD}"

echo "Base commit: $BASE"
echo "Rebasing $(git rev-parse $UPTO) onto $BASE"
echo ""

# Build sequence: oldest (first to apply) to newest.
# pick = keep this commit (and its message); fixup = merge into previous pick.
REBASE_TODO=$(mktemp)
trap 'rm -f "$REBASE_TODO"' EXIT

cat > "$REBASE_TODO" << 'SEQUENCE'
# 1) Task + Node skeleton (init, base server, DB + health)
pick df8960c init: implement repository
fixup 1bff6d5 feat: add base server
fixup f081e5f feat: initial databases and health check

# 2) Contracts, DI, routes, error handling, app factory
pick dfc1ce3 feat: add repository and service contracts
fixup 5c9b310 feat: complete project setup with contracts, DI, and routes
fixup 6c47a96 feat: update path for alias
fixup 33018a0 feat(inventory): add centralized error handling and 404 for missing product
fixup aafccc4 refactor(server): extract app factory and health plugin; add /api/v1/inventory prefix

# 3) Biome and formatting
pick de2e8d1 chore(node): add Biome, remove errors.contracts, apply formatting

# 4) Postgres ProductRepository
pick 264d8d2 feat(inventory): implement ProductRepository with Postgres
fixup 6413af2 WIP: update format

# 5) TransactionRepository + naive reserve + test scripts
pick 733b781 feat(inventory): implement TransactionRepository with Postgres
fixup a54f972 chore(01-atomic-inventory): add test scripts and DB reset, fix TransactionRepository DI

# 6) Pessimistic locking
pick eee6548 feat(inventory): add pessimistic locking and load-test scripts

# 7) Optimistic locking
pick 3c16979 chore(inventory): deprecate naive flow, parallel load test
fixup 0bbab22 feat(atomic-inventory): README path to best practices + optimistic locking
fixup d5cfd7c refactor(inventory): config-based optimistic retries and service cleanup

# 8) Redis atomic reserve
pick fe6a0e8 feat(01-atomic-inventory): add Redis atomic counter reserve strategy

# 9) Load-test scripts and subtask docs
pick 1c1eaab chore(01-atomic-inventory): load-test scripts and docs
fixup feed0b2 docs(01-atomic-inventory): concise README and step-by-step subtask guides

# 10) Logging and README cleanup
pick 5e87f48 chore: app logging config and root README cleanup

# 11) Shared infra, node/ and go/
pick e0a1ea2 refactor(01-atomic-inventory): shared infra at task root, node/ and go/ impls
SEQUENCE

# Overwrite git's todo with our sequence (editor is called with path to todo as $1)
export GIT_SEQUENCE_EDITOR="cp '$REBASE_TODO' \"\$1\""
git rebase -i "$BASE"

echo ""
echo "Rebase finished. Push with: git push --force-with-lease origin $(git branch --show-current)"
