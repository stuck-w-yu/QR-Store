<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Order } from '$lib/types';
	import { 
		Flame, Check, BellRing, Clock, AlertTriangle, 
		UtensilsCrossed, RefreshCw, Volume2, VolumeX, CheckCircle2 
	} from '@lucide/svelte';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let soundEnabled = $state(true);
	let ws: WebSocket | null = null;

	// Audio notification using Web Audio API synthesizer
	function playChime() {
		if (!soundEnabled || typeof window === 'undefined') return;
		try {
			const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
			const osc = ctx.createOscillator();
			const gain = ctx.createGain();

			osc.type = 'sine';
			osc.frequency.setValueAtTime(587.33, ctx.currentTime); // D5
			osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.3); // A5

			gain.gain.setValueAtTime(0.3, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.6);

			osc.connect(gain);
			gain.connect(ctx.destination);

			osc.start();
			osc.stop(ctx.currentTime + 0.6);
		} catch (e) {
			console.error('Audio play error', e);
		}
	}

	async function loadOrders() {
		try {
			const data = await api.get<Order[]>('/kitchen/orders');
			// Preserve completed orders if backend query has not yet refreshed
			const existingCompleted = orders.filter((o) => o.status === 'COMPLETED');
			const returnedIds = new Set(data.map((o) => o.id));
			const missingCompleted = existingCompleted.filter((o) => !returnedIds.has(o.id));
			orders = [...data, ...missingCompleted];
		} catch (e) {
			console.error('Failed to load kitchen orders', e);
		} finally {
			loading = false;
		}
	}

	function connectWebSocket() {
		const restoId = auth.user?.restaurant_id || 'rst_nusantara';
		const channel = `restaurant:${restoId}:kitchen`;

		const wsURL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
		try {
			ws = new WebSocket(`${wsURL}?channel=${channel}`);

			ws.onmessage = (event) => {
				try {
					const msg = JSON.parse(event.data);
					if (msg.event === 'NEW_ORDER_CONFIRMED' || msg.event === 'ORDER_STATUS_CHANGED') {
						playChime();
						loadOrders();
					}
				} catch (e) {
					console.error('Invalid ws message in kitchen', e);
				}
			};

			ws.onclose = () => {
				setTimeout(() => {
					if (ws && ws.readyState === WebSocket.CLOSED) {
						connectWebSocket();
					}
				}, 3000);
			};
		} catch (e) {
			console.error('Failed to connect WebSocket in kitchen', e);
		}
	}

	async function handleAccept(orderId: string) {
		await api.post(`/kitchen/orders/${orderId}/accept`);
		await loadOrders();
	}

	async function handleReady(orderId: string) {
		await api.post(`/kitchen/orders/${orderId}/ready`);
		await loadOrders();
	}

	async function handleComplete(orderId: string) {
		try {
			await api.post(`/kitchen/orders/${orderId}/complete`);
			// Optimistically update order status to COMPLETED
			orders = orders.map((o) =>
				o.id === orderId ? { ...o, status: 'COMPLETED', updated_at: new Date().toISOString() } : o
			);
			await loadOrders();
		} catch (e) {
			console.error('Failed to complete order', e);
		}
	}

	function getElapsedMinutes(createdAt: string): number {
		const created = new Date(createdAt).getTime();
		const now = Date.now();
		return Math.floor((now - created) / 60000);
	}

	let confirmedOrders = $derived(orders.filter((o) => o.status === 'CONFIRMED'));
	let preparingOrders = $derived(orders.filter((o) => o.status === 'PREPARING'));
	let readyOrders = $derived(orders.filter((o) => o.status === 'READY'));
	let completedOrders = $derived(orders.filter((o) => o.status === 'COMPLETED'));
	let activeOrdersCount = $derived(confirmedOrders.length + preparingOrders.length + readyOrders.length);

	onMount(async () => {
		if (!auth.initialized) {
			await auth.init();
		}
		if (!auth.user) {
			goto('/login');
			return;
		}

		await loadOrders();
		connectWebSocket();
	});

	onDestroy(() => {
		if (ws) ws.close();
	});
</script>

<div class="h-full flex flex-col space-y-6">
	<!-- Kitchen Subheader Controls -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
		<div class="flex flex-wrap items-center gap-2.5">
			<div class="flex items-center gap-2">
				<span class="text-xs font-bold text-slate-400">Total Tiket Aktif:</span>
				<span class="px-3 py-1 bg-orange-500/20 text-orange-400 font-extrabold rounded-lg text-sm border border-orange-500/30">
					{activeOrdersCount} Pesanan
				</span>
			</div>
			{#if completedOrders.length > 0}
				<div class="flex items-center gap-1.5 px-3 py-1 bg-emerald-500/15 text-emerald-400 font-bold rounded-lg text-xs border border-emerald-500/25">
					<CheckCircle2 class="w-3.5 h-3.5" />
					<span>{completedOrders.length} Selesai Diantar</span>
				</div>
			{/if}
		</div>

		<div class="flex items-center gap-2 sm:gap-3 w-full sm:w-auto">
			<button
				type="button"
				onclick={() => (soundEnabled = !soundEnabled)}
				class="flex-1 sm:flex-initial justify-center flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold border transition-colors {soundEnabled
					? 'bg-emerald-950/60 border-emerald-500/40 text-emerald-400'
					: 'bg-slate-800 border-slate-700 text-slate-400'}"
			>
				{#if soundEnabled}
					<Volume2 class="w-4 h-4" />
					<span>Suara Notifikasi ON</span>
				{:else}
					<VolumeX class="w-4 h-4" />
					<span>Suara OFF</span>
				{/if}
			</button>

			<button
				type="button"
				onclick={loadOrders}
				class="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl border border-slate-700 transition-colors shrink-0"
			>
				<RefreshCw class="w-4 h-4" />
			</button>
		</div>
	</div>

	<!-- 4-Column Kanban Board -->
	<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 sm:gap-6 flex-1 items-start">
		<!-- Column 1: Pesanan Baru (CONFIRMED) -->
		<div class="bg-slate-900/60 rounded-3xl p-4 border border-slate-800 flex flex-col space-y-4">
			<div class="flex items-center justify-between pb-3 border-b border-slate-800">
				<div class="flex items-center gap-2">
					<div class="w-3 h-3 rounded-full bg-blue-500 animate-pulse"></div>
					<h2 class="font-extrabold text-sm text-slate-200 uppercase tracking-wider">Pesanan Baru</h2>
				</div>
				<span class="px-2 py-0.5 bg-blue-500/20 text-blue-400 font-bold text-xs rounded-full">
					{confirmedOrders.length}
				</span>
			</div>

			<div class="space-y-4 overflow-y-auto max-h-[75vh] pr-1">
				{#each confirmedOrders as ticket (ticket.id)}
					{@const mins = getElapsedMinutes(ticket.created_at)}
					<div class="bg-slate-800/90 rounded-2xl p-4 border-2 border-blue-500/40 shadow-lg space-y-3">
						<div class="flex items-start justify-between">
							<div>
								<span class="text-xs font-mono font-bold text-blue-400 block">{ticket.order_number}</span>
								<h3 class="font-extrabold text-base text-white">{ticket.table_name || 'Meja'}</h3>
							</div>
							<div class="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded-md {mins > 15 ? 'bg-red-500/20 text-red-400 font-bold' : 'bg-slate-700 text-slate-300'}">
								<Clock class="w-3 h-3" />
								<span>{mins}m lalu</span>
							</div>
						</div>

						<!-- Items -->
						<div class="divide-y divide-slate-700/60 text-xs">
							{#each ticket.items || [] as it}
								<div class="py-2">
									<div class="flex items-baseline justify-between font-bold text-slate-200">
										<span class="text-sm text-orange-300 font-black mr-2">{it.quantity}x</span>
										<span class="flex-1">{it.menu_name_snapshot}</span>
									</div>
									{#if it.selected_modifiers && it.selected_modifiers.length > 0}
										<div class="text-[11px] text-slate-400 ml-6 mt-0.5">
											{it.selected_modifiers.map(m => m.option_name).join(', ')}
										</div>
									{/if}
									{#if it.notes}
										<div class="text-[11px] text-amber-400 font-medium italic ml-6 mt-0.5">
											Catatan: {it.notes}
										</div>
									{/if}
								</div>
							{/each}
						</div>

						{#if ticket.notes}
							<div class="bg-amber-950/40 border border-amber-500/30 p-2 rounded-xl text-[11px] text-amber-300">
								<strong>Instruksi Meja:</strong> {ticket.notes}
							</div>
						{/if}

						<button
							type="button"
							onclick={() => handleAccept(ticket.id)}
							class="w-full bg-blue-600 hover:bg-blue-500 text-white font-extrabold py-3 rounded-xl shadow-md text-xs flex items-center justify-center gap-2 transition-all active:scale-[0.98]"
						>
							<Flame class="w-4 h-4" />
							<span>Terima & Mulai Masak</span>
						</button>
					</div>
				{:else}
					<div class="text-center py-16 text-slate-600 text-xs">Tidak ada pesanan baru</div>
				{/each}
			</div>
		</div>

		<!-- Column 2: Sedang Dimasak (PREPARING) -->
		<div class="bg-slate-900/60 rounded-3xl p-4 border border-slate-800 flex flex-col space-y-4">
			<div class="flex items-center justify-between pb-3 border-b border-slate-800">
				<div class="flex items-center gap-2">
					<div class="w-3 h-3 rounded-full bg-amber-500 animate-pulse"></div>
					<h2 class="font-extrabold text-sm text-slate-200 uppercase tracking-wider">Sedang Dimasak</h2>
				</div>
				<span class="px-2 py-0.5 bg-amber-500/20 text-amber-400 font-bold text-xs rounded-full">
					{preparingOrders.length}
				</span>
			</div>

			<div class="space-y-4 overflow-y-auto max-h-[75vh] pr-1">
				{#each preparingOrders as ticket (ticket.id)}
					{@const mins = getElapsedMinutes(ticket.created_at)}
					<div class="bg-slate-800/90 rounded-2xl p-4 border-2 border-amber-500/40 shadow-lg space-y-3">
						<div class="flex items-start justify-between">
							<div>
								<span class="text-xs font-mono font-bold text-amber-400 block">{ticket.order_number}</span>
								<h3 class="font-extrabold text-base text-white">{ticket.table_name || 'Meja'}</h3>
							</div>
							<div class="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded-md {mins > 20 ? 'bg-red-500/20 text-red-400 font-bold' : 'bg-slate-700 text-slate-300'}">
								<Clock class="w-3 h-3" />
								<span>{mins}m lalu</span>
							</div>
						</div>

						<div class="divide-y divide-slate-700/60 text-xs">
							{#each ticket.items || [] as it}
								<div class="py-2">
									<div class="flex items-baseline justify-between font-bold text-slate-200">
										<span class="text-sm text-amber-400 font-black mr-2">{it.quantity}x</span>
										<span class="flex-1">{it.menu_name_snapshot}</span>
									</div>
									{#if it.selected_modifiers && it.selected_modifiers.length > 0}
										<div class="text-[11px] text-slate-400 ml-6 mt-0.5">
											{it.selected_modifiers.map(m => m.option_name).join(', ')}
										</div>
									{/if}
									{#if it.notes}
										<div class="text-[11px] text-amber-400 font-medium italic ml-6 mt-0.5">
											Catatan: {it.notes}
										</div>
									{/if}
								</div>
							{/each}
						</div>

						<button
							type="button"
							onclick={() => handleReady(ticket.id)}
							class="w-full bg-amber-500 hover:bg-amber-400 text-slate-950 font-extrabold py-3 rounded-xl shadow-md text-xs flex items-center justify-center gap-2 transition-all active:scale-[0.98]"
						>
							<BellRing class="w-4 h-4" />
							<span>Tandai Siap Disajikan</span>
						</button>
					</div>
				{:else}
					<div class="text-center py-16 text-slate-600 text-xs">Belum ada hidangan yang dimasak</div>
				{/each}
			</div>
		</div>

		<!-- Column 3: Siap Disajikan (READY) -->
		<div class="bg-slate-900/60 rounded-3xl p-4 border border-slate-800 flex flex-col space-y-4">
			<div class="flex items-center justify-between pb-3 border-b border-slate-800">
				<div class="flex items-center gap-2">
					<div class="w-3 h-3 rounded-full bg-emerald-500 animate-pulse"></div>
					<h2 class="font-extrabold text-sm text-slate-200 uppercase tracking-wider">Siap Disajikan</h2>
				</div>
				<span class="px-2 py-0.5 bg-emerald-500/20 text-emerald-400 font-bold text-xs rounded-full">
					{readyOrders.length}
				</span>
			</div>

			<div class="space-y-4 overflow-y-auto max-h-[75vh] pr-1">
				{#each readyOrders as ticket (ticket.id)}
					<div class="bg-slate-800/90 rounded-2xl p-4 border-2 border-emerald-500/40 shadow-lg space-y-3">
						<div class="flex items-start justify-between">
							<div>
								<span class="text-xs font-mono font-bold text-emerald-400 block">{ticket.order_number}</span>
								<h3 class="font-extrabold text-base text-white">{ticket.table_name || 'Meja'}</h3>
							</div>
							<span class="px-2.5 py-0.5 bg-emerald-500/20 text-emerald-400 text-[10px] font-bold rounded-full uppercase border border-emerald-500/30">
								Siap Antar
							</span>
						</div>

						<div class="divide-y divide-slate-700/60 text-xs">
							{#each ticket.items || [] as it}
								<div class="py-1.5 flex justify-between">
									<span class="text-slate-300 font-medium">{it.quantity}x {it.menu_name_snapshot}</span>
								</div>
							{/each}
						</div>

						<button
							type="button"
							onclick={() => handleComplete(ticket.id)}
							class="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-extrabold py-2.5 px-3 rounded-xl shadow-md text-xs flex flex-col items-center justify-center gap-0.5 transition-all active:scale-[0.98]"
						>
							<div class="flex items-center gap-1.5 font-black">
								<Check class="w-4 h-4 stroke-3" />
								<span>Pesanan Selesai (Sudah Diantar)</span>
							</div>
							<span class="text-[10px] text-emerald-100 font-medium opacity-90">
								Klik saat pesanan telah diantar ke meja
							</span>
						</button>
					</div>
				{:else}
					<div class="text-center py-16 text-slate-600 text-xs">Tidak ada hidangan menunggu diantar</div>
				{/each}
			</div>
		</div>

		<!-- Column 4: Pesanan Selesai / Sudah Diantar (COMPLETED) -->
		<div class="bg-slate-900/60 rounded-3xl p-4 border border-slate-800 flex flex-col space-y-4">
			<div class="flex items-center justify-between pb-3 border-b border-slate-800">
				<div class="flex items-center gap-2">
					<div class="w-3 h-3 rounded-full bg-emerald-400"></div>
					<h2 class="font-extrabold text-sm text-slate-200 uppercase tracking-wider">Selesai (Diantar)</h2>
				</div>
				<span class="px-2 py-0.5 bg-slate-700/60 text-slate-300 font-bold text-xs rounded-full">
					{completedOrders.length}
				</span>
			</div>

			<div class="space-y-4 overflow-y-auto max-h-[75vh] pr-1">
				{#each completedOrders as ticket (ticket.id)}
					<div class="bg-slate-800/60 rounded-2xl p-4 border border-slate-700/60 opacity-90 space-y-3">
						<div class="flex items-start justify-between">
							<div>
								<span class="text-xs font-mono font-bold text-slate-400 block">{ticket.order_number}</span>
								<h3 class="font-extrabold text-base text-slate-200">{ticket.table_name || 'Meja'}</h3>
							</div>
							<span class="px-2.5 py-0.5 bg-emerald-500/20 text-emerald-300 text-[10px] font-bold rounded-full uppercase flex items-center gap-1 border border-emerald-500/30">
								<Check class="w-3 h-3" />
								Sudah Diantar
							</span>
						</div>

						<div class="divide-y divide-slate-700/40 text-xs text-slate-400">
							{#each ticket.items || [] as it}
								<div class="py-1.5 flex justify-between">
									<span>{it.quantity}x {it.menu_name_snapshot}</span>
								</div>
							{/each}
						</div>

						<div class="pt-2 border-t border-slate-700/50 flex items-center justify-between text-[11px] text-slate-400">
							<span class="flex items-center gap-1 text-emerald-400 font-semibold">
								<CheckCircle2 class="w-3.5 h-3.5" />
								Pesanan Selesai
							</span>
							<span class="font-mono text-[10px] text-slate-500">
								{new Date(ticket.updated_at || ticket.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
							</span>
						</div>
					</div>
				{:else}
					<div class="text-center py-16 text-slate-600 text-xs">Belum ada pesanan yang diantar</div>
				{/each}
			</div>
		</div>
	</div>
</div>
