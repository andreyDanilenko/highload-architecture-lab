# 1. Naive Non-Idempotent Endpoint

**What:** Simple HTTP handler for a side-effectful operation (e.g. payment processing) without any idempotency guarantees.

**Why:** To demonstrate the problem: repeated requests lead to repeated side effects (double charge, duplicate order, etc.).

---

## Implementation steps

1. **Define a basic request/response**
   - `PaymentRequest` with fields like `UserID`, `Amount`.
   - Stub `paymentService.Charge(userID, amount)` that simulates external billing.

2. **Naive handler**
   - Parse JSON request.
   - Call `paymentService.Charge`.
   - On success, return `200 OK` with payment info.
   - On error, `500 Internal Server Error`.

3. **Demonstrate the issue**
   - Write a simple test or script:
     - Send the **same** payment request multiple times (e.g. due to simulated timeout).
     - Show that `Charge` is called multiple times (e.g. by logging or counting calls).
   - Document scenarios:
     - User double-clicks "Pay" button.
     - HTTP client retries after network timeout.

---

## What will be done

- Implement a non-idempotent `/payments` (or similar) endpoint.
- Show via logs/tests how repeated requests repeat the side effect.
- Use this as a baseline to justify adding an idempotency layer.

