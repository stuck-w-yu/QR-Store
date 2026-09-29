<script lang="ts">
	import { 
		QrCode, ChefHat, LayoutDashboard, Utensils, 
		CheckCircle, ArrowRight, Store, ShieldCheck, Zap, Sparkles 
	} from '@lucide/svelte';
</script>

<div class="min-h-screen bg-slate-900 text-white selection:bg-orange-500 font-['Plus_Jakarta_Sans',sans-serif] relative overflow-hidden">
	<!-- Ambient Background Glows -->
	<div class="absolute top-0 left-1/2 -translate-x-1/2 w-200 h-125 bg-linear-to-b from-orange-600/20 to-transparent blur-3xl pointer-events-none"></div>

	<!-- Navigation Bar -->
	<header class="max-w-6xl mx-auto p-4 sm:p-6 flex items-center justify-between relative z-10 gap-2">
		<div class="flex items-center gap-2.5 sm:gap-3">
			<div class="w-10 h-10 sm:w-11 sm:h-11 rounded-2xl bg-linear-to-tr from-orange-600 to-amber-500 flex items-center justify-center shadow-lg shadow-orange-500/30 shrink-0">
				<Store class="w-5 h-5 sm:w-6 sm:h-6 text-white" />
			</div>
			<div>
				<span class="text-base sm:text-lg font-black tracking-tight font-['Outfit'] block">QR-Store</span>
				<span class="text-[9px] sm:text-[10px] text-orange-400 font-bold uppercase tracking-wider block -mt-1">Self-Order System</span>
			</div>
		</div>

		<div class="flex items-center gap-2 sm:gap-3 shrink-0">
			<a
				href="/superadmin"
				class="text-xs font-bold text-indigo-300 hover:text-indigo-200 bg-indigo-950/60 border border-indigo-700/50 px-2.5 sm:px-3 py-1.5 rounded-xl transition-colors hidden sm:inline-flex items-center gap-1.5"
			>
				<Sparkles class="w-3.5 h-3.5 text-indigo-400" />
				<span>Superadmin HQ</span>
			</a>
			<a
				href="/login"
				class="text-xs font-bold text-slate-300 hover:text-white px-2.5 sm:px-4 py-2 rounded-xl transition-colors"
			>
				Staff Login
			</a>
			<a
				href="/order?token=demo-qr-token-table-01"
				class="text-xs font-bold bg-orange-600 hover:bg-orange-500 text-white px-3.5 sm:px-5 py-2.5 rounded-xl shadow-lg shadow-orange-600/30 transition-all active:scale-95"
			>
				Mulai Demo
			</a>
		</div>
	</header>

	<!-- Hero Section -->
	<main class="max-w-5xl mx-auto px-4 sm:px-6 pt-8 sm:pt-12 pb-20 sm:pb-24 relative z-10 text-center space-y-6 sm:space-y-8">
		<div class="inline-flex items-center gap-2 px-3.5 sm:px-4 py-1.5 rounded-full bg-white/10 border border-white/15 backdrop-blur-md text-xs font-semibold text-orange-300">
			<Zap class="w-3.5 h-3.5 text-amber-400 shrink-0" />
			<span class="text-[11px] sm:text-xs">Platform Pemesanan Meja Mandiri & Kasir Cashless</span>
		</div>

		<h1 class="text-3xl sm:text-5xl md:text-6xl font-black font-['Outfit'] tracking-tight max-w-3xl mx-auto leading-[1.15]">
			Solusi Digitalisasi Restoran <br class="hidden sm:inline" />
			<span class="bg-linear-to-r from-orange-400 via-amber-300 to-yellow-400 bg-clip-text text-transparent">
				Cepat, Praktis, & Real-time
			</span>
		</h1>

		<p class="text-slate-400 text-sm sm:text-base max-w-2xl mx-auto leading-relaxed">
			Pelanggan memindai QR Code di meja, memilih hidangan favorit, membayar via QRIS secara instan, dan pesanan otomatis langsung diterima di dapur tanpa antrean.
		</p>

		<!-- Interactive Demo Portals Grid -->
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6 pt-8 text-left">
			<!-- 1. Customer Self-Order -->
			<a
				href="/order?token=demo-qr-token-table-01"
				class="bg-slate-800/80 border border-slate-700/80 hover:border-orange-500/80 rounded-3xl p-6 transition-all duration-200 hover:-translate-y-1 shadow-lg hover:shadow-orange-500/10 group flex flex-col justify-between space-y-6"
			>
				<div class="space-y-4">
					<div class="w-12 h-12 rounded-2xl bg-orange-500/20 text-orange-400 flex items-center justify-center ring-1 ring-orange-500/30">
						<QrCode class="w-6 h-6" />
					</div>
					<div>
						<h3 class="font-bold text-lg text-white font-['Outfit'] group-hover:text-orange-400 transition-colors">
							Pelanggan (Meja 01)
						</h3>
						<p class="text-xs text-slate-400 mt-1 leading-relaxed">
							Simulasi scan QR meja, lihat katalog digital, atur modifier hidangan, keranjang, dan bayar dengan QRIS.
						</p>
					</div>
				</div>

				<div class="flex items-center gap-2 text-xs font-bold text-orange-400 group-hover:translate-x-1 transition-transform">
					<span>Buka Halaman Pelanggan</span>
					<ArrowRight class="w-4 h-4" />
				</div>
			</a>

			<!-- 2. Kitchen Display Board -->
			<a
				href="/kitchen/display"
				class="bg-slate-800/80 border border-slate-700/80 hover:border-amber-500/80 rounded-3xl p-6 transition-all duration-200 hover:-translate-y-1 shadow-lg hover:shadow-amber-500/10 group flex flex-col justify-between space-y-6"
			>
				<div class="space-y-4">
					<div class="w-12 h-12 rounded-2xl bg-amber-500/20 text-amber-400 flex items-center justify-center ring-1 ring-amber-500/30">
						<ChefHat class="w-6 h-6" />
					</div>
					<div>
						<h3 class="font-bold text-lg text-white font-['Outfit'] group-hover:text-amber-400 transition-colors">
							Kitchen Display (KDS)
						</h3>
						<p class="text-xs text-slate-400 mt-1 leading-relaxed">
							Papan kanban interaktif dapur. Terima pesanan yang sudah dibayar, mulai proses masak, dan tandai hidangan siap saji.
						</p>
					</div>
				</div>

				<div class="flex items-center gap-2 text-xs font-bold text-amber-400 group-hover:translate-x-1 transition-transform">
					<span>Buka Layar Dapur</span>
					<ArrowRight class="w-4 h-4" />
				</div>
			</a>

			<!-- 3. Admin & Owner Portal -->
			<a
				href="/admin/dashboard"
				class="bg-slate-800/80 border border-slate-700/80 hover:border-blue-500/80 rounded-3xl p-6 transition-all duration-200 hover:-translate-y-1 shadow-lg hover:shadow-blue-500/10 group flex flex-col justify-between space-y-6"
			>
				<div class="space-y-4">
					<div class="w-12 h-12 rounded-2xl bg-blue-500/20 text-blue-400 flex items-center justify-center ring-1 ring-blue-500/30">
						<LayoutDashboard class="w-6 h-6" />
					</div>
					<div>
						<h3 class="font-bold text-lg text-white font-['Outfit'] group-hover:text-blue-400 transition-colors">
							Admin & Kasir Portal
						</h3>
						<p class="text-xs text-slate-400 mt-1 leading-relaxed">
							Dashboard statistik omset harian, CRUD menu, pengaturan meja, cetak & download QR Code, dan monitoring transaksi.
						</p>
					</div>
				</div>

				<div class="flex items-center gap-2 text-xs font-bold text-blue-400 group-hover:translate-x-1 transition-transform">
					<span>Buka Dashboard Staff</span>
					<ArrowRight class="w-4 h-4" />
				</div>
			</a>
		</div>

		<!-- 4. Superadmin Platform Console Banner -->
		<a
			href="/superadmin"
			class="block bg-linear-to-r from-indigo-950/80 via-slate-900 to-purple-950/80 border border-indigo-500/30 hover:border-indigo-400/70 rounded-3xl p-6 sm:p-7 text-left transition-all duration-200 hover:-translate-y-1 shadow-xl hover:shadow-indigo-500/10 group relative overflow-hidden"
		>
			<div class="absolute -right-8 -bottom-8 w-48 h-48 bg-indigo-500/10 rounded-full blur-2xl pointer-events-none"></div>
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-5 relative z-10">
				<div class="flex items-start gap-4">
					<div class="w-12 h-12 rounded-2xl bg-indigo-600/30 border border-indigo-500/40 text-indigo-400 flex items-center justify-center shrink-0">
						<Sparkles class="w-6 h-6" />
					</div>
					<div>
						<div class="flex items-center gap-2">
							<h3 class="font-bold text-lg text-white font-['Outfit'] group-hover:text-indigo-300 transition-colors">
								Superadmin HQ & Tenant Console
							</h3>
							<span class="px-2 py-0.5 rounded-full text-[10px] font-extrabold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 uppercase tracking-wide">
								NEW
							</span>
						</div>
						<p class="text-xs text-slate-400 mt-1 max-w-2xl leading-relaxed">
							Monitoring metrik platform QR Store se-Indonesia, GMV nasional, pembekuan status outlet (freeze/suspend), konfigurasi paket langganan (Starter, Pro, Enterprise), serta pendaftaran business owner & cabang restoran baru secara otomatis.
						</p>
					</div>
				</div>
				<div class="flex items-center gap-2 text-xs font-bold text-indigo-400 group-hover:translate-x-1 transition-transform shrink-0">
					<span>Buka Superadmin Console</span>
					<ArrowRight class="w-4 h-4" />
				</div>
			</div>
		</a>

		<!-- System Architecture Highlights -->
		<div class="pt-16 border-t border-slate-800 grid grid-cols-2 sm:grid-cols-4 gap-6 text-slate-400 text-xs">
			<div class="flex flex-col items-center gap-1.5">
				<ShieldCheck class="w-5 h-5 text-orange-400" />
				<span class="font-bold text-white">Backend Price Integrity</span>
				<span>Perhitungan harga di sisi server</span>
			</div>
			<div class="flex flex-col items-center gap-1.5">
				<Zap class="w-5 h-5 text-amber-400" />
				<span class="font-bold text-white">Real-time WebSockets</span>
				<span>Notifikasi instan dapur & pelanggan</span>
			</div>
			<div class="flex flex-col items-center gap-1.5">
				<CheckCircle class="w-5 h-5 text-emerald-400" />
				<span class="font-bold text-white">Webhook Idempotency</span>
				<span>Proteksi duplikasi payment gateway</span>
			</div>
			<div class="flex flex-col items-center gap-1.5">
				<Store class="w-5 h-5 text-purple-400" />
				<span class="font-bold text-white">Multi-Tenant Ready</span>
				<span>Arsitektur modular monolith Go</span>
			</div>
		</div>
	</main>
</div>
