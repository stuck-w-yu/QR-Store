<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { Order } from '$lib/types';
	import { 
		Wallet, Search, RefreshCw, CheckCircle2, Clock, 
		XCircle, AlertCircle, Printer, ArrowRight, Banknote,
		Sparkles, Utensils, User, MapPin, Volume2, VolumeX,
		Check, Receipt, DollarSign, Calculator, ChevronRight,
		Camera, Eye, Image as ImageIcon, Trash2, SlidersHorizontal,
		RotateCcw, AlertTriangle, Plus, ShieldAlert
	} from '@lucide/svelte';

	let orders = $state<Order[]>([]);
	let loading = $state(true);
	let searchQuery = $state('');
	let activeTab = $state<'PENDING' | 'PAID_TODAY' | 'CANCELLED' | 'ALL'>('PENDING');
	let soundEnabled = $state(true);
	let wsConnected = $state(false);

	// CRUD Edit & Delete Order State
	let editModalOrder = $state<Order | null>(null);
	let deleteModalOrder = $state<Order | null>(null);
	let editStatus = $state<string>('');
	let editPaymentStatus = $state<string>('');
	let editNotes = $state<string>('');
	let submittingEdit = $state(false);
	let deletingOrderId = $state<string | null>(null);

	// Proof of Payment Inspection State
	let viewProofModalData = $state<{
		image: string;
		fileName: string;
		fileSize: string;
		uploadedAt: string;
		orderNumber: string;
		gdriveUrl?: string;
	} | null>(null);

	function getOrderProof(orderOrId: string | Order | null | undefined): { image: string; fileName: string; fileSize: string; uploadedAt: string; gdriveUrl?: string; gdriveId?: string } | null {
		if (!orderOrId) return null;
		const orderId = typeof orderOrId === 'string' ? orderOrId : orderOrId.id;
		const orderObj = typeof orderOrId !== 'string' ? orderOrId : orders.find(o => o.id === orderId);
		const proofUrl = orderObj?.proof_url;

		let localData: any = null;
		if (typeof window !== 'undefined') {
			try {
				const saved = localStorage.getItem(`payment_proof_${orderId}`);
				if (saved) localData = JSON.parse(saved);
			} catch (_) {}
		}

		if (proofUrl) {
			return {
				image: proofUrl,
				fileName: localData?.fileName || 'bukti-pembayaran.jpg',
				fileSize: localData?.fileSize || 'Foto Bukti Pelanggan',
				uploadedAt: localData?.uploadedAt || 'Terkirim ke Kasir',
				gdriveUrl: proofUrl.startsWith('http') ? proofUrl : (localData?.gdriveUrl || undefined),
				gdriveId: localData?.gdriveId
			};
		}

		if (localData?.image) {
			return localData;
		}

		return null;
	}

	// WebSocket & Polling
	let ws: WebSocket | null = null;
	let pollInterval: any = null;

	// Payment Confirmation Modal State
	let paymentModalOrder = $state<Order | null>(null);
	let selectedMethod = $state<'CASH' | 'QRIS_MANUAL' | 'DEBIT'>('CASH');
	let paidAmount = $state<number>(0);
	let submittingPayment = $state(false);

	// Success / Receipt Modal State
	let successModalData = $state<{
		order: Order;
		paidAmount: number;
		change: number;
		paymentMethod: string;
	} | null>(null);

	// Cancel Order Modal State
	let cancelModalOrder = $state<Order | null>(null);
	let cancelSubmitting = $state(false);

	// Audio notification using Web Audio API synthesizer
	function playChime() {
		if (!soundEnabled || typeof window === 'undefined') return;
		try {
			const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
			const osc = ctx.createOscillator();
			const gain = ctx.createGain();

			osc.type = 'triangle';
			osc.frequency.setValueAtTime(523.25, ctx.currentTime); // C5
			osc.frequency.exponentialRampToValueAtTime(783.99, ctx.currentTime + 0.15); // G5
			osc.frequency.exponentialRampToValueAtTime(1046.50, ctx.currentTime + 0.35); // C6

			gain.gain.setValueAtTime(0.35, ctx.currentTime);
			gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.7);

			osc.connect(gain);
			gain.connect(ctx.destination);

			osc.start();
			osc.stop(ctx.currentTime + 0.7);
		} catch (e) {
			console.error('Audio play error', e);
		}
	}

	async function loadOrders() {
		try {
			const data = await api.get<Order[]>('/orders');
			orders = data;
		} catch (e) {
			console.error('Failed to load orders', e);
		} finally {
			loading = false;
		}
	}

	function connectWebSocket() {
		const restoId = auth.user?.restaurant_id || 'rst_nusantara';
		const channel = `restaurant:${restoId}:cashier`;

		const wsURL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
		try {
			if (ws) {
				try { ws.close(); } catch (_) {}
			}
			ws = new WebSocket(`${wsURL}?channel=${channel}`);

			ws.onopen = () => {
				wsConnected = true;
			};

			ws.onmessage = (event) => {
				try {
					const msg = JSON.parse(event.data);
					const relevantEvents = [
						'NEW_ORDER_PENDING',
						'order.created',
						'ORDER_STATUS_CHANGED',
						'PAYMENT_PAID',
						'PAYMENT_CONFIRMED',
						'order.confirmed',
						'PAYMENT_PROOF_SUBMITTED',
						'ORDER_DELETED'
					];
					if (relevantEvents.includes(msg.event)) {
						if (msg.event === 'NEW_ORDER_PENDING' || msg.event === 'order.created' || msg.event === 'PAYMENT_PROOF_SUBMITTED') {
							playChime();
						}
						if (msg.event === 'ORDER_DELETED') {
							const delId = msg.data?.order_id;
							if (delId) {
								orders = orders.filter((o) => o.id !== delId);
							}
						}
						// Direct update in state if order object is supplied
						if (msg.data?.order) {
							const updated = msg.data.order as Order;
							const idx = orders.findIndex((o) => o.id === updated.id);
							if (idx !== -1) {
								orders[idx] = { ...orders[idx], ...updated };
							} else {
								orders = [updated, ...orders];
							}
						}
						loadOrders();
					}
				} catch (e) {
					console.error('Invalid ws message in cashier confirmation', e);
				}
			};

			ws.onerror = () => {
				wsConnected = false;
			};

			ws.onclose = () => {
				wsConnected = false;
				setTimeout(() => {
					if (typeof window !== 'undefined' && (!ws || ws.readyState === WebSocket.CLOSED)) {
						connectWebSocket();
					}
				}, 3000);
			};
		} catch (e) {
			wsConnected = false;
			console.error('Failed to connect WebSocket in cashier confirmation', e);
		}
	}

	// Filter orders
	let pendingOrders = $derived(
		orders.filter((o) => (o.status === 'WAITING_PAYMENT' || o.payment_status === 'UNPAID') && o.status !== 'CANCELLED')
	);

	let paidTodayOrders = $derived(
		orders.filter((o) => o.payment_status === 'PAID' && o.status !== 'CANCELLED')
	);

	let cancelledOrders = $derived(
		orders.filter((o) => o.status === 'CANCELLED')
	);

	let displayedOrders = $derived(
		orders.filter((o) => {
			// Tab filtering
			if (activeTab === 'PENDING') {
				if (o.status === 'CANCELLED' || (o.status !== 'WAITING_PAYMENT' && o.payment_status !== 'UNPAID')) return false;
			} else if (activeTab === 'PAID_TODAY') {
				if (o.payment_status !== 'PAID' || o.status === 'CANCELLED') return false;
			} else if (activeTab === 'CANCELLED') {
				if (o.status !== 'CANCELLED') return false;
			}

			// Search filtering
			if (!searchQuery.trim()) return true;
			const q = searchQuery.toLowerCase();
			const matchNum = o.order_number?.toLowerCase().includes(q);
			const matchTable = (o.table_name || o.table_id || '').toLowerCase().includes(q);
			const matchItem = o.items?.some((it) => it.menu_name_snapshot?.toLowerCase().includes(q));
			return matchNum || matchTable || matchItem;
		})
	);

	let totalPendingAmount = $derived(
		pendingOrders.reduce((sum, o) => sum + (o.total || 0), 0)
	);

	let totalPaidTodayAmount = $derived(
		paidTodayOrders.reduce((sum, o) => sum + (o.total || 0), 0)
	);

	// Quick Change Calculation
	let calculatedChange = $derived(
		paymentModalOrder ? Math.max(0, paidAmount - paymentModalOrder.total) : 0
	);

	let isAmountSufficient = $derived(
		paymentModalOrder ? (selectedMethod === 'CASH' ? paidAmount >= paymentModalOrder.total : true) : false
	);

	function openPaymentModal(o: Order) {
		paymentModalOrder = o;
		if (o.proof_url || o.payment_method === 'QRIS' || o.payment_method === 'QRIS_MANUAL') {
			selectedMethod = 'QRIS_MANUAL';
		} else {
			selectedMethod = 'CASH';
		}
		paidAmount = o.total; // Default to exact amount
	}

	function setQuickAmount(amount: number) {
		paidAmount = amount;
	}

	async function submitPaymentConfirmation() {
		if (!paymentModalOrder) return;
		if (selectedMethod === 'CASH' && paidAmount < paymentModalOrder.total) {
			alert('Uang yang diterima kurang dari total tagihan!');
			return;
		}

		submittingPayment = true;
		const currentOrder = paymentModalOrder;
		const currentPaid = selectedMethod === 'CASH' ? paidAmount : currentOrder.total;
		const change = currentPaid - currentOrder.total;

		try {
			await api.post(`/cashier/orders/${currentOrder.id}/payment`, {
				payment_method: selectedMethod,
				paid_amount: currentPaid
			});

			// Show receipt modal
			successModalData = {
				order: currentOrder,
				paidAmount: currentPaid,
				change: change,
				paymentMethod: selectedMethod
			};

			paymentModalOrder = null;
			await loadOrders();
		} catch (e: any) {
			alert(e?.message || 'Gagal memproses konfirmasi pembayaran');
		} finally {
			submittingPayment = false;
		}
	}

	async function handleCancelOrder(orderId: string) {
		if (!confirm('Yakin ingin membatalkan pesanan ini? Status pesanan akan menjadi CANCELLED.')) return;
		cancelSubmitting = true;
		try {
			await api.post(`/cashier/orders/${orderId}/cancel`);
			await loadOrders();
		} catch (e: any) {
			alert(e?.message || 'Gagal membatalkan pesanan');
		} finally {
			cancelSubmitting = false;
		}
	}

	function openEditModal(o: Order) {
		editModalOrder = o;
		editStatus = o.status || 'WAITING_PAYMENT';
		editPaymentStatus = (o.payment_status as string) || 'UNPAID';
		editNotes = o.notes || '';
	}

	async function saveEditOrder() {
		if (!editModalOrder) return;
		submittingEdit = true;
		try {
			await api.patch(`/orders/${editModalOrder.id}`, {
				status: editStatus,
				payment_status: editPaymentStatus,
				notes: editNotes ? editNotes.trim() : null
			});
			editModalOrder = null;
			await loadOrders();
		} catch (e: any) {
			alert(e?.message || 'Gagal menyimpan perubahan pesanan');
		} finally {
			submittingEdit = false;
		}
	}

	function openDeleteModal(order: Order) {
		deleteModalOrder = order;
	}

	async function executeDeleteOrder(order: Order) {
		deletingOrderId = order.id;
		try {
			await api.delete(`/orders/${order.id}`);
			if (editModalOrder?.id === order.id) {
				editModalOrder = null;
			}
			if (deleteModalOrder?.id === order.id) {
				deleteModalOrder = null;
			}
			orders = orders.filter((o) => o.id !== order.id);
			await loadOrders();
		} catch (e: any) {
			alert(e?.message || 'Gagal menghapus pesanan');
		} finally {
			deletingOrderId = null;
		}
	}

	async function handleDeleteOrder(order: Order) {
		openDeleteModal(order);
	}

	async function handleRestoreOrder(order: Order, targetStatus: 'WAITING_PAYMENT' | 'CONFIRMED' = 'CONFIRMED') {
		const targetLabel = targetStatus === 'CONFIRMED' ? 'Diterima Dapur (Sudah Lunas)' : 'Menunggu Pembayaran';
		if (!confirm(`Pulihkan pesanan #${order.order_number} kembali menjadi "${targetLabel}"?`)) return;
		try {
			await api.patch(`/orders/${order.id}`, {
				status: targetStatus,
				payment_status: targetStatus === 'CONFIRMED' ? 'PAID' : 'UNPAID'
			});
			await loadOrders();
		} catch (e: any) {
			alert(e?.message || 'Gagal memulihkan pesanan');
		}
	}

	function printReceipt() {
		window.print();
	}

	function getElapsedMinutes(createdAt: string): number {
		const created = new Date(createdAt).getTime();
		const now = Date.now();
		return Math.max(0, Math.floor((now - created) / 60000));
	}

	onMount(async () => {
		if (!auth.initialized) {
			await auth.init();
		}
		if (!auth.user) {
			return;
		}

		await loadOrders();
		connectWebSocket();
		pollInterval = setInterval(loadOrders, 5000);
	});

	onDestroy(() => {
		if (ws) ws.close();
		if (pollInterval) clearInterval(pollInterval);
	});
</script>

<svelte:head>
	<title>Konfirmasi Pembayaran Kasir & Tunai | Resto Nusantara</title>
</svelte:head>

<div class="space-y-6 max-w-7xl mx-auto pb-12">
	<!-- Top Header & Live Indicators -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-6 rounded-3xl border border-slate-200/80 shadow-xs">
		<div class="space-y-1">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 rounded-2xl bg-orange-500/10 text-orange-600 flex items-center justify-center font-bold">
					<Wallet class="w-5 h-5" />
				</div>
				<div>
					<h1 class="text-xl sm:text-2xl font-black text-slate-900 font-['Outfit'] flex items-center gap-2">
						Konfirmasi Kasir & Tunai
						{#if pendingOrders.length > 0}
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-black bg-rose-600 text-white animate-bounce">
								{pendingOrders.length} Baru
							</span>
						{/if}
					</h1>
					<p class="text-xs text-slate-500 font-medium">
						Verifikasi dan terima pembayaran tunai (cash) pelanggan meja secara instan
					</p>
				</div>
			</div>
		</div>

		<div class="flex flex-wrap items-center gap-2 sm:gap-3">
			<!-- WebSocket Live Badge -->
			<div class="flex items-center gap-1.5 px-3 py-1.5 {wsConnected ? 'bg-emerald-50 text-emerald-700 border-emerald-200' : 'bg-amber-50 text-amber-700 border-amber-200'} font-bold rounded-xl text-xs border">
				<span class="w-2 h-2 rounded-full {wsConnected ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500 animate-ping'}"></span>
				<span>{wsConnected ? 'Live WebSocket' : 'Menghubungkan WS...'}</span>
			</div>

			<!-- Audio Notification Toggle -->
			<button
				type="button"
				onclick={() => (soundEnabled = !soundEnabled)}
				class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold border transition-colors {soundEnabled
					? 'bg-slate-900 text-white border-slate-800'
					: 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50'}"
				title="Toggle Suara Notifikasi"
			>
				{#if soundEnabled}
					<Volume2 class="w-3.5 h-3.5 text-orange-400" />
					<span>Suara ON</span>
				{:else}
					<VolumeX class="w-3.5 h-3.5 text-slate-400" />
					<span>Suara OFF</span>
				{/if}
			</button>

			<!-- Manual Refresh -->
			<button
				type="button"
				onclick={loadOrders}
				class="p-2 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 transition-colors shadow-xs cursor-pointer"
				title="Muat Ulang Data"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>

			<!-- Quick POS Shortcut -->
			<a
				href="/admin/pos"
				class="px-3.5 py-1.5 bg-orange-600 hover:bg-orange-700 active:scale-95 text-white font-bold rounded-xl text-xs flex items-center gap-1.5 shadow-md shadow-orange-600/20 transition-all"
				title="Buka Kasir POS untuk Tambah Pesanan Baru"
			>
				<Plus class="w-3.5 h-3.5" />
				<span>Buat Pesanan Baru</span>
			</a>
		</div>
	</div>

	<!-- Metric Highlights -->
	<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
		<!-- Card 1: Menunggu Pembayaran -->
		<div class="bg-linear-to-br from-amber-500 to-orange-600 rounded-3xl p-5 text-white shadow-lg shadow-orange-500/20 relative overflow-hidden">
			<div class="absolute -right-4 -bottom-4 w-28 h-28 bg-white/10 rounded-full blur-xl pointer-events-none"></div>
			<div class="flex items-center justify-between text-xs font-bold uppercase tracking-wider text-orange-100">
				<span>Menunggu Pembayaran</span>
				<Clock class="w-4 h-4" />
			</div>
			<div class="mt-3">
				<div class="text-3xl font-black font-['Outfit'] leading-none">
					{pendingOrders.length} Pesanan
				</div>
				<div class="text-xs text-orange-100/90 font-bold mt-1.5">
					Total Tagihan: {formatRupiah(totalPendingAmount)}
				</div>
			</div>
		</div>

		<!-- Card 2: Diterima Hari Ini -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs">
			<div class="flex items-center justify-between text-xs font-bold uppercase tracking-wider text-slate-400">
				<span>Tunai Diterima Hari Ini</span>
				<CheckCircle2 class="w-4 h-4 text-emerald-500" />
			</div>
			<div class="mt-3">
				<div class="text-3xl font-black text-slate-900 font-['Outfit'] leading-none">
					{paidTodayOrders.length} Transaksi
				</div>
				<div class="text-xs text-emerald-600 font-bold mt-1.5">
					Total: {formatRupiah(totalPaidTodayAmount)}
				</div>
			</div>
		</div>

		<!-- Card 3: Akses Konfirmasi -->
		<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs flex flex-col justify-between">
			<div class="flex items-center justify-between text-xs font-bold uppercase tracking-wider text-slate-400">
				<span>Otoritas Petugas</span>
				<User class="w-4 h-4 text-blue-500" />
			</div>
			<div class="mt-3 flex items-center justify-between">
				<div>
					<div class="text-sm font-black text-slate-900">
						{auth.user?.name || 'Administrator'}
					</div>
					<div class="text-xs text-slate-500 font-medium">
						Role: <span class="font-bold text-orange-600">{auth.user?.role || 'ADMIN'}</span> (Akses Penuh Kasir)
					</div>
				</div>
				{#if cancelledOrders.length > 0}
					<button
						type="button"
						onclick={() => (activeTab = 'CANCELLED')}
						class="px-2.5 py-1 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-700 text-xs font-bold border border-rose-200 transition-colors flex items-center gap-1 cursor-pointer"
						title="Lihat pesanan yang dibatalkan / nyangkut"
					>
						<XCircle class="w-3.5 h-3.5 text-rose-600" />
						<span>{cancelledOrders.length} Batal</span>
					</button>
				{/if}
			</div>
		</div>
	</div>

	<!-- Controls & Filter Tabs -->
	<div class="flex flex-col md:flex-row items-stretch md:items-center justify-between gap-4">
		<!-- Tabs -->
		<div class="flex items-center gap-1.5 p-1 bg-slate-200/70 rounded-2xl w-full sm:w-auto overflow-x-auto">
			<button
				type="button"
				onclick={() => (activeTab = 'PENDING')}
				class="flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center justify-center gap-2 whitespace-nowrap cursor-pointer {activeTab === 'PENDING'
					? 'bg-white text-orange-600 shadow-xs'
					: 'text-slate-600 hover:text-slate-900'}"
			>
				<span>Menunggu Tunai</span>
				{#if pendingOrders.length > 0}
					<span class="px-2 py-0.5 rounded-full text-[10px] bg-rose-500 text-white font-black">
						{pendingOrders.length}
					</span>
				{/if}
			</button>

			<button
				type="button"
				onclick={() => (activeTab = 'PAID_TODAY')}
				class="flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center justify-center gap-2 whitespace-nowrap cursor-pointer {activeTab === 'PAID_TODAY'
					? 'bg-white text-slate-900 shadow-xs'
					: 'text-slate-600 hover:text-slate-900'}"
			>
				<span>Sudah Lunas Hari Ini</span>
				<span class="text-[10px] text-slate-400">({paidTodayOrders.length})</span>
			</button>

			<button
				type="button"
				onclick={() => (activeTab = 'CANCELLED')}
				class="flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center justify-center gap-2 whitespace-nowrap cursor-pointer {activeTab === 'CANCELLED'
					? 'bg-white text-rose-600 shadow-xs'
					: 'text-slate-600 hover:text-slate-900'}"
			>
				<span>Dibatalkan</span>
				{#if cancelledOrders.length > 0}
					<span class="px-2 py-0.5 rounded-full text-[10px] bg-rose-100 text-rose-700 font-black">
						{cancelledOrders.length}
					</span>
				{/if}
			</button>

			<button
				type="button"
				onclick={() => (activeTab = 'ALL')}
				class="flex-1 sm:flex-initial px-4 py-2 rounded-xl text-xs font-black transition-all flex items-center justify-center gap-2 whitespace-nowrap cursor-pointer {activeTab === 'ALL'
					? 'bg-white text-slate-900 shadow-xs'
					: 'text-slate-600 hover:text-slate-900'}"
			>
				<span>Semua Pesanan</span>
			</button>
		</div>

		<!-- Search Bar -->
		<div class="relative w-full md:w-80">
			<Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Cari No. Pesanan (ORD-...) / Meja..."
				class="w-full pl-10 pr-4 py-2 rounded-2xl bg-white border border-slate-200 text-xs font-medium focus:outline-hidden focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all shadow-xs"
			/>
		</div>
	</div>

	<!-- Orders Grid -->
	{#if loading && orders.length === 0}
		<div class="text-center py-24 bg-white rounded-3xl border border-slate-200/80">
			<RefreshCw class="w-8 h-8 text-slate-300 animate-spin mx-auto mb-3" />
			<p class="text-xs font-bold text-slate-400">Memuat daftar pesanan menunggu konfirmasi...</p>
		</div>
	{:else if displayedOrders.length === 0}
		<div class="text-center py-20 bg-white rounded-3xl border border-slate-200/80 p-8 space-y-3">
			<div class="w-14 h-14 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto">
				<CheckCircle2 class="w-7 h-7" />
			</div>
			<h3 class="text-base font-black text-slate-900 font-['Outfit']">
				{activeTab === 'PENDING' ? 'Tidak Ada Tagihan Tunai Tertunda!' : 'Tidak Ditemukan Pesanan'}
			</h3>
			<p class="text-xs text-slate-500 max-w-md mx-auto">
				{activeTab === 'PENDING'
					? 'Semua pesanan meja telah terbayar atau belum ada pesanan baru yang memilih metode bayar tunai di kasir.'
					: 'Coba ubah kata kunci pencarian atau ganti tab filter di atas.'}
			</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
			{#each displayedOrders as order (order.id)}
				<div class="bg-white rounded-3xl border transition-all duration-200 overflow-hidden flex flex-col justify-between {order.status === 'CANCELLED'
					? 'border-rose-300 bg-rose-50/15 shadow-xs'
					: order.status === 'WAITING_PAYMENT' || order.payment_status === 'UNPAID'
					? 'border-amber-300 shadow-md shadow-amber-500/5 hover:border-amber-400'
					: 'border-slate-200/80 shadow-xs hover:border-slate-300'}">
					
					<!-- Header Card -->
					<div class="p-5 border-b border-slate-100 bg-slate-50/50 space-y-2.5">
						<div class="flex items-start justify-between gap-2">
							<div class="flex items-center gap-2">
								<span class="px-2.5 py-1 rounded-xl bg-orange-600 text-white font-black text-xs font-['Outfit'] flex items-center gap-1 shadow-xs">
									<MapPin class="w-3.5 h-3.5" />
									{order.table_name || `Meja ${order.table_id || '?'}`}
								</span>
								<span class="font-mono text-xs font-bold text-slate-800">
									{order.order_number}
								</span>
							</div>

							<div class="flex items-center gap-1.5 flex-wrap justify-end">
								{#if order.status === 'CANCELLED'}
									<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-rose-100 text-rose-700 border border-rose-200">
										<XCircle class="w-3 h-3" />
										DIBATALKAN
									</span>
								{:else if order.payment_status === 'PAID'}
									<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
										<CheckCircle2 class="w-3 h-3" />
										LUNAS
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200 animate-pulse">
										<Clock class="w-3 h-3" />
										BELUM BAYAR
									</span>
								{/if}

								<!-- Tombol Cepat Hapus di Header -->
								<button
									type="button"
									onclick={() => openDeleteModal(order)}
									class="p-1 rounded-lg bg-rose-50 hover:bg-rose-100 text-rose-600 border border-rose-200/80 transition-colors cursor-pointer shrink-0"
									title="Hapus Pesanan Ini"
								>
									<Trash2 class="w-3.5 h-3.5" />
								</button>
							</div>
						</div>

						<div class="flex items-center justify-between text-[11px] text-slate-400 font-medium">
							<span>Waktu pesan: {new Date(order.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} WIB</span>
							<span class="text-orange-600 font-bold">{getElapsedMinutes(order.created_at.toString())} mnt lalu</span>
						</div>

						<!-- Status Alur Dapur / Pesanan & Tombol CRUD -->
						<div class="pt-2 border-t border-slate-200/60 flex items-center justify-between gap-2 text-[11px]">
							<span class="text-slate-500 font-medium truncate">
								Alur: 
								<span class="font-bold {
									order.status === 'CANCELLED' ? 'text-rose-600' :
									order.status === 'PREPARING' ? 'text-orange-600' :
									order.status === 'READY' ? 'text-emerald-600' :
									order.status === 'COMPLETED' ? 'text-slate-700' :
									order.status === 'CONFIRMED' ? 'text-blue-600' : 'text-amber-600'
								}">
									{order.status === 'CANCELLED' ? 'Dibatalkan' :
									 order.status === 'PREPARING' ? 'Sedang Dimasak' :
									 order.status === 'READY' ? 'Siap Diantar' :
									 order.status === 'COMPLETED' ? 'Selesai' :
									 order.status === 'CONFIRMED' ? 'Diteruskan ke Dapur' : 'Menunggu Bayar'}
								</span>
							</span>

							<button
								type="button"
								onclick={() => openDeleteModal(order)}
								class="px-2.5 py-1 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-600 font-bold text-[11px] border border-rose-200 shadow-2xs transition-colors flex items-center gap-1 cursor-pointer shrink-0"
								title="Hapus Pesanan Ini"
							>
								<Trash2 class="w-3.5 h-3.5" />
								<span>Hapus</span>
							</button>
						</div>
					</div>

					<!-- Items List -->
					<div class="p-5 flex-1 space-y-3">
						<div class="space-y-2 max-h-48 overflow-y-auto pr-1">
							{#each (order.items || []) as item}
								<div class="flex items-start justify-between text-xs py-1 border-b border-slate-50 last:border-0">
									<div class="flex items-start gap-2">
										<span class="w-5 h-5 rounded-md bg-slate-100 text-slate-700 font-bold flex items-center justify-center text-[11px] shrink-0">
											{item.quantity}x
										</span>
										<div>
											<span class="font-bold text-slate-800">{item.menu_name_snapshot}</span>
											{#if item.notes}
												<p class="text-[10px] text-amber-600 italic">Catatan: {item.notes}</p>
											{/if}
										</div>
									</div>
									<span class="font-semibold text-slate-600 shrink-0">
										{formatRupiah(item.subtotal)}
									</span>
								</div>
							{/each}
						</div>

						{#if order.notes}
							<div class="p-2.5 bg-amber-50/80 rounded-xl border border-amber-200/60 text-[11px] text-amber-800">
								<span class="font-bold">Catatan Meja:</span> {order.notes}
							</div>
						{/if}
					</div>

					<!-- Pricing & Actions Footer -->
					<div class="p-5 pt-4 bg-slate-50/80 border-t border-slate-100 space-y-4">
						<div class="flex items-baseline justify-between">
							<span class="text-xs font-bold text-slate-500 uppercase tracking-wider">Total Tagihan</span>
							<span class="text-xl font-black font-['Outfit'] text-orange-600">
								{formatRupiah(order.total)}
							</span>
						</div>

						<!-- Proof of Payment Badge (if uploaded by customer) -->
						{#if getOrderProof(order.id)}
							{@const proof = getOrderProof(order.id)!}
							<div class="p-2.5 bg-amber-50/90 border border-amber-200/90 rounded-2xl flex items-center justify-between gap-2 shadow-2xs">
								<div class="flex items-center gap-2 min-w-0">
									<img src={proof.image} alt="Bukti" class="w-8 h-8 rounded-lg object-cover border border-amber-200 shrink-0" />
									<div class="text-[11px] truncate">
										<span class="font-bold text-amber-950 truncate flex items-center gap-1">
											<Camera class="w-3 h-3 text-amber-600" />
											Ada Bukti Foto Pelanggan
										</span>
										<span class="text-amber-700 text-[10px] block truncate">{proof.fileName || 'Foto Bukti'} &bull; {proof.fileSize}</span>
									</div>
								</div>
								<button
									type="button"
									onclick={() => {
										viewProofModalData = {
											...proof,
											orderNumber: order.order_number
										};
									}}
									class="px-2.5 py-1 rounded-xl bg-white hover:bg-amber-100 text-amber-900 font-bold text-[11px] border border-amber-300 shadow-2xs transition-colors shrink-0 flex items-center gap-1 cursor-pointer"
								>
									<Eye class="w-3 h-3" />
									<span>Lihat Foto</span>
								</button>
							</div>
						{/if}

						<!-- Action Buttons -->
						{#if order.status === 'CANCELLED'}
							<!-- Tampilan Aksi untuk Pesanan Dibatalkan / Nyangkut -->
							<div class="space-y-2 pt-1">
								<div class="p-2 bg-rose-50 border border-rose-200 rounded-xl text-rose-800 text-[11px] flex items-center justify-between">
									<span class="font-bold flex items-center gap-1">
										<XCircle class="w-3.5 h-3.5 text-rose-600" />
										Pesanan Dibatalkan
									</span>
									<span class="text-[10px] text-rose-600 font-medium">Data bisa dibersihkan</span>
								</div>
								<div class="flex items-center gap-2">
									<button
										type="button"
										onclick={() => handleRestoreOrder(order)}
										class="flex-1 py-2 px-2.5 bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 rounded-xl font-bold text-xs transition-colors flex items-center justify-center gap-1 cursor-pointer"
										title="Pulihkan pesanan ini kembali ke dapur"
									>
										<RotateCcw class="w-3.5 h-3.5 text-blue-600" />
										<span>Pulihkan</span>
									</button>
									<button
										type="button"
										onclick={() => handleDeleteOrder(order)}
										disabled={deletingOrderId === order.id}
										class="flex-1 py-2 px-2.5 bg-rose-600 hover:bg-rose-700 active:scale-98 text-white rounded-xl font-bold text-xs shadow-xs transition-all flex items-center justify-center gap-1 cursor-pointer disabled:opacity-50"
										title="Hapus permanen pesanan nyangkut ini"
									>
										{#if deletingOrderId === order.id}
											<RefreshCw class="w-3.5 h-3.5 animate-spin" />
											<span>Hapus...</span>
										{:else}
											<Trash2 class="w-3.5 h-3.5" />
											<span>Hapus</span>
										{/if}
									</button>
									<button
										type="button"
										onclick={() => openEditModal(order)}
										class="p-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-600 transition-colors cursor-pointer"
										title="Kelola Lengkap"
									>
										<SlidersHorizontal class="w-3.5 h-3.5" />
									</button>
								</div>
							</div>
						{:else if order.payment_status !== 'PAID'}
							<div class="flex items-center gap-2">
								<button
									type="button"
									onclick={() => openPaymentModal(order)}
									class="flex-1 py-2.5 px-3 bg-orange-600 hover:bg-orange-700 active:scale-98 text-white rounded-2xl font-black text-xs shadow-md shadow-orange-600/25 transition-all flex items-center justify-center gap-1.5 cursor-pointer"
								>
									<Check class="w-4 h-4" />
									<span>Terima Kasir / Tunai</span>
								</button>
								<button
									type="button"
									onclick={() => openEditModal(order)}
									class="p-2.5 rounded-2xl bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 transition-colors cursor-pointer"
									title="Kelola / Edit Status Pesanan (CRUD)"
								>
									<SlidersHorizontal class="w-4 h-4 text-orange-600" />
								</button>
								<button
									type="button"
									onclick={() => openDeleteModal(order)}
									class="p-2.5 rounded-2xl bg-rose-50 hover:bg-rose-100 text-rose-600 border border-rose-200 transition-colors cursor-pointer"
									title="Hapus Pesanan Ini"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							</div>
						{:else}
							<div class="flex items-center justify-between text-xs text-slate-500 font-semibold pt-1">
								<span class="flex items-center gap-1 text-emerald-600 font-bold">
									<CheckCircle2 class="w-4 h-4" />
									{#if order.status === 'COMPLETED'}
										Selesai Diantar
									{:else if order.status === 'READY'}
										Siap Diantar
									{:else if order.status === 'PREPARING'}
										Sedang Dimasak
									{:else}
										Diteruskan ke Dapur
									{/if}
								</span>
								<div class="flex items-center gap-1.5">
									<button
										type="button"
										onclick={() => {
											successModalData = {
												order: order,
												paidAmount: order.total,
												change: 0,
												paymentMethod: order.payment_method || 'CASH'
											};
										}}
										class="text-orange-600 hover:underline font-bold text-[11px] cursor-pointer mr-1"
									>
										Lihat Struk
									</button>
									<button
										type="button"
										onclick={() => openEditModal(order)}
										class="px-2 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-[11px] transition-colors flex items-center gap-1 cursor-pointer"
										title="Kelola Status Pesanan (CRUD)"
									>
										<SlidersHorizontal class="w-3.5 h-3.5 text-orange-600" />
										<span>Kelola</span>
									</button>
									<button
										type="button"
										onclick={() => openDeleteModal(order)}
										class="px-2 py-1.5 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-600 border border-rose-200 font-bold text-[11px] transition-colors flex items-center gap-1 cursor-pointer"
										title="Hapus Pesanan Ini"
									>
										<Trash2 class="w-3.5 h-3.5" />
										<span>Hapus</span>
									</button>
								</div>
							</div>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- ==================== SMART CASH PAYMENT MODAL ==================== -->
{#if paymentModalOrder}
	<div class="fixed inset-0 bg-slate-950/70 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-3xl p-6 sm:p-8 w-full max-w-lg space-y-6 shadow-2xl animate-in fade-in zoom-in-95 duration-150 max-h-[95vh] overflow-y-auto">
			<!-- Modal Header -->
			<div class="flex items-start justify-between pb-4 border-b border-slate-100">
				<div>
					<span class="text-[11px] font-bold text-orange-600 uppercase tracking-wider">
						Konfirmasi Pembayaran
					</span>
					<h3 class="text-xl font-black text-slate-900 font-['Outfit'] flex items-center gap-2">
						{paymentModalOrder.table_name || `Meja ${paymentModalOrder.table_id || '?'}`}
						<span class="text-xs font-mono font-bold text-slate-400">({paymentModalOrder.order_number})</span>
					</h3>
				</div>
				<button
					type="button"
					onclick={() => (paymentModalOrder = null)}
					class="p-1 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100"
				>
					✕
				</button>
			</div>

			<!-- Method Selector -->
			<div class="space-y-2">
				<div class="text-xs font-bold text-slate-600 uppercase tracking-wider">Pilih Metode Pembayaran</div>
				<div class="grid grid-cols-3 gap-2">
					<button
						type="button"
						onclick={() => (selectedMethod = 'CASH')}
						class="py-2.5 px-3 rounded-2xl border text-xs font-bold transition-all flex flex-col items-center gap-1 {selectedMethod === 'CASH'
							? 'bg-orange-50 border-orange-500 text-orange-700 shadow-xs'
							: 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
					>
						<Banknote class="w-4 h-4" />
						<span>Uang Tunai</span>
					</button>

					<button
						type="button"
						onclick={() => (selectedMethod = 'QRIS_MANUAL')}
						class="py-2.5 px-3 rounded-2xl border text-xs font-bold transition-all flex flex-col items-center gap-1 {selectedMethod === 'QRIS_MANUAL'
							? 'bg-orange-50 border-orange-500 text-orange-700 shadow-xs'
							: 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
					>
						<Sparkles class="w-4 h-4" />
						<span>QRIS Kasir</span>
					</button>

					<button
						type="button"
						onclick={() => (selectedMethod = 'DEBIT')}
						class="py-2.5 px-3 rounded-2xl border text-xs font-bold transition-all flex flex-col items-center gap-1 {selectedMethod === 'DEBIT'
							? 'bg-orange-50 border-orange-500 text-orange-700 shadow-xs'
							: 'border-slate-200 text-slate-600 hover:bg-slate-50'}"
					>
						<DollarSign class="w-4 h-4" />
						<span>Debit / EDC</span>
					</button>
				</div>
			</div>

			<!-- Total Tagihan Banner -->
			<div class="p-4 rounded-2xl bg-slate-900 text-white flex items-center justify-between">
				<div>
					<span class="text-[10px] uppercase font-bold text-slate-400">Total Tagihan Wajib Bayar</span>
					<div class="text-2xl font-black font-['Outfit'] text-orange-400">
						{formatRupiah(paymentModalOrder.total)}
					</div>
				</div>
				<div class="text-right text-[11px] text-slate-400">
					<span>{paymentModalOrder.items?.length || 0} Menu Item</span>
				</div>
			</div>

			{#if selectedMethod === 'CASH'}
				<!-- Quick Proof notice for Cash if customer uploaded something -->
				{#if getOrderProof(paymentModalOrder)}
					{@const proof = getOrderProof(paymentModalOrder)!}
					<div class="p-3 bg-amber-50/90 border border-amber-200/90 rounded-2xl flex items-center justify-between gap-2 text-left">
						<div class="flex items-center gap-2 min-w-0">
							<img src={proof.image} alt="Bukti" class="w-8 h-8 rounded-lg object-cover border border-amber-200 shrink-0" />
							<div class="text-[11px] truncate">
								<span class="font-bold text-amber-950 truncate flex items-center gap-1">
									<Camera class="w-3 h-3 text-amber-600" />
									Pelanggan mengunggah foto bukti
								</span>
								<span class="text-amber-700 text-[10px] block truncate">{proof.fileName || 'Foto Bukti'} &bull; {proof.uploadedAt}</span>
							</div>
						</div>
						<button
							type="button"
							onclick={() => {
								viewProofModalData = {
									...proof,
									orderNumber: paymentModalOrder!.order_number
								};
							}}
							class="px-2.5 py-1 rounded-xl bg-white hover:bg-amber-100 text-amber-900 font-bold text-[11px] border border-amber-300 shadow-2xs transition-colors shrink-0 flex items-center gap-1 cursor-pointer"
						>
							<Eye class="w-3 h-3" />
							<span>Lihat</span>
						</button>
					</div>
				{/if}

				<!-- Cash Calculator -->
				<div class="space-y-3">
					<div class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center justify-between">
						<span>Nominal Uang Diterima Pelanggan</span>
						<span class="text-[11px] text-orange-600 font-bold">Kalkulator Kasir</span>
					</div>

					<!-- Quick Nominal Presets -->
					<div class="grid grid-cols-3 sm:grid-cols-5 gap-1.5">
						<button
							type="button"
							onclick={() => setQuickAmount(paymentModalOrder!.total)}
							class="py-2 px-1 text-[11px] font-bold rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-700"
						>
							Uang Pas
						</button>
						<button
							type="button"
							onclick={() => setQuickAmount(10000)}
							class="py-2 px-1 text-[11px] font-bold rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-700"
						>
							Rp 10.000
						</button>
						<button
							type="button"
							onclick={() => setQuickAmount(20000)}
							class="py-2 px-1 text-[11px] font-bold rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-700"
						>
							Rp 20.000
						</button>
						<button
							type="button"
							onclick={() => setQuickAmount(50000)}
							class="py-2 px-1 text-[11px] font-bold rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-700"
						>
							Rp 50.000
						</button>
						<button
							type="button"
							onclick={() => setQuickAmount(100000)}
							class="py-2 px-1 text-[11px] font-bold rounded-xl border border-slate-200 hover:bg-slate-100 text-slate-700"
						>
							Rp 100.000
						</button>
					</div>

					<!-- Custom Paid Amount Input -->
					<div class="relative">
						<span class="absolute left-4 top-1/2 -translate-y-1/2 font-bold text-slate-400 text-sm">Rp</span>
						<input
							type="number"
							bind:value={paidAmount}
							min={paymentModalOrder.total}
							step="1000"
							class="w-full pl-12 pr-4 py-3 rounded-2xl bg-slate-50 border border-slate-200 text-base font-black text-slate-900 focus:outline-hidden focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500"
						/>
					</div>

					<!-- Change Amount Result Box -->
					<div class="p-4 rounded-2xl border transition-colors {isAmountSufficient
						? 'bg-emerald-50/70 border-emerald-200 text-emerald-900'
						: 'bg-rose-50/70 border-rose-200 text-rose-900'}">
						<div class="flex items-center justify-between">
							<span class="text-xs font-bold uppercase tracking-wider">
								{isAmountSufficient ? 'Uang Kembalian Pelanggan' : 'Status Pembayaran'}
							</span>
							{#if isAmountSufficient}
								<span class="text-lg font-black font-['Outfit'] text-emerald-700">
									{formatRupiah(calculatedChange)}
								</span>
							{:else}
								<span class="text-xs font-bold text-rose-600 flex items-center gap-1">
									<AlertCircle class="w-3.5 h-3.5" />
									Uang Masih Kurang
								</span>
							{/if}
						</div>
					</div>
				</div>
			{:else if selectedMethod === 'QRIS_MANUAL'}
				<!-- QRIS Verification Section with Customer Proof of Payment -->
				<div class="space-y-3">
					<div class="text-xs font-bold text-slate-700 uppercase tracking-wider flex items-center justify-between">
						<span class="flex items-center gap-1.5">
							<Sparkles class="w-4 h-4 text-orange-600" />
							Konfirmasi Pembayaran QRIS
						</span>
						<span class="text-[11px] text-slate-400 font-medium">QRIS Kasir</span>
					</div>

					{#if getOrderProof(paymentModalOrder)}
						{@const proof = getOrderProof(paymentModalOrder)!}
						<div class="p-4 bg-gradient-to-br from-amber-50/90 to-orange-50/80 border-2 border-orange-200/90 rounded-2xl space-y-3 shadow-xs text-left">
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-1.5 text-xs font-extrabold text-amber-950">
									<Camera class="w-4 h-4 text-orange-600" />
									<span>Bukti Foto dari Pelanggan</span>
								</div>
								<span class="text-[10px] bg-white text-amber-800 font-bold px-2 py-0.5 rounded-full border border-amber-200">
									{proof.uploadedAt || 'Terkirim'}
								</span>
							</div>

							<!-- Image Preview Card -->
							<div class="flex items-center gap-3.5 bg-white p-3 rounded-xl border border-orange-150 shadow-2xs">
								<button
									type="button"
									onclick={() => {
										viewProofModalData = {
											...proof,
											orderNumber: paymentModalOrder!.order_number
										};
									}}
									class="relative w-16 h-16 rounded-xl overflow-hidden bg-slate-900 border-2 border-orange-400 group shrink-0 cursor-pointer shadow-xs"
									title="Klik untuk memperbesar foto bukti"
								>
									<img src={proof.image} alt="Bukti Transfer QRIS" class="w-full h-full object-cover group-hover:scale-110 transition-transform duration-200" />
									<div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center text-white">
										<Eye class="w-4 h-4" />
									</div>
								</button>

								<div class="min-w-0 flex-1 space-y-1">
									<p class="text-xs font-black text-slate-900 truncate" title={proof.fileName}>
										{proof.fileName || 'bukti-qris.jpg'}
									</p>
									<p class="text-[11px] text-slate-500 font-medium">
										{proof.fileSize}
									</p>
									<div class="pt-1 flex items-center gap-2 flex-wrap">
										<button
											type="button"
											onclick={() => {
												viewProofModalData = {
													...proof,
													orderNumber: paymentModalOrder!.order_number
												};
											}}
											class="px-2.5 py-1 bg-orange-600 hover:bg-orange-700 text-white rounded-lg text-[11px] font-bold flex items-center gap-1 shadow-2xs transition-colors cursor-pointer"
										>
											<Eye class="w-3 h-3" />
											<span>Lihat Foto Penuh</span>
										</button>
										{#if proof.gdriveUrl}
											<a
												href={proof.gdriveUrl}
												target="_blank"
												rel="noopener noreferrer"
												class="px-2.5 py-1 bg-white hover:bg-slate-100 text-blue-600 border border-blue-200 rounded-lg text-[11px] font-bold flex items-center gap-1 transition-colors"
											>
												<span>Folder GDrive</span>
											</a>
										{/if}
									</div>
								</div>
							</div>

							<div class="bg-amber-100/70 border border-amber-300/80 rounded-xl p-2.5 flex items-start gap-2">
								<CheckCircle2 class="w-4 h-4 text-emerald-600 shrink-0 mt-0.5" />
								<p class="text-[11px] text-amber-900 leading-tight">
									Pastikan nominal pada foto bukti pembayaran sesuai dengan total tagihan <strong>{formatRupiah(paymentModalOrder.total)}</strong>.
								</p>
							</div>
						</div>
					{:else}
						<div class="p-4 bg-slate-50 border border-slate-200 rounded-2xl space-y-2 text-left">
							<div class="flex items-center gap-2 text-slate-700 font-bold text-xs">
								<AlertCircle class="w-4 h-4 text-amber-500 shrink-0" />
								<span>Belum Ada Foto Bukti dari Pelanggan</span>
							</div>
							<p class="text-[11px] text-slate-500 leading-relaxed">
								Pelanggan belum mengirimkan foto bukti pembayaran QRIS. Anda tetap dapat mengonfirmasi jika saldo rekening / notifikasi soundbox QRIS sudah masuk.
							</p>
						</div>
					{/if}
				</div>
			{:else if selectedMethod === 'DEBIT'}
				<!-- Debit / EDC Section -->
				<div class="p-4 bg-slate-50 border border-slate-200 rounded-2xl space-y-2 text-left">
					<div class="flex items-center gap-2 text-slate-700 font-bold text-xs">
						<DollarSign class="w-4 h-4 text-orange-600 shrink-0" />
						<span>Pembayaran Kartu Debit / Mesin EDC</span>
					</div>
					<p class="text-[11px] text-slate-500 leading-relaxed">
						Pastikan transaksi pada mesin EDC telah berhasil (APPROVED) dan struk EDC telah keluar sebelum menekan tombol konfirmasi.
					</p>
				</div>
			{/if}

			<!-- Submit & Cancel Buttons -->
			<div class="pt-2 flex items-center gap-3">
				<button
					type="button"
					onclick={() => (paymentModalOrder = null)}
					class="w-1/3 py-3 rounded-2xl border border-slate-200 text-slate-700 font-bold text-xs hover:bg-slate-100 transition-colors"
				>
					Batal
				</button>

				<button
					type="button"
					onclick={submitPaymentConfirmation}
					disabled={submittingPayment || (selectedMethod === 'CASH' && !isAmountSufficient)}
					class="w-2/3 py-3 rounded-2xl bg-orange-600 hover:bg-orange-700 active:scale-98 disabled:opacity-50 text-white font-black text-xs shadow-lg shadow-orange-600/30 transition-all flex items-center justify-center gap-2 cursor-pointer"
				>
					{#if submittingPayment}
						<RefreshCw class="w-4 h-4 animate-spin" />
						<span>Memproses...</span>
					{:else}
						<CheckCircle2 class="w-4 h-4" />
						<span>Konfirmasi & Teruskan ke Dapur</span>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== RECEIPT / CONFIRMATION SUCCESS MODAL ==================== -->
{#if successModalData}
	<div class="fixed inset-0 bg-slate-950/70 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-3xl p-6 sm:p-8 w-full max-w-sm space-y-6 shadow-2xl animate-in fade-in zoom-in-95 duration-150 text-center">
			<div class="w-16 h-16 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center mx-auto">
				<CheckCircle2 class="w-8 h-8" />
			</div>

			<div class="space-y-1">
				<h3 class="text-xl font-black text-slate-900 font-['Outfit']">Pembayaran Berhasil!</h3>
				<p class="text-xs text-slate-500 font-medium">
					Pesanan meja <span class="font-bold text-slate-800">{successModalData.order.table_name || 'Meja'}</span> telah lunas dan otomatis dikirim ke Kitchen Board.
				</p>
			</div>

			<!-- Thermal Receipt Card -->
			<div class="p-4 bg-slate-50 rounded-2xl border border-dashed border-slate-300 text-left text-xs space-y-2">
				<div class="flex justify-between font-mono font-bold text-slate-800">
					<span>No: {successModalData.order.order_number}</span>
					<span>{successModalData.paymentMethod}</span>
				</div>
				<div class="border-t border-slate-200 pt-2 space-y-1">
					<div class="flex justify-between text-slate-500">
						<span>Total Tagihan:</span>
						<span class="font-bold text-slate-800">{formatRupiah(successModalData.order.total)}</span>
					</div>
					<div class="flex justify-between text-slate-500">
						<span>Uang Diterima:</span>
						<span class="font-bold text-slate-800">{formatRupiah(successModalData.paidAmount)}</span>
					</div>
					<div class="flex justify-between text-emerald-700 font-black border-t border-slate-200 pt-1 text-sm">
						<span>Kembalian:</span>
						<span>{formatRupiah(successModalData.change)}</span>
					</div>
				</div>
			</div>

			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={printReceipt}
					class="flex-1 py-2.5 px-3 rounded-2xl border border-slate-200 text-slate-700 font-bold text-xs hover:bg-slate-100 flex items-center justify-center gap-1.5 transition-colors"
				>
					<Printer class="w-4 h-4" />
					<span>Cetak Struk</span>
				</button>
				<button
					type="button"
					onclick={() => (successModalData = null)}
					class="flex-1 py-2.5 px-3 rounded-2xl bg-slate-900 text-white font-bold text-xs hover:bg-slate-800 transition-colors"
				>
					Tutup
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== VIEW PROOF OF PAYMENT MODAL ==================== -->
{#if viewProofModalData}
	<div class="fixed inset-0 bg-slate-950/80 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-3xl max-w-lg w-full overflow-hidden shadow-2xl animate-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
			<div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<div class="w-8 h-8 rounded-xl bg-orange-50 border border-orange-200 flex items-center justify-center text-orange-600">
						<ImageIcon class="w-4 h-4" />
					</div>
					<div>
						<h3 class="text-xs font-black text-slate-900 truncate max-w-xs">
							Bukti Pembayaran ({viewProofModalData.orderNumber})
						</h3>
						<p class="text-[10px] text-slate-400 font-medium">
							{viewProofModalData.fileName || 'Foto Bukti'} &bull; {viewProofModalData.fileSize} &bull; Diunggah {viewProofModalData.uploadedAt || 'Baru saja'}
						</p>
					</div>
				</div>
				<button
					type="button"
					onclick={() => (viewProofModalData = null)}
					class="w-8 h-8 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 flex items-center justify-center cursor-pointer"
				>
					<XCircle class="w-4 h-4" />
				</button>
			</div>

			<div class="p-4 bg-slate-950 flex items-center justify-center flex-1 overflow-auto">
				<img 
					src={viewProofModalData.image} 
					alt="Bukti Transfer Penuh" 
					class="max-w-full max-h-[65vh] object-contain rounded-lg shadow-md"
				/>
			</div>

			<div class="p-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between gap-2">
				<span class="text-[11px] text-slate-500 font-medium truncate">
					Verifikasi nominal & tujuan transfer
				</span>
				<div class="flex items-center gap-2 shrink-0">
					{#if viewProofModalData.gdriveUrl}
						<a
							href={viewProofModalData.gdriveUrl}
							target="_blank"
							rel="noopener noreferrer"
							class="px-3 py-2 bg-blue-50 hover:bg-blue-100 text-blue-700 font-bold text-xs rounded-xl border border-blue-200 transition-colors inline-flex items-center gap-1.5"
						>
							<span>Buka di Google Drive</span>
						</a>
					{/if}
					<button
						type="button"
						onclick={() => (viewProofModalData = null)}
						class="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs rounded-xl transition-colors cursor-pointer"
					>
						Tutup
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== CRUD: KELOLA & EDIT STATUS PESANAN MODAL ==================== -->
{#if editModalOrder}
	<div class="fixed inset-0 bg-slate-950/75 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-white rounded-3xl p-6 sm:p-7 w-full max-w-lg space-y-5 shadow-2xl animate-in zoom-in-95 duration-150 max-h-[92vh] overflow-y-auto">
			<!-- Modal Header -->
			<div class="flex items-start justify-between pb-3 border-b border-slate-100">
				<div>
					<span class="text-[11px] font-bold text-orange-600 uppercase tracking-wider flex items-center gap-1">
						<SlidersHorizontal class="w-3.5 h-3.5" />
						Manajemen / CRUD Pesanan
					</span>
					<h2 class="text-lg font-black text-slate-900 font-['Outfit'] mt-0.5">
						#{editModalOrder.order_number} &bull; {editModalOrder.table_name || `Meja ${editModalOrder.table_id || '?'}`}
					</h2>
					<p class="text-xs text-slate-400 font-medium">
						Waktu: {new Date(editModalOrder.created_at).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })}
					</p>
				</div>
				<button
					type="button"
					onclick={() => (editModalOrder = null)}
					class="w-8 h-8 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-500 flex items-center justify-center cursor-pointer transition-colors"
				>
					<XCircle class="w-5 h-5" />
				</button>
			</div>

			<!-- Status Alur Pesanan -->
			<div class="space-y-2">
				<div class="block text-xs font-bold text-slate-700 uppercase tracking-wider">
					Ubah Status Alur Dapur / Pesanan:
				</div>
				<div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
					{#each [
						{ key: 'WAITING_PAYMENT', label: 'Menunggu Bayar', color: 'border-amber-400 bg-amber-50 text-amber-800' },
						{ key: 'CONFIRMED', label: 'Diteruskan Dapur', color: 'border-blue-400 bg-blue-50 text-blue-800' },
						{ key: 'PREPARING', label: 'Sedang Dimasak', color: 'border-orange-400 bg-orange-50 text-orange-800' },
						{ key: 'READY', label: 'Siap Diantar', color: 'border-emerald-400 bg-emerald-50 text-emerald-800' },
						{ key: 'COMPLETED', label: 'Selesai', color: 'border-slate-400 bg-slate-100 text-slate-800' },
						{ key: 'CANCELLED', label: 'Dibatalkan', color: 'border-rose-400 bg-rose-50 text-rose-800' }
					] as st}
						<button
							type="button"
							onclick={() => (editStatus = st.key)}
							class="p-2.5 rounded-xl border text-left text-xs font-bold transition-all cursor-pointer {editStatus === st.key
								? `${st.color} ring-2 ring-orange-500 shadow-xs font-black`
								: 'border-slate-200 bg-white hover:bg-slate-50 text-slate-600'}"
						>
							<div class="flex items-center justify-between">
								<span>{st.label}</span>
								{#if editStatus === st.key}
									<Check class="w-3.5 h-3.5 shrink-0" />
								{/if}
							</div>
						</button>
					{/each}
				</div>
			</div>

			<!-- Status Pembayaran -->
			<div class="space-y-2">
				<div class="block text-xs font-bold text-slate-700 uppercase tracking-wider">
					Status Pembayaran Tagihan:
				</div>
				<div class="grid grid-cols-2 gap-2">
					<button
						type="button"
						onclick={() => (editPaymentStatus = 'UNPAID')}
						class="p-2.5 rounded-xl border text-xs font-bold transition-all cursor-pointer flex items-center justify-between {editPaymentStatus === 'UNPAID'
							? 'border-amber-400 bg-amber-50 text-amber-900 ring-2 ring-amber-500 font-black'
							: 'border-slate-200 bg-white hover:bg-slate-50 text-slate-600'}"
					>
						<span class="flex items-center gap-1.5">
							<Clock class="w-3.5 h-3.5 text-amber-600" />
							BELUM BAYAR (UNPAID)
						</span>
						{#if editPaymentStatus === 'UNPAID'}
							<Check class="w-3.5 h-3.5 text-amber-700" />
						{/if}
					</button>

					<button
						type="button"
						onclick={() => (editPaymentStatus = 'PAID')}
						class="p-2.5 rounded-xl border text-xs font-bold transition-all cursor-pointer flex items-center justify-between {editPaymentStatus === 'PAID'
							? 'border-emerald-400 bg-emerald-50 text-emerald-900 ring-2 ring-emerald-500 font-black'
							: 'border-slate-200 bg-white hover:bg-slate-50 text-slate-600'}"
					>
						<span class="flex items-center gap-1.5">
							<CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" />
							SUDAH LUNAS (PAID)
						</span>
						{#if editPaymentStatus === 'PAID'}
							<Check class="w-3.5 h-3.5 text-emerald-700" />
						{/if}
					</button>
				</div>
			</div>

			<!-- Catatan Meja -->
			<div class="space-y-1.5">
				<label for="edit-notes-input" class="block text-xs font-bold text-slate-700 uppercase tracking-wider">
					Catatan Meja / Pesanan:
				</label>
				<input
					id="edit-notes-input"
					type="text"
					bind:value={editNotes}
					placeholder="Catatan tambahan meja (opsional)..."
					class="w-full px-3 py-2 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium focus:outline-hidden focus:ring-2 focus:ring-orange-500/20 focus:border-orange-500 transition-all"
				/>
			</div>

			<!-- Ringkasan Item Pesanan -->
			<div class="p-3 bg-slate-50 rounded-2xl border border-slate-200/80 space-y-2">
				<div class="flex items-center justify-between text-xs font-bold text-slate-600 pb-1 border-b border-slate-200">
					<span>Rincian Item ({editModalOrder.items?.length || 0})</span>
					<span class="text-orange-600 font-black font-['Outfit']">{formatRupiah(editModalOrder.total)}</span>
				</div>
				<div class="max-h-32 overflow-y-auto space-y-1 pr-1 text-xs text-slate-700">
					{#each (editModalOrder.items || []) as item}
						<div class="flex items-center justify-between py-0.5">
							<span class="truncate">{item.quantity}x {item.menu_name_snapshot}</span>
							<span class="font-semibold text-slate-500 shrink-0">{formatRupiah(item.subtotal)}</span>
						</div>
					{/each}
				</div>
			</div>

			<!-- Footer Actions -->
			<div class="pt-2 border-t border-slate-100 flex flex-col sm:flex-row items-center justify-between gap-3">
				<!-- Tombol Hapus Permanen (untuk pesanan nyangkut/spam) -->
				<button
					type="button"
					onclick={() => handleDeleteOrder(editModalOrder!)}
					disabled={deletingOrderId === editModalOrder.id}
					class="w-full sm:w-auto px-3.5 py-2.5 rounded-xl bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors cursor-pointer disabled:opacity-50"
					title="Hapus permanen pesanan nyangkut ini dari database"
				>
					<Trash2 class="w-3.5 h-3.5 text-rose-600" />
					<span>Hapus Permanen Pesanan</span>
				</button>

				<div class="flex items-center gap-2 w-full sm:w-auto">
					<button
						type="button"
						onclick={() => (editModalOrder = null)}
						class="flex-1 sm:flex-initial px-4 py-2.5 rounded-xl border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-100 transition-colors cursor-pointer"
					>
						Batal
					</button>
					<button
						type="button"
						onclick={saveEditOrder}
						disabled={submittingEdit}
						class="flex-1 sm:flex-initial px-5 py-2.5 rounded-xl bg-orange-600 hover:bg-orange-700 active:scale-95 text-white font-bold text-xs shadow-md shadow-orange-600/25 transition-all flex items-center justify-center gap-1.5 cursor-pointer disabled:opacity-50"
					>
						{#if submittingEdit}
							<RefreshCw class="w-3.5 h-3.5 animate-spin" />
							<span>Menyimpan...</span>
						{:else}
							<Check class="w-3.5 h-3.5" />
							<span>Simpan Perubahan</span>
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- ==================== DELETE ORDER CONFIRMATION MODAL ==================== -->
{#if deleteModalOrder}
	<div
		class="fixed inset-0 bg-slate-950/70 backdrop-blur-xs z-50 flex items-center justify-center p-4"
		onclick={(e) => {
			if (e.target === e.currentTarget && !deletingOrderId) deleteModalOrder = null;
		}}
		onkeydown={(e) => {
			if (e.key === 'Escape' && !deletingOrderId) deleteModalOrder = null;
		}}
		tabindex="-1"
		role="dialog"
		aria-modal="true"
	>
		<div class="bg-white rounded-3xl p-6 sm:p-7 w-full max-w-md space-y-5 shadow-2xl animate-in fade-in zoom-in-95 duration-150 border border-rose-100">
			<!-- Header -->
			<div class="flex items-center gap-3">
				<div class="w-12 h-12 rounded-2xl bg-rose-100 text-rose-600 flex items-center justify-center shrink-0">
					<Trash2 class="w-6 h-6" />
				</div>
				<div>
					<span class="text-[10px] font-black uppercase tracking-wider text-rose-600 bg-rose-50 px-2 py-0.5 rounded-md border border-rose-200">
						Hapus Pesanan Permanen
					</span>
					<h3 class="text-lg font-black text-slate-900 font-['Outfit'] mt-0.5">
						#{deleteModalOrder.order_number}
					</h3>
				</div>
			</div>

			<!-- Order Summary -->
			<div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-100 space-y-2 text-xs text-slate-600">
				<div class="flex justify-between">
					<span class="text-slate-400">Meja:</span>
					<span class="font-bold text-slate-800">{deleteModalOrder.table_name || `Meja ${deleteModalOrder.table_id || '?'}`}</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Total Tagihan:</span>
					<span class="font-bold text-orange-600 font-['Outfit']">{formatRupiah(deleteModalOrder.total)}</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Status Pesanan:</span>
					<span class="font-bold text-slate-800">{deleteModalOrder.status} ({deleteModalOrder.payment_status})</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Waktu Pesan:</span>
					<span class="font-medium text-slate-700">{new Date(deleteModalOrder.created_at).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })}</span>
				</div>
				{#if deleteModalOrder.items && deleteModalOrder.items.length > 0}
					<div class="pt-1.5 border-t border-slate-200/60">
						<span class="text-slate-400 block mb-1">Menu yang dipesan ({deleteModalOrder.items.length} item):</span>
						<div class="max-h-24 overflow-y-auto space-y-1 pr-1 text-[11px]">
							{#each deleteModalOrder.items as it}
								<div class="flex justify-between text-slate-700">
									<span>{it.quantity}x {it.menu_name_snapshot}</span>
									<span class="font-semibold text-slate-500">{formatRupiah(it.subtotal)}</span>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>

			<!-- Warning Message -->
			<div class="p-3 bg-rose-50/90 rounded-2xl border border-rose-200 text-[11px] text-rose-800 flex items-start gap-2">
				<AlertTriangle class="w-4 h-4 text-rose-600 shrink-0 mt-0.5" />
				<div>
					<strong>Perhatian:</strong> Seluruh data pesanan ini (termasuk tagihan, status dapur, dan transaksi) akan <strong>dihapus permanen</strong> dari sistem dan database. Tindakan ini tidak dapat dibatalkan.
				</div>
			</div>

			<!-- Action Buttons -->
			<div class="flex items-center gap-2 pt-1">
				<button
					type="button"
					onclick={() => (deleteModalOrder = null)}
					disabled={deletingOrderId === deleteModalOrder.id}
					class="flex-1 py-3 px-4 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs transition-colors cursor-pointer"
				>
					Batal
				</button>
				<button
					type="button"
					onclick={() => executeDeleteOrder(deleteModalOrder!)}
					disabled={deletingOrderId === deleteModalOrder.id}
					class="flex-1 py-3 px-4 rounded-xl bg-rose-600 hover:bg-rose-700 active:scale-95 text-white font-black text-xs shadow-md shadow-rose-600/30 transition-all flex items-center justify-center gap-2 cursor-pointer disabled:opacity-50"
				>
					{#if deletingOrderId === deleteModalOrder.id}
						<RefreshCw class="w-4 h-4 animate-spin" />
						<span>Menghapus...</span>
					{:else}
						<Trash2 class="w-4 h-4" />
						<span>Ya, Hapus Sekarang</span>
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

