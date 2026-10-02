<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { api, formatRupiah } from '$lib/api/client';
	import type { Order, Payment } from '$lib/types';
	import confetti from 'canvas-confetti';
	import { 
		Clock, CheckCircle2, ChefHat, BellRing, Sparkles, 
		QrCode, ArrowLeft, RefreshCw, AlertCircle, Check,
		Receipt, Printer, Store, X, Eye, Wallet
	} from '@lucide/svelte';

	const orderId = page.params.orderId;

	let loading = $state(true);
	let error = $state<string | null>(null);
	let order = $state<Order | null>(null);
	let payment = $state<Payment | null>(null);
	let showInvoiceModal = $state(false);
	let invoiceAutoShown = $state(false);
	let selectedMethod = $state<'QRIS' | 'CASH'>('QRIS');

	const isPaid = $derived(!!(order && order.status !== 'WAITING_PAYMENT' && order.status !== 'CANCELLED'));

	let ws: WebSocket | null = null;

	const steps = [
		{ key: 'WAITING_PAYMENT', label: 'Menunggu Pembayaran', desc: 'Selesaikan transaksi Anda' },
		{ key: 'CONFIRMED', label: 'Pesanan Diterima', desc: 'Diteruskan ke sistem dapur' },
		{ key: 'PREPARING', label: 'Sedang Dimasak', desc: 'Chef sedang menyiapkan hidangan' },
		{ key: 'READY', label: 'Pesanan Siap', desc: 'Siap diantar ke meja Anda' },
		{ key: 'COMPLETED', label: 'Pesanan Selesai (Sudah Diantar)', desc: 'Hidangan sudah diantar ke meja Anda. Selamat menikmati!' }
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

	function formatDateTime(dateStr?: string | null): string {
		if (!dateStr) return new Date().toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' });
		try {
			const d = new Date(dateStr);
			return d.toLocaleString('id-ID', {
				day: 'numeric',
				month: 'short',
				year: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return String(dateStr);
		}
	}

	function printInvoice() {
		showInvoiceModal = true;
		setTimeout(() => {
			window.print();
		}, 200);
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
			} else {
				// Fetch existing payment details for the invoice
				try {
					const p = await api.get<Payment>(`/orders/${orderId}/payment`);
					payment = p;
				} catch (e) {
					console.log('Payment record not found or not created yet');
				}

				// Automatically show invoice once upon visiting completed order
				if (!invoiceAutoShown) {
					invoiceAutoShown = true;
					showInvoiceModal = true;
				}
			}
		} catch (err: any) {
			error = err?.message || 'Gagal memuat status pesanan.';
		} finally {
			loading = false;
		}
	}

	function connectWebSocket() {
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

							// Confetti celebration & show invoice on payment success!
							if (prevStatus === 'WAITING_PAYMENT' && order?.status === 'CONFIRMED') {
								api.get<Payment>(`/orders/${orderId}/payment`).then((p) => {
									payment = p;
								}).catch(() => {});
								invoiceAutoShown = true;
								showInvoiceModal = true;
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
			console.error('Failed to connect order WebSocket', e);
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
			<div class="bg-linear-to-br from-slate-900 via-slate-800 to-orange-950 text-white rounded-3xl p-5 shadow-lg relative overflow-hidden">
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

			<!-- Payment Method & Instructions (If Waiting Payment) -->
			{#if order.status === 'WAITING_PAYMENT'}
				<div class="bg-white rounded-3xl p-5 border-2 border-orange-500/40 shadow-md text-center space-y-4">
					<!-- Method Tabs Switcher -->
					<div>
						<span class="block text-[11px] font-bold text-slate-400 uppercase tracking-wider mb-2 text-left">
							Pilih Metode Pembayaran:
						</span>
						<div class="grid grid-cols-2 gap-1.5 p-1 bg-slate-100 rounded-2xl">
							<button
								type="button"
								onclick={() => (selectedMethod = 'QRIS')}
								class="flex items-center justify-center gap-2 py-2.5 px-3 rounded-xl text-xs font-bold transition-all {selectedMethod === 'QRIS'
									? 'bg-white text-orange-600 shadow-xs'
									: 'text-slate-500 hover:text-slate-800'}"
							>
								<QrCode class="w-4 h-4" />
								<span>QRIS (Cashless)</span>
							</button>
							<button
								type="button"
								onclick={() => (selectedMethod = 'CASH')}
								class="flex items-center justify-center gap-2 py-2.5 px-3 rounded-xl text-xs font-bold transition-all {selectedMethod === 'CASH'
									? 'bg-white text-emerald-700 shadow-xs'
									: 'text-slate-500 hover:text-slate-800'}"
							>
								<Wallet class="w-4 h-4" />
								<span>Tunai di Kasir</span>
							</button>
						</div>
					</div>

					{#if selectedMethod === 'QRIS'}
						<!-- QRIS View -->
						<div class="space-y-4 animate-in fade-in duration-150">
							<div class="inline-flex items-center gap-1.5 px-3 py-1 bg-amber-50 text-amber-800 rounded-full text-xs font-bold">
								<Clock class="w-3.5 h-3.5 text-amber-600" />
								<span>Menunggu Pembayaran Cashless</span>
							</div>

							<div class="bg-slate-50 p-4 rounded-2xl border border-slate-200 inline-block mx-auto">
								<div class="w-48 h-48 bg-white border border-slate-300 rounded-xl p-2 mx-auto flex flex-col items-center justify-center">
									<QrCode class="w-36 h-36 text-slate-800" />
									<span class="text-[10px] font-mono text-slate-400 font-bold mt-1">QRIS NASIONAL</span>
								</div>
							</div>

							<p class="text-xs text-slate-500 max-w-xs mx-auto">
								Buka aplikasi e-wallet (GoPay, OVO, ShopeePay, DANA) atau mobile banking untuk memindai QRIS di atas.
							</p>


						</div>
					{:else}
						<!-- Cash / Tunai View -->
						<div class="space-y-4 animate-in fade-in duration-150 text-left">
							<div class="flex items-center justify-between">
								<div class="inline-flex items-center gap-1.5 px-3 py-1 bg-amber-50 text-amber-800 rounded-full text-xs font-bold">
									<Clock class="w-3.5 h-3.5 text-amber-600" />
									<span>Menunggu Pembayaran Tunai</span>
								</div>
								<span class="px-2.5 py-0.5 rounded-md text-[10px] font-extrabold uppercase bg-amber-100 text-amber-800 border border-amber-200">
									BAYAR DI KASIR
								</span>
							</div>

							<!-- Instruction Notice -->
							<div class="bg-amber-50/90 border border-amber-200/90 rounded-2xl p-4 space-y-1.5">
								<div class="flex items-center gap-2 text-amber-900 font-bold text-xs">
									<AlertCircle class="w-4 h-4 text-amber-600 shrink-0" />
									<span>Tunjukkan Invoice Tagihan ke Kasir</span>
								</div>
								<p class="text-[11px] text-amber-800 leading-relaxed font-medium">
									Silakan menuju meja kasir dan tunjukkan <strong>nomor pesanan</strong> atau <strong>invoice tagihan</strong> di bawah ini untuk membayar tunai. Setelah kasir menerima pembayaran, pesanan Anda akan langsung diproses oleh dapur.
								</p>
							</div>

							<!-- Billing Slip / Unpaid Invoice Card Preview -->
							<div class="bg-slate-50 border border-slate-200 rounded-2xl p-4 space-y-3">
								<div class="flex items-center justify-between border-b border-dashed border-slate-200 pb-2">
									<div class="flex items-center gap-1.5">
										<Store class="w-4 h-4 text-orange-600" />
										<span class="font-bold text-slate-800 text-xs">Tagihan Meja {order.table_name || 'Meja'}</span>
									</div>
									<span class="text-[10px] font-extrabold font-mono text-amber-800 bg-amber-100 px-2 py-0.5 rounded border border-amber-200">
										BELUM BAYAR
									</span>
								</div>

								<div class="flex items-center justify-between text-xs">
									<div>
										<span class="text-[10px] text-slate-400 block font-bold uppercase">No. Pesanan</span>
										<span class="font-mono font-bold text-slate-900 text-sm">{order.order_number}</span>
									</div>
									<div class="text-right">
										<span class="text-[10px] text-slate-400 block font-bold uppercase">Total Tagihan Tunai</span>
										<span class="font-extrabold text-orange-600 font-['Outfit'] text-base">{formatRupiah(order.total)}</span>
									</div>
								</div>

								<!-- Inline Cash Payment Invoice -->
								<div class="bg-white rounded-xl p-3.5 border border-slate-200/90 shadow-2xs space-y-3">
									<div class="flex items-center justify-between pb-2 border-b border-dashed border-slate-200">
										<div class="flex items-center gap-2">
											<div class="w-6 h-6 rounded-lg bg-orange-50 border border-orange-200 flex items-center justify-center text-orange-600">
												<Receipt class="w-3.5 h-3.5" />
											</div>
											<span class="text-xs font-black text-slate-800 tracking-tight font-['Outfit']">INVOICE PEMBAYARAN</span>
										</div>
										<span class="text-[10px] text-slate-400 font-mono font-medium">{formatDateTime(order.created_at)}</span>
									</div>

									<!-- Item Breakdown -->
									<div class="space-y-2 max-h-52 overflow-y-auto pr-0.5 divide-y divide-slate-100">
										<div class="flex justify-between text-[10px] font-bold text-slate-400 uppercase tracking-wider pb-1">
											<span>Menu ({order.items?.length || 0})</span>
											<span>Subtotal</span>
										</div>

										{#each order.items || [] as item}
											<div class="pt-1.5 flex items-start justify-between gap-2 text-xs">
												<div class="min-w-0">
													<div class="font-bold text-slate-800 flex items-center gap-1.5">
														<span class="text-orange-600 font-mono text-[11px] font-bold">{item.quantity}x</span>
														<span class="truncate">{item.menu_name_snapshot}</span>
													</div>
													{#if item.selected_modifiers && item.selected_modifiers.length > 0}
														<div class="text-[10px] text-slate-500 pl-4 mt-0.5">
															+ {item.selected_modifiers.map(m => m.option_name).join(', ')}
														</div>
													{/if}
													{#if item.notes}
														<div class="text-[10px] text-slate-400 italic pl-4">"{item.notes}"</div>
													{/if}
												</div>
												<span class="font-bold text-slate-800 font-['Outfit'] shrink-0 text-xs">
													{formatRupiah(item.subtotal)}
												</span>
											</div>
										{/each}
									</div>

									<!-- Calculation Summary -->
									<div class="pt-2 border-t border-dashed border-slate-200 space-y-1 text-xs text-slate-600">
										<div class="flex justify-between text-[11px]">
											<span>Subtotal</span>
											<span class="font-semibold text-slate-800">{formatRupiah(order.subtotal)}</span>
										</div>
										<div class="flex justify-between text-[11px]">
											<span>Pajak (PB1)</span>
											<span class="font-semibold text-slate-800">{formatRupiah(order.tax)}</span>
										</div>
										{#if order.service_charge > 0}
											<div class="flex justify-between text-[11px]">
												<span>Biaya Layanan</span>
												<span class="font-semibold text-slate-800">{formatRupiah(order.service_charge)}</span>
											</div>
										{/if}
										{#if order.discount > 0}
											<div class="flex justify-between text-[11px] text-emerald-600 font-semibold">
												<span>Diskon</span>
												<span>-{formatRupiah(order.discount)}</span>
											</div>
										{/if}
										<div class="pt-1.5 border-t border-slate-200 flex justify-between items-center text-xs font-black text-slate-900">
											<span class="uppercase tracking-tight text-slate-700">Total Wajib Bayar</span>
											<span class="text-orange-600 font-['Outfit'] text-base font-extrabold">{formatRupiah(order.total)}</span>
										</div>
									</div>
								</div>
							</div>

							<!-- Action Buttons for Cash Method -->
							<div class="space-y-2 pt-1">
								<div class="flex gap-2">
									<button
										type="button"
										onclick={() => (showInvoiceModal = true)}
										class="flex-1 bg-amber-600 hover:bg-amber-700 text-white font-bold py-2.5 px-4 rounded-xl text-xs flex items-center justify-center gap-2 shadow-sm transition-all active:scale-[0.98]"
									>
										<Receipt class="w-4 h-4" />
										<span>Buka Invoice Tagihan Kasir</span>
									</button>
									<button
										type="button"
										onclick={printInvoice}
										class="bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold py-2.5 px-3 rounded-xl text-xs flex items-center justify-center gap-1.5 transition-colors"
										title="Cetak Tagihan"
									>
										<Printer class="w-4 h-4" />
										<span>Cetak</span>
									</button>
								</div>


							</div>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Paid Invoice / Receipt Notification Card (Shown when payment is complete) -->
			{#if isPaid}
				<div class="bg-white border-2 border-emerald-500/40 rounded-3xl p-5 shadow-xs space-y-3.5">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2.5">
							<div class="w-9 h-9 rounded-2xl bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-600 shadow-xs">
								<CheckCircle2 class="w-5 h-5" />
							</div>
							<div>
								<div class="text-xs font-black text-slate-900 flex items-center gap-1.5">
									<span>Pembayaran Berhasil</span>
									<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-extrabold bg-emerald-100 text-emerald-800">
										LUNAS
									</span>
								</div>
								<div class="text-[11px] text-slate-500 font-medium">
									{payment?.payment_method || 'QRIS (Cashless)'} &bull; {formatDateTime(payment?.paid_at || order.updated_at)}
								</div>
							</div>
						</div>
					</div>

					<div class="bg-slate-50 rounded-2xl p-3 border border-slate-100 flex items-center justify-between text-xs">
						<div>
							<span class="text-slate-400 text-[10px] block font-bold uppercase tracking-wider">No. Invoice</span>
							<span class="font-mono font-bold text-slate-800">INV-{order.order_number}</span>
						</div>
						<div class="text-right">
							<span class="text-slate-400 text-[10px] block font-bold uppercase tracking-wider">Total Lunas</span>
							<span class="font-black font-['Outfit'] text-emerald-600 text-sm">{formatRupiah(order.total)}</span>
						</div>
					</div>

					<div class="flex gap-2 pt-1">
						<button
							type="button"
							onclick={() => (showInvoiceModal = true)}
							class="flex-1 bg-emerald-600 hover:bg-emerald-700 text-white font-bold py-2.5 px-4 rounded-xl text-xs flex items-center justify-center gap-2 shadow-sm shadow-emerald-600/20 transition-all active:scale-[0.98]"
						>
							<Receipt class="w-4 h-4" />
							<span>Tampilkan Invoice</span>
						</button>
						<button
							type="button"
							onclick={printInvoice}
							class="bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold py-2.5 px-3 rounded-xl text-xs flex items-center justify-center gap-1.5 transition-colors"
							title="Cetak Struk"
						>
							<Printer class="w-4 h-4" />
							<span>Cetak</span>
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
									<Check class="w-3.5 h-3.5 stroke-3" />
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

<!-- Digital Invoice Modal Popup (Auto shown on payment completed & can be opened anytime) -->
{#if showInvoiceModal && order}
	<div
		class="fixed inset-0 z-50 bg-slate-950/70 backdrop-blur-xs flex items-center justify-center p-4 overflow-y-auto printable-invoice-modal"
		onclick={(e) => {
			if (e.target === e.currentTarget) showInvoiceModal = false;
		}}
		onkeydown={(e) => {
			if (e.key === 'Escape') showInvoiceModal = false;
		}}
		tabindex="-1"
		role="dialog"
		aria-modal="true"
	>
		<div
			class="bg-white w-full max-w-md rounded-3xl shadow-2xl border border-slate-200 overflow-hidden my-auto printable-invoice-content"
		>
			<!-- Top Action Bar (hidden on print) -->
			<div class="px-6 pt-5 pb-3 flex items-center justify-between border-b border-slate-100 no-print">
				<div class="flex items-center gap-2">
					<div class="w-7 h-7 rounded-lg bg-orange-600 flex items-center justify-center text-white shadow-xs">
						<Store class="w-4 h-4" />
					</div>
					<span class="text-xs font-bold text-slate-800 font-['Outfit']">
						{isPaid ? 'Invoice Pembayaran' : 'Tagihan Kasir (Billing)'}
					</span>
				</div>
				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={printInvoice}
						class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-bold bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors"
					>
						<Printer class="w-3.5 h-3.5" />
						<span>Cetak</span>
					</button>
					<button
						type="button"
						onclick={() => (showInvoiceModal = false)}
						class="w-7 h-7 rounded-lg text-slate-400 hover:text-slate-700 hover:bg-slate-100 flex items-center justify-center transition-colors"
						aria-label="Tutup"
					>
						<X class="w-4 h-4" />
					</button>
				</div>
			</div>

			<!-- Printable Receipt Paper Content -->
			<div class="p-6 sm:p-7 space-y-4 text-slate-800">
				<!-- Brand Header -->
				<div class="text-center space-y-1">
					<div class="w-12 h-12 bg-linear-to-tr from-orange-600 to-amber-500 rounded-2xl flex items-center justify-center mx-auto text-white shadow-md shadow-orange-500/20">
						<Store class="w-6 h-6" />
					</div>
					<h2 class="text-xl font-black font-['Outfit'] tracking-tight text-slate-900 mt-2">Resto Nusantara</h2>
					<p class="text-[11px] text-slate-500 font-medium">Sistem Pemesanan Self-Order Digital</p>
				</div>

				<!-- Stamp Badge -->
				<div class="flex items-center justify-center pt-1">
					{#if !isPaid}
						<div class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-full border-2 border-amber-500/60 bg-amber-50 text-amber-800 font-extrabold text-xs tracking-wider uppercase shadow-xs">
							<Clock class="w-4 h-4 text-amber-600" />
							<span>BELUM DIBAYAR / TAGIHAN KASIR</span>
						</div>
					{:else}
						<div class="inline-flex items-center gap-1.5 px-4 py-1.5 rounded-full border-2 border-emerald-500/60 bg-emerald-50 text-emerald-700 font-extrabold text-xs tracking-wider uppercase shadow-xs">
							<CheckCircle2 class="w-4 h-4 text-emerald-600" />
							<span>LUNAS / TELAH DIBAYAR</span>
						</div>
					{/if}
				</div>

				{#if !isPaid}
					<div class="bg-amber-50 border border-amber-200/80 rounded-2xl p-3 text-center text-xs text-amber-900 font-semibold leading-relaxed">
						Tunjukkan invoice tagihan ini kepada kasir di meja kasir untuk melakukan pembayaran tunai.
					</div>
				{/if}

				<!-- Meta Data Grid -->
				<div class="grid grid-cols-2 gap-y-2 text-xs pt-3 border-t border-dashed border-slate-300">
					<div>
						<span class="text-[10px] text-slate-400 uppercase font-bold block">No. Invoice</span>
						<span class="font-mono font-bold text-slate-900">{isPaid ? 'INV-' + order.order_number : 'BILL-' + order.order_number}</span>
					</div>
					<div class="text-right">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">No. Pesanan</span>
						<span class="font-mono font-bold text-slate-900">{order.order_number}</span>
					</div>
					<div>
						<span class="text-[10px] text-slate-400 uppercase font-bold block">{isPaid ? 'Waktu Bayar' : 'Waktu Terbit'}</span>
						<span class="font-medium text-slate-800">{formatDateTime(isPaid ? (payment?.paid_at || order.updated_at) : order.created_at)}</span>
					</div>
					<div class="text-right">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">Meja</span>
						<span class="font-bold text-slate-900">{order.table_name || 'Meja'}</span>
					</div>
					<div>
						<span class="text-[10px] text-slate-400 uppercase font-bold block">Metode Bayar</span>
						<span class="font-bold text-slate-900">{isPaid ? (payment?.payment_method || (selectedMethod === 'CASH' ? 'Tunai (Cash)' : 'QRIS')) : (selectedMethod === 'CASH' ? 'Tunai di Kasir' : 'QRIS (Menunggu Bayar)')}</span>
					</div>
					<div class="text-right">
						{#if isPaid}
							<span class="text-[10px] text-slate-400 uppercase font-bold block">Ref ID Transaksi</span>
							<span class="font-mono text-[11px] text-slate-600 truncate max-w-[140px] inline-block">{payment?.provider_transaction_id || payment?.id || '-'}</span>
						{:else}
							<span class="text-[10px] text-slate-400 uppercase font-bold block">Status Tagihan</span>
							<span class="text-amber-700 font-bold text-[11px] uppercase">Menunggu Kasir</span>
						{/if}
					</div>
				</div>

				<!-- Items Table -->
				<div class="pt-3 border-t-2 border-dashed border-slate-300 space-y-2">
					<div class="flex justify-between text-[11px] font-bold text-slate-400 uppercase tracking-wider pb-1">
						<span>Menu</span>
						<span>Subtotal</span>
					</div>

					{#each order.items || [] as item}
						<div class="flex items-start justify-between gap-3 text-xs">
							<div class="min-w-0">
								<div class="font-bold text-slate-900">
									{item.quantity}x {item.menu_name_snapshot}
								</div>
								{#if item.selected_modifiers && item.selected_modifiers.length > 0}
									<div class="text-[10px] text-slate-500 mt-0.5">
										+ {item.selected_modifiers.map(m => m.option_name).join(', ')}
									</div>
								{/if}
								{#if item.notes}
									<div class="text-[10px] text-slate-400 italic">"{item.notes}"</div>
								{/if}
							</div>
							<span class="font-bold text-slate-900 font-['Outfit'] shrink-0">{formatRupiah(item.subtotal)}</span>
						</div>
					{/each}
				</div>

				<!-- Calculations -->
				<div class="pt-3 border-t-2 border-dashed border-slate-300 space-y-1.5 text-xs text-slate-600">
					<div class="flex justify-between">
						<span>Subtotal</span>
						<span class="font-semibold text-slate-900">{formatRupiah(order.subtotal)}</span>
					</div>
					<div class="flex justify-between">
						<span>Pajak Restoran</span>
						<span class="font-semibold text-slate-900">{formatRupiah(order.tax)}</span>
					</div>
					{#if order.service_charge > 0}
						<div class="flex justify-between">
							<span>Biaya Layanan</span>
							<span class="font-semibold text-slate-900">{formatRupiah(order.service_charge)}</span>
						</div>
					{/if}
					{#if order.discount > 0}
						<div class="flex justify-between text-emerald-600 font-semibold">
							<span>Diskon</span>
							<span>-{formatRupiah(order.discount)}</span>
						</div>
					{/if}

					<div class="pt-2 border-t border-slate-200 flex justify-between text-sm font-black text-slate-900 font-['Outfit']">
						<span>{isPaid ? 'TOTAL DIBAYAR' : 'TOTAL TAGIHAN'}</span>
						<span class="text-orange-600 text-base">{formatRupiah(order.total)}</span>
					</div>
					{#if isPaid}
						<div class="flex justify-between text-xs text-slate-500">
							<span>Kembalian</span>
							<span class="font-semibold">Rp 0</span>
						</div>
					{:else}
						<div class="flex justify-between text-xs text-slate-500">
							<span>Status</span>
							<span class="font-semibold text-amber-700">Belum Dibayar (Kasir)</span>
						</div>
					{/if}
				</div>

				<!-- Footer note / Verification QR -->
				<div class="pt-3 border-t border-dashed border-slate-300 text-center space-y-2">
					<div class="inline-flex flex-col items-center justify-center p-2.5 rounded-xl bg-slate-50 border border-slate-200">
						<QrCode class="w-16 h-16 text-slate-700" />
						<span class="font-mono font-bold text-[10px] text-slate-500 mt-1">KASIR: {order.order_number}</span>
					</div>
					<p class="text-[10px] text-slate-400 max-w-xs mx-auto">
						{isPaid
							? 'Struk ini diterbitkan secara otomatis dan berlaku sebagai bukti pembayaran sah dari Resto Nusantara.'
							: 'Tunjukkan struk/invoice tagihan ini kepada kasir untuk memproses pembayaran tunai Anda.'}
					</p>
					<p class="text-xs font-bold text-slate-700 font-['Outfit']">
						{isPaid ? 'Terima kasih atas pesanan Anda!' : 'Mohon selesaikan pembayaran di kasir.'}
					</p>
				</div>
			</div>

			<!-- Bottom Close & Print Action Buttons (hidden on print) -->
			<div class="p-4 bg-slate-50 border-t border-slate-100 flex gap-2 no-print">
				<button
					type="button"
					onclick={printInvoice}
					class="flex-1 bg-orange-600 hover:bg-orange-700 text-white font-bold py-2.5 px-4 rounded-xl text-xs flex items-center justify-center gap-2 shadow-sm transition-all"
				>
					<Printer class="w-4 h-4" />
					<span>{isPaid ? 'Cetak Struk / Invoice' : 'Cetak Invoice Tagihan'}</span>
				</button>

				<button
					type="button"
					onclick={() => (showInvoiceModal = false)}
					class="px-4 py-2.5 rounded-xl border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-100 transition-colors"
				>
					Tutup
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	@media print {
		:global(body) {
			background: white !important;
		}
		:global(header),
		:global(main),
		:global(.no-print) {
			display: none !important;
		}
		.printable-invoice-modal {
			position: static !important;
			background: white !important;
			padding: 0 !important;
			display: block !important;
			inset: auto !important;
			overflow: visible !important;
		}
		.printable-invoice-content {
			box-shadow: none !important;
			border: 1px solid #cbd5e1 !important;
			max-width: 100% !important;
			width: 100% !important;
			margin: 0 !important;
			border-radius: 0 !important;
		}
	}
</style>

