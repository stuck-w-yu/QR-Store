<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import { 
		DollarSign, ShoppingCart, TrendingUp, CheckCircle, 
		XCircle, QrCode, Utensils, RefreshCw, ChefHat 
	} from '@lucide/svelte';

	let loading = $state(true);
	let stats = $state<{
		today_revenue: number;
		today_orders: number;
		today_paid: number;
		today_cancelled: number;
		average_order: number;
	}>({
		today_revenue: 0,
		today_orders: 0,
		today_paid: 0,
		today_cancelled: 0,
		average_order: 0
	});

	async function loadStats() {
		try {
			loading = true;
			const data = await api.get<any>('/orders/analytics/today');
			stats = data;
		} catch (e) {
			console.error('Failed to load stats', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadStats();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Dashboard Bisnis</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Ringkasan performa penjualan dan transaksi Resto Nusantara hari ini
			</p>
		</div>

		<button
			type="button"
			onclick={loadStats}
			class="self-start sm:self-auto flex items-center gap-2 bg-white hover:bg-slate-50 text-slate-700 px-4 py-2 rounded-xl border border-slate-200 shadow-xs text-xs font-bold transition-colors"
		>
			<RefreshCw class="w-3.5 h-3.5 {loading ? 'animate-spin' : ''}" />
			<span>Perbarui Data</span>
		</button>
	</div>

	<!-- Metric Cards Grid -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
		<!-- Revenue -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-3">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Pendapatan Hari Ini</span>
				<div class="w-9 h-9 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center">
					<DollarSign class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-2xl font-black text-slate-900 font-['Outfit'] tracking-tight">
					{formatRupiah(stats.today_revenue)}
				</div>
				<span class="text-[11px] text-emerald-600 font-semibold flex items-center gap-1 mt-1">
					<TrendingUp class="w-3 h-3" />
					<span>Cashless Real-time</span>
				</span>
			</div>
		</div>

		<!-- Total Orders -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-3">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Total Pesanan</span>
				<div class="w-9 h-9 rounded-xl bg-blue-100 text-blue-600 flex items-center justify-center">
					<ShoppingCart class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-2xl font-black text-slate-900 font-['Outfit'] tracking-tight">
					{stats.today_orders}
				</div>
				<span class="text-[11px] text-slate-500 font-medium block mt-1">
					Semua order yang dibuat
				</span>
			</div>
		</div>

		<!-- Average Order -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-3">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Rata-rata Order</span>
				<div class="w-9 h-9 rounded-xl bg-purple-100 text-purple-600 flex items-center justify-center">
					<TrendingUp class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-2xl font-black text-slate-900 font-['Outfit'] tracking-tight">
					{formatRupiah(stats.average_order)}
				</div>
				<span class="text-[11px] text-slate-500 font-medium block mt-1">
					Per transaksi sukses
				</span>
			</div>
		</div>

		<!-- Paid vs Cancelled -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-3">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Status Pembayaran</span>
				<div class="w-9 h-9 rounded-xl bg-emerald-100 text-emerald-600 flex items-center justify-center">
					<CheckCircle class="w-5 h-5" />
				</div>
			</div>
			<div class="flex items-center gap-4">
				<div>
					<div class="text-xl font-black text-emerald-600 font-['Outfit']">{stats.today_paid}</div>
					<span class="text-[10px] text-slate-400 font-bold uppercase">Berhasil</span>
				</div>
				<div class="w-px h-8 bg-slate-200"></div>
				<div>
					<div class="text-xl font-black text-red-500 font-['Outfit']">{stats.today_cancelled}</div>
					<span class="text-[10px] text-slate-400 font-bold uppercase">Batal</span>
				</div>
			</div>
		</div>
	</div>

	<!-- Quick Action Shortcuts -->
	<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
		<a
			href="/admin/orders"
			class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs hover:border-orange-500 hover:shadow-md transition-all group"
		>
			<div class="w-12 h-12 rounded-2xl bg-orange-50 text-orange-600 flex items-center justify-center mb-4 group-hover:scale-105 transition-transform">
				<ShoppingCart class="w-6 h-6" />
			</div>
			<h3 class="font-bold text-slate-900 text-base mb-1">Pesanan Masuk</h3>
			<p class="text-xs text-slate-500">Lihat dan monitor status pesanan pelanggan secara langsung.</p>
		</a>

		<a
			href="/kitchen/display"
			target="_blank"
			class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs hover:border-amber-500 hover:shadow-md transition-all group"
		>
			<div class="w-12 h-12 rounded-2xl bg-amber-50 text-amber-600 flex items-center justify-center mb-4 group-hover:scale-105 transition-transform">
				<ChefHat class="w-6 h-6" />
			</div>
			<h3 class="font-bold text-slate-900 text-base mb-1">Kitchen Board (KDS)</h3>
			<p class="text-xs text-slate-500">Buka tampilan layar sentuh dapur untuk proses pembuatan makanan.</p>
		</a>

		<a
			href="/admin/tables"
			class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs hover:border-blue-500 hover:shadow-md transition-all group"
		>
			<div class="w-12 h-12 rounded-2xl bg-blue-50 text-blue-600 flex items-center justify-center mb-4 group-hover:scale-105 transition-transform">
				<QrCode class="w-6 h-6" />
			</div>
			<h3 class="font-bold text-slate-900 text-base mb-1">Meja & QR Code</h3>
			<p class="text-xs text-slate-500">Cetak QR Code unik untuk ditempelkan pada masing-masing meja restoran.</p>
		</a>
	</div>
</div>
