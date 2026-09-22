# PRD — Self-Order Restaurant & Cashier System

**Version:** 1.0
**Status:** Draft
**Platform:** Web / Mobile Web
**Architecture:** Modular Monolith
**Frontend:** SvelteKit + TypeScript + Tailwind CSS
**Backend:** Golang + Gin
**Database:** PostgreSQL
**Realtime:** WebSocket / SSE
**Payment:** Payment Gateway
**Deployment:** Docker

---

# 1. Product Overview

## 1.1 Product Name

**Self-Order Restaurant System**

Nama produk dapat diganti sesuai branding aplikasi.

## 1.2 Product Description

Self-Order Restaurant System adalah platform pemesanan makanan berbasis QR Code yang memungkinkan pelanggan melakukan pemesanan secara mandiri melalui perangkat smartphone tanpa harus memanggil pelayan atau melakukan pemesanan langsung di kasir.

Setiap meja restoran memiliki QR Code unik. Pelanggan melakukan scan QR Code tersebut untuk mengakses menu restoran, memilih makanan/minuman, memasukkan pesanan ke keranjang, melakukan checkout, dan membayar secara non-tunai melalui payment gateway.

Setelah pembayaran berhasil dikonfirmasi oleh backend melalui webhook payment gateway, pesanan secara otomatis diteruskan ke sistem dapur.

Restoran memiliki dashboard untuk mengelola menu, meja, pesanan, pembayaran, pengguna, serta laporan transaksi.

---

# 2. Problem Statement

Sistem restoran konvensional memiliki beberapa permasalahan:

1. Pelanggan harus menunggu pelayan untuk memesan.
2. Kesalahan pencatatan pesanan dapat terjadi.
3. Kasir harus memasukkan ulang pesanan.
4. Proses pembayaran dapat menyebabkan antrean.
5. Pelanggan tidak dapat mengetahui status pesanan secara real-time.
6. Restoran membutuhkan proses manual untuk meneruskan order ke dapur.
7. Rekap transaksi membutuhkan pekerjaan administratif.
8. Sulit melakukan pemisahan akses antara owner, kasir, dan kitchen.

Sistem ini dirancang untuk mengurangi proses manual tersebut melalui digitalisasi proses order dan pembayaran.

---

# 3. Goals

## 3.1 Primary Goals

Sistem harus memungkinkan:

* Customer melakukan scan QR meja.
* Customer melihat menu digital.
* Customer memilih makanan/minuman.
* Customer melakukan checkout.
* Customer melakukan pembayaran cashless.
* Sistem menerima webhook payment gateway.
* Sistem memvalidasi pembayaran.
* Order otomatis masuk ke kitchen.
* Kitchen mengubah status order.
* Customer melihat status order secara real-time.
* Admin mengelola restoran.
* Kasir melihat dan mengelola transaksi.
* Owner melihat laporan bisnis.

## 3.2 Secondary Goals

* Multi-restaurant / multi-tenant.
* Manajemen meja.
* Manajemen QR Code.
* Manajemen kategori dan menu.
* Manajemen modifier/add-on.
* Riwayat transaksi.
* Laporan penjualan.
* Audit log.
* Role-based access control.

---

# 4. Non-Goals

Versi MVP tidak wajib menyediakan:

* Delivery.
* Driver management.
* Marketplace.
* Loyalty point kompleks.
* Inventory accounting penuh.
* Payroll karyawan.
* Akuntansi perusahaan.
* Integrasi hardware POS khusus.
* Offline ordering.

Fitur tersebut dapat dimasukkan pada versi berikutnya.

---

# 5. Target Users

## 5.1 Customer

Pelanggan restoran yang melakukan pemesanan.

## 5.2 Admin

Pengguna yang mengelola operasional restoran dan konfigurasi sistem.

## 5.3 Kasir

Pengguna yang menangani transaksi dan operasional pembayaran.

## 5.4 Kitchen

Pengguna yang menerima dan memproses order.

## 5.5 Owner

Pemilik restoran yang membutuhkan akses laporan dan monitoring bisnis.

---

# 6. User Roles

| Role     | Menu | Order       | Payment | Kitchen     | Report  | User Management |
| -------- | ---- | ----------- | ------- | ----------- | ------- | --------------- |
| Owner    | Full | Full        | Full    | View        | Full    | Full            |
| Admin    | Full | Full        | Full    | View        | Full    | Manage          |
| Kasir    | View | Manage      | Manage  | View        | Limited | No              |
| Kitchen  | View | Process     | View    | Manage      | No      | No              |
| Customer | View | Create/View | Pay     | View Status | No      | No              |

---

# 7. Core User Journey

```text
Customer
   |
   v
Scan QR Meja
   |
   v
Restaurant Menu
   |
   v
Select Menu
   |
   v
Cart
   |
   v
Checkout
   |
   v
Create Order
   |
   v
Create Payment
   |
   v
Payment Gateway
   |
   v
Customer Pays
   |
   v
Payment Webhook
   |
   v
Backend Validation
   |
   v
Payment PAID
   |
   v
Order CONFIRMED
   |
   v
Kitchen
   |
   v
Cooking
   |
   v
Ready
   |
   v
Completed
```

---

# 8. Customer Features

## 8.1 Scan QR

Customer melakukan scan QR yang terdapat pada meja.

QR harus mengandung identifier meja/restoran.

Contoh:

```text
https://app.example.com/order/rst_123/tbl_001
```

Backend harus memvalidasi:

* Restaurant exists.
* Table exists.
* Table aktif.
* Restaurant aktif.
* QR valid.

---

# 9. Customer Menu

Customer dapat melihat:

* Restaurant name.
* Restaurant logo.
* Categories.
* Menu.
* Price.
* Image.
* Description.
* Availability.
* Modifier.
* Add-on.

Contoh:

```text
Nasi Goreng Spesial

Rp 25.000

Nasi goreng dengan telur dan ayam.

Add-on:
[ ] Telur +Rp5.000
[ ] Keju +Rp4.000

[ Tambah ]
```

---

# 10. Cart

Cart menyimpan:

```text
Menu
Quantity
Modifier
Notes
Price
Subtotal
```

Customer dapat:

* Menambah quantity.
* Mengurangi quantity.
* Menghapus item.
* Menambahkan catatan.

Contoh:

```text
Nasi Goreng
2 × Rp25.000

Catatan:
Tidak pedas

Subtotal:
Rp50.000
```

---

# 11. Checkout

Checkout menampilkan:

```text
Restaurant
Table
Items
Subtotal
Tax
Service Charge
Discount
Total
Payment Method
```

Backend harus menghitung ulang harga.

Frontend tidak boleh dipercaya untuk menentukan:

```text
price
subtotal
tax
discount
total
```

Frontend hanya mengirim:

```text
menu_id
quantity
modifier_id
notes
```

Backend mengambil harga dari database.

---

# 12. Payment

Semua pembayaran dilakukan secara cashless melalui payment gateway.

Metode pembayaran tergantung provider yang digunakan, misalnya:

* QRIS
* E-wallet
* Virtual Account
* Card
* Metode lain yang disediakan provider

Flow:

```text
Frontend
   |
   v
Backend
   |
   v
Payment Gateway
   |
   v
Payment Page / QR / Payment Method
```

---

# 13. Payment Webhook

Webhook merupakan komponen penting.

Payment gateway mengirim notifikasi ke:

```text
POST /api/v1/payments/webhook
```

Backend melakukan:

1. Validate signature.
2. Validate transaction ID.
3. Validate order ID.
4. Validate amount.
5. Validate merchant/account.
6. Check transaction status.
7. Check idempotency.
8. Update payment.
9. Update order.

---

# 14. Payment State Machine

```text
PENDING
   |
   +------> PAID
   |
   +------> FAILED
   |
   +------> EXPIRED
   |
   +------> CANCELLED
```

Order tidak boleh dianggap lunas hanya karena frontend menerima redirect success.

Sumber kebenaran pembayaran berasal dari hasil validasi backend terhadap payment gateway/webhook.

---

# 15. Order State Machine

```text
WAITING_PAYMENT
       |
       v
CONFIRMED
       |
       v
PREPARING
       |
       v
READY
       |
       v
COMPLETED
```

Alternative:

```text
WAITING_PAYMENT
      |
      +----> CANCELLED
```

Setelah paid:

```text
CONFIRMED
      |
      +----> CANCELLED
```

Jika restoran mendukung refund, status refund sebaiknya dipisahkan dari order status.

---

# 16. Kitchen Display System

Kitchen dashboard menampilkan:

```text
NEW ORDERS
```

Contoh:

```text
ORDER #ORD-000123

TABLE 12

2x Nasi Goreng
1x Es Teh

Note:
Tidak pedas

[ TERIMA ]
```

Setelah diterima:

```text
PREPARING
```

Setelah selesai:

```text
READY
```

Kitchen tidak memiliki permission untuk mengubah harga atau pembayaran.

---

# 17. Customer Order Tracking

Customer dapat melihat:

```text
Order #000123

✓ Pembayaran berhasil
✓ Pesanan diterima
● Sedang dibuat
○ Pesanan siap
○ Selesai
```

Realtime menggunakan:

```text
WebSocket
```

atau:

```text
SSE
```

---

# 18. Admin Dashboard

Admin dapat mengelola:

## Restaurant

* Nama restoran.
* Logo.
* Alamat.
* Nomor telepon.
* Tax.
* Service charge.
* Operating hours.

## Menu

* Category.
* Product.
* Price.
* Image.
* Description.
* Availability.
* Modifier.
* Add-on.

## Table

* Create.
* Update.
* Delete.
* Activate/deactivate.
* Generate QR.
* Download QR.

## Order

* View.
* Filter.
* Search.
* Cancel.
* View detail.

## Payment

* Transaction.
* Status.
* Payment method.
* Amount.
* Reference.

---

# 19. Cashier Dashboard

Kasir dapat:

* Melihat order.
* Melihat payment.
* Melihat status transaksi.
* Membantu customer.
* Melihat transaksi hari ini.
* Melihat detail order.
* Melakukan tindakan yang diizinkan sistem terhadap order.

Jika sistem benar-benar cashless, kasir **tidak melakukan input pembayaran tunai**.

---

# 20. Owner Dashboard

Owner dapat melihat:

```text
Today's Sales
Today's Orders
Average Order Value
Paid Transactions
Cancelled Orders
Popular Menu
Sales by Period
```

Contoh:

```text
TODAY

Revenue
Rp8.450.000

Orders
182

Average Order
Rp46.428

Paid
175

Cancelled
7
```

---

# 21. Multi-Tenant Architecture

Jika sistem dibuat sebagai SaaS, satu instance aplikasi dapat digunakan banyak restoran.

```text
Platform
│
├── Restaurant A
│   ├── Users
│   ├── Tables
│   ├── Menus
│   └── Orders
│
├── Restaurant B
│   ├── Users
│   ├── Tables
│   ├── Menus
│   └── Orders
│
└── Restaurant C
    ├── Users
    ├── Tables
    ├── Menus
    └── Orders
```

Semua data tenant harus memiliki:

```text
restaurant_id
```

Backend wajib melakukan tenant isolation pada setiap query yang relevan.

---

# 22. ERD Database

## 22.1 Core ERD

```text
┌─────────────────┐
│   restaurants   │
├─────────────────┤
│ id PK           │
│ name            │
│ slug            │
│ logo_url        │
│ status          │
│ tax_percent     │
│ service_percent │
│ created_at      │
└───────┬─────────┘
        │
        ├──────────────────┐
        │                  │
        ▼                  ▼
┌───────────────┐    ┌───────────────┐
│    users      │    │    tables     │
├───────────────┤    ├───────────────┤
│ id PK         │    │ id PK         │
│ restaurant_id │    │ restaurant_id │
│ role          │    │ name          │
│ name          │    │ qr_token      │
│ email         │    │ status        │
└───────────────┘    └───────┬───────┘
                             │
                             ▼
                     ┌────────────────┐
                     │ table_sessions │
                     ├────────────────┤
                     │ id PK          │
                     │ restaurant_id  │
                     │ table_id FK    │
                     │ status         │
                     │ started_at     │
                     │ ended_at       │
                     └───────┬────────┘
                             │
                             ▼
                       ┌─────────────┐
                       │   orders    │
                       ├─────────────┤
                       │ id PK       │
                       │ restaurant  │
                       │ table_id    │
                       │ session_id  │
                       │ status      │
                       │ subtotal    │
                       │ tax         │
                       │ service     │
                       │ discount    │
                       │ total       │
                       └──────┬──────┘
                              │
                     ┌────────┴────────┐
                     ▼                 ▼
              ┌─────────────┐   ┌─────────────┐
              │order_items  │   │  payments   │
              ├─────────────┤   ├─────────────┤
              │ id PK       │   │ id PK       │
              │ order_id FK │   │ order_id FK │
              │ menu_id FK  │   │ gateway     │
              │ quantity    │   │ reference   │
              │ unit_price  │   │ amount      │
              │ subtotal    │   │ status      │
              └──────┬──────┘   └─────────────┘
                     │
                     ▼
               ┌───────────┐
               │   menus   │
               ├───────────┤
               │ id PK     │
               │ category  │
               │ name      │
               │ price     │
               │ available │
               └───────────┘
```

---

# 23. Database Tables

## restaurants

```sql
id
name
slug
logo_url
address
phone
tax_percent
service_percent
status
created_at
updated_at
```

## users

```sql
id
restaurant_id
name
email
password_hash
role
status
created_at
updated_at
```

Roles:

```text
OWNER
ADMIN
CASHIER
KITCHEN
```

## tables

```sql
id
restaurant_id
name
qr_token
status
created_at
updated_at
```

## table_sessions

```sql
id
restaurant_id
table_id
status
started_at
ended_at
```

## categories

```sql
id
restaurant_id
name
sort_order
status
created_at
updated_at
```

## menus

```sql
id
restaurant_id
category_id
name
description
price
image_url
available
sort_order
created_at
updated_at
```

## menu_modifiers

```sql
id
restaurant_id
name
type
required
created_at
updated_at
```

## menu_modifier_options

```sql
id
modifier_id
name
additional_price
created_at
updated_at
```

## orders

```sql
id
restaurant_id
table_id
table_session_id
order_number
status
subtotal
tax
service_charge
discount
total
notes
created_at
updated_at
```

## order_items

```sql
id
order_id
menu_id
menu_name_snapshot
unit_price
quantity
subtotal
notes
created_at
```

Harga dan nama menu sebaiknya disimpan sebagai snapshot pada order item.

Tujuannya agar perubahan menu di masa depan tidak mengubah histori transaksi.

## payments

```sql
id
restaurant_id
order_id
provider
provider_transaction_id
payment_method
amount
status
expired_at
paid_at
created_at
updated_at
```

## payment_webhook_events

```sql
id
provider
event_id
event_type
payload
processed
processed_at
created_at
```

Tabel ini penting untuk idempotency dan audit webhook.

## order_status_history

```sql
id
order_id
from_status
to_status
changed_by
created_at
```

---

# 24. Index Database

Index yang direkomendasikan:

```sql
orders(restaurant_id)
orders(restaurant_id, created_at)
orders(order_number)
orders(status)
payments(order_id)
payments(provider_transaction_id)
menus(restaurant_id)
menus(category_id)
tables(restaurant_id)
users(restaurant_id)
```

Untuk SaaS:

```sql
orders(restaurant_id, status)
orders(restaurant_id, created_at)
menus(restaurant_id, available)
```

---

# 25. API Architecture

Base URL:

```text
/api/v1
```

Authentication:

```text
Authorization: Bearer <token>
```

atau session cookie untuk web application.

---

# 26. Authentication API

## Login

```http
POST /api/v1/auth/login
```

Request:

```json
{
  "email": "admin@example.com",
  "password": "password"
}
```

Response:

```json
{
  "user": {
    "id": "usr_001",
    "name": "Admin",
    "role": "ADMIN"
  },
  "access_token": "..."
}
```

## Current User

```http
GET /api/v1/auth/me
```

## Logout

```http
POST /api/v1/auth/logout
```

---

# 27. Restaurant API

```http
GET    /api/v1/restaurants/:id
PATCH  /api/v1/restaurants/:id
```

Admin/Owner only.

---

# 28. Table API

```http
GET    /api/v1/tables
POST   /api/v1/tables
GET    /api/v1/tables/:id
PATCH  /api/v1/tables/:id
DELETE /api/v1/tables/:id
```

QR:

```http
GET /api/v1/tables/:id/qr
POST /api/v1/tables/:id/qr/regenerate
```

---

# 29. Public Table API

Customer tidak perlu login.

```http
GET /api/v1/public/tables/:qr_token
```

Response:

```json
{
  "restaurant": {
    "id": "rst_001",
    "name": "Restaurant ABC"
  },
  "table": {
    "id": "tbl_001",
    "name": "Table 01"
  }
}
```

---

# 30. Category API

```http
GET    /api/v1/categories
POST   /api/v1/categories
GET    /api/v1/categories/:id
PATCH  /api/v1/categories/:id
DELETE /api/v1/categories/:id
```

---

# 31. Menu API

Public:

```http
GET /api/v1/public/restaurants/:restaurant_id/menu
```

Admin:

```http
GET    /api/v1/menus
POST   /api/v1/menus
GET    /api/v1/menus/:id
PATCH  /api/v1/menus/:id
DELETE /api/v1/menus/:id
```

Availability:

```http
PATCH /api/v1/menus/:id/availability
```

---

# 32. Order API

Create order:

```http
POST /api/v1/public/orders
```

Request:

```json
{
  "qr_token": "qr_xxx",
  "items": [
    {
      "menu_id": "menu_001",
      "quantity": 2,
      "notes": "Tidak pedas"
    }
  ]
}
```

Backend melakukan:

```text
Validate QR
     ↓
Validate Restaurant
     ↓
Validate Table
     ↓
Validate Menu
     ↓
Get Current Price
     ↓
Calculate Total
     ↓
Create Order
```

Response:

```json
{
  "order_id": "ord_001",
  "order_number": "ORD-000001",
  "status": "WAITING_PAYMENT",
  "total": 69000
}
```

---

# 33. Order Detail

```http
GET /api/v1/public/orders/:id
```

Customer dapat melihat:

```json
{
  "order_number": "ORD-000001",
  "status": "PREPARING",
  "total": 69000,
  "items": []
}
```

---

# 34. Admin Order API

```http
GET /api/v1/orders
GET /api/v1/orders/:id
PATCH /api/v1/orders/:id/status
POST /api/v1/orders/:id/cancel
```

---

# 35. Kitchen API

```http
GET /api/v1/kitchen/orders
POST /api/v1/kitchen/orders/:id/accept
POST /api/v1/kitchen/orders/:id/ready
```

Backend harus memastikan hanya role KITCHEN/ADMIN/OWNER yang dapat menjalankan endpoint tertentu.

---

# 36. Payment API

Create payment:

```http
POST /api/v1/orders/:id/payment
```

Response:

```json
{
  "payment_id": "pay_001",
  "status": "PENDING",
  "payment_url": "..."
}
```

Payment status:

```http
GET /api/v1/payments/:id
```

---

# 37. Payment Webhook API

```http
POST /api/v1/payments/webhook
```

Webhook tidak menggunakan authentication user biasa.

Authentication webhook menggunakan mekanisme yang diberikan payment gateway, misalnya:

```text
Signature
Secret
Token
Provider verification API
```

---

# 38. Payment Flow

## Step 1 — Customer Checkout

```text
Customer
   |
   v
POST /orders
```

Backend:

```text
Create Order
status = WAITING_PAYMENT
```

---

## Step 2 — Create Payment

```text
POST /orders/:id/payment
```

Backend:

```text
Get Order
    ↓
Verify status
    ↓
Calculate amount
    ↓
Create Payment Gateway Transaction
    ↓
Save Payment
    ↓
Return Payment URL / QR
```

---

## Step 3 — Customer Pays

```text
Customer
    |
    v
Payment Gateway
    |
    v
Payment Provider
```

---

## Step 4 — Webhook

```text
Payment Gateway
      |
      | webhook
      v
POST /payments/webhook
```

---

## Step 5 — Validate

Backend:

```text
Verify Signature
       ↓
Find Payment
       ↓
Check Amount
       ↓
Check Order
       ↓
Check Event ID
       ↓
Check Idempotency
```

---

## Step 6 — Update

Jika valid:

```text
payments.status = PAID

orders.status = CONFIRMED
```

Kemudian publish realtime event:

```text
ORDER_CONFIRMED
```

---

# 39. Idempotency

Webhook dapat dikirim lebih dari satu kali.

Backend tidak boleh membuat order atau payment kedua.

Contoh:

```text
Webhook #123
      ↓
Process
      ↓
processed = true

Webhook #123 lagi
      ↓
Already processed
      ↓
Ignore
```

Gunakan:

```text
provider + event_id
```

sebagai unique constraint jika provider menyediakan event ID yang dapat diandalkan.

---

# 40. Payment Security

Backend tidak boleh mempercayai:

```text
amount dari frontend
status dari frontend
payment success dari frontend
```

Frontend hanya digunakan sebagai UI.

Source of truth:

```text
Database + Payment Gateway verification
```

---

# 41. API Response Standard

Semua API menggunakan format konsisten.

Success:

```json
{
  "success": true,
  "data": {},
  "message": "Success"
}
```

Error:

```json
{
  "success": false,
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "Order not found"
  }
}
```

---

# 42. HTTP Status

Gunakan:

```text
200 OK
201 Created
204 No Content
400 Bad Request
401 Unauthorized
403 Forbidden
404 Not Found
409 Conflict
422 Unprocessable Entity
429 Too Many Requests
500 Internal Server Error
```

---

# 43. SvelteKit Folder Structure

```text
frontend/
│
├── src/
│   │
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts
│   │   │   ├── auth.ts
│   │   │   ├── menu.ts
│   │   │   ├── order.ts
│   │   │   ├── payment.ts
│   │   │   └── table.ts
│   │   │
│   │   ├── components/
│   │   │   ├── Button.svelte
│   │   │   ├── Modal.svelte
│   │   │   ├── MenuCard.svelte
│   │   │   ├── CartItem.svelte
│   │   │   └── OrderStatus.svelte
│   │   │
│   │   ├── stores/
│   │   │   ├── cart.ts
│   │   │   ├── auth.ts
│   │   │   └── order.ts
│   │   │
│   │   ├── types/
│   │   │   ├── menu.ts
│   │   │   ├── order.ts
│   │   │   ├── payment.ts
│   │   │   └── user.ts
│   │   │
│   │   └── utils/
│   │
│   ├── routes/
│   │   ├── +layout.svelte
│   │   ├── +page.svelte
│   │   │
│   │   ├── order/
│   │   │   └── [restaurant]/
│   │   │       └── [table]/
│   │   │           ├── +page.svelte
│   │   │           ├── menu/
│   │   │           ├── cart/
│   │   │           ├── checkout/
│   │   │           └── order/
│   │   │
│   │   ├── admin/
│   │   │   ├── +layout.svelte
│   │   │   ├── dashboard/
│   │   │   ├── menus/
│   │   │   ├── categories/
│   │   │   ├── tables/
│   │   │   ├── orders/
│   │   │   ├── payments/
│   │   │   ├── users/
│   │   │   └── settings/
│   │   │
│   │   ├── kitchen/
│   │   │   ├── +layout.svelte
│   │   │   ├── orders/
│   │   │   └── display/
│   │   │
│   │   └── login/
│   │
│   └── app.html
│
├── static/
├── tests/
├── package.json
├── svelte.config.js
├── vite.config.ts
├── tailwind.config.ts
└── Dockerfile
```

---

# 44. Go Backend Structure

Gunakan modular monolith terlebih dahulu.

```text
backend/
│
├── cmd/
│   └── server/
│       └── main.go
│
├── internal/
│   │
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── model.go
│   │   └── dto.go
│   │
│   ├── restaurant/
│   │
│   ├── user/
│   │
│   ├── table/
│   │
│   ├── menu/
│   │
│   ├── order/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── model.go
│   │   └── dto.go
│   │
│   ├── payment/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── gateway.go
│   │   ├── webhook.go
│   │   └── model.go
│   │
│   ├── kitchen/
│   │
│   ├── report/
│   │
│   └── notification/
│
├── pkg/
│   ├── database/
│   ├── logger/
│   ├── response/
│   ├── validator/
│   ├── auth/
│   └── websocket/
│
├── migrations/
│
├── configs/
│
├── docs/
│
├── tests/
│
├── go.mod
├── go.sum
└── Dockerfile
```

---

# 45. Backend Layer

Gunakan pattern:

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
Database
```

Contoh:

```text
POST /orders
      ↓
OrderHandler
      ↓
OrderService
      ↓
OrderRepository
      ↓
PostgreSQL
```

Payment:

```text
PaymentHandler
      ↓
PaymentService
      ↓
PaymentGateway
      ↓
Provider
```

---

# 46. Payment Gateway Abstraction

Jangan membuat seluruh aplikasi tergantung pada satu provider.

Gunakan interface:

```go
type PaymentGateway interface {
    CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error)
    GetPayment(ctx context.Context, transactionID string) (*PaymentStatus, error)
    VerifyWebhook(req WebhookRequest) (*WebhookResult, error)
}
```

Kemudian:

```text
PaymentGateway
      │
      ├── MidtransGateway
      │
      ├── XenditGateway
      │
      └── OtherGateway
```

Dengan demikian provider dapat diganti tanpa mengubah seluruh business logic.

---

# 47. Role-Based Access Control

## OWNER

Full access:

```text
Restaurant
Users
Menu
Tables
Orders
Payments
Reports
Settings
```

## ADMIN

```text
Menu
Tables
Orders
Payments
Users
Settings
```

Tidak boleh mengubah data tertentu yang khusus owner jika sistem menerapkannya.

## CASHIER

```text
Orders
Payments
Transaction History
```

Tidak boleh:

```text
Delete Menu
Manage Owner
Manage System
```

## KITCHEN

```text
View Orders
Accept Order
Start Cooking
Mark Ready
```

Tidak boleh:

```text
View Financial Reports
Manage Payment
Manage Users
```

---

# 48. Permission Matrix

```text
                     OWNER ADMIN CASHIER KITCHEN
------------------------------------------------
Dashboard              ✓     ✓      ✓       ✓
Restaurant Settings    ✓     ✓      -       -
Users                  ✓     ✓      -       -
Menu                   ✓     ✓      -       -
Tables                 ✓     ✓      -       -
Orders                 ✓     ✓      ✓       ✓
Payment                ✓     ✓      ✓       -
Kitchen                ✓     ✓      ✓       ✓
Reports                ✓     ✓      limited -
```

---

# 49. Realtime Architecture

WebSocket endpoint:

```text
/ws
```

Customer subscribe:

```text
order:{order_id}
```

Kitchen subscribe:

```text
restaurant:{restaurant_id}:kitchen
```

Event:

```json
{
  "event": "ORDER_STATUS_CHANGED",
  "data": {
    "order_id": "ord_001",
    "status": "READY"
  }
}
```

---

# 50. Order Event Flow

```text
Payment Webhook
       |
       v
Payment Service
       |
       v
Order Service
       |
       +----> Database
       |
       +----> Event Bus
                    |
             ┌──────┴──────┐
             ▼             ▼
         Kitchen        Customer
```

Untuk MVP event bus internal sederhana sudah cukup.

Tidak perlu Kafka.

---

# 51. QR Code Security

QR sebaiknya tidak hanya menggunakan:

```text
table_id=1
```

Gunakan token acak:

```text
qr_token = random secure token
```

Contoh:

```text
/order/rst_123/tbl_001?token=8f92...
```

Token dapat diregenerasi jika QR dicurigai disalahgunakan.

---

# 52. Customer Session

Customer tidak wajib membuat akun.

Flow:

```text
Scan QR
   ↓
Generate Anonymous Session
   ↓
Browse Menu
   ↓
Create Order
```

Session dapat menggunakan:

```text
HttpOnly Secure Cookie
```

atau anonymous session ID.

---

# 53. Cart Security

Cart frontend boleh menyimpan data sementara.

Namun ketika checkout:

```text
Frontend
   |
   | menu_id + quantity
   v
Backend
   |
   ├── Check Menu
   ├── Check Price
   ├── Check Availability
   ├── Calculate Tax
   ├── Calculate Service
   └── Calculate Total
```

Backend menghasilkan final total.

---

# 54. Prevent Price Manipulation

Jangan menerima:

```json
{
  "price": 1000
}
```

dari frontend.

Frontend cukup:

```json
{
  "menu_id": "menu_001",
  "quantity": 2
}
```

Backend:

```text
menu_001
    ↓
Database
    ↓
price = 25000
```

---

# 55. Order Number

Gunakan nomor order yang mudah dibaca manusia:

```text
ORD-000001
ORD-000002
ORD-000003
```

Namun primary key database tetap menggunakan UUID/ULID.

Contoh:

```text
id:
01JABC...

order_number:
ORD-000123
```

---

# 56. UUID / ULID

Untuk ID database dapat menggunakan:

```text
UUID
```

atau:

```text
ULID
```

ULID memiliki keuntungan untuk ordering berdasarkan waktu dan tetap unik.

---

# 57. Audit Log

Action penting dicatat:

```text
USER_LOGIN
MENU_CREATED
MENU_UPDATED
MENU_DELETED
ORDER_CANCELLED
PAYMENT_CREATED
PAYMENT_PAID
PAYMENT_REFUNDED
TABLE_CREATED
USER_CREATED
```

Contoh:

```text
audit_logs

id
restaurant_id
user_id
action
entity_type
entity_id
metadata
created_at
```

---

# 58. Security Requirements

Sistem harus menerapkan:

* HTTPS.
* Password hashing.
* Secure cookies.
* Input validation.
* SQL injection protection.
* CSRF protection jika menggunakan cookie authentication.
* Rate limiting.
* Webhook signature verification.
* Authorization per role.
* Tenant isolation.
* Audit logging.
* Secure secret management.
* Payment amount verification.
* Idempotency.

---

# 59. Observability

Minimal:

```text
Application Logs
Error Tracking
Health Check
Database Monitoring
```

Endpoint:

```http
GET /health
GET /ready
```

Contoh:

```json
{
  "status": "ok",
  "database": "ok"
}
```

---

# 60. Docker Architecture

Development:

```text
docker-compose.yml

frontend
backend
postgres
```

Production:

```text
Caddy
   |
   ├── Frontend
   ├── Backend
   └── PostgreSQL
```

Jika database menggunakan managed database, PostgreSQL tidak perlu berada di server aplikasi.

---

# 61. Docker Services

```yaml
services:

  frontend:
    build: ./frontend

  backend:
    build: ./backend

  postgres:
    image: postgres
```

Untuk production, database lebih baik dipisahkan dari container aplikasi jika menggunakan layanan managed database.

---

# 62. Environment Variables

Backend:

```text
APP_ENV
APP_PORT

DATABASE_URL

JWT_SECRET

PAYMENT_PROVIDER
PAYMENT_API_KEY
PAYMENT_SECRET
PAYMENT_WEBHOOK_SECRET

STORAGE_ENDPOINT
STORAGE_ACCESS_KEY
STORAGE_SECRET_KEY
STORAGE_BUCKET

CORS_ORIGINS
```

Frontend:

```text
PUBLIC_API_URL
PUBLIC_WS_URL
```

Jangan memasukkan secret payment gateway ke frontend.

---

# 63. API Security Boundary

Public:

```text
GET /public/menu
GET /public/table
POST /public/orders
GET /public/orders/:id
POST /orders/:id/payment
```

Authenticated:

```text
/admin/*
/kitchen/*
```

Webhook:

```text
/payments/webhook
```

Webhook memiliki mekanisme authentication tersendiri.

---

# 64. Error Handling

Contoh:

```text
MENU_UNAVAILABLE
TABLE_NOT_FOUND
RESTAURANT_INACTIVE
ORDER_NOT_FOUND
ORDER_ALREADY_PAID
PAYMENT_EXPIRED
PAYMENT_FAILED
INVALID_WEBHOOK
INVALID_PAYMENT_AMOUNT
FORBIDDEN
```

---

# 65. Business Rules

## Rule 1

Menu unavailable tidak dapat dipesan.

## Rule 2

Harga selalu diambil backend.

## Rule 3

Order tidak dianggap paid berdasarkan frontend.

## Rule 4

Payment success harus diverifikasi.

## Rule 5

Webhook harus idempotent.

## Rule 6

Order hanya dapat diproses kitchen jika pembayaran sudah valid, kecuali restoran mengaktifkan konfigurasi khusus.

## Rule 7

Order yang sudah completed tidak boleh diedit.

## Rule 8

Historical order mempertahankan snapshot harga.

## Rule 9

User hanya dapat mengakses data restaurant yang menjadi tenant-nya.

## Rule 10

Customer anonymous hanya dapat melihat order berdasarkan secure order/session identifier.

---

# 66. MVP Scope

### Phase 1

Customer:

* QR scan.
* Menu.
* Cart.
* Checkout.
* Payment.
* Order status.

Admin:

* Login.
* Menu CRUD.
* Category CRUD.
* Table CRUD.
* QR generation.
* Order dashboard.

Kitchen:

* Order list.
* Accept.
* Preparing.
* Ready.

Payment:

* Payment creation.
* Webhook.
* Payment verification.

---

# 67. Phase 2

* Owner dashboard.
* Sales reports.
* Multi-tenant.
* Modifier.
* Add-ons.
* Discounts.
* Order history.
* Audit log.
* Better analytics.

---

# 68. Phase 3

* Inventory.
* Loyalty.
* Customer account.
* Reservation.
* Promotion.
* Integration printer.
* POS hardware.
* Accounting integration.
* Advanced analytics.

---

# 69. Recommended Development Order

## Sprint 1 — Foundation

```text
Repository
Docker
PostgreSQL
Go
SvelteKit
Authentication
Migration
```

## Sprint 2 — Restaurant

```text
Restaurant
Users
Roles
Tables
QR
```

## Sprint 3 — Menu

```text
Category
Menu
Availability
Public Menu
```

## Sprint 4 — Order

```text
Cart
Checkout
Order
Order Items
Order Status
```

## Sprint 5 — Payment

```text
Payment Gateway
Create Payment
Webhook
Verification
Idempotency
```

## Sprint 6 — Kitchen

```text
Kitchen Dashboard
Realtime
Order Status
```

## Sprint 7 — Admin

```text
Dashboard
Orders
Payments
Tables
Menus
```

## Sprint 8 — Production

```text
Security
Logging
Monitoring
Testing
Docker
CI/CD
Backup
```

---

# 70. Testing Strategy

## Unit Test

Test:

```text
Price calculation
Tax
Service charge
Discount
Order state
Payment state
Permission
```

## Integration Test

Test:

```text
API → Service → Repository → Database
```

## Payment Test

Test:

```text
Payment pending
Payment success
Payment failed
Payment expired
Duplicate webhook
Invalid webhook
Wrong amount
```

## E2E Test

Test:

```text
Scan QR
→ Menu
→ Cart
→ Checkout
→ Payment
→ Webhook
→ Kitchen
→ Ready
```

---

# 71. Critical Payment Test Cases

### Case 1 — Successful payment

```text
Order
WAITING_PAYMENT

Payment
PENDING

Webhook
PAID

Result:

Payment = PAID
Order = CONFIRMED
```

### Case 2 — Duplicate webhook

```text
Webhook A
Webhook A
Webhook A
```

Result:

```text
Only one transaction update
```

### Case 3 — Wrong amount

```text
Order total = Rp50.000
Webhook amount = Rp10.000
```

Result:

```text
Reject
```

### Case 4 — Invalid signature

```text
Webhook
+
Invalid signature
```

Result:

```text
Reject
```

### Case 5 — Payment expired

```text
Payment = EXPIRED
Order = WAITING_PAYMENT
```

Order dapat dibatalkan otomatis berdasarkan business rule.

---

# 72. Recommended Final Architecture

```text
                           INTERNET
                               │
                               ▼
                     ┌──────────────────┐
                     │      CADDY       │
                     │ Reverse Proxy    │
                     └────────┬─────────┘
                              │
                 ┌────────────┴────────────┐
                 │                         │
                 ▼                         ▼
        ┌────────────────┐        ┌────────────────┐
        │    SVELTEKIT   │        │   GO BACKEND   │
        │                │        │                │
        │ Customer       │        │ REST API       │
        │ Admin          │        │ Auth           │
        │ Kitchen        │        │ Order          │
        │ Owner          │        │ Payment        │
        └────────────────┘        │ Webhook        │
                                  │ WebSocket      │
                                  └───────┬────────┘
                                          │
                    ┌─────────────────────┼───────────────────┐
                    │                     │                   │
                    ▼                     ▼                   ▼
             ┌─────────────┐      ┌──────────────┐    ┌─────────────┐
             │ PostgreSQL  │      │   Payment    │    │   Storage   │
             │             │      │   Gateway    │    │             │
             └─────────────┘      └──────┬───────┘    └─────────────┘
                                         │
                                         ▼
                               ┌──────────────────┐
                               │ QRIS / E-Wallet │
                               │ VA / Card       │
                               └──────────────────┘
```

---

# 73. Recommended Technology Stack

| Layer             | Technology              |
| ----------------- | ----------------------- |
| Customer Web      | SvelteKit               |
| Admin Web         | SvelteKit               |
| Kitchen Web       | SvelteKit               |
| Language Frontend | TypeScript              |
| Styling           | Tailwind CSS            |
| Backend           | Golang                  |
| HTTP Framework    | Gin                     |
| API               | REST                    |
| Realtime          | WebSocket / SSE         |
| Database          | PostgreSQL              |
| ORM/Query         | sqlc / pgx / GORM       |
| Authentication    | Session/JWT             |
| Payment           | Payment Gateway         |
| File Storage      | S3-compatible           |
| Reverse Proxy     | Caddy                   |
| Container         | Docker                  |
| CI/CD             | GitHub Actions          |
| Monitoring        | Sentry / Prometheus     |
| Logging           | Structured JSON logging |

---

# 74. Architectural Recommendation

Untuk versi pertama, gunakan:

```text
SvelteKit
       +
Tailwind CSS
       +
Go
       +
Gin
       +
PostgreSQL
       +
Payment Gateway
       +
WebSocket/SSE
       +
Docker
```

Gunakan **modular monolith**, bukan microservices.

Struktur:

```text
                 GO APPLICATION
                       │
       ┌───────────────┼────────────────┐
       │               │                │
      Auth           Orders           Payment
       │               │                │
      Menu            Kitchen         Webhook
       │               │                │
      Table          Reports        Notification
       └───────────────┼────────────────┘
                       │
                   PostgreSQL
```

Ketika traffic dan kompleksitas sudah besar, modul tertentu dapat dipisahkan menjadi service tersendiri.

---

# 75. Definition of Done — MVP

MVP dianggap selesai apabila:

* [ ] Customer dapat scan QR.
* [ ] Sistem mengetahui restaurant dan table.
* [ ] Customer dapat melihat menu.
* [ ] Customer dapat memasukkan menu ke cart.
* [ ] Customer dapat checkout.
* [ ] Backend menghitung total.
* [ ] Payment gateway dapat membuat transaksi.
* [ ] Customer dapat melakukan pembayaran.
* [ ] Webhook dapat diterima.
* [ ] Webhook diverifikasi.
* [ ] Payment dapat berubah menjadi PAID.
* [ ] Order berubah menjadi CONFIRMED.
* [ ] Kitchen menerima order.
* [ ] Kitchen dapat mengubah status.
* [ ] Customer dapat melihat status order.
* [ ] Admin dapat mengelola menu.
* [ ] Admin dapat mengelola meja.
* [ ] Admin dapat generate QR.
* [ ] Role permission berjalan.
* [ ] Tenant isolation berjalan.
* [ ] Duplicate webhook aman.
* [ ] Price manipulation tidak dapat dilakukan.
* [ ] Semua transaksi tercatat di database.
* [ ] Sistem dapat dijalankan menggunakan Docker.
* [ ] HTTPS digunakan pada production.

---

# 76. Kesimpulan

Self-Order Restaurant System dibangun sebagai platform yang menghubungkan:

```text
CUSTOMER
   ↓
QR TABLE
   ↓
DIGITAL MENU
   ↓
ORDER
   ↓
PAYMENT GATEWAY
   ↓
WEBHOOK
   ↓
PAYMENT VERIFICATION
   ↓
KITCHEN
   ↓
ORDER COMPLETED
```

Arsitektur yang direkomendasikan untuk MVP adalah:

```text
SvelteKit
+
Tailwind CSS
+
Golang
+
Gin
+
PostgreSQL
+
Payment Gateway
+
WebSocket/SSE
+
Docker
```

Pendekatan modular monolith dipilih agar pengembangan awal tetap sederhana, tetapi struktur modul memungkinkan sistem berkembang menjadi platform SaaS multi-tenant di kemudian hari.
