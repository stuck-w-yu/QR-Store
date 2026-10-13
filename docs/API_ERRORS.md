# API Error Taxonomy — QR-Store

Semua respons error memakai envelope yang sama:

```json
{
  "success": false,
  "error": {
    "code": "ORDER_NOT_FOUND",
    "message": "Human-readable message"
  }
}
```

`code` bersifat stabil dan aman untuk diprogram; `message` hanya untuk ditampilkan.

## Pemetaan Status HTTP

| HTTP | Arti | Kapan dipakai |
|---|---|---|
| 400 | Bad Request | Body/query tidak valid (`INVALID_REQUEST`), binding gagal, nominal kurang (`INSUFFICIENT_PAID_AMOUNT`) |
| 401 | Unauthorized | Token JWT hilang/invalid/kedaluwarsa; webhook tanpa signature valid |
| 403 | Forbidden | Role tidak diizinkan untuk endpoint; akses lintas-tenant |
| 404 | Not Found | Resource tidak ada (`ORDER_NOT_FOUND`, `REGISTER_NOT_FOUND`, dll.) |
| 409 | Conflict | Melanggar aturan state (`ORDER_ALREADY_PAID`, `SHIFT_ALREADY_CLOSED`, transisi status ilegal) |
| 429 | Too Many Requests | Rate limit terlampaui (login/webhook/public) |
| 500 | Internal Server Error | Kesalahan tak terduga; pesan selalu generik (`An internal error occurred. Please try again later.`) |
| 502 | Bad Gateway | Kegagalan dari payment gateway (opsional, mapping masa depan) |

## Kode Error Umum

| Code | HTTP | Sumber |
|---|---|---|
| `INVALID_REQUEST` | 400 | Binding request gagal |
| `UNAUTHORIZED` | 401 | Middleware auth |
| `FORBIDDEN` | 403 | `RequireRoles` |
| `VALIDATION_ERROR` | 400 | Validasi domain |
| `INTERNAL_SERVER_ERROR` | 500 | Fallback; detail hanya di log server |

## Kode Error Domain

| Code | HTTP | Domain |
|---|---|---|
| `ORDER_NOT_FOUND` | 404 | order |
| `ORDER_ALREADY_PAID` | 409 | payment |
| `ORDER_CANCELLED` | 409 | payment |
| `INVALID_STATUS_TRANSITION` | 409 | order (state machine) |
| `INSUFFICIENT_PAID_AMOUNT` | 400 | payment |
| `INVALID_PAYMENT_METHOD` | 400 | payment |
| `PAYMENT_NOT_FOUND` | 404 | payment |
| `REGISTER_NOT_FOUND` | 404 | cashier |
| `REGISTER_HAS_ACTIVE_SHIFT` | 409 | cashier |
| `SHIFT_NOT_FOUND` | 404 | cashier |
| `SHIFT_ALREADY_CLOSED` | 409 | cashier |
| `TABLE_NOT_FOUND` | 404 | table |
| `MENU_NOT_FOUND` | 404 | menu |
| `RESTAURANT_NOT_FOUND` | 404 | restaurant |

## Prinsip Keamanan

- **Tidak ada kebocoran internal**: respons 5xx tidak pernah memuat `err.Error()`; gunakan `response.ServerError(c, code, err)` yang mencatat detail ke log server.
- **Fail-closed**: transisi status tak dikenal → ditolak (`409`); channel WebSocket non-publik tanpa token valid → ditolak.
- **Tenant isolation**: query selalu di-scope `restaurant_id` dari klaim JWT, bukan dari body request.
