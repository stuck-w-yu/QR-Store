<script lang="ts">
	import { cart } from '$lib/stores/cart.svelte';
	import { formatRupiah, api } from '$lib/api/client';
	import { goto } from '$app/navigation';
	import { ArrowLeft, Trash2, Plus, Minus, ShoppingBag, ShieldCheck } from '@lucide/svelte';

	let generalNotes = $state('');
	let submitting = $state(false);
	let errorMessage = $state<string | null>(null);

	async function handleCheckout() {
		if (cart.items.length === 0) return;
		if (!cart.qrToken) {
			errorMessage = 'Sesi meja tidak valid. Silakan scan ulang QR Code.';
			return;
		}

		try {
			submitting = true;
			errorMessage = null;

			// 1. Prepare items payload for backend pricing engine
			const payload = {
				qr_token: cart.qrToken,
				notes: generalNotes || undefined,
				items: cart.items.map((it) => ({
					menu_id: it.menu.id,
					quantity: it.quantity,
					modifier_option_ids: it.selectedOptions.map((o) => o.id),
					notes: it.notes || undefined
				}))
			};

			// 2. Submit order to backend (Rule 2 & 54: Backend calculates final prices)
			const orderResult = await api.post<{ id: string; order_number: string; total: number }>('/public/orders', payload);

			// 3. Clear cart and navigate to live tracking page
			cart.clear();
			goto(`/order/status/${orderResult.id}`);
		} catch (err: any) {
			errorMessage = err?.message || 'Gagal memproses pesanan. Silakan coba lagi.';
		} finally {
			submitting = false;
		}
	}
</script>

<div class="min-h-screen bg-slate-50 pb-28">
	<!-- Header -->
	<header class="sticky top-0 z-30 bg-white/90 backdrop-blur-md border-b border-slate-200/80 px-4 py-3.5">
		<div class="max-w-lg mx-auto flex items-center justify-between">
			<a
				href="/order?token={cart.qrToken}"
				class="w-9 h-9 rounded-xl bg-slate-100 text-slate-700 flex items-center justify-center hover:bg-slate-200 transition-colors"
			>
				<ArrowLeft class="w-4 h-4" />
			</a>
			<h1 class="font-extrabold text-base text-slate-800 font-['Outfit']">Keranjang Pesanan</h1>
			<button
				type="button"
				onclick={() => cart.clear()}
				class="text-xs font-semibold text-red-600 hover:text-red-700 transition-colors"
			>
				Hapus Semua
			</button>
		</div>
	</header>

	<main class="max-w-lg mx-auto p-4 space-y-4">
		{#if errorMessage}
			<div class="p-3.5 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs font-medium">
				{errorMessage}
			</div>
		{/if}

		{#if cart.items.length === 0}
			<div class="text-center py-20">
				<div class="w-16 h-16 bg-orange-100 text-orange-600 rounded-full flex items-center justify-center mx-auto mb-3">
					<ShoppingBag class="w-8 h-8" />
				</div>
				<h2 class="font-bold text-slate-800 text-base mb-1">Keranjang Masih Kosong</h2>
				<p class="text-xs text-slate-500 mb-6">Pilih menu favorit Anda dan tambahkan ke keranjang.</p>
				<a
					href="/order?token={cart.qrToken}"
					class="inline-block bg-orange-600 text-white font-semibold text-xs px-5 py-2.5 rounded-xl shadow-md"
				>
					Kembali ke Menu
				</a>
			</div>
		{:else}
			<!-- Items List -->
			<div class="space-y-3">
				{#each cart.items as item}
					<div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs flex flex-col gap-3">
						<div class="flex items-start justify-between gap-2">
							<div class="flex-1">
								<h3 class="font-bold text-sm text-slate-900 leading-snug">{item.menu.name}</h3>
								
								<!-- Selected Modifiers -->
								{#if item.selectedOptions && item.selectedOptions.length > 0}
									<div class="flex flex-wrap gap-1 mt-1.5">
										{#each item.selectedOptions as opt}
											<span class="text-[11px] bg-orange-50 text-orange-700 font-medium px-2 py-0.5 rounded-md border border-orange-200/60">
												+{opt.name} {opt.additional_price > 0 ? `(${formatRupiah(opt.additional_price)})` : ''}
											</span>
										{/each}
									</div>
								{/if}

								<!-- Note -->
								{#if item.notes}
									<p class="text-[11px] text-slate-500 italic mt-1.5">"{item.notes}"</p>
								{/if}
							</div>

							<button
								type="button"
								onclick={() => cart.removeItem(item.key)}
								class="text-slate-400 hover:text-red-600 p-1 transition-colors"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</div>

						<div class="flex items-center justify-between pt-2 border-t border-slate-100">
							<span class="font-extrabold text-sm text-orange-600 font-['Outfit']">
								{formatRupiah(item.subtotal)}
							</span>

							<!-- Quantity Control -->
							<div class="flex items-center gap-2 bg-slate-50 px-2 py-1 rounded-xl border border-slate-200/80">
								<button
									type="button"
									onclick={() => cart.updateQuantity(item.key, -1)}
									class="w-6 h-6 rounded-lg bg-white text-slate-700 flex items-center justify-center hover:bg-slate-200 active:scale-95 shadow-2xs"
								>
									<Minus class="w-3 h-3" />
								</button>
								<span class="w-5 text-center font-bold text-xs">{item.quantity}</span>
								<button
									type="button"
									onclick={() => cart.updateQuantity(item.key, 1)}
									class="w-6 h-6 rounded-lg bg-orange-500 text-white flex items-center justify-center hover:bg-orange-600 active:scale-95 shadow-2xs"
								>
									<Plus class="w-3 h-3" />
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- General Order Notes -->
			<div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs">
				<label for="general-notes" class="block text-xs font-bold text-slate-800 mb-1.5">Catatan Pesanan Keseluruhan</label>
				<textarea
					id="general-notes"
					bind:value={generalNotes}
					rows="2"
					placeholder="Tulis instruksi tambahan untuk meja ini..."
					class="w-full text-xs p-3 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 resize-none"
				></textarea>
			</div>

			<!-- Payment Summary Breakdown -->
			<div class="bg-white rounded-2xl p-4 border border-slate-200/80 shadow-xs space-y-2.5">
				<h3 class="font-bold text-xs uppercase tracking-wider text-slate-400">Ringkasan Pembayaran</h3>

				<div class="flex justify-between text-xs text-slate-600">
					<span>Subtotal ({cart.count} item)</span>
					<span class="font-semibold text-slate-900 font-['Outfit']">{formatRupiah(cart.subtotal)}</span>
				</div>

				<div class="flex justify-between text-xs text-slate-600">
					<span>Pajak Restoran ({cart.tableInfo?.restaurant?.tax_percent ?? 10}%)</span>
					<span class="font-semibold text-slate-900 font-['Outfit']">{formatRupiah(cart.tax)}</span>
				</div>

				{#if cart.serviceCharge > 0}
					<div class="flex justify-between text-xs text-slate-600">
						<span>Biaya Layanan ({cart.tableInfo?.restaurant?.service_percent}%)</span>
						<span class="font-semibold text-slate-900 font-['Outfit']">{formatRupiah(cart.serviceCharge)}</span>
					</div>
				{/if}

				<div class="pt-2 border-t border-slate-100 flex justify-between items-baseline">
					<span class="font-bold text-sm text-slate-900">Total Tagihan</span>
					<span class="font-extrabold text-lg text-orange-600 font-['Outfit']">{formatRupiah(cart.total)}</span>
				</div>
			</div>

			<div class="flex items-center gap-2 text-[11px] text-slate-500 justify-center pt-1">
				<ShieldCheck class="w-4 h-4 text-emerald-600" />
				<span>Harga final divalidasi langsung oleh sistem restoran</span>
			</div>
		{/if}
	</main>

	<!-- Fixed Bottom Checkout Button -->
	{#if cart.items.length > 0}
		<div class="fixed bottom-0 inset-x-0 bg-white/90 backdrop-blur-md border-t border-slate-200/80 p-4 z-40">
			<div class="max-w-lg mx-auto flex items-center gap-4">
				<div>
					<div class="text-[10px] text-slate-500 uppercase tracking-wider font-bold">Total Pembayaran</div>
					<div class="text-base font-extrabold text-slate-900 font-['Outfit']">{formatRupiah(cart.total)}</div>
				</div>

				<button
					type="button"
					onclick={handleCheckout}
					disabled={submitting}
					class="flex-1 bg-orange-600 hover:bg-orange-700 text-white font-bold py-3.5 px-4 rounded-xl shadow-lg shadow-orange-500/30 text-center text-sm disabled:opacity-50 transition-all active:scale-[0.98]"
				>
					{#if submitting}
						Memproses Pesanan...
					{:else}
						Konfirmasi & Bayar
					{/if}
				</button>
			</div>
		</div>
	{/if}
</div>
