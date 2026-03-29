## Naive rate-limit load test

This directory contains a simple shell script to exercise the naive in-memory
rate limiter of the Anti‑Bruteforce service over HTTP.

### Prerequisites

- Anti‑Bruteforce server is running (by default on `http://localhost:3000`), e.g.:

```bash
cd 02-anti-bruteforce/go
make -f ../Makefile dev
```

### Script: `race-test.sh`

Sends multiple parallel `POST` requests to a chosen endpoint and reports how
many returned **200 OK** vs **429 Too Many Requests**.

**Usage:**

```bash
cd 02-anti-bruteforce/scripts/load-test
chmod +x race-test.sh

# Default: POST http://localhost:3000/login with 20 requests
./race-test.sh

# Custom base URL and path
./race-test.sh http://localhost:3000 /login 50
./race-test.sh http://localhost:3000 /resource/naive 50
```
TOTAL_REQUESTS=${3:-20}
**Output:**

- Total number of completed requests.
- Count of 200/429/other status codes, so you can visually confirm that the
  naive limiter starts returning 429 after the configured threshold.

for i in {1..2000}; do
  echo -n "$i -> "
  curl -s -o /dev/null -w "%{http_code}\n" -X POST "http://localhost:3000/resource/naive" \
    -H "X-Forwarded-For: 1.2.3.4"
done


for i in {1..20}; do
  echo -n "$i -> "
  curl -s -o /dev/null -w "%{http_code}\n" -X POST "http://localhost:3000/vault/login" \
    -H "X-Forwarded-For: 1.2.3.4"
done

for i in {1..2000}; do
  echo -n "$i -> "
  curl -s -o /dev/null -w "%{http_code}\n" -X POST "http://localhost:3000/resource/pessimistic" \
    -H "X-Forwarded-For: 1.2.3.4"
done

for i in {1..20}; do
  echo -n "$i -> "
  curl -s -o /dev/null -w "%{http_code}\n" -X POST "http://localhost:3000/resource/optimistic" \
    -H "X-Forwarded-For: 1.2.3.4"
done

for i in {1..20}; do
  echo -n "$i -> "
  curl -s -o /dev/null -w "%{http_code}\n" -X POST "http://localhost:3000/login" \
    -H "X-Forwarded-For: 1.2.3.4"
done
