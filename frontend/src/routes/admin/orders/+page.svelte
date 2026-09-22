<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { Order, OrderStatus } from '$lib/types';
	import { 
		ShoppingCart, RefreshCw, Eye, XCircle, 
		Clock, CheckCircle, AlertCircle, Filter 
	} from '@lucide/svelte';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let selectedStatus = $state<string>('');
	let selectedOrder = $state<Order | null>(null);

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
				return { bg: 'bg-slate-100 text-slate-700 border-slate-200', label: 'Selesai' };
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

		<div class="flex items-center gap-3">
			<!-- Status Filter -->
			<select
				bind:value={selectedStatus}
				onchange={loadOrders}
				class="text-xs font-semibold px-3 py-2 rounded-xl border border-slate-200 bg-white focus:outline-none focus:border-orange-500"
			>
				<option value="">Semua Status</option>
				<option value="WAITING_PAYMENT">Menunggu Bayar</option>
				<option value="CONFIRMED">Diterima</option>
				<option value="PREPARING">Sedang Dimasak</option>
				<option value="READY">Siap Saji</option>
				<option value="COMPLETED">Selesai</option>
				<option value="CANCELLED">Dibatalkan</option>
			</select>

			<button
				type="button"
				onclick={loadOrders}
				class="p-2 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-colors"
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
			<div class="overflow-x-auto">
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
								<td class="p-4 pr-6 text-right space-x-2">
									<button
										type="button"
										onclick={() => (selectedOrder = o)}
										class="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors"
										title="Lihat Rincian"
									>
										<Eye class="w-3.5 h-3.5" />
									</button>

									{#if o.status !== 'COMPLETED' && o.status !== 'CANCELLED'}
										<button
											type="button"
											onclick={() => handleCancelOrder(o.id)}
											class="p-1.5 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
											title="Batalkan Pesanan"
										>
											<XCircle class="w-3.5 h-3.5" />
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
			<div class="bg-white rounded-3xl p-6 w-full max-w-md space-y-4 shadow-xl">
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
			</div>
		</div>
	{/if}
</div>
