# QR-Store — Self-Order Restaurant & Cashier System

Sistem pemesanan restoran dan kasir mandiri berbasis QR Code sesuai spesifikasi [PRD/prd.md](file:///home/stuckwyu/dev/code/QR-Store/PRD/prd.md).

## 🚀 Arsitektur & Teknologi

- **Backend**: Golang (Gin), Modular Monolith
- **Database**: PostgreSQL dengan skema multi-tenant & migrasi SQL bernomor (single source of truth, di-embed ke binary)
- **Realtime**: WebSockets (channel pub/sub untuk notifikasi instan dapur & pelanggan)
- **Payment Gateway**: Abstraksi Gateway (Midtrans + Mock Gateway dengan proteksi idempotency webhook)
- **Frontend**: SvelteKit 5 + TypeScript + Tailwind CSS (lucide-svelte, runes mode)
- **Container**: Docker multi-stage & Docker Compose

---

## 💻 Menjalankan dengan Docker Compose (Satu Perintah)

```bash
docker compose up --build
```

Setelah berjalan:
- **Web Portal Utama**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080](http://localhost:8080)
- **Health Check**: [http://localhost:8080/health](http://localhost:8080/health)

---

## 🛠️ Menjalankan Secara Lokal (Development)

### 1. Database & Backend
Pastikan PostgreSQL berjalan lokal pada port 5432:
```bash
# Di direktori backend/
cd backend
go run ./cmd/server
```
*Backend otomatis menjalankan migrasi skema dan melakukan seeding data demo pertama kali.*

### 2. Frontend
```bash
# Di direktori frontend/
cd frontend
npm run dev
```

---

## 📱 URL Pengujian & Simulasi

1. **Simulasi Pelanggan (Scan Meja 01)**:
   - Akses: [http://localhost:3000/order?token=demo-qr-token-table-01](http://localhost:3000/order?token=demo-qr-token-table-01)
   - Pilih hidangan, atur tingkat kepedasan / topping, masukkan ke keranjang, dan lakukan checkout.
   - Pada halaman status, klik tombol **"Simulasi Bayar Instan"** untuk menguji pembayaran QRIS.

2. **Kitchen Display System (Layar Dapur)**:
   - Akses: [http://localhost:3000/kitchen/display](http://localhost:3000/kitchen/display)
   - Tiket pesanan baru yang sudah dibayar akan otomatis muncul secara *real-time* dengan suara lonceng.
   - Koki dapat menekan tombol **"Mulai Masak"** dan **"Siap Disajikan"**. Status pelanggan akan terupdate secara langsung via WebSocket tanpa refresh!

3. **Staff & Admin Dashboard**:
   - Akses: [http://localhost:3000/login](http://localhost:3000/login) atau [http://localhost:3000/admin/dashboard](http://localhost:3000/admin/dashboard)
   - **Akun Demo Bawaan**:
     - **Owner**: `owner@resto.com` / `password123`
     - **Admin**: `admin@resto.com` / `password123`
     - **Kasir**: `cashier@resto.com` / `password123`
     - **Dapur**: `kitchen@resto.com` / `password123`

---

## 🧪 Menjalankan Pengujian Otomatis (Testing)

```bash
cd backend
go test -v ./tests/unit/...
```
Pengujian mencakup:
- Validasi *State Machine* pesanan (mencegah lompatan status ilegal).
- Validasi *Signature* SHA-512 Payment Gateway (Midtrans).
- Verifikasi simulasi pembayaran Mock Gateway.

---

## ☁️ Deployment

### Docker / VPS (Direkomendasikan untuk produksi)

Cara paling mudah: `docker compose up --build` di VPS. Semua layanan (PostgreSQL, backend, frontend, Caddy) berjalan dalam satu jaringan, WebSocket realtime aktif penuh.

### Vercel (⚠️ keterbatasan realtime)

`vercel.json` memetakan `/api`, `/health`, dan `/ws` ke service `backend`. Namun **Vercel Serverless Functions tidak mendukung koneksi WebSocket persisten** — fungsi dibatasi durasi eksekusinya. Konsekuensinya di Vercel:

- Notifikasi realtime (dapur/pelanggan) **tidak akan bertahan**; client akan terus mencoba reconnect lalu menyerah.
- REST API tetap berfungsi normal.
- Untuk realtime, arahkan `VITE_WS_URL` ke backend yang di-host di VPS/Docker (mis. `wss://api.contoh.com/ws`), atau gunakan platform yang mendukung WebSocket (Fly.io, Railway, Render, VPS).

Jika realtime tidak diperlukan, frontend sudah aman: kegagalan koneksi WS tidak memblokir alur pemesanan (data tetap bisa diambil ulang via REST).

### Environment Penting

| Variabel | Wajib | Catatan |
|---|---|---|
| `JWT_SECRET` | ✅ | Minimal 32 karakter. App **gagal start** di produksi bila tidak aman. |
| `PAYMENT_WEBHOOK_SECRET` | ✅ | Dipakai HMAC-SHA256 untuk memverifikasi webhook. |
| `DATABASE_URL` | ✅ | Koneksi PostgreSQL; backend retry dengan backoff saat DB belum siap. |
| `CORS_ORIGINS` | ✅ | Daftar origin dipisah koma. Wildcard + credentials dilarang di produksi. |
| `SEED_DEMO_DATA` | — | Default `false`. Jangan aktifkan di produksi. |
| `VITE_WS_URL` | — | URL WebSocket absolut untuk realtime lintas-domain. |
| `BUCKET_DELETE_SECRET` (GAS) | ✅ | Script Property di Google Apps Script untuk otorisasi hapus file. |
