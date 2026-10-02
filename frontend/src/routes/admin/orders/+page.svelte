<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { Order, OrderStatus } from '$lib/types';
	import { 
		ShoppingCart, RefreshCw, Eye, XCircle, 
		Clock, CheckCircle, AlertCircle, Filter,
		RotateCcw, Ban, Wallet
	} from '@lucide/svelte';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let selectedStatus = $state<string>('');
	let selectedOrder = $state<Order | null>(null);

	// Refund Modal State
	let refundModalOrder = $state<Order | null>(null);
	let refundAmount = $state(0);
	let refundReason = $state('');
	let refundSubmitting = $state(false);

	// Void Modal State
	let voidModalOrder = $state<Order | null>(null);
	let voidReason = $state('');
	let voidSubmitting = $state(false);

	async function handleAcceptCash(o: Order) {
		if (!confirm(`Konfirmasi terima pembayaran TUNAI sebesar ${formatRupiah(o.total)} untuk ${o.order_number}?`)) {
			return;
		}
		try {
			await api.post(`/cashier/orders/${o.id}/payment`, {
				payment_method: 'CASH',
				paid_amount: o.total
			});
			await loadOrders();
			if (selectedOrder?.id === o.id) {
				selectedOrder = null;
			}
			alert(`Pembayaran tunai pesanan ${o.order_number} berhasil diterima! Pesanan diteruskan ke dapur.`);
		} catch (e: any) {
			alert(e?.message || 'Gagal memproses pembayaran kasir');
		}
	}

	async function loadOrders() {
		try {
			loading = true;
			const query = selectedStatus ? `?status=${selectedStatus}` : '';
			const data = await api.get<Order[]>(`/orders${query}`);
			orders = data;
		} catch (e) {
			console.error('Failed to load orders', e);
		} finally {
			loading = false;
		}
	}

	function openRefundModal(o: Order) {
		refundModalOrder = o;
		refundAmount = o.total;
		refundReason = '';
	}

	async function handleProcessRefund() {
		if (!refundModalOrder) return;
		if (refundAmount <= 0 || refundAmount > refundModalOrder.total) {
			alert('Nominal refund tidak valid');
			return;
		}
		if (!refundReason.trim()) {
			alert('Alasan refund wajib diisi');
			return;
		}

		refundSubmitting = true;
		try {
			await api.post(`/orders/${refundModalOrder.id}/refund`, {
				amount: refundAmount,
				reason: refundReason
			});
			refundModalOrder = null;
			await loadOrders();
			alert('Refund berhasil diproses');
		} catch (e: any) {
			alert(e?.message || 'Gagal memproses refund');
		} finally {
			refundSubmitting = false;
		}
	}

	function openVoidModal(o: Order) {
		voidModalOrder = o;
		voidReason = '';
	}

	async function handleProcessVoid() {
		if (!voidModalOrder) return;
		if (!voidReason.trim()) {
			alert('Alasan void/pembatalan wajib diisi');
			return;
		}

		voidSubmitting = true;
		try {
			await api.post(`/orders/${voidModalOrder.id}/void`, {
				reason: voidReason
			});
			voidModalOrder = null;
			await loadOrders();
			alert('Pesanan berhasil divoid');
		} catch (e: any) {
			alert(e?.message || 'Gagal melakukan void');
		} finally {
			voidSubmitting = false;
		}
	}

	async function handleCancelOrder(orderId: string) {
		if (!confirm('Batalkan pesanan ini?')) return;
		try {
			await api.post(`/orders/${orderId}/cancel`);
			await loadOrders();
			if (selectedOrder?.id === orderId) {
				selectedOrder = null;
			}
		} catch (e) {
			alert('Gagal membatalkan pesanan');
		}
	}

	function getStatusBadge(status: OrderStatus) {
		switch (status) {
			case 'WAITING_PAYMENT':
				return { bg: 'bg-amber-50 text-amber-700 border-amber-200', label: 'Menunggu Bayar' };
			case 'CONFIRMED':
				return { bg: 'bg-blue-50 text-blue-700 border-blue-200', label: 'Diterima' };
			case 'PREPARING':
				return { bg: 'bg-purple-50 text-purple-700 border-purple-200', label: 'Dimasak' };
			case 'READY':
				return { bg: 'bg-emerald-50 text-emerald-700 border-emerald-200', label: 'Siap Saji' };
			case 'COMPLETED':
				return { bg: 'bg-slate-100 text-slate-700 border-slate-200', label: 'Selesai (Diantar)' };
			case 'CANCELLED':
				return { bg: 'bg-red-50 text-red-700 border-red-200', label: 'Dibatalkan' };
			default:
				return { bg: 'bg-slate-100 text-slate-700 border-slate-200', label: status };
		}
	}

	onMount(() => {
		loadOrders();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Pesanan Masuk</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Pantau dan kelola seluruh riwayat transaksi pesanan restoran
			</p>
		</div>

		<div class="flex items-center gap-2 sm:gap-3 w-full sm:w-auto">
			<!-- Status Filter -->
			<select
				bind:value={selectedStatus}
				onchange={loadOrders}
				class="text-xs font-semibold px-3 py-2 rounded-xl border border-slate-200 bg-white focus:outline-none focus:border-orange-500 flex-1 sm:flex-initial"
			>
				<option value="">Semua Status</option>
				<option value="WAITING_PAYMENT">Menunggu Bayar</option>
				<option value="CONFIRMED">Diterima</option>
				<option value="PREPARING">Sedang Dimasak</option>
				<option value="READY">Siap Saji</option>
				<option value="COMPLETED">Selesai (Sudah Diantar)</option>
				<option value="CANCELLED">Dibatalkan</option>
			</select>

			<button
				type="button"
				onclick={loadOrders}
				class="p-2 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-colors shrink-0"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Orders Table -->
	<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
		{#if loading}
			<div class="text-center py-20 text-slate-500 text-xs">Memuat data pesanan...</div>
		{:else if orders.length === 0}
			<div class="text-center py-20 text-slate-500 text-xs">Belum ada pesanan pada filter ini</div>
		{:else}
			<!-- Mobile Card List View -->
			<div class="md:hidden divide-y divide-slate-100">
				{#each orders as o}
					{@const badge = getStatusBadge(o.status)}
					<div class="p-4 space-y-3">
						<div class="flex items-start justify-between gap-2">
							<div>
								<span class="font-mono font-bold text-xs text-slate-900 block">{o.order_number}</span>
								<h3 class="font-bold text-sm text-slate-900">{o.table_name || 'Meja'}</h3>
							</div>
							<span class="px-2.5 py-0.5 rounded-full text-[10px] font-bold border {badge.bg} shrink-0">
								{badge.label}
							</span>
						</div>

						<div class="flex items-center justify-between text-xs pt-1">
							<div>
								<span class="text-[11px] text-slate-400 block">Total Transaksi</span>
								<span class="font-black text-sm text-orange-600 font-['Outfit']">{formatRupiah(o.total)}</span>
							</div>
							<div class="text-right">
								<span class="text-[11px] text-slate-400 block">Waktu Pesan</span>
								<span class="font-mono text-[11px] text-slate-600 font-medium">
									{new Date(o.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
								</span>
							</div>
						</div>

						<div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-100">
							{#if o.status === 'WAITING_PAYMENT'}
								<button
									type="button"
									onclick={() => handleAcceptCash(o)}
									class="py-1.5 px-3 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs flex items-center justify-center gap-1.5 transition-colors shadow-xs"
									title="Terima Pembayaran Tunai Kasir"
								>
									<Wallet class="w-3.5 h-3.5" />
									<span>Terima Tunai</span>
								</button>
							{/if}

							<button
								type="button"
								onclick={() => (selectedOrder = o)}
								class="flex-1 py-1.5 px-3 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors"
							>
								<Eye class="w-3.5 h-3.5" />
								<span>Detail</span>
							</button>

							{#if o.status === 'CONFIRMED' || o.status === 'PREPARING' || o.status === 'READY' || o.status === 'COMPLETED'}
								<button
									type="button"
									onclick={() => openRefundModal(o)}
									class="py-1.5 px-3 rounded-xl bg-orange-50 hover:bg-orange-100 text-orange-600 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors"
								>
									<RotateCcw class="w-3.5 h-3.5" />
									<span>Refund</span>
								</button>
							{/if}

							{#if o.status !== 'COMPLETED' && o.status !== 'CANCELLED'}
								<button
									type="button"
									onclick={() => openVoidModal(o)}
									class="py-1.5 px-3 rounded-xl bg-red-50 hover:bg-red-100 text-red-600 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors"
								>
									<Ban class="w-3.5 h-3.5" />
									<span>Void</span>
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>

			<!-- Desktop Table View -->
			<div class="hidden md:block overflow-x-auto">
				<table class="w-full text-left text-xs border-collapse">
					<thead>
						<tr class="bg-slate-50/80 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
							<th class="p-4 pl-6">No. Pesanan</th>
							<th class="p-4">Meja</th>
							<th class="p-4">Total</th>
							<th class="p-4">Status</th>
							<th class="p-4">Waktu</th>
							<th class="p-4 pr-6 text-right">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 font-medium text-slate-700">
						{#each orders as o}
							{@const badge = getStatusBadge(o.status)}
							<tr class="hover:bg-slate-50/50 transition-colors">
								<td class="p-4 pl-6 font-mono font-bold text-slate-900">{o.order_number}</td>
								<td class="p-4 font-bold text-slate-900">{o.table_name || 'Meja'}</td>
								<td class="p-4 font-extrabold text-orange-600 font-['Outfit']">{formatRupiah(o.total)}</td>
								<td class="p-4">
									<span class="px-2.5 py-1 rounded-full text-[10px] font-bold border {badge.bg}">
										{badge.label}
									</span>
								</td>
								<td class="p-4 text-slate-400 font-mono text-[11px]">
									{new Date(o.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
								</td>
								<td class="p-4 pr-6 text-right space-x-1.5">
									{#if o.status === 'WAITING_PAYMENT'}
										<button
											type="button"
											onclick={() => handleAcceptCash(o)}
											class="p-1.5 px-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-[11px] inline-flex items-center gap-1 transition-colors shadow-xs"
											title="Terima Pembayaran Tunai Kasir"
										>
											<Wallet class="w-3 h-3" />
											<span>Terima Kasir</span>
										</button>
									{/if}

									<button
										type="button"
										onclick={() => (selectedOrder = o)}
										class="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors"
										title="Lihat Rincian"
									>
										<Eye class="w-3.5 h-3.5" />
									</button>

									{#if o.status === 'CONFIRMED' || o.status === 'PREPARING' || o.status === 'READY' || o.status === 'COMPLETED'}
										<button
											type="button"
											onclick={() => openRefundModal(o)}
											class="p-1.5 rounded-lg bg-orange-50 hover:bg-orange-100 text-orange-600 transition-colors"
											title="Refund Transaksi"
										>
											<RotateCcw class="w-3.5 h-3.5" />
										</button>
									{/if}

									{#if o.status !== 'COMPLETED' && o.status !== 'CANCELLED'}
										<button
											type="button"
											onclick={() => openVoidModal(o)}
											class="p-1.5 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
											title="Void / Batalkan Pesanan"
										>
											<Ban class="w-3.5 h-3.5" />
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Detail Modal -->
	{#if selectedOrder}
		<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-6 w-full max-w-md space-y-4 shadow-xl max-h-[90vh] overflow-y-auto">
				<div class="flex items-start justify-between border-b border-slate-100 pb-3">
					<div>
						<span class="text-xs font-mono text-slate-400 font-bold block">{selectedOrder.order_number}</span>
						<h3 class="text-base font-extrabold text-slate-900">{selectedOrder.table_name || 'Meja'}</h3>
					</div>
					<button
						type="button"
						onclick={() => (selectedOrder = null)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold"
					>
						Tutup
					</button>
				</div>

				<div class="divide-y divide-slate-100 max-h-60 overflow-y-auto pr-1 text-xs">
					{#each selectedOrder.items || [] as item}
						<div class="py-2 flex justify-between">
							<div>
								<span class="font-bold text-slate-900">{item.quantity}x {item.menu_name_snapshot}</span>
								{#if item.notes}
									<p class="text-[11px] text-slate-400 italic">"{item.notes}"</p>
								{/if}
							</div>
							<span class="font-bold text-slate-700 font-['Outfit']">{formatRupiah(item.subtotal)}</span>
						</div>
					{/each}
				</div>

				<div class="pt-3 border-t border-slate-100 space-y-1 text-xs text-slate-500">
					<div class="flex justify-between">
						<span>Subtotal</span>
						<span class="font-semibold text-slate-900">{formatRupiah(selectedOrder.subtotal)}</span>
					</div>
					<div class="flex justify-between">
						<span>Pajak</span>
						<span class="font-semibold text-slate-900">{formatRupiah(selectedOrder.tax)}</span>
					</div>
					<div class="flex justify-between font-bold text-sm text-slate-900 pt-1">
						<span>Total</span>
						<span class="text-orange-600 font-extrabold font-['Outfit']">{formatRupiah(selectedOrder.total)}</span>
					</div>
				</div>

				{#if selectedOrder.status === 'WAITING_PAYMENT'}
					<div class="pt-2 border-t border-slate-100">
						<button
							type="button"
							onclick={() => selectedOrder && handleAcceptCash(selectedOrder)}
							class="w-full py-2.5 px-4 rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs flex items-center justify-center gap-2 shadow-sm transition-all active:scale-[0.98]"
						>
							<Wallet class="w-4 h-4" />
							<span>Terima Pembayaran Tunai (Kasir)</span>
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Refund Modal -->
	{#if refundModalOrder}
		<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-7 w-full max-w-md space-y-5 shadow-2xl animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
				<div class="flex items-start justify-between border-b border-slate-100 pb-3">
					<div>
						<h3 class="text-base font-extrabold text-slate-900">Refund Transaksi Pesanan</h3>
						<span class="text-xs font-mono text-slate-400 font-bold block">{refundModalOrder.order_number}</span>
					</div>
					<button
						type="button"
						onclick={() => (refundModalOrder = null)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold"
					>
						✕
					</button>
				</div>

				<div class="space-y-4 text-xs font-medium">
					<div class="space-y-1.5">
						<label for="ref-amount" class="font-bold text-slate-700">Nominal Pengembalian Dana</label>
						<div class="relative">
							<span class="absolute left-3.5 top-2.5 font-bold text-slate-400">Rp</span>
							<input
								id="ref-amount"
								type="number"
								min="1"
								max={refundModalOrder.total}
								bind:value={refundAmount}
								class="w-full pl-10 pr-3.5 py-2.5 rounded-xl border border-slate-200 font-bold focus:outline-none focus:border-orange-500"
							/>
						</div>
						<p class="text-[11px] text-slate-400">Maksimum refund: {formatRupiah(refundModalOrder.total)}</p>
					</div>

					<div class="space-y-1.5">
						<label for="ref-reason" class="font-bold text-slate-700">Alasan Refund</label>
						<textarea
							id="ref-reason"
							bind:value={refundReason}
							rows="2"
							class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
							placeholder="Contoh: Menu habis atau pembatalan atas permintaan pelanggan"
							required
						></textarea>
					</div>

					<div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (refundModalOrder = null)}
							class="px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="button"
							onclick={handleProcessRefund}
							disabled={refundSubmitting}
							class="px-5 py-2 bg-orange-600 hover:bg-orange-700 text-white rounded-xl font-bold shadow-md shadow-orange-600/20 disabled:opacity-50 transition-all"
						>
							{refundSubmitting ? 'Memproses...' : 'Proses Refund'}
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- Void Modal -->
	{#if voidModalOrder}
		<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-7 w-full max-w-md space-y-5 shadow-2xl animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
				<div class="flex items-start justify-between border-b border-slate-100 pb-3">
					<div>
						<h3 class="text-base font-extrabold text-slate-900">Void / Batalkan Pesanan</h3>
						<span class="text-xs font-mono text-slate-400 font-bold block">{voidModalOrder.order_number}</span>
					</div>
					<button
						type="button"
						onclick={() => (voidModalOrder = null)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold"
					>
						✕
					</button>
				</div>

				<div class="space-y-4 text-xs font-medium">
					<div class="p-3 bg-red-50 border border-red-200 rounded-xl text-[11px] text-red-800 leading-relaxed font-semibold">
						Pesanan akan dibatalkan (VOID) dan tercatat dalam ledger rekonsiliasi kasir aktif & audit trail.
					</div>

					<div class="space-y-1.5">
						<label for="void-reason" class="font-bold text-slate-700">Alasan Void</label>
						<textarea
							id="void-reason"
							bind:value={voidReason}
							rows="2"
							class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
							placeholder="Contoh: Pesanan ganda (duplicate order) atau salah input meja"
							required
						></textarea>
					</div>

					<div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (voidModalOrder = null)}
							class="px-4 py-2 text-slate-600 hover:bg-slate-100 rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="button"
							onclick={handleProcessVoid}
							disabled={voidSubmitting}
							class="px-5 py-2 bg-red-600 hover:bg-red-700 text-white rounded-xl font-bold shadow-md shadow-red-600/20 disabled:opacity-50 transition-all"
						>
							{voidSubmitting ? 'Memproses...' : 'Konfirmasi Void'}
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
