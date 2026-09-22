<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import { shiftStore } from '$lib/stores/shift.svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Register, ShiftTransaction } from '$lib/types';
	import { 
		CreditCard, Store, Clock, ArrowUpRight, ArrowDownRight, 
		Lock, RefreshCw, CheckCircle2, AlertCircle, AlertTriangle, 
		TrendingUp, Wallet, QrCode, Receipt, Filter
	} from '@lucide/svelte';

	let loading = $state(true);
	let registers = $state<Register[]>([]);
	let transactions = $state<ShiftTransaction[]>([]);
	let txFilter = $state<string>('');

	// Modal states
	let showOpenModal = $state(false);
	let showCloseModal = $state(false);

	// Open Shift Form
	let selectedRegisterId = $state('');
	let openingBalance = $state(0);
	let openError = $state<string | null>(null);
	let openSubmitting = $state(false);

	// Close Shift Form
	let actualAmount = $state<number>(0);
	let closingNotes = $state('');
	let closeError = $state<string | null>(null);
	let closeSubmitting = $state(false);
	let difference = $derived(
		actualAmount - (shiftStore.summary?.expected_amount || 0)
	);

	async function loadDashboard() {
		loading = true;
		try {
			await shiftStore.loadCurrentShift();
			const regs = await api.get<Register[]>('/registers');
			registers = regs.filter(r => r.status === 'ACTIVE');
			if (registers.length > 0 && !selectedRegisterId) {
				selectedRegisterId = registers[0].id;
			}

			if (shiftStore.currentShift) {
				actualAmount = shiftStore.summary?.expected_amount || 0;
				await loadTransactions();
			}
		} catch (e) {
			console.error('Failed to load dashboard:', e);
		} finally {
			loading = false;
		}
	}

	async function loadTransactions() {
		if (!shiftStore.currentShift) return;
		try {
			const query = txFilter ? `?type=${txFilter}` : '';
			const res = await api.get<{ transactions: ShiftTransaction[] }>(
				`/cashier/shifts/${shiftStore.currentShift.id}/transactions${query}`
			);
			transactions = res.transactions || [];
		} catch (e) {
			console.error('Failed to load transactions:', e);
		}
	}

	async function handleOpenShift() {
		if (!selectedRegisterId) {
			openError = 'Pilih mesin kasir terlebih dahulu';
			return;
		}
		if (openingBalance < 0) {
			openError = 'Saldo awal tidak boleh bernilai negatif';
			return;
		}

		openSubmitting = true;
		openError = null;
		try {
			await shiftStore.openShift(selectedRegisterId, openingBalance);
			showOpenModal = false;
			await loadDashboard();
		} catch (e: any) {
			openError = e?.message || 'Gagal membuka shift';
		} finally {
			openSubmitting = false;
		}
	}

	async function handleCloseShift() {
		closeSubmitting = true;
		closeError = null;
		try {
			await shiftStore.closeShift(actualAmount, closingNotes);
			showCloseModal = false;
			await loadDashboard();
		} catch (e: any) {
			closeError = e?.message || 'Gagal menutup shift';
		} finally {
			closeSubmitting = false;
		}
	}

	onMount(() => {
		loadDashboard();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<div class="flex items-center gap-3">
				<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Kasir POS & Shift</h1>
				{#if shiftStore.currentShift}
					<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-emerald-100 text-emerald-800 border border-emerald-200 animate-pulse">
						<span class="w-2 h-2 rounded-full bg-emerald-600"></span>
						SHIFT AKTIF (OPEN)
					</span>
				{:else}
					<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-slate-100 text-slate-600 border border-slate-200">
						<span class="w-2 h-2 rounded-full bg-slate-400"></span>
						BELUM BUKA SHIFT
					</span>
				{/if}
			</div>
			<p class="text-xs text-slate-500 font-medium mt-1">
				Monitoring transaksi kasir real-time, rekonsiliasi pembayaran, dan penutupan shift
			</p>
		</div>

		<div class="flex items-center gap-3">
			<button
				type="button"
				onclick={loadDashboard}
				class="p-2.5 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-colors flex items-center gap-2 text-xs font-bold"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
				<span class="hidden sm:inline">Segarkan</span>
			</button>

			{#if shiftStore.currentShift}
				<button
					type="button"
					onclick={() => {
						actualAmount = shiftStore.summary?.expected_amount || 0;
						showCloseModal = true;
					}}
					class="px-4 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-xl shadow-md shadow-red-600/20 transition-all font-bold text-xs flex items-center gap-2"
				>
					<Lock class="w-4 h-4" />
					<span>Tutup Kasir</span>
				</button>
			{:else}
				<button
					type="button"
					onclick={() => (showOpenModal = true)}
					class="px-4 py-2.5 bg-orange-600 hover:bg-orange-700 text-white rounded-xl shadow-md shadow-orange-600/20 transition-all font-bold text-xs flex items-center gap-2"
				>
					<Store class="w-4 h-4" />
					<span>Buka Kasir Sekarang</span>
				</button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="bg-white rounded-3xl p-16 text-center border border-slate-200 text-slate-400 text-xs font-semibold">
			<RefreshCw class="w-6 h-6 animate-spin mx-auto mb-3 text-orange-600" />
			Memuat status shift kasir...
		</div>
	{:else if !shiftStore.currentShift}
		<!-- Empty State: No Active Shift -->
		<div class="bg-white rounded-3xl p-10 md:p-14 text-center border border-slate-200/90 shadow-sm relative overflow-hidden">
			<div class="max-w-md mx-auto space-y-4">
				<div class="w-16 h-16 rounded-2xl bg-orange-50 border border-orange-200 flex items-center justify-center mx-auto text-orange-600 shadow-inner">
					<Store class="w-8 h-8" />
				</div>
				<h2 class="text-xl font-black text-slate-900 font-['Outfit']">Belum Ada Shift Kasir yang Dibuka</h2>
				<p class="text-xs text-slate-500 leading-relaxed font-medium">
					Buka shift kasir terlebih dahulu untuk mencatat rekonsiliasi dan transaksi kasir harian.
				</p>
				<div class="p-4 bg-emerald-50/80 border border-emerald-200/80 rounded-2xl text-left text-xs text-emerald-800 space-y-1">
					<div class="font-bold flex items-center gap-1.5 text-emerald-900">
						<CheckCircle2 class="w-4 h-4 text-emerald-600" />
						Operasi Self-Order Tetap Berjalan
					</div>
					<p class="text-[11px] text-emerald-700">
						Pelanggan tetap dapat memesan makanan dan bayar secara mandiri melalui QR code meja tanpa terikat shift kasir.
					</p>
				</div>
				<div class="pt-2">
					<button
						type="button"
						onclick={() => (showOpenModal = true)}
						class="px-6 py-3 bg-orange-600 hover:bg-orange-700 text-white rounded-2xl shadow-lg shadow-orange-600/30 transition-all font-bold text-xs inline-flex items-center gap-2"
					>
						<Store class="w-4 h-4" />
						<span>Buka Shift Kasir Baru</span>
					</button>
				</div>
			</div>
		</div>
	{:else}
		<!-- Active Shift Information Card -->
		<div class="bg-slate-900 text-white rounded-3xl p-6 md:p-8 shadow-xl relative overflow-hidden">
			<div class="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
				<div class="space-y-2">
					<div class="flex items-center gap-2 text-xs font-mono font-bold text-orange-400">
						<span>SHIFT ID: #{shiftStore.currentShift.id.slice(-6).toUpperCase()}</span>
						<span>•</span>
						<span class="text-emerald-400">● LIVE</span>
					</div>
					<h2 class="text-2xl font-black font-['Outfit']">{shiftStore.currentShift.register_name || 'POS Utama'}</h2>
					<div class="flex flex-wrap items-center gap-4 text-xs text-slate-400 font-medium">
						<div class="flex items-center gap-1.5">
							<span class="text-slate-500">Kasir:</span>
							<span class="text-white font-bold">{shiftStore.currentShift.cashier_name || auth.user?.name}</span>
						</div>
						<div class="flex items-center gap-1.5">
							<Clock class="w-3.5 h-3.5 text-slate-500" />
							<span>Dibuka: {new Date(shiftStore.currentShift.opened_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} WIB</span>
						</div>
						<div class="flex items-center gap-1.5">
							<span class="text-slate-500">Saldo Awal:</span>
							<span class="text-white font-bold">{formatRupiah(shiftStore.currentShift.opening_balance)}</span>
						</div>
					</div>
				</div>

				<div class="bg-slate-800/80 backdrop-blur-xs border border-slate-700/80 rounded-2xl p-4 md:p-5 flex items-center gap-6">
					<div>
						<span class="text-[11px] font-bold text-slate-400 block uppercase tracking-wider">Net Penjualan</span>
						<span class="text-2xl md:text-3xl font-black text-orange-400 font-['Outfit']">
							{formatRupiah(shiftStore.summary?.net_sales || 0)}
						</span>
					</div>
					<div class="border-l border-slate-700 pl-6">
						<span class="text-[11px] font-bold text-slate-400 block uppercase tracking-wider">Total Pesanan</span>
						<span class="text-2xl font-black text-white font-['Outfit']">
							{shiftStore.summary?.orders_count || 0}
						</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Summary Metrics Grid -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
				<div class="flex items-center justify-between text-slate-400 mb-2">
					<span class="text-xs font-bold uppercase tracking-wider">Gross Sales</span>
					<ArrowUpRight class="w-4 h-4 text-emerald-500" />
				</div>
				<div class="text-xl font-black text-slate-900 font-['Outfit']">
					{formatRupiah(shiftStore.summary?.gross_sales || 0)}
				</div>
				<p class="text-[11px] text-slate-500 mt-1">Total kotor transaksi pesanan</p>
			</div>

			<div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
				<div class="flex items-center justify-between text-slate-400 mb-2">
					<span class="text-xs font-bold uppercase tracking-wider">Total Refund</span>
					<ArrowDownRight class="w-4 h-4 text-red-500" />
				</div>
				<div class="text-xl font-black text-red-600 font-['Outfit']">
					-{formatRupiah(shiftStore.summary?.refund_total || 0)}
				</div>
				<p class="text-[11px] text-slate-500 mt-1">Pengembalian dana sukses</p>
			</div>

			<div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
				<div class="flex items-center justify-between text-slate-400 mb-2">
					<span class="text-xs font-bold uppercase tracking-wider">Total Void</span>
					<AlertTriangle class="w-4 h-4 text-amber-500" />
				</div>
				<div class="text-xl font-black text-amber-600 font-['Outfit']">
					-{formatRupiah(shiftStore.summary?.void_total || 0)}
				</div>
				<p class="text-[11px] text-slate-500 mt-1">Pembatalan pesanan</p>
			</div>

			<div class="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
				<div class="flex items-center justify-between text-slate-400 mb-2">
					<span class="text-xs font-bold uppercase tracking-wider">Expected Total</span>
					<TrendingUp class="w-4 h-4 text-blue-500" />
				</div>
				<div class="text-xl font-black text-blue-600 font-['Outfit']">
					{formatRupiah(shiftStore.summary?.expected_amount || 0)}
				</div>
				<p class="text-[11px] text-slate-500 mt-1">Net Sales + Saldo Awal</p>
			</div>
		</div>

		<!-- Payment Breakdown -->
		<div class="bg-white rounded-3xl p-6 border border-slate-200/80 shadow-xs space-y-4">
			<div class="flex items-center justify-between">
				<h3 class="font-extrabold text-slate-900 text-sm flex items-center gap-2">
					<Wallet class="w-4 h-4 text-orange-600" />
					Rincian Metode Pembayaran Cashless
				</h3>
				<span class="text-xs text-slate-400 font-medium">Auto-Reconciled</span>
			</div>

			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
				<div class="p-4 rounded-2xl bg-slate-50 border border-slate-100 flex items-center gap-4">
					<div class="w-10 h-10 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center shrink-0">
						<QrCode class="w-5 h-5" />
					</div>
					<div>
						<span class="text-[11px] font-bold text-slate-400 uppercase">QRIS Statis/Dinamis</span>
						<div class="text-base font-black text-slate-900 font-['Outfit']">
							{formatRupiah(shiftStore.summary?.payment_methods?.['QRIS'] || 0)}
						</div>
					</div>
				</div>

				<div class="p-4 rounded-2xl bg-slate-50 border border-slate-100 flex items-center gap-4">
					<div class="w-10 h-10 rounded-xl bg-blue-100 text-blue-600 flex items-center justify-center shrink-0">
						<Wallet class="w-5 h-5" />
					</div>
					<div>
						<span class="text-[11px] font-bold text-slate-400 uppercase">E-Wallet (GoPay/OVO)</span>
						<div class="text-base font-black text-slate-900 font-['Outfit']">
							{formatRupiah(shiftStore.summary?.payment_methods?.['E-Wallet'] || 0)}
						</div>
					</div>
				</div>

				<div class="p-4 rounded-2xl bg-slate-50 border border-slate-100 flex items-center gap-4">
					<div class="w-10 h-10 rounded-xl bg-purple-100 text-purple-600 flex items-center justify-center shrink-0">
						<CreditCard class="w-5 h-5" />
					</div>
					<div>
						<span class="text-[11px] font-bold text-slate-400 uppercase">Virtual Account / Bank</span>
						<div class="text-base font-black text-slate-900 font-['Outfit']">
							{formatRupiah(shiftStore.summary?.payment_methods?.['Virtual Account'] || 0)}
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Shift Transactions Feed -->
		<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
			<div class="p-5 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
				<div>
					<h3 class="font-extrabold text-slate-900 text-sm flex items-center gap-2">
						<Receipt class="w-4 h-4 text-orange-600" />
						Aktivitas & Log Transaksi Shift
					</h3>
					<p class="text-[11px] text-slate-400 font-medium">Rekaman real-time SALE, REFUND, dan VOID pada shift ini</p>
				</div>

				<div class="flex items-center gap-2">
					<select
						bind:value={txFilter}
						onchange={loadTransactions}
						class="text-xs font-semibold px-3 py-1.5 rounded-xl border border-slate-200 bg-slate-50 focus:outline-none focus:border-orange-500"
					>
						<option value="">Semua Tipe</option>
						<option value="SALE">Hanya Penjualan (SALE)</option>
						<option value="REFUND">Hanya Refund</option>
						<option value="VOID">Hanya Void</option>
					</select>
				</div>
			</div>

			{#if transactions.length === 0}
				<div class="text-center py-16 text-slate-400 text-xs font-medium">
					Belum ada transaksi yang tercatat pada shift ini.
				</div>
			{:else}
				<div class="divide-y divide-slate-100">
					{#each transactions as t}
						<div class="p-4 flex items-center justify-between hover:bg-slate-50/50 transition-colors text-xs">
							<div class="flex items-center gap-3">
								{#if t.type === 'SALE'}
									<div class="w-8 h-8 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-[10px]">
										SALE
									</div>
								{:else if t.type === 'REFUND'}
									<div class="w-8 h-8 rounded-xl bg-red-50 text-red-600 flex items-center justify-center font-bold text-[10px]">
										RFND
									</div>
								{:else}
									<div class="w-8 h-8 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center font-bold text-[10px]">
										VOID
									</div>
								{/if}

								<div>
									<div class="font-bold text-slate-900 flex items-center gap-2">
										<span>{t.type}</span>
										{#if t.order_id}
											<span class="text-[11px] font-mono text-slate-400">Order: {t.order_id.slice(-6)}</span>
										{/if}
									</div>
									<div class="text-[11px] text-slate-400">
										{new Date(t.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })} WIB
										{#if t.metadata?.payment_method}
											• via {t.metadata.payment_method}
										{/if}
										{#if t.metadata?.reason}
											• Alasan: "{t.metadata.reason}"
										{/if}
									</div>
								</div>
							</div>

							<div class="text-right">
								<span class="font-black font-['Outfit'] {t.type === 'SALE' ? 'text-emerald-600' : 'text-red-600'}">
									{t.type === 'SALE' ? '+' : '-'}{formatRupiah(t.amount)}
								</span>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Modal: Buka Shift -->
	{#if showOpenModal}
		<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-6 sm:p-8 w-full max-w-md space-y-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
				<div class="flex items-center justify-between border-b border-slate-100 pb-4">
					<div>
						<h3 class="text-lg font-black text-slate-900 font-['Outfit']">Buka Shift Kasir</h3>
						<p class="text-xs text-slate-500 font-medium">Mulai sesi kasir baru untuk monitoring & rekonsiliasi</p>
					</div>
					<button
						type="button"
						onclick={() => (showOpenModal = false)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold p-1 rounded-lg"
					>
						✕
					</button>
				</div>

				{#if openError}
					<div class="p-3 bg-red-50 text-red-600 text-xs rounded-xl flex items-center gap-2 border border-red-200">
						<AlertCircle class="w-4 h-4 shrink-0" />
						<span>{openError}</span>
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); handleOpenShift(); }} class="space-y-4 text-xs font-medium">
					<div class="space-y-1.5">
						<label for="reg-select" class="font-bold text-slate-700">Pilih Mesin Kasir (Register)</label>
						<select
							id="reg-select"
							bind:value={selectedRegisterId}
							class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 bg-slate-50 font-semibold focus:outline-none focus:border-orange-500"
							required
						>
							{#each registers as reg}
								<option value={reg.id}>{reg.name}</option>
							{/each}
						</select>
					</div>

					<div class="space-y-1.5">
						<label for="open-balance" class="font-bold text-slate-700">Saldo Awal Kasir (Cash Float)</label>
						<div class="relative">
							<span class="absolute left-3.5 top-2.5 font-bold text-slate-400">Rp</span>
							<input
								id="open-balance"
								type="number"
								min="0"
								bind:value={openingBalance}
								class="w-full pl-10 pr-3.5 py-2.5 rounded-xl border border-slate-200 font-bold focus:outline-none focus:border-orange-500"
								placeholder="0"
							/>
						</div>
						<p class="text-[11px] text-slate-400">Untuk sistem restoran cashless, saldo awal dapat diisi Rp0.</p>
					</div>

					<div class="pt-4 flex items-center justify-end gap-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (showOpenModal = false)}
							class="px-4 py-2.5 text-slate-600 hover:bg-slate-100 rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="submit"
							disabled={openSubmitting}
							class="px-5 py-2.5 bg-orange-600 hover:bg-orange-700 text-white rounded-xl font-bold shadow-md shadow-orange-600/20 disabled:opacity-50 transition-all"
						>
							{openSubmitting ? 'Membuka...' : 'Konfirmasi Buka Kasir'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Modal: Tutup Shift -->
	{#if showCloseModal && shiftStore.currentShift}
		<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-6 sm:p-8 w-full max-w-lg space-y-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
				<div class="flex items-center justify-between border-b border-slate-100 pb-4">
					<div>
						<h3 class="text-lg font-black text-slate-900 font-['Outfit']">Tutup Shift Kasir (Closing)</h3>
						<p class="text-xs text-slate-500 font-medium">Validasi ringkasan rekonsiliasi sebelum menutup shift kasir</p>
					</div>
					<button
						type="button"
						onclick={() => (showCloseModal = false)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold p-1 rounded-lg"
					>
						✕
					</button>
				</div>

				{#if closeError}
					<div class="p-3 bg-red-50 text-red-600 text-xs rounded-xl flex items-center gap-2 border border-red-200">
						<AlertCircle class="w-4 h-4 shrink-0" />
						<span>{closeError}</span>
					</div>
				{/if}

				<!-- Reconciliation Box -->
				<div class="p-4 bg-slate-50 rounded-2xl border border-slate-100 space-y-2 text-xs">
					<div class="flex justify-between text-slate-500">
						<span>Gross Penjualan</span>
						<span class="font-bold text-slate-800">{formatRupiah(shiftStore.summary?.gross_sales || 0)}</span>
					</div>
					<div class="flex justify-between text-slate-500">
						<span>Total Refund</span>
						<span class="font-bold text-red-600">-{formatRupiah(shiftStore.summary?.refund_total || 0)}</span>
					</div>
					<div class="flex justify-between text-slate-500">
						<span>Total Void</span>
						<span class="font-bold text-amber-600">-{formatRupiah(shiftStore.summary?.void_total || 0)}</span>
					</div>
					<div class="border-t border-slate-200 pt-2 flex justify-between font-extrabold text-slate-900 text-sm">
						<span>Net Penjualan</span>
						<span class="text-orange-600 font-['Outfit']">{formatRupiah(shiftStore.summary?.net_sales || 0)}</span>
					</div>
					<div class="flex justify-between text-slate-500 pt-1 border-t border-slate-200/60">
						<span>Expected Total (Net + Saldo Awal)</span>
						<span class="font-bold text-slate-900">{formatRupiah(shiftStore.summary?.expected_amount || 0)}</span>
					</div>
				</div>

				<form onsubmit={(e) => { e.preventDefault(); handleCloseShift(); }} class="space-y-4 text-xs font-medium">
					<div class="space-y-1.5">
						<label for="close-actual" class="font-bold text-slate-700">Nominal Aktual Tercatat</label>
						<div class="relative">
							<span class="absolute left-3.5 top-2.5 font-bold text-slate-400">Rp</span>
							<input
								id="close-actual"
								type="number"
								min="0"
								bind:value={actualAmount}
								class="w-full pl-10 pr-3.5 py-2.5 rounded-xl border border-slate-200 font-bold focus:outline-none focus:border-orange-500"
							/>
						</div>
						{#if difference !== 0}
							<div class="text-[11px] font-bold {difference > 0 ? 'text-emerald-600' : 'text-red-600'}">
								Selisih (Variance): {difference > 0 ? '+' : ''}{formatRupiah(difference)}
							</div>
						{/if}
					</div>

					<div class="space-y-1.5">
						<label for="close-notes" class="font-bold text-slate-700">Catatan Penutupan Shift (Opsional)</label>
						<textarea
							id="close-notes"
							bind:value={closingNotes}
							rows="2"
							class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 font-normal"
							placeholder="Contoh: Shift berjalan lancar, semua transaksi QRIS sesuai settlement."
						></textarea>
					</div>

					<div class="p-3 bg-amber-50 border border-amber-200 rounded-xl text-[11px] text-amber-800 leading-relaxed font-semibold flex gap-2 items-start">
						<AlertTriangle class="w-4 h-4 shrink-0 text-amber-600 mt-0.5" />
						<span>Peringatan: Shift yang telah ditutup tidak dapat menerima transaksi kasir baru dan tidak dapat dibuka kembali.</span>
					</div>

					<div class="pt-4 flex items-center justify-end gap-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (showCloseModal = false)}
							class="px-4 py-2.5 text-slate-600 hover:bg-slate-100 rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="submit"
							disabled={closeSubmitting}
							class="px-5 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-xl font-bold shadow-md shadow-red-600/20 disabled:opacity-50 transition-all"
						>
							{closeSubmitting ? 'Menutup...' : 'Tutup Shift Sekarang'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
