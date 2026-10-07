<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api, formatRupiah, getWebSocketURL } from '$lib/api/client';
	import { cart } from '$lib/stores/cart.svelte';
	import type { Order, Payment, PublicTableInfo, PublicActiveOrderSummary } from '$lib/types';
	import confetti from 'canvas-confetti';
	import { 
		Clock, CheckCircle2, ChefHat, BellRing, Sparkles, 
		QrCode, ArrowLeft, RefreshCw, AlertCircle, Check,
		Receipt, Printer, Store, X, Eye, Wallet,
		Upload, Camera, Trash2, Image as ImageIcon, Send,
		UtensilsCrossed, Plus, ArrowRight
	} from '@lucide/svelte';
	import { uploadToGDrive, deleteFromGDrive } from '$lib/services/gdriveBucket';

	const orderId = $derived(page.params.orderId || '');
	let prevOrderId = $state('');

	let loading = $state(true);
	let error = $state<string | null>(null);
	let order = $state<Order | null>(null);
	let payment = $state<Payment | null>(null);
	let showInvoiceModal = $state(false);
	let invoiceAutoShown = $state(false);
	let selectedMethod = $state<'QRIS' | 'CASH'>('QRIS');
	let activeOrdersList = $state<PublicActiveOrderSummary[]>([]);
	let tableInfo = $state<PublicTableInfo | null>(null);
	let isRefreshing = $state(false);
	let refreshSuccess = $state(false);

	const qrToken = $derived(
		page.url.searchParams.get('token') ||
		cart.qrToken ||
		(typeof localStorage !== 'undefined' ? localStorage.getItem('last_qr_token') : '') ||
		''
	);

	const orderMenuUrl = $derived(
		qrToken ? `/order?token=${encodeURIComponent(qrToken)}&new_order=true` : '/order?new_order=true'
	);

	function getMiniStatusInfo(status: string) {
		switch (status) {
			case 'PREPARING':
				return { label: 'Sedang Dimasak', badge: 'bg-orange-500 text-white', icon: '🍳', border: 'border-orange-500' };
			case 'CONFIRMED':
				return { label: 'Belum Dimasak', badge: 'bg-blue-600 text-white', icon: '⏱️', border: 'border-blue-500' };
			case 'WAITING_PAYMENT':
				return { label: 'Belum Bayar', badge: 'bg-amber-500 text-white', icon: '💳', border: 'border-amber-500' };
			case 'READY':
				return { label: 'Siap Diantar', badge: 'bg-emerald-600 text-white', icon: '🍽️', border: 'border-emerald-500' };
			default:
				return { label: status, badge: 'bg-slate-600 text-white', icon: '📋', border: 'border-slate-500' };
		}
	}

	async function loadActiveOrders() {
		const token = qrToken;
		if (!token) return;
		try {
			const info = await api.get<PublicTableInfo>(`/public/tables/${encodeURIComponent(token)}?_t=${Date.now()}`);
			if (info) {
				tableInfo = info;
				if (Array.isArray(info.active_orders)) {
					activeOrdersList = info.active_orders;
				} else {
					activeOrdersList = [];
				}
			}
		} catch (e) {
			console.warn('Failed to load active orders list', e);
		}
	}

	// Proof of Payment State
	let proofImage = $state<string | null>(null);
	let proofFileName = $state<string>('');
	let proofFileSize = $state<string>('');
	let proofUploadedAt = $state<string>('');
	let isUploadingProof = $state(false);
	let isSubmittingProof = $state(false);
	let proofSubmitted = $state(false);
	let showProofModal = $state(false);
	let gdriveFileUrl = $state<string | null>(null);
	let gdriveFileId = $state<string | null>(null);

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

	async function loadOrder(isManual = false) {
		const currentId = orderId;
		if (!currentId) return;

		if (isManual) {
			isRefreshing = true;
		}
		const startTime = Date.now();
		try {
			const o = await api.get<Order>(`/public/orders/${currentId}?_t=${Date.now()}`);
			order = o;

			// Reset proof state untuk pesanan aktif
			proofImage = null;
			proofFileName = '';
			proofFileSize = '';
			proofUploadedAt = '';
			gdriveFileUrl = null;
			gdriveFileId = null;
			proofSubmitted = false;

			if (o.proof_url) {
				proofImage = o.proof_url;
				proofSubmitted = true;
				if (o.proof_url.startsWith('http')) {
					gdriveFileUrl = o.proof_url;
				}
			} else {
				try {
					const saved = localStorage.getItem(`payment_proof_${currentId}`);
					if (saved) {
						const d = JSON.parse(saved);
						proofImage = d.image;
						proofFileName = d.fileName;
						proofFileSize = d.fileSize;
						proofUploadedAt = d.uploadedAt;
						gdriveFileUrl = d.gdriveUrl || null;
						gdriveFileId = d.gdriveId || null;
					}
				} catch (_) {}
			}

			if (o.status === 'WAITING_PAYMENT') {
				try {
					const p = await api.post<Payment>(`/orders/${currentId}/payment`);
					payment = p;
				} catch (e) {
					console.error('Failed to create payment session', e);
				}
			} else {
				try {
					const p = await api.get<Payment>(`/orders/${currentId}/payment?_t=${Date.now()}`);
					payment = p;
				} catch (e) {
					console.log('Payment record not found or not created yet');
				}

				if (!invoiceAutoShown) {
					invoiceAutoShown = true;
					showInvoiceModal = true;
				}
			}

			if (o.status === 'COMPLETED' || o.status === 'CANCELLED') {
				try {
					const saved = localStorage.getItem(`active_order_${qrToken}`);
					if (saved === currentId) {
						localStorage.removeItem(`active_order_${qrToken}`);
					}
				} catch (_) {}
			}

			// Muat seluruh pesanan aktif meja untuk multi-order tracking
			await loadActiveOrders();

			if (isManual) {
				refreshSuccess = true;
				setTimeout(() => {
					refreshSuccess = false;
				}, 2200);
			}
		} catch (err: any) {
			error = err?.message || 'Gagal memuat status pesanan.';
		} finally {
			if (isManual) {
				const elapsed = Date.now() - startTime;
				if (elapsed < 400) {
					await new Promise((resolve) => setTimeout(resolve, 400 - elapsed));
				}
				isRefreshing = false;
			}
			loading = false;
		}
	}

	async function switchToOrder(targetId: string) {
		if (!targetId || targetId === orderId) return;
		loading = true;
		prevOrderId = targetId;
		await goto(`/order/status/${targetId}?token=${encodeURIComponent(qrToken)}`, {
			replaceState: false,
			noScroll: true,
			keepFocus: true
		});
		await loadOrder();
		connectWebSocket();
	}

	$effect(() => {
		const currentId = page.params.orderId;
		if (currentId && currentId !== prevOrderId) {
			prevOrderId = currentId;
			loadOrder();
			connectWebSocket();
		}
	});

	function connectWebSocket() {
		const currentId = orderId;
		if (!currentId) return;

		if (ws) {
			try {
				ws.close();
			} catch (_) {}
			ws = null;
		}

		const wsURL = getWebSocketURL();
		const channel = `order:${currentId}`;

		try {
			ws = new WebSocket(`${wsURL}?channel=${channel}`);

			ws.onopen = () => {
				console.log(`Connected to order websocket: ${currentId}`);
			};

			ws.onmessage = (event) => {
				try {
					const msg = JSON.parse(event.data);
					if (msg.event === 'ORDER_STATUS_CHANGED' || msg.event === 'PAYMENT_PAID') {
						if (msg.data?.order) {
							const prevStatus = order?.status;
							order = msg.data.order;
							loadActiveOrders();

							if (prevStatus === 'WAITING_PAYMENT' && order?.status === 'CONFIRMED') {
								api.get<Payment>(`/orders/${currentId}/payment?_t=${Date.now()}`).then((p) => {
									payment = p;
								}).catch(() => {});
								invoiceAutoShown = true;
								showInvoiceModal = true;
								confetti({ particleCount: 80, spread: 60, origin: { y: 0.6 } });
							} else if (order?.status === 'READY') {
								confetti({ particleCount: 120, spread: 80, origin: { y: 0.5 } });
							} else if (order?.status === 'COMPLETED') {
								try {
									const saved = localStorage.getItem(`active_order_${qrToken}`);
									if (saved === currentId) {
										localStorage.removeItem(`active_order_${qrToken}`);
									}
								} catch (_) {}
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

	async function handleFileUpload(e: Event) {
		const target = e.target as HTMLInputElement;
		if (!target.files || target.files.length === 0) return;
		const file = target.files[0];
		if (!file.type.startsWith('image/')) {
			alert('Mohon pilih file gambar (JPG, PNG, WEBP, HEIC, GIF, dll).');
			return;
		}

		isUploadingProof = true;
		proofSubmitted = false;
		proofFileName = file.name;
		proofFileSize = (file.size / 1024).toFixed(1) + ' KB';
		proofUploadedAt = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });

		// Tampilkan preview lokal instan
		const reader = new FileReader();
		reader.onload = () => {
			proofImage = reader.result as string;
		};
		reader.readAsDataURL(file);

		// Unggah secara asinkron ke bucket Google Drive
		try {
			// Jika sudah ada bukti sebelumnya yang terunggah ke GDrive, bersihkan file lama
			if (gdriveFileId || gdriveFileUrl) {
				deleteFromGDrive(gdriveFileId || gdriveFileUrl);
			}

			const result = await uploadToGDrive(file, {
				folder: 'bukti_pembayaran',
				filename: `bukti_${order?.order_number || orderId}_${file.name}`,
				compress: true
			});

			if (result.status === 'success') {
				gdriveFileUrl = result.directUrl || result.fileUrl || null;
				gdriveFileId = result.fileId || null;
				// Jika sebelumnya sudah terkirim, update otomatis ke kasir dengan URL GDrive
				if (proofSubmitted) {
					api.post<Order>(`/public/orders/${orderId}/proof`, {
						proof_url: gdriveFileUrl
					}).catch((e) => console.warn('Auto-update proof URL error:', e));
				}
			}
		} catch (err) {
			console.error('Google Drive upload error (fallback to local preview)', err);
		} finally {
			isUploadingProof = false;
			try {
				localStorage.setItem(`payment_proof_${orderId}`, JSON.stringify({
					image: proofImage,
					fileName: file.name,
					fileSize: proofFileSize,
					uploadedAt: proofUploadedAt,
					gdriveUrl: gdriveFileUrl,
					gdriveId: gdriveFileId
				}));
			} catch (_) {}
		}
	}

	async function submitProofToCashier() {
		const targetUrl = gdriveFileUrl || proofImage;
		if (!targetUrl) {
			alert('Silakan pilih foto bukti pembayaran terlebih dahulu.');
			return;
		}

		isSubmittingProof = true;
		try {
			const updated = await api.post<Order>(`/public/orders/${orderId}/proof`, {
				proof_url: targetUrl
			});
			if (updated) {
				order = updated;
			}
			proofSubmitted = true;
		} catch (err: any) {
			console.error('Gagal mengirim bukti pembayaran:', err);
			alert(err?.message || 'Gagal mengirim bukti pembayaran ke kasir. Silakan coba lagi.');
		} finally {
			isSubmittingProof = false;
		}
	}

	async function removeProof() {
		// Hapus juga file gambar bukti dari Google Drive
		if (gdriveFileId || gdriveFileUrl) {
			try {
				await deleteFromGDrive(gdriveFileId || gdriveFileUrl);
			} catch (err) {
				console.warn('Gagal menghapus bukti dari Google Drive:', err);
			}
		}
		proofImage = null;
		proofFileName = '';
		proofFileSize = '';
		proofUploadedAt = '';
		gdriveFileUrl = null;
		gdriveFileId = null;
		proofSubmitted = false;
		try {
			localStorage.removeItem(`payment_proof_${orderId}`);
		} catch (_) {}
	}

	onMount(() => {
		loadOrder();
		loadActiveOrders();
		connectWebSocket();
		try {
			const saved = localStorage.getItem(`payment_proof_${orderId}`);
			if (saved) {
				const d = JSON.parse(saved);
				proofImage = d.image;
				proofFileName = d.fileName;
				proofFileSize = d.fileSize;
				proofUploadedAt = d.uploadedAt;
				gdriveFileUrl = d.gdriveUrl || null;
				gdriveFileId = d.gdriveId || null;
			}
		} catch (_) {}
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
				href={orderMenuUrl}
				class="w-8 h-8 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center hover:bg-slate-200"
				title="Kembali ke Pilihan Menu"
			>
				<ArrowLeft class="w-4 h-4" />
			</a>
			<h1 class="font-bold text-sm text-slate-800 font-['Outfit']">Status Pesanan</h1>
			<button
				type="button"
				onclick={() => loadOrder(true)}
				disabled={isRefreshing}
				class="w-8 h-8 rounded-lg bg-slate-100 hover:bg-slate-200 active:scale-95 disabled:opacity-60 text-slate-700 flex items-center justify-center transition-all cursor-pointer relative"
				title="Perbarui Status Pesanan"
			>
				<RefreshCw class="w-4 h-4 transition-transform {isRefreshing ? 'animate-spin text-orange-600' : 'text-slate-700'}" />
			</button>
		</div>
	</header>

	{#if refreshSuccess}
		<div class="fixed top-14 left-1/2 -translate-x-1/2 z-50 bg-emerald-600 text-white text-xs font-bold px-3.5 py-1.5 rounded-full shadow-lg flex items-center gap-1.5 animate-in fade-in slide-in-from-top-2 duration-150">
			<Check class="w-3.5 h-3.5" />
			<span>Status pesanan diperbarui</span>
		</div>
	{/if}

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
				<a href={orderMenuUrl} class="inline-block text-xs font-bold text-orange-600">Pilih Menu Restoran</a>
			</div>
		{:else if order}
			<!-- Multi-Order Switcher Bar (Jika ada lebih dari 1 pesanan aktif di meja ini) -->
			{#if activeOrdersList.length > 1}
				<div class="bg-white rounded-3xl p-3.5 sm:p-4 border border-slate-200/90 shadow-sm space-y-2.5 animate-in fade-in slide-in-from-top-2">
					<div class="flex items-center justify-between px-1">
						<div class="flex items-center gap-2">
							<span class="flex h-2 w-2 relative">
								<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75"></span>
								<span class="relative inline-flex rounded-full h-2 w-2 bg-orange-500"></span>
							</span>
							<h3 class="text-xs font-black uppercase tracking-wider text-slate-800 font-['Outfit']">
								{activeOrdersList.length} Pesanan Aktif Meja Ini
							</h3>
						</div>
						<span class="text-[10px] sm:text-[11px] font-bold text-orange-600 flex items-center gap-0.5">
							Klik untuk lacak pesanan secara rinci &rarr;
						</span>
					</div>

					<div class="grid grid-cols-2 gap-2">
						{#each activeOrdersList as act, idx}
							{@const info = getMiniStatusInfo(act.status)}
							{@const isSelected = act.id === orderId}
							<button
								type="button"
								onclick={() => switchToOrder(act.id)}
								class="p-2.5 sm:p-3 rounded-2xl transition-all border flex flex-col justify-between gap-2 text-left relative cursor-pointer {
									isSelected
										? 'bg-linear-to-br from-orange-500/10 to-amber-500/5 border-orange-500 ring-2 ring-orange-500/25 shadow-md scale-[1.01]'
										: 'bg-slate-50 hover:bg-orange-50/40 hover:border-orange-300 border-slate-200/80 text-slate-600 active:scale-98'
								}"
							>
								<!-- Baris 1: Index + Order Number + Pill Sedang Dilihat / Lacak -->
								<div class="flex items-center justify-between gap-1">
									<div class="flex items-center gap-1.5 min-w-0">
										<span class="px-1.5 py-0.5 rounded-md text-[10px] font-black font-['Outfit'] shrink-0 {
											isSelected ? 'bg-orange-600 text-white' : 'bg-slate-200 text-slate-700'
										}">
											#{idx + 1}
										</span>
										<span class="font-mono text-xs font-bold truncate {isSelected ? 'text-slate-900 font-black' : 'text-slate-700'}">
											{act.order_number}
										</span>
									</div>
									{#if isSelected}
										<span class="shrink-0 px-1.5 py-0.5 rounded-md text-[9px] font-black uppercase tracking-wider bg-orange-600 text-white shadow-2xs">
											Aktif
										</span>
									{:else}
										<span class="shrink-0 text-[10px] text-orange-600 font-bold hover:underline flex items-center">
											Lacak &rarr;
										</span>
									{/if}
								</div>

								<!-- Baris 2: Mini Badge Status + Total Harga -->
								<div class="flex items-center justify-between gap-1 pt-1.5 border-t {isSelected ? 'border-orange-200/70' : 'border-slate-200/70'}">
									<span class="inline-flex items-center gap-1 px-1.5 sm:px-2 py-0.5 rounded-lg text-[9px] sm:text-[10px] font-black {info.badge} shadow-2xs truncate">
										<span>{info.icon}</span>
										<span class="truncate">{info.label}</span>
									</span>
									<span class="text-[11px] sm:text-xs font-black font-['Outfit'] shrink-0 {isSelected ? 'text-orange-600' : 'text-slate-600'}">
										{formatRupiah(act.total)}
									</span>
								</div>
							</button>
						{/each}
					</div>
				</div>
			{/if}

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

			<!-- Status Memasak / Dapur Highlight Card (Hanya tampil saat proses berlangsung / belum selesai diantar) -->
			{#if order.status !== 'COMPLETED'}
			<div class="rounded-3xl p-4.5 border shadow-sm transition-all {
				order.status === 'PREPARING'
					? 'bg-linear-to-r from-orange-500/15 via-amber-500/10 to-orange-500/5 border-orange-500/40 text-orange-950'
					: order.status === 'CONFIRMED'
					? 'bg-linear-to-r from-blue-500/15 via-indigo-500/10 to-blue-500/5 border-blue-500/40 text-blue-950'
					: order.status === 'WAITING_PAYMENT'
					? 'bg-linear-to-r from-amber-500/15 via-yellow-500/10 to-amber-500/5 border-amber-500/40 text-amber-950'
					: order.status === 'READY'
					? 'bg-linear-to-r from-emerald-500/15 via-teal-500/10 to-emerald-500/5 border-emerald-500/40 text-emerald-950'
					: 'bg-white border-slate-200 text-slate-800'
			}">
				<div class="flex items-center gap-3.5">
					<div class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 shadow-sm {
						order.status === 'PREPARING'
							? 'bg-orange-600 text-white animate-pulse'
							: order.status === 'CONFIRMED'
							? 'bg-blue-600 text-white'
							: order.status === 'WAITING_PAYMENT'
							? 'bg-amber-600 text-white'
							: order.status === 'READY'
							? 'bg-emerald-600 text-white'
							: 'bg-slate-700 text-white'
					}">
						{#if order.status === 'PREPARING'}
							<ChefHat class="w-6 h-6 animate-bounce" />
						{:else if order.status === 'CONFIRMED'}
							<Clock class="w-6 h-6" />
						{:else if order.status === 'WAITING_PAYMENT'}
							<Wallet class="w-6 h-6" />
						{:else if order.status === 'READY'}
							<BellRing class="w-6 h-6 animate-bounce" />
						{:else}
							<Sparkles class="w-6 h-6" />
						{/if}
					</div>

					<div class="min-w-0 flex-1">
						<div class="flex items-center gap-2 flex-wrap">
							<span class="text-xs uppercase tracking-wider font-extrabold {
								order.status === 'PREPARING' ? 'text-orange-700' :
								order.status === 'CONFIRMED' ? 'text-blue-700' :
								order.status === 'WAITING_PAYMENT' ? 'text-amber-800' :
								order.status === 'READY' ? 'text-emerald-700' : 'text-slate-600'
							}">
								Proses Dapur
							</span>
							<span class="text-[10px] font-black px-2 py-0.5 rounded-full {
								order.status === 'PREPARING' ? 'bg-orange-500 text-white' :
								order.status === 'CONFIRMED' ? 'bg-blue-600 text-white' :
								order.status === 'WAITING_PAYMENT' ? 'bg-amber-500 text-white' :
								order.status === 'READY' ? 'bg-emerald-600 text-white' : 'bg-slate-500 text-white'
							}">
								{order.status === 'PREPARING' ? 'SEDANG DIMASAK' :
								 order.status === 'CONFIRMED' ? 'BELUM DIMASAK' :
								 order.status === 'WAITING_PAYMENT' ? 'BELUM DIMASAK' :
								 order.status === 'READY' ? 'SUDAH MATANG & SIAP' : 'SELESAI'}
							</span>
						</div>

						<h3 class="font-extrabold text-sm text-slate-900 mt-0.5 leading-snug font-['Outfit']">
							{#if order.status === 'PREPARING'}
								🍳 Sedang Dimasak oleh Chef
							{:else if order.status === 'CONFIRMED'}
								📋 Menunggu Giliran Masak (Antrean Dapur)
							{:else if order.status === 'WAITING_PAYMENT'}
								💳 Menunggu Pembayaran Selesai
							{:else if order.status === 'READY'}
								🍽️ Hidangan Siap Diantar ke Meja
							{:else}
								Pesanan Dibatalkan
							{/if}
						</h3>

						<p class="text-xs text-slate-600 mt-0.5 leading-relaxed">
							{#if order.status === 'PREPARING'}
								Chef sedang menyiapkan hidangan segar untuk meja Anda. Mohon tunggu ya!
							{:else if order.status === 'CONFIRMED'}
								Pesanan sudah diterima dapur dan sedang menunggu giliran untuk dimasak.
							{:else if order.status === 'WAITING_PAYMENT'}
								Pesanan akan langsung dimasak segera setelah pembayaran Anda diselesaikan.
							{:else if order.status === 'READY'}
								Makanan dan minuman sudah matang dan sedang diantar ke meja Anda.
							{:else}
								Pesanan ini telah dibatalkan.
							{/if}
						</p>

						{#if activeOrdersList.length > 1}
							<div class="mt-2.5 pt-2 border-t border-slate-200/60 flex items-center justify-between text-[11px] text-slate-500 font-medium">
								<span>Melacak status <strong>#{order.order_number}</strong></span>
								<span class="text-orange-600 font-bold">Tersedia {activeOrdersList.length} pesanan di atas &uarr;</span>
							</div>
						{/if}
					</div>
				</div>
			</div>
			{/if}

			<!-- Action Card: Tetap Masih Bisa Memesan Menu Baru Lagi di Sini -->
			<div class="bg-linear-to-r from-orange-50 via-amber-50 to-orange-50/50 rounded-3xl p-4 border border-orange-200 shadow-xs flex items-center justify-between gap-3">
				<div class="flex items-center gap-3 min-w-0">
					<div class="w-10 h-10 rounded-2xl bg-orange-600 text-white flex items-center justify-center shrink-0 shadow-md shadow-orange-500/20">
						<UtensilsCrossed class="w-5 h-5" />
					</div>
					<div class="min-w-0">
						<h4 class="font-bold text-xs text-slate-900">Ingin Tambah Pesanan Lain?</h4>
						<p class="text-[11px] text-slate-500 truncate mt-0.5">
							Pesan menu makanan atau minuman baru untuk meja ini
						</p>
					</div>
				</div>

				<a
					href={orderMenuUrl}
					class="shrink-0 bg-orange-600 hover:bg-orange-700 active:scale-95 text-white font-extrabold text-xs px-3.5 py-2.5 rounded-xl shadow-md shadow-orange-500/25 flex items-center gap-1.5 transition-all"
				>
					<Plus class="w-4 h-4" />
					<span>Pesan Lagi</span>
				</a>
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
								<div class="w-52 h-52 bg-white border border-slate-300 rounded-xl p-2 mx-auto flex flex-col items-center justify-center overflow-hidden shadow-xs">
									{#if tableInfo?.restaurant?.qris_image_url}
										<img
											src={tableInfo.restaurant.qris_image_url}
											alt="QRIS Toko"
											class="w-full h-full object-contain rounded-lg"
										/>
									{:else}
										<QrCode class="w-36 h-36 text-slate-800" />
										<span class="text-[10px] font-mono text-slate-400 font-bold mt-1">QRIS NASIONAL</span>
									{/if}
								</div>
							</div>

							<p class="text-xs text-slate-500 max-w-xs mx-auto">
								Buka aplikasi e-wallet (GoPay, OVO, ShopeePay, DANA) atau mobile banking untuk memindai QRIS di atas.
							</p>

							<!-- Placeholder Upload Bukti Pembayaran Foto (Hanya untuk QRIS / Cashless) -->
							<div class="mt-4 pt-4 border-t border-slate-200/80 text-left">
								<div class="flex items-center justify-between mb-2.5">
									<div>
										<h4 class="text-xs font-black text-slate-900 tracking-tight flex items-center gap-1.5">
											<Camera class="w-4 h-4 text-orange-600" />
											<span>Upload Bukti Pembayaran (Foto)</span>
										</h4>
										<p class="text-[11px] text-slate-500">
											Unggah bukti transfer atau tangkapan layar pembayaran QRIS
										</p>
									</div>
									{#if proofImage}
										<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-800 border border-amber-200">
											<span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></span>
											Menunggu Verifikasi Admin
										</span>
									{/if}
								</div>

								{#if !proofImage}
									<!-- Dropzone Upload Placeholder -->
									<label class="group relative flex flex-col items-center justify-center p-5 sm:p-6 border-2 border-dashed border-slate-300 hover:border-orange-500 bg-slate-50/80 hover:bg-orange-50/20 rounded-2xl cursor-pointer transition-all duration-200 text-center">
										<input 
											type="file" 
											accept="image/*" 
											onchange={handleFileUpload} 
											class="sr-only" 
										/>
										<div class="w-12 h-12 rounded-2xl bg-white shadow-xs border border-slate-200 flex items-center justify-center text-slate-500 group-hover:text-orange-600 group-hover:scale-105 group-hover:border-orange-200 transition-all mb-2.5">
											{#if isUploadingProof}
												<RefreshCw class="w-5 h-5 animate-spin text-orange-600" />
											{:else}
												<Upload class="w-5 h-5" />
											{/if}
										</div>
										<span class="text-xs font-bold text-slate-800 group-hover:text-orange-600 transition-colors">
											{isUploadingProof ? 'Memproses Foto...' : 'Pilih atau Seret Foto Bukti Pembayaran'}
										</span>
										<span class="text-[10px] sm:text-[11px] text-slate-400 mt-1 max-w-xs">
											Mendukung semua format foto (JPG, PNG, WEBP, HEIC, GIF, dll)
										</span>
									</label>
								{:else}
									<!-- Uploaded Proof Card Preview -->
									<div class="bg-white border border-slate-200 rounded-2xl p-3 shadow-xs space-y-3">
										<div class="flex items-center gap-3">
											<button 
												type="button" 
												onclick={() => (showProofModal = true)}
												class="relative w-16 h-16 rounded-xl overflow-hidden bg-slate-100 border border-slate-200 group shrink-0"
												title="Klik untuk memperbesar foto"
											>
												<img 
													src={proofImage} 
													alt="Bukti Pembayaran" 
													class="w-full h-full object-cover transition-transform group-hover:scale-110" 
												/>
												<div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white">
													<Eye class="w-4 h-4" />
												</div>
											</button>

											<div class="min-w-0 flex-1">
												<div class="flex items-center gap-1.5 flex-wrap">
													<p class="text-xs font-bold text-slate-900 truncate" title={proofFileName}>
														{proofFileName || 'bukti-pembayaran.jpg'}
													</p>
													{#if isUploadingProof}
														<span class="inline-flex items-center gap-1 text-[10px] text-orange-700 bg-orange-50 px-2 py-0.5 rounded-full font-bold border border-orange-200">
															<RefreshCw class="w-2.5 h-2.5 animate-spin text-orange-600" />
															Memproses foto bukti...
														</span>
													{:else if proofImage}
														<span class="inline-flex items-center gap-1 text-[10px] text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full font-bold border border-emerald-200">
															<CheckCircle2 class="w-2.5 h-2.5 text-emerald-600" />
															Foto Siap Dikirim
														</span>
													{/if}
												</div>
												<p class="text-[11px] text-slate-400 font-medium">
													{proofFileSize} &bull; Diunggah {proofUploadedAt || 'Baru saja'}
												</p>
												<div class="mt-1.5 flex items-center gap-2 flex-wrap">
													<button
														type="button"
														onclick={() => (showProofModal = true)}
														class="text-[11px] font-bold text-orange-600 hover:text-orange-700 flex items-center gap-1 cursor-pointer"
													>
														<Eye class="w-3.5 h-3.5" />
														Lihat Foto
													</button>
													<span class="text-slate-200">&bull;</span>
													<label class="text-[11px] font-bold text-slate-600 hover:text-slate-900 cursor-pointer">
														<span>Ganti Foto</span>
														<input 
															type="file" 
															accept="image/*" 
															onchange={handleFileUpload} 
															class="sr-only" 
														/>
													</label>
													<span class="text-slate-200">&bull;</span>
													<button
														type="button"
														onclick={removeProof}
														class="text-[11px] font-bold text-rose-600 hover:text-rose-700 flex items-center gap-0.5 cursor-pointer"
													>
														<Trash2 class="w-3 h-3" />
														Hapus
													</button>
												</div>
											</div>
										</div>

										<!-- Tombol Aksi Kirim Bukti Pembayaran -->
										<div class="pt-2 border-t border-slate-100 space-y-2">
											{#if proofSubmitted || order?.proof_url}
												<div class="flex items-center justify-between gap-2 p-2.5 bg-emerald-50 border border-emerald-200 rounded-xl text-emerald-800">
													<div class="flex items-center gap-2 min-w-0">
														<div class="w-6 h-6 rounded-full bg-emerald-100 flex items-center justify-center shrink-0">
															<CheckCircle2 class="w-4 h-4 text-emerald-600" />
														</div>
														<div class="text-left min-w-0">
															<p class="text-xs font-bold leading-tight truncate">Bukti Terkirim ke Kasir</p>
															<p class="text-[10px] text-emerald-600 font-medium">Kasir sedang memverifikasi</p>
														</div>
													</div>
													<button
														type="button"
														onclick={submitProofToCashier}
														disabled={isSubmittingProof || isUploadingProof}
														class="px-2.5 py-1 text-[11px] font-bold text-emerald-700 hover:text-emerald-900 bg-white border border-emerald-300 rounded-lg shadow-2xs hover:bg-emerald-50 transition-colors shrink-0 disabled:opacity-50 cursor-pointer"
													>
														{isSubmittingProof ? 'Mengirim...' : 'Kirim Ulang'}
													</button>
												</div>
											{:else}
												<button
													type="button"
													onclick={submitProofToCashier}
													disabled={isSubmittingProof || isUploadingProof}
													class="w-full py-2.5 px-4 bg-orange-600 hover:bg-orange-700 active:scale-98 disabled:opacity-50 text-white rounded-xl font-bold text-xs shadow-md shadow-orange-600/25 transition-all flex items-center justify-center gap-2 cursor-pointer"
												>
													{#if isSubmittingProof}
														<RefreshCw class="w-4 h-4 animate-spin" />
														<span>Mengirim Bukti ke Kasir...</span>
													{:else}
														<Send class="w-4 h-4" />
														<span>Kirim Bukti Pembayaran</span>
													{/if}
												</button>
											{/if}
										</div>

										<div class="bg-amber-50/80 border border-amber-200/70 rounded-xl p-2.5 flex items-start gap-2">
											<AlertCircle class="w-3.5 h-3.5 text-amber-600 shrink-0 mt-0.5" />
											<p class="text-[11px] text-amber-800 leading-tight">
												Setelah status dikonfirmasi oleh admin/kasir, pembayaran otomatis <strong>selesai</strong> dan pesanan langsung masuk ke antrian dapur.
											</p>
										</div>
									</div>
								{/if}
							</div>
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

					{#if proofImage}
						<div class="pt-3 border-t border-slate-100 flex items-center justify-between">
							<div class="flex items-center gap-2.5 min-w-0">
								<img src={proofImage} alt="Bukti Foto" class="w-9 h-9 rounded-lg object-cover border border-slate-200 shrink-0" />
								<div class="text-[11px] text-left truncate">
									<span class="font-bold text-slate-800 block truncate">Bukti Pembayaran Terlampir</span>
									<span class="text-slate-400 text-[10px] block truncate">{proofFileName || 'foto-bukti'} &bull; {proofFileSize}</span>
								</div>
							</div>
							<button
								type="button"
								onclick={() => (showProofModal = true)}
								class="text-xs font-bold text-emerald-700 hover:text-emerald-800 shrink-0 ml-2"
							>
								Lihat Foto
							</button>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Lacak Status Pesanan (Progress Stepper) - Hilang bila pesanan sudah selesai diantar semua -->
			{#if order.status !== 'COMPLETED'}
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
			{:else}
				<!-- Status Pesanan Selesai Diantar -->
				<div class="bg-linear-to-br from-emerald-500/10 via-teal-500/10 to-emerald-500/5 border border-emerald-500/30 rounded-3xl p-6 text-center shadow-xs space-y-3">
					<div class="w-14 h-14 mx-auto rounded-2xl bg-emerald-600 text-white flex items-center justify-center shadow-md shadow-emerald-600/20">
						<Check class="w-7 h-7 stroke-3" />
					</div>
					<div>
						<h2 class="font-extrabold text-base text-slate-900 font-['Outfit']">Semua Hidangan Selesai Diantar</h2>
						<p class="text-xs text-slate-600 mt-1 max-w-xs mx-auto leading-relaxed">
							Semua pesanan makanan dan minuman telah disajikan di meja Anda. Selamat menikmati santapan Anda!
						</p>
					</div>
					<div class="pt-2 flex flex-col sm:flex-row items-center justify-center gap-2">
						<button
							type="button"
							onclick={() => (showInvoiceModal = true)}
							class="w-full sm:w-auto px-4 py-2.5 rounded-xl bg-white border border-slate-200 text-slate-700 font-bold text-xs hover:bg-slate-50 transition-all flex items-center justify-center gap-2 shadow-2xs"
						>
							<Receipt class="w-4 h-4 text-slate-500" />
							<span>Lihat Struk / Invoice</span>
						</button>
						<a
							href={orderMenuUrl}
							class="w-full sm:w-auto px-4 py-2.5 rounded-xl bg-orange-600 hover:bg-orange-500 active:scale-95 text-white font-bold text-xs transition-all flex items-center justify-center gap-2 shadow-sm"
						>
							<Plus class="w-4 h-4" />
							<span>Pesan Menu Baru Lagi</span>
						</a>
					</div>
				</div>
			{/if}

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

				<div class="pt-2">
					<a
						href={orderMenuUrl}
						class="w-full bg-white hover:bg-orange-50 active:scale-[0.98] border-2 border-dashed border-orange-300 hover:border-orange-500 text-orange-600 font-bold py-3.5 px-4 rounded-2xl text-xs flex items-center justify-center gap-2 shadow-xs transition-all"
					>
						<Plus class="w-4 h-4 text-orange-600" />
						<span>Pesan Menu Baru Lagi untuk Meja Ini</span>
					</a>
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
							<span class="font-mono text-[11px] text-slate-600 truncate max-w-35 inline-block">{payment?.provider_transaction_id || payment?.id || '-'}</span>
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

				<!-- Footer note -->
				<div class="pt-3 border-t border-dashed border-slate-300 text-center space-y-2">
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

<!-- Modal Preview Foto Bukti Pembayaran -->
{#if showProofModal && proofImage}
	<div class="fixed inset-0 z-50 bg-black/80 backdrop-blur-xs flex items-center justify-center p-4">
		<div class="bg-white rounded-3xl max-w-lg w-full overflow-hidden shadow-2xl animate-in zoom-in-95 duration-150 flex flex-col max-h-[90vh]">
			<div class="px-5 py-4 border-b border-slate-100 flex items-center justify-between">
				<div class="flex items-center gap-2">
					<div class="w-8 h-8 rounded-xl bg-orange-50 border border-orange-200 flex items-center justify-center text-orange-600">
						<ImageIcon class="w-4 h-4" />
					</div>
					<div>
						<h3 class="text-xs font-black text-slate-900 truncate max-w-xs">{proofFileName || 'Bukti Pembayaran'}</h3>
						<p class="text-[10px] text-slate-400 font-medium">{proofFileSize} &bull; {proofUploadedAt || 'Baru saja'}</p>
					</div>
				</div>
				<button
					type="button"
					onclick={() => (showProofModal = false)}
					class="w-8 h-8 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 flex items-center justify-center"
				>
					<X class="w-4 h-4" />
				</button>
			</div>

			<div class="p-4 bg-slate-950 flex items-center justify-center flex-1 overflow-auto">
				<img 
					src={proofImage} 
					alt="Bukti Pembayaran Penuh" 
					class="max-w-full max-h-[65vh] object-contain rounded-lg shadow-md"
				/>
			</div>

			<div class="p-4 bg-slate-50 border-t border-slate-100 flex items-center justify-between">
				<span class="text-[11px] text-slate-500 font-medium">
					{isPaid ? 'Status: Pembayaran Sudah Lunas' : 'Status: Menunggu Konfirmasi Admin / Kasir'}
				</span>
				<button
					type="button"
					onclick={() => (showProofModal = false)}
					class="px-4 py-2 bg-slate-900 hover:bg-slate-800 text-white font-bold text-xs rounded-xl transition-colors"
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

