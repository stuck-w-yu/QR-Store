<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { api, formatRupiah } from '$lib/api/client';
	import { mockSubscribe } from '$lib/api/mock';
	import type { Order, Payment } from '$lib/types';
	import confetti from 'canvas-confetti';
	import { 
		Clock, CheckCircle2, ChefHat, BellRing, Sparkles, 
		QrCode, ArrowLeft, RefreshCw, AlertCircle, Check 
	} from '@lucide/svelte';

	const orderId = page.params.orderId;

	let loading = $state(true);
	let error = $state<string | null>(null);
	let order = $state<Order | null>(null);
	let payment = $state<Payment | null>(null);
	let paying = $state(false);

	let ws: WebSocket | null = null;

	const steps = [
		{ key: 'WAITING_PAYMENT', label: 'Menunggu Pembayaran', desc: 'Selesaikan transaksi Anda' },
		{ key: 'CONFIRMED', label: 'Pesanan Diterima', desc: 'Diteruskan ke sistem dapur' },
		{ key: 'PREPARING', label: 'Sedang Dimasak', desc: 'Chef sedang menyiapkan hidangan' },
		{ key: 'READY', label: 'Pesanan Siap', desc: 'Siap diantar ke meja Anda' },
		{ key: 'COMPLETED', label: 'Selesai', desc: 'Terima kasih, selamat menikmati!' }
	];

	function getStepIndex(status: string): number {
		switch (status) {
			case 'WAITING_PAYMENT': return 0;
			case 'CONFIRMED': return 1;
			case 'PREPARING': return 2;
			case 'READY': return 3;
			case 'COMPLETED': return 4;
			case 'CANCELLED': return -1;
			default: return 0;
		}
	}

	async function loadOrder() {
		try {
			const o = await api.get<Order>(`/public/orders/${orderId}`);
			order = o;

			if (o.status === 'WAITING_PAYMENT') {
				try {
					const p = await api.post<Payment>(`/orders/${orderId}/payment`);
					payment = p;
				} catch (e) {
					console.error('Failed to create payment session', e);
				}
			}
		} catch (err: any) {
			error = err?.message || 'Gagal memuat status pesanan.';
		} finally {
			loading = false;
		}
	}

	function connectWebSocket() {
		// 1. Mock Realtime Bus (Cross-tab instant sync)
		const unsubscribeMock = mockSubscribe(`order:${orderId}`, (msg) => {
			if (msg.event === 'ORDER_STATUS_CHANGED' || msg.event === 'PAYMENT_PAID') {
				if (msg.data?.order) {
					const prevStatus = order?.status;
					order = msg.data.order;

					if (prevStatus === 'WAITING_PAYMENT' && order?.status === 'CONFIRMED') {
						confetti({ particleCount: 80, spread: 60, origin: { y: 0.6 } });
					} else if (order?.status === 'READY') {
						confetti({ particleCount: 120, spread: 80, origin: { y: 0.5 } });
					}
				}
			}
		});

		// 2. Real Backend WebSocket (when backend is running)
		const wsURL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
		const channel = `order:${orderId}`;

		try {
			ws = new WebSocket(`${wsURL}?channel=${channel}`);

			ws.onopen = () => {
				console.log('Connected to order websocket');
			};

			ws.onmessage = (event) => {
				try {
					const msg = JSON.parse(event.data);
					if (msg.event === 'ORDER_STATUS_CHANGED' || msg.event === 'PAYMENT_PAID') {
						if (msg.data?.order) {
							const prevStatus = order?.status;
							order = msg.data.order;

							// Confetti celebration on payment success or order ready!
							if (prevStatus === 'WAITING_PAYMENT' && order?.status === 'CONFIRMED') {
								confetti({ particleCount: 80, spread: 60, origin: { y: 0.6 } });
							} else if (order?.status === 'READY') {
								confetti({ particleCount: 120, spread: 80, origin: { y: 0.5 } });
							}
						}
					}
				} catch (e) {
					console.error('Invalid ws message', e);
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
			// ws failed, mock bus will handle it
		}

		return unsubscribeMock;
	}

	async function handleSimulatePayment() {
		try {
			paying = true;
			await api.post('/payments/simulate-pay', { order_id: orderId });
			await loadOrder();
			confetti({ particleCount: 100, spread: 70, origin: { y: 0.6 } });
		} catch (err: any) {
			alert(err?.message || 'Gagal simulasi bayar');
		} finally {
			paying = false;
		}
	}

	onMount(() => {
		loadOrder();
		connectWebSocket();
	});

	onDestroy(() => {
		if (ws) {
			ws.close();
		}
	});
</script>

<div class="min-h-screen bg-slate-50 pb-20">
	<!-- Top Bar -->
	<header class="bg-white border-b border-slate-200/80 px-4 py-3.5 sticky top-0 z-30">
		<div class="max-w-md mx-auto flex items-center justify-between">
			<a
				href="/order"
				class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center hover:bg-slate-200"
			>
				<ArrowLeft class="w-4 h-4" />
			</a>
			<h1 class="font-bold text-sm text-slate-800 font-['Outfit']">Status Pesanan</h1>
			<button
				type="button"
				onclick={loadOrder}
				class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center hover:bg-slate-200"
			>
				<RefreshCw class="w-4 h-4" />
			</button>
		</div>
	</header>

	<main class="max-w-md mx-auto p-4 space-y-4">
		{#if loading}
			<div class="text-center py-24 text-slate-500">
				<div class="w-10 h-10 border-4 border-orange-500 border-t-transparent rounded-full animate-spin mx-auto mb-3"></div>
				<p class="text-xs font-semibold">Mengambil informasi pesanan...</p>
			</div>
		{:else if error}
			<div class="p-6 text-center bg-white rounded-2xl border border-red-100 mt-8">
				<AlertCircle class="w-12 h-12 text-red-500 mx-auto mb-2" />
				<h2 class="font-bold text-slate-900 text-base">Pesanan Tidak Ditemukan</h2>
				<p class="text-xs text-slate-500 mt-1 mb-4">{error}</p>
				<a href="/order" class="inline-block text-xs font-bold text-orange-600">Kembali ke Beranda</a>
			</div>
		{:else if order}
			<!-- Order Header Card -->
			<div class="bg-gradient-to-br from-slate-900 via-slate-800 to-orange-950 text-white rounded-3xl p-5 shadow-lg relative overflow-hidden">
				<div class="absolute -right-8 -bottom-8 w-32 h-32 bg-orange-500/20 rounded-full blur-2xl"></div>

				<div class="flex items-start justify-between relative z-10">
					<div>
						<div class="text-[11px] text-orange-300 font-medium">Nomor Pesanan</div>
						<div class="text-lg font-black font-['Outfit'] tracking-wide">{order.order_number}</div>
						{#if order.table_name}
							<div class="text-xs text-slate-300 mt-0.5">{order.table_name}</div>
						{/if}
					</div>

					<div class="text-right">
						<div class="text-[11px] text-orange-300 font-medium">Total Tagihan</div>
						<div class="text-lg font-extrabold text-white font-['Outfit']">{formatRupiah(order.total)}</div>
					</div>
				</div>
			</div>

			<!-- QRIS Payment Box (If Waiting Payment) -->
			{#if order.status === 'WAITING_PAYMENT'}
				<div class="bg-white rounded-3xl p-5 border-2 border-orange-500/40 shadow-md text-center space-y-4">
					<div class="inline-flex items-center gap-1.5 px-3 py-1 bg-amber-50 text-amber-800 rounded-full text-xs font-bold">
						<Clock class="w-3.5 h-3.5" />
						<span>Menunggu Pembayaran Cashless</span>
					</div>

					<div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 inline-block mx-auto">
						{#if payment?.qr_string}
							<div class="w-48 h-48 bg-white border border-slate-300 rounded-xl p-2 mx-auto flex flex-col items-center justify-center">
								<QrCode class="w-36 h-36 text-slate-800" />
								<span class="text-[10px] font-mono text-slate-400 font-bold mt-1">QRIS NASIONAL</span>
							</div>
						{:else}
							<div class="w-48 h-48 bg-white border border-slate-300 rounded-xl p-2 mx-auto flex flex-col items-center justify-center">
								<QrCode class="w-36 h-36 text-slate-800" />
								<span class="text-[10px] font-mono text-slate-400 font-bold mt-1">QRIS NASIONAL</span>
							</div>
						{/if}
					</div>

					<p class="text-xs text-slate-500 max-w-xs mx-auto">
						Buka aplikasi e-wallet (GoPay, OVO, ShopeePay, DANA) atau mobile banking untuk memindai QRIS di atas.
					</p>

					<!-- Instant Mock Payment Trigger for Development / Testing -->
					<div class="pt-2 border-t border-slate-100">
						<button
							type="button"
							onclick={handleSimulatePayment}
							disabled={paying}
							class="w-full bg-emerald-600 hover:bg-emerald-700 text-white font-bold py-3 rounded-xl shadow-md text-xs flex items-center justify-center gap-2 disabled:opacity-50 transition-all"
						>
							{#if paying}
								<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
								<span>Memverifikasi Pembayaran...</span>
							{:else}
								<Check class="w-4 h-4" />
								<span>Simulasi Bayar Instan (Dev/Demo)</span>
							{/if}
						</button>
					</div>
				</div>
			{/if}

			<!-- Progress Stepper -->
			<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-6">
				<h2 class="font-extrabold text-sm text-slate-900 font-['Outfit']">Lacak Status Pesanan</h2>

				<div class="relative pl-7 space-y-7 before:absolute before:left-3 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200">
					{#each steps as step, idx}
						{@const currentIdx = getStepIndex(order.status)}
						{@const isPast = currentIdx > idx}
						{@const isCurrent = currentIdx === idx}
						{@const isPending = currentIdx < idx}

						<div class="relative">
							<!-- Circle Indicator -->
							<div
								class="absolute -left-7 top-0.5 w-6 h-6 rounded-full flex items-center justify-center border-2 text-xs transition-all {isPast
									? 'bg-orange-600 border-orange-600 text-white'
									: isCurrent
									? 'bg-white border-orange-600 text-orange-600 ring-4 ring-orange-100 font-bold'
									: 'bg-white border-slate-300 text-slate-300'}"
							>
								{#if isPast}
									<Check class="w-3.5 h-3.5 stroke-[3]" />
								{:else}
									<span class="text-[11px]">{idx + 1}</span>
								{/if}
							</div>

							<div>
								<div class="text-sm font-bold {isCurrent ? 'text-orange-600 font-extrabold' : isPast ? 'text-slate-800' : 'text-slate-400'}">
									{step.label}
								</div>
								<div class="text-xs {isCurrent ? 'text-slate-600' : 'text-slate-400'} mt-0.5">
									{step.desc}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Order Items Details -->
			<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs space-y-3">
				<h3 class="font-extrabold text-xs uppercase tracking-wider text-slate-400">Rincian Item</h3>

				<div class="divide-y divide-slate-100">
					{#each order.items || [] as item}
						<div class="py-2.5 flex items-start justify-between gap-3 text-xs">
							<div>
								<div class="font-bold text-slate-900">{item.quantity}x {item.menu_name_snapshot}</div>
								{#if item.selected_modifiers && item.selected_modifiers.length > 0}
									<div class="text-[11px] text-slate-500 mt-0.5">
										{item.selected_modifiers.map(m => m.option_name).join(', ')}
									</div>
								{/if}
								{#if item.notes}
									<div class="text-[11px] text-slate-400 italic">"{item.notes}"</div>
								{/if}
							</div>
							<span class="font-bold text-slate-800 font-['Outfit']">{formatRupiah(item.subtotal)}</span>
						</div>
					{/each}
				</div>

				<div class="pt-2 border-t border-slate-100 space-y-1 text-xs text-slate-500">
					<div class="flex justify-between">
						<span>Subtotal</span>
						<span class="font-semibold text-slate-900">{formatRupiah(order.subtotal)}</span>
					</div>
					<div class="flex justify-between">
						<span>Pajak</span>
						<span class="font-semibold text-slate-900">{formatRupiah(order.tax)}</span>
					</div>
					{#if order.service_charge > 0}
						<div class="flex justify-between">
							<span>Layanan</span>
							<span class="font-semibold text-slate-900">{formatRupiah(order.service_charge)}</span>
						</div>
					{/if}
					<div class="flex justify-between text-sm font-bold text-slate-900 pt-1">
						<span>Total</span>
						<span class="text-orange-600 font-extrabold font-['Outfit']">{formatRupiah(order.total)}</span>
					</div>
				</div>
			</div>
		{/if}
	</main>
</div>
