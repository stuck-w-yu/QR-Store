# Plan — Cashier Shift Management

## 1. Overview

Dokumen ini menjelaskan rencana implementasi fitur **Cashier Shift Management** pada aplikasi Self-Order Restaurant & Cashier System.

Fitur ini memungkinkan restoran untuk mengelola:

* Buka kasir.
* Shift kasir.
* Pergantian kasir.
* Monitoring transaksi per shift.
* Rekonsiliasi transaksi.
* Refund.
* Void/cancel.
* Tutup kasir.
* Laporan shift.
* Audit aktivitas kasir.

### Prinsip utama

Sistem harus memisahkan:

```text
Restaurant Operation
        │
        ├── Customer Orders
        ├── Kitchen
        └── Payments

Cashier Shift
        │
        ├── Cashier
        ├── Register
        ├── Sales
        ├── Refund
        └── Reconciliation
```

**Cashier Shift tidak boleh menjadi dependency utama untuk customer order.**

Customer tetap dapat melakukan self-order meskipun:

* Belum ada kasir yang membuka shift.
* Kasir sedang melakukan pergantian shift.
* Shift kasir sebelumnya sudah ditutup.

---

# 2. Goals

## 2.1 Primary Goals

Implementasi fitur:

* Cashier Shift.
* Open Cashier.
* Close Cashier.
* Shift transaction tracking.
* Payment reconciliation.
* Refund tracking.
* Void tracking.
* Shift report.
* Role-based access.
* Audit trail.

## 2.2 Secondary Goals

Mendukung:

* Multiple cashier.
* Multiple shift.
* Multiple register.
* Multi-outlet.
* Cashless restaurant.
* Future cash payment.
* Future hardware POS.

---

# 3. Non-Goals

Fitur berikut belum menjadi bagian dari implementasi awal:

* Payroll.
* Employee attendance.
* HR management.
* Accounting system.
* Full inventory accounting.
* Bank reconciliation.
* Tax accounting.
* Automatic payout reconciliation dari rekening bank.

---

# 4. Existing Architecture

Sistem saat ini:

```text
SvelteKit
    │
    ▼
Go Backend
    │
    ├── Authentication
    ├── Restaurant
    ├── Table
    ├── Menu
    ├── Order
    ├── Payment
    ├── Kitchen
    └── Report
    │
    ▼
PostgreSQL
```

Payment:

```text
Customer
    │
    ▼
Order
    │
    ▼
Payment Gateway
    │
    ▼
Webhook
    │
    ▼
Payment Verification
    │
    ▼
Order CONFIRMED
```

---

# 5. New Architecture

Tambahkan:

```text
Cashier Shift
Register
Shift Transaction
Shift Closing
Reconciliation
Audit Log
```

Architecture:

```text
                    RESTAURANT
                         │
          ┌──────────────┴──────────────┐
          │                             │
          ▼                             ▼
   Restaurant Operation           Cashier Management
          │                             │
          ├── Orders                    ├── Register
          ├── Kitchen                   ├── Shift
          └── Payments                  ├── Reconciliation
                                        └── Closing
```

---

# 6. Business Flow

## 6.1 Open Cashier

```text
Cashier Login
     │
     ▼
Select Register
     │
     ▼
Check Existing Open Shift
     │
     ▼
Input Opening Balance
     │
     ▼
Confirm
     │
     ▼
Shift OPEN
```

Untuk restoran cashless:

```text
Opening Balance = Rp0
```

Tetapi field tetap disediakan agar sistem dapat mendukung cash payment di masa depan.

---

# 7. Open Shift Rules

Sebelum membuat shift:

1. User harus authenticated.
2. User harus memiliki role yang sesuai.
3. Restaurant harus aktif.
4. Register harus aktif.
5. User tidak boleh memiliki shift aktif lain pada register yang sama.
6. Register tidak boleh memiliki shift aktif jika sistem menggunakan satu-shift-per-register.
7. Opening balance harus valid.
8. Opening balance tidak boleh negatif.

---

# 8. Shift Status

Gunakan:

```text
OPEN
CLOSING
CLOSED
CANCELLED
```

Normal flow:

```text
OPEN
  │
  ▼
CLOSING
  │
  ▼
CLOSED
```

---

# 9. Active Shift

Cashier dashboard harus menampilkan shift aktif:

```text
SHIFT #SHIFT-001

Cashier:
Wahyu

Register:
POS-01

Opened:
09:00

Status:
OPEN
```

---

# 10. Shift Transaction Tracking

Setiap transaksi yang relevan harus dapat dikaitkan dengan shift.

Jenis transaksi:

```text
SALE
REFUND
VOID
ADJUSTMENT
```

Contoh:

```text
SHIFT #001

SALE        +Rp100.000
SALE        +Rp50.000
SALE        +Rp75.000
REFUND      -Rp25.000
VOID        -Rp50.000
```

---

# 11. Hubungan Order dan Shift

Order tetap dapat dibuat tanpa shift.

Namun jika order berasal dari aktivitas kasir atau transaksi yang perlu dicatat pada shift, sistem dapat menghubungkannya:

```text
Order
   │
   ├── Payment
   │
   └── Cashier Shift
```

Untuk self-order:

```text
Customer
   │
   ▼
Order
   │
   ▼
Payment
```

Tidak wajib:

```text
Order → Cashier Shift
```

---

# 12. Payment dan Shift

Payment yang berhasil selama periode shift dapat digunakan untuk reconciliation.

Contoh:

```text
Shift:
09:00 - 17:00

Payments:

QRIS
Rp3.000.000

E-Wallet
Rp1.500.000

Virtual Account
Rp500.000

TOTAL
Rp5.000.000
```

---

# 13. Cashless Reconciliation

Karena sistem menggunakan pembayaran non-tunai, rekonsiliasi utama berdasarkan payment.

Formula:

```text
Gross Sales
- Refund
- Void
+ Adjustment
= Net Sales
```

Contoh:

```text
Gross Sales       Rp5.000.000
Refund              Rp200.000
Void                Rp100.000
Adjustment            Rp0
--------------------------------
Net Sales         Rp4.700.000
```

---

# 14. Payment Method Breakdown

Closing harus menampilkan:

```text
QRIS
E-Wallet
Virtual Account
Card
Other
```

Contoh:

```text
Payment Summary

QRIS             Rp3.000.000
GoPay               Rp700.000
OVO                 Rp300.000
DANA                Rp200.000
VA                  Rp500.000
--------------------------------
TOTAL             Rp4.700.000
```

---

# 15. Close Cashier Flow

```text
Cashier
   │
   ▼
Click "Tutup Kasir"
   │
   ▼
System calculates summary
   │
   ▼
Display reconciliation
   │
   ▼
Cashier confirms
   │
   ▼
Shift CLOSING
   │
   ▼
Final validation
   │
   ▼
Shift CLOSED
```

---

# 16. Closing Screen

Contoh:

```text
=================================
          TUTUP KASIR
=================================

Cashier:
Wahyu

Register:
POS-01

Shift:
09:00 - 17:00

---------------------------------

Orders                 152

Gross Sales
Rp5.000.000

Refund
-Rp200.000

Void
-Rp100.000

---------------------------------

NET SALES
Rp4.700.000

---------------------------------

PAYMENT METHODS

QRIS
Rp3.000.000

E-Wallet
Rp1.200.000

VA
Rp500.000

---------------------------------

TOTAL
Rp4.700.000

[ TUTUP KASIR ]
```

---

# 17. Closing Confirmation

Sebelum closing:

```text
Apakah Anda yakin ingin menutup shift?

Shift yang ditutup tidak dapat menerima transaksi kasir baru.

[ Batal ]
[ Tutup Kasir ]
```

---

# 18. Closing Validation

Backend melakukan:

```text
Validate Shift
       │
       ├── Is shift OPEN?
       ├── Is cashier authorized?
       ├── Is register valid?
       ├── Are pending payments?
       ├── Are pending transactions?
       └── Is reconciliation valid?
```

Jika valid:

```text
OPEN
 ↓
CLOSING
 ↓
CLOSED
```

---

# 19. Pending Payment Handling

Jika terdapat payment:

```text
PENDING
```

saat closing, sistem harus memiliki policy.

Default:

```text
CLOSING BLOCKED
```

Contoh:

```text
Tidak dapat menutup kasir.

Masih terdapat 3 pembayaran pending.
```

Alternatif future:

```text
Allow Close
```

dengan payment tetap tercatat berdasarkan waktu transaksi/payment confirmation.

Untuk MVP gunakan:

```text
BLOCK CLOSING IF CRITICAL PENDING TRANSACTIONS EXIST
```

---

# 20. Refund Handling

Refund harus dicatat sebagai transaksi terpisah.

```text
Payment
   │
   ▼
Refund
```

Database:

```text
refunds
```

atau menggunakan transaction ledger.

Contoh:

```text
SALE
+100.000

REFUND
-100.000
```

Refund harus memiliki:

* ID.
* Payment ID.
* Order ID.
* Amount.
* Reason.
* Requested by.
* Approved by.
* Created at.

---

# 21. Void Handling

Void berbeda dengan refund.

### Void

Transaksi/order dibatalkan sebelum finalisasi sesuai business rule.

### Refund

Payment yang sudah berhasil dikembalikan.

Contoh:

```text
Order
   │
   ├── CANCELLED → Void
   │
   └── PAID → Refund
```

Semua tindakan harus memiliki audit log.

---

# 22. Database Changes

Tambahkan tabel:

```text
registers
cashier_shifts
cashier_shift_transactions
shift_closings
refunds
audit_logs
```

---

# 23. Table: registers

```sql
CREATE TABLE registers (
    id UUID PRIMARY KEY,
    restaurant_id UUID NOT NULL,
    outlet_id UUID NULL,

    name VARCHAR(100) NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

---

# 24. Table: cashier_shifts

```sql
CREATE TABLE cashier_shifts (
    id UUID PRIMARY KEY,

    restaurant_id UUID NOT NULL,
    outlet_id UUID NULL,
    register_id UUID NOT NULL,
    cashier_id UUID NOT NULL,

    status VARCHAR(20) NOT NULL,

    opening_balance NUMERIC(15,2) NOT NULL DEFAULT 0,

    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP NULL,

    expected_total NUMERIC(15,2) NOT NULL DEFAULT 0,
    actual_total NUMERIC(15,2) NULL,
    difference NUMERIC(15,2) NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

---

# 25. Table: cashier_shift_transactions

```sql
CREATE TABLE cashier_shift_transactions (
    id UUID PRIMARY KEY,

    shift_id UUID NOT NULL,

    order_id UUID NULL,
    payment_id UUID NULL,

    type VARCHAR(30) NOT NULL,

    amount NUMERIC(15,2) NOT NULL,

    metadata JSONB NULL,

    created_at TIMESTAMP NOT NULL
);
```

Transaction types:

```text
SALE
REFUND
VOID
ADJUSTMENT
```

---

# 26. Table: shift_closings

```sql
CREATE TABLE shift_closings (
    id UUID PRIMARY KEY,

    shift_id UUID NOT NULL UNIQUE,

    gross_sales NUMERIC(15,2) NOT NULL DEFAULT 0,
    refund_total NUMERIC(15,2) NOT NULL DEFAULT 0,
    void_total NUMERIC(15,2) NOT NULL DEFAULT 0,
    adjustment_total NUMERIC(15,2) NOT NULL DEFAULT 0,

    net_sales NUMERIC(15,2) NOT NULL DEFAULT 0,

    expected_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    actual_amount NUMERIC(15,2) NULL,
    difference NUMERIC(15,2) NULL,

    notes TEXT NULL,

    closed_by UUID NOT NULL,
    closed_at TIMESTAMP NOT NULL
);
```

---

# 27. Table: refunds

```sql
CREATE TABLE refunds (
    id UUID PRIMARY KEY,

    restaurant_id UUID NOT NULL,
    order_id UUID NOT NULL,
    payment_id UUID NOT NULL,

    amount NUMERIC(15,2) NOT NULL,

    reason TEXT NOT NULL,

    status VARCHAR(30) NOT NULL,

    requested_by UUID NOT NULL,
    approved_by UUID NULL,

    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NULL
);
```

---

# 28. Audit Logs

Gunakan:

```text
audit_logs
```

Fields:

```text
id
restaurant_id
user_id
action
entity_type
entity_id
metadata
created_at
```

Contoh:

```text
CASHIER_SHIFT_OPENED
CASHIER_SHIFT_CLOSED
REFUND_CREATED
REFUND_APPROVED
ORDER_VOIDED
```

---

# 29. ERD Update

```text
                    restaurants
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ▼              ▼              ▼
        users         registers       orders
          │              │              │
          │              ▼              │
          │        cashier_shifts        │
          │              │              │
          │              ▼              │
          │   shift_transactions        │
          │              │              │
          │              ├───────────────┤
          │              │
          │              ▼
          │       shift_closings
          │
          ▼
      audit_logs

orders
   │
   ▼
payments
   │
   ▼
refunds
```

---

# 30. API Endpoints

## Register

```http
GET    /api/v1/registers
POST   /api/v1/registers
GET    /api/v1/registers/:id
PATCH  /api/v1/registers/:id
DELETE /api/v1/registers/:id
```

---

# 31. Shift API

## Get Current Shift

```http
GET /api/v1/cashier/shifts/current
```

## Open Shift

```http
POST /api/v1/cashier/shifts
```

Request:

```json
{
  "register_id": "reg_001",
  "opening_balance": 0
}
```

Response:

```json
{
  "id": "shift_001",
  "status": "OPEN",
  "register_id": "reg_001",
  "opening_balance": 0
}
```

---

# 32. Shift Detail

```http
GET /api/v1/cashier/shifts/:id
```

---

# 33. Shift Transactions

```http
GET /api/v1/cashier/shifts/:id/transactions
```

Query:

```text
?page=1
&limit=50
&type=SALE
```

---

# 34. Shift Summary

```http
GET /api/v1/cashier/shifts/:id/summary
```

Response:

```json
{
  "gross_sales": 5000000,
  "refund_total": 200000,
  "void_total": 100000,
  "net_sales": 4700000,
  "payment_methods": {
    "qris": 3000000,
    "ewallet": 1200000,
    "va": 500000
  }
}
```

---

# 35. Close Shift

```http
POST /api/v1/cashier/shifts/:id/close
```

Request:

```json
{
  "actual_amount": 0,
  "notes": "Shift normal"
}
```

Untuk cashless:

```text
actual_amount = 0
```

atau field tersebut dapat dibuat nullable jika rekonsiliasi dilakukan berdasarkan payment provider.

---

# 36. Refund API

```http
POST /api/v1/orders/:id/refund
```

Request:

```json
{
  "amount": 50000,
  "reason": "Customer meminta pembatalan"
}
```

---

# 37. Void API

```http
POST /api/v1/orders/:id/void
```

Request:

```json
{
  "reason": "Duplicate order"
}
```

---

# 38. Shift Report API

```http
GET /api/v1/reports/shifts
GET /api/v1/reports/shifts/:id
```

Filter:

```text
cashier
register
outlet
date
status
```

---

# 39. Role Permission

## OWNER

```text
Open Shift             ✓
Close Shift            ✓
View Shift             ✓
View All Shift         ✓
Refund                 ✓
Void                   ✓
Reports                ✓
Register Management    ✓
```

## ADMIN

```text
Open Shift             ✓
Close Shift            ✓
View Shift             ✓
View All Shift         ✓
Refund                 ✓*
Void                   ✓*
Reports                ✓
Register Management    ✓
```

`*` dapat dikonfigurasi berdasarkan policy restoran.

## CASHIER

```text
Open Shift             ✓
Close Own Shift        ✓
View Own Shift         ✓
View Other Shift       -
Refund                 Restricted
Void                   Restricted
Reports                Limited
Register Management    -
```

## KITCHEN

```text
Open Shift             -
Close Shift            -
View Shift             -
Orders                 ✓
Kitchen Status         ✓
```

---

# 40. Frontend Routes

Tambahkan:

```text
/admin/cashier
/admin/cashier/open
/admin/cashier/current
/admin/cashier/close

/admin/registers
/admin/registers/[id]

/admin/shifts
/admin/shifts/[id]
```

Untuk cashier:

```text
/cashier
/cashier/open
/cashier/current
/cashier/close
```

---

# 41. Cashier Dashboard

```text
┌─────────────────────────────────────┐
│ CASHIER                             │
├─────────────────────────────────────┤
│                                     │
│ Status: ● OPEN                      │
│ Register: POS-01                    │
│ Cashier: Wahyu                      │
│                                     │
│ Opened: 09:00                       │
│                                     │
│ Orders: 152                         │
│ Net Sales: Rp4.700.000              │
│                                     │
│ QRIS: Rp3.000.000                   │
│ E-Wallet: Rp1.200.000               │
│ VA: Rp500.000                       │
│                                     │
│ [ VIEW TRANSACTIONS ]               │
│                                     │
│ [ TUTUP KASIR ]                     │
└─────────────────────────────────────┘
```

---

# 42. Open Cashier UI

```text
┌─────────────────────────────┐
│       BUKA KASIR            │
├─────────────────────────────┤
│                             │
│ Register                    │
│ [ POS-01             ▼ ]    │
│                             │
│ Saldo Awal                  │
│ Rp [ 0                 ]    │
│                             │
│        [ BUKA KASIR ]       │
└─────────────────────────────┘
```

---

# 43. Closing UI

```text
┌─────────────────────────────────┐
│          TUTUP KASIR            │
├─────────────────────────────────┤
│                                 │
│ Gross Sales        Rp5.000.000  │
│ Refund              Rp200.000   │
│ Void                Rp100.000   │
│                                 │
│ Net Sales          Rp4.700.000  │
│                                 │
│ QRIS               Rp3.000.000  │
│ E-Wallet           Rp1.200.000  │
│ VA                  Rp500.000   │
│                                 │
│ Notes                           │
│ [____________________________] │
│                                 │
│ [ BATAL ] [ TUTUP KASIR ]      │
└─────────────────────────────────┘
```

---

# 44. Shift History

Admin/Owner:

```text
SHIFT HISTORY

Date        Cashier   Register  Sales
-----------------------------------------
22 Sep      Wahyu     POS-01    Rp4.7 jt
22 Sep      Budi      POS-02    Rp3.2 jt
21 Sep      Wahyu     POS-01    Rp5.1 jt
```

Filter:

```text
Date
Cashier
Register
Outlet
Status
```

---

# 45. Reconciliation Architecture

```text
                  PAYMENT
                     │
                     ▼
              Payment Gateway
                     │
                     ▼
                  Webhook
                     │
                     ▼
                Payment DB
                     │
                     ▼
              Shift Calculator
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
       Expected             Actual
          │                     │
          └──────────┬──────────┘
                     ▼
                 Difference
```

---

# 46. Reconciliation Rules

Untuk cashless:

```text
Expected Payment
=
Successful Payments
-
Refunds
-
Voids
```

Seluruh nominal harus berasal dari database/payment provider yang tervalidasi.

Frontend tidak boleh menentukan angka reconciliation.

---

# 47. Multiple Cashiers

Sistem harus mendukung:

```text
Register POS-01
    │
    ├── Shift Wahyu
    │
    ├── CLOSED
    │
    └── Shift Budi
```

Tidak boleh:

```text
POS-01
 ├── Wahyu OPEN
 └── Budi OPEN
```

jika policy sistem menggunakan satu active shift per register.

Database dapat menggunakan unique constraint/partial index untuk mencegah kondisi ini.

---

# 48. Shift Handover

Future feature:

```text
Shift A
   │
   ▼
Handover
   │
   ▼
Shift B
```

Handover report:

```text
Previous Cashier
New Cashier
Opening Time
Closing Time
Sales
Refund
Void
Payment Summary
```

Fitur ini belum wajib untuk MVP.

---

# 49. Concurrency

Backend harus menangani kemungkinan:

```text
Cashier A → Open Register
Cashier B → Open Register
```

secara bersamaan.

Gunakan database transaction dan unique constraint.

Jangan hanya melakukan:

```text
SELECT active_shift
```

kemudian:

```text
INSERT shift
```

tanpa protection.

---

# 50. Idempotency

Open shift:

```text
POST /cashier/shifts
```

harus memiliki protection agar double click tidak membuat dua shift.

Close shift juga harus idempotent.

Contoh:

```text
Click Close
Click Close
Click Close
```

hasil:

```text
1 CLOSED SHIFT
```

bukan:

```text
3 CLOSED SHIFTS
```

---

# 51. Audit Requirements

Audit setiap:

```text
SHIFT_OPENED
SHIFT_CLOSED
SHIFT_CLOSING
SHIFT_REOPENED
REFUND_CREATED
REFUND_APPROVED
ORDER_VOIDED
ADJUSTMENT_CREATED
```

Data audit:

```text
user
timestamp
IP
entity
entity_id
action
metadata
```

---

# 52. Testing

## Unit Test

Test:

* Shift state transition.
* Sales calculation.
* Refund calculation.
* Void calculation.
* Net sales.
* Payment breakdown.
* Difference calculation.
* Permission.

## Integration Test

Test:

```text
Open Shift
→ Order
→ Payment
→ Shift Transaction
→ Closing
```

## Concurrency Test

Test:

```text
Two users open same register simultaneously.
```

Expected:

```text
Only one succeeds.
```

## Idempotency Test

Test:

```text
Close request sent multiple times.
```

Expected:

```text
Only one close operation.
```

---

# 53. Critical Test Scenarios

### Scenario 1

```text
Cashier opens shift
Opening = Rp0
```

Expected:

```text
OPEN
```

### Scenario 2

```text
Customer makes successful payment
```

Expected:

```text
Payment = PAID
Order = CONFIRMED
Shift transaction = SALE
```

### Scenario 3

```text
Payment refunded
```

Expected:

```text
Refund = CREATED/COMPLETED
Shift transaction = REFUND
```

### Scenario 4

```text
Order voided
```

Expected:

```text
Shift transaction = VOID
Audit log created
```

### Scenario 5

```text
Cashier closes shift
```

Expected:

```text
Shift = CLOSED
Closing record created
```

### Scenario 6

```text
Cashier tries to open second shift on same register
```

Expected:

```text
409 CONFLICT
```

---

# 54. Migration Plan

## Migration 001

Create:

```text
registers
```

## Migration 002

Create:

```text
cashier_shifts
```

## Migration 003

Create:

```text
cashier_shift_transactions
```

## Migration 004

Create:

```text
shift_closings
```

## Migration 005

Create:

```text
refunds
```

## Migration 006

Create:

```text
audit_logs
```

## Migration 007

Add optional:

```text
cashier_shift_id
```

to relevant transactional tables.

---

# 55. Backend Implementation Order

## Step 1

Create models:

```text
Register
CashierShift
ShiftTransaction
ShiftClosing
Refund
```

## Step 2

Create migrations.

## Step 3

Create repositories.

## Step 4

Create shift service.

Services:

```text
OpenShift()
GetCurrentShift()
GetShiftSummary()
CloseShift()
```

## Step 5

Create shift handlers.

## Step 6

Implement RBAC.

## Step 7

Connect payment transactions.

## Step 8

Implement refund/void.

## Step 9

Implement audit logs.

## Step 10

Implement reports.

---

# 56. Frontend Implementation Order

## Step 1

Create cashier layout.

## Step 2

Create current shift store.

```text
cashierShiftStore
```

## Step 3

Create Open Shift page.

## Step 4

Create Current Shift dashboard.

## Step 5

Create Transaction History.

## Step 6

Create Closing Summary.

## Step 7

Create Closing Confirmation.

## Step 8

Create Shift History.

## Step 9

Create Owner/Admin reports.

---

# 57. API Integration Order

Frontend integration:

```text
GET /cashier/shifts/current
       │
       ├── null
       │     ↓
       │  Open Shift
       │
       └── OPEN
             ↓
          Dashboard
```

Closing:

```text
GET /shift/:id/summary
        ↓
Closing Screen
        ↓
POST /shift/:id/close
        ↓
CLOSED
```

---

# 58. Deployment

No infrastructure tambahan diperlukan untuk MVP.

Existing:

```text
SvelteKit
Go
PostgreSQL
Docker
```

tetap digunakan.

Optional future:

```text
Redis
Queue
Event Bus
```

tidak diperlukan untuk tahap pertama.

---

# 59. Rollout Strategy

## Development

Feature flag:

```text
CASHIER_SHIFT_ENABLED=true
```

## Staging

Test:

```text
Open
Order
Payment
Refund
Void
Close
Report
```

## Production

Enable untuk:

```text
Pilot Restaurant
```

Kemudian lakukan monitoring sebelum rollout penuh.

---

# 60. Definition of Done

Fitur dianggap selesai apabila:

* [ ] Register dapat dibuat.
* [ ] Cashier dapat membuka shift.
* [ ] Sistem mencegah duplicate active shift.
* [ ] Current shift dapat dilihat.
* [ ] Transaksi dapat dikaitkan dengan shift jika relevan.
* [ ] Payment summary dapat dihitung.
* [ ] Refund tercatat.
* [ ] Void tercatat.
* [ ] Closing summary dapat dibuat.
* [ ] Cashier dapat menutup shift.
* [ ] Shift berubah menjadi CLOSED.
* [ ] Shift tidak dapat digunakan kembali.
* [ ] Owner dapat melihat shift history.
* [ ] Admin dapat melihat shift history.
* [ ] Cashier hanya dapat melihat shift yang diizinkan.
* [ ] Audit log berjalan.
* [ ] Webhook payment tetap berjalan.
* [ ] Self-order tetap dapat berjalan tanpa active cashier shift.
* [ ] Idempotency diterapkan.
* [ ] Concurrency protection diterapkan.
* [ ] Unit test tersedia.
* [ ] Integration test tersedia.
* [ ] Docker build berhasil.

---

# 61. Final Target Flow

```text
                    CUSTOMER
                       │
                       ▼
                   Scan QR
                       │
                       ▼
                     MENU
                       │
                       ▼
                     ORDER
                       │
                       ▼
                  PAYMENT
                       │
                       ▼
              PAYMENT WEBHOOK
                       │
                       ▼
                  PAYMENT PAID
                       │
                       ▼
                   KITCHEN
                       │
                       ▼
                    READY


             ─────────────────────
                 CASHIER FLOW
             ─────────────────────

                   CASHIER
                       │
                       ▼
                 LOGIN
                       │
                       ▼
                OPEN SHIFT
                       │
                       ▼
                 SHIFT OPEN
                       │
                       ▼
              MONITOR TRANSACTION
                       │
                       ├── SALE
                       ├── REFUND
                       └── VOID
                       │
                       ▼
                SHIFT SUMMARY
                       │
                       ▼
               RECONCILIATION
                       │
                       ▼
                CLOSE SHIFT
                       │
                       ▼
                SHIFT CLOSED
```

---

# 62. Final Architectural Principle

Sistem harus mempertahankan pemisahan berikut:

```text
CUSTOMER ORDER
      ≠
CASHIER SHIFT
```

Customer order merupakan bagian dari **restaurant operation**, sedangkan cashier shift merupakan bagian dari **cashier accountability dan reconciliation**.

Dengan desain ini:

```text
Kasir belum buka
      ↓
Customer tetap bisa order
      ↓
Payment tetap diproses
      ↓
Kitchen tetap menerima order
```

Sementara:

```text
Cashier buka shift
      ↓
Cashier dapat memonitor transaksi
      ↓
Cashier melakukan reconciliation
      ↓
Cashier tutup shift
```

Hal ini membuat sistem tetap cocok untuk model **self-order + cashless payment**, tetapi tetap memiliki mekanisme POS yang lazim digunakan restoran.

# 63. Future Extensions

Setelah Cashier Shift stabil, fitur berikut dapat ditambahkan:

1. Cash payment.
2. Cash drawer management.
3. Cash in / cash out.
4. Shift handover.
5. Multi-register.
6. Multi-outlet.
7. Daily closing.
8. Accounting integration.
9. Bank reconciliation.
10. Automated payment settlement reconciliation.
11. Printer integration.
12. Hardware POS.
13. Inventory integration.
14. Employee attendance.
15. Advanced financial reporting.
