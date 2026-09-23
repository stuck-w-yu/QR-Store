<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api, formatRupiah } from '$lib/api/client';
	import { cart } from '$lib/stores/cart.svelte';
	import type { PublicTableInfo, CategoryWithMenus, Menu, ModifierOption } from '$lib/types';
	import { 
		ShoppingBag, UtensilsCrossed, Plus, Minus, X, Check, 
		Clock, AlertCircle, Sparkles, ChevronRight, Store 
	} from '@lucide/svelte';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let catalog = $state<CategoryWithMenus[]>([]);
	let activeCategory = $state<string>('');
	let tableInfo = $state<PublicTableInfo | null>(null);

	// Modal State
	let selectedMenu = $state<Menu | null>(null);
	let selectedOptions = $state<ModifierOption[]>([]);
	let itemQuantity = $state(1);
	let itemNotes = $state('');

	onMount(async () => {
		const token = page.url.searchParams.get('token') || 'demo-qr-token-table-01';

		try {
			loading = true;
			// 1. Fetch table and restaurant
			const info = await api.get<PublicTableInfo>(`/public/tables/${token}`);
			tableInfo = info;
			cart.setSession(token, info);

			// 2. Fetch digital menu catalog
			const catData = await api.get<CategoryWithMenus[]>(`/public/restaurants/${info.restaurant.id}/menu`);
			catalog = catData;
			if (catData.length > 0) {
				activeCategory = catData[0].id;
			}
		} catch (err: any) {
			error = err?.message || 'Gagal memuat informasi meja. Pastikan QR code valid.';
		} finally {
			loading = false;
		}
	});

	function openItemModal(menuItem: Menu) {
		selectedMenu = menuItem;
		itemQuantity = 1;
		itemNotes = '';
		selectedOptions = [];

		// Auto-select first option for required single modifiers
		if (menuItem.modifiers) {
			for (const mod of menuItem.modifiers) {
				if (mod.required && mod.type === 'SINGLE' && mod.options.length > 0) {
					selectedOptions.push(mod.options[0]);
				}
			}
		}
	}

	function closeModal() {
		selectedMenu = null;
	}

	function toggleOption(modType: string, option: ModifierOption) {
		if (modType === 'SINGLE') {
			// Replace existing option from same modifier
			selectedOptions = selectedOptions.filter((o) => o.modifier_id !== option.modifier_id);
			selectedOptions.push(option);
		} else {
			// Multiple toggle
			const exists = selectedOptions.some((o) => o.id === option.id);
			if (exists) {
				selectedOptions = selectedOptions.filter((o) => o.id !== option.id);
			} else {
				selectedOptions.push(option);
			}
		}
	}

	function isOptionSelected(optionId: string): boolean {
		return selectedOptions.some((o) => o.id === optionId);
	}

	function getModalItemTotal(): number {
		if (!selectedMenu) return 0;
		let unit = selectedMenu.price;
		for (const opt of selectedOptions) {
			unit += opt.additional_price;
		}
		return unit * itemQuantity;
	}

	function handleAddToCart() {
		if (!selectedMenu) return;
		cart.addItem(selectedMenu, itemQuantity, selectedOptions, itemNotes);
		closeModal();
	}
</script>

<div class="min-h-screen bg-slate-50 pb-28">
	{#if loading}
		<div class="flex flex-col items-center justify-center min-h-[60vh] gap-3 text-slate-500">
			<div class="w-10 h-10 border-4 border-orange-500 border-t-transparent rounded-full animate-spin"></div>
			<p class="font-medium text-sm">Menyiapkan menu restoran...</p>
		</div>
	{:else if error}
		<div class="max-w-md mx-auto p-6 mt-12 text-center">
			<div class="w-16 h-16 bg-red-100 text-red-600 rounded-full flex items-center justify-center mx-auto mb-4">
				<AlertCircle class="w-8 h-8" />
			</div>
			<h2 class="text-xl font-bold text-slate-800 mb-2">QR Code Tidak Valid</h2>
			<p class="text-slate-600 text-sm mb-6">{error}</p>
			<a
				href="/order?token=demo-qr-token-table-01"
				class="inline-block bg-orange-600 text-white font-semibold px-6 py-2.5 rounded-xl shadow-lg shadow-orange-500/30 text-sm"
			>
				Gunakan Demo Meja 01
			</a>
		</div>
	{:else if tableInfo}
		<!-- Restaurant Header Banner -->
		<header class="bg-linear-to-br from-slate-900 via-slate-800 to-orange-950 text-white pt-8 pb-14 px-4 relative overflow-hidden">
			<div class="absolute -right-10 -bottom-10 w-48 h-48 bg-orange-500/20 rounded-full blur-3xl pointer-events-none"></div>
			
			<div class="max-w-lg mx-auto flex items-center justify-between relative z-10">
				<div class="flex items-center gap-3.5">
					{#if tableInfo.restaurant.logo_url}
						<img
							src={tableInfo.restaurant.logo_url}
							alt={tableInfo.restaurant.name}
							class="w-13 h-13 rounded-2xl object-cover ring-2 ring-orange-500/50 shadow-md"
						/>
					{:else}
						<div class="w-13 h-13 rounded-2xl bg-orange-600 flex items-center justify-center shadow-md">
							<Store class="w-7 h-7 text-white" />
						</div>
					{/if}
					<div>
						<h1 class="text-lg font-extrabold tracking-tight font-['Outfit']">{tableInfo.restaurant.name}</h1>
						<p class="text-xs text-orange-200/80 font-medium">Self-Order & Cashless QR</p>
					</div>
				</div>

				<!-- Table Badge -->
				<div class="bg-white/10 backdrop-blur-md border border-white/20 px-3.5 py-1.5 rounded-full text-right shadow-sm">
					<span class="text-[10px] text-orange-300 uppercase tracking-wider block font-semibold">Nomor Meja</span>
					<span class="text-sm font-bold text-white tracking-wide">{tableInfo.table.name}</span>
				</div>
			</div>
		</header>

		<!-- Main Content Container -->
		<main class="max-w-lg mx-auto px-4 -mt-7 relative z-20">
			<!-- Sticky Category Navigation Bar -->
			<div class="sticky top-2 z-30 bg-white/95 backdrop-blur-md rounded-2xl p-1.5 shadow-md border border-slate-200/60 mb-6 overflow-x-auto scrollbar-none flex gap-1.5">
				{#each catalog as cat}
					<button
						onclick={() => (activeCategory = cat.id)}
						class="px-4 py-2 rounded-xl text-xs font-bold whitespace-nowrap transition-all duration-200 {activeCategory === cat.id
							? 'bg-orange-600 text-white shadow-md shadow-orange-500/30'
							: 'text-slate-600 hover:bg-slate-100'}"
					>
						{cat.name}
					</button>
				{/each}
			</div>

			<!-- Menu Items Section -->
			<div class="space-y-8">
				{#each catalog as cat}
					{#if activeCategory === '' || activeCategory === cat.id}
						<section class="space-y-3">
							<div class="flex items-center justify-between">
								<h2 class="text-base font-extrabold text-slate-800 font-['Outfit'] flex items-center gap-2">
									<span>{cat.name}</span>
									<span class="text-xs font-semibold px-2 py-0.5 bg-slate-200 text-slate-600 rounded-full">{cat.menus.length}</span>
								</h2>
							</div>

							<div class="grid grid-cols-1 gap-3.5">
								{#each cat.menus as menuItem}
									<div
										role="button"
										tabindex="0"
										onclick={() => openItemModal(menuItem)}
										onkeydown={(e) => e.key === 'Enter' && openItemModal(menuItem)}
										class="bg-white rounded-2xl p-3.5 border border-slate-200/80 shadow-sm hover:shadow-md transition-all flex gap-3.5 cursor-pointer text-left group"
									>
										<!-- Item Image -->
										{#if menuItem.image_url}
											<div class="relative w-24 h-24 rounded-xl overflow-hidden shrink-0 bg-slate-100">
												<img
													src={menuItem.image_url}
													alt={menuItem.name}
													class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
													loading="lazy"
												/>
											</div>
										{/if}

										<!-- Details -->
										<div class="flex-1 flex flex-col justify-between">
											<div>
												<h3 class="font-bold text-slate-900 text-sm leading-snug group-hover:text-orange-600 transition-colors">
													{menuItem.name}
												</h3>
												{#if menuItem.description}
													<p class="text-xs text-slate-500 line-clamp-2 mt-1 leading-relaxed">
														{menuItem.description}
													</p>
												{/if}
											</div>

											<div class="flex items-center justify-between mt-2 pt-2 border-t border-slate-100">
												<span class="text-sm font-extrabold text-orange-600 font-['Outfit']">
													{formatRupiah(menuItem.price)}
												</span>
												<button
													type="button"
													class="bg-orange-50 text-orange-600 hover:bg-orange-600 hover:text-white font-semibold text-xs px-3 py-1.5 rounded-lg flex items-center gap-1 transition-colors"
												>
													<Plus class="w-3.5 h-3.5" />
													<span>Tambah</span>
												</button>
											</div>
										</div>
									</div>
								{/each}
							</div>
						</section>
					{/if}
				{/each}
			</div>
		</main>

		<!-- Floating Bottom Cart Bar -->
		{#if cart.count > 0}
			<div class="fixed bottom-4 inset-x-0 z-40 px-4 max-w-lg mx-auto">
				<a
					href="/order/cart"
					class="bg-slate-900 text-white p-3.5 rounded-2xl shadow-xl shadow-slate-900/40 flex items-center justify-between border border-slate-800 hover:bg-slate-800 transition-all active:scale-[0.98]"
				>
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-orange-600 flex items-center justify-center font-bold text-white text-sm shadow-md">
							{cart.count}
						</div>
						<div>
							<div class="text-[11px] text-slate-400 font-medium">Total Pesanan</div>
							<div class="text-sm font-extrabold font-['Outfit'] text-white">
								{formatRupiah(cart.total)}
							</div>
						</div>
					</div>

					<div class="flex items-center gap-1 font-bold text-xs bg-orange-500 text-white px-4 py-2.5 rounded-xl shadow-md">
						<span>Lihat Keranjang</span>
						<ChevronRight class="w-4 h-4" />
					</div>
				</a>
			</div>
		{/if}

		<!-- Item Selection & Modifier Modal -->
		{#if selectedMenu}
			<div class="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-end sm:items-center justify-center p-0 sm:p-4">
				<div class="bg-white w-full max-w-md rounded-t-3xl sm:rounded-3xl max-h-[85vh] flex flex-col overflow-hidden animate-in slide-in-from-bottom duration-200">
					<!-- Modal Header -->
					<div class="relative">
						{#if selectedMenu.image_url}
							<img
								src={selectedMenu.image_url}
								alt={selectedMenu.name}
								class="w-full h-44 object-cover"
							/>
						{/if}
						<button
							type="button"
							onclick={closeModal}
							class="absolute top-3 right-3 w-8 h-8 rounded-full bg-black/50 text-white flex items-center justify-center hover:bg-black/70 backdrop-blur-md"
						>
							<X class="w-4 h-4" />
						</button>
					</div>

					<div class="p-5 overflow-y-auto flex-1 space-y-5">
						<div>
							<h3 class="text-lg font-bold text-slate-900">{selectedMenu.name}</h3>
							<p class="text-sm text-slate-500 mt-1">{selectedMenu.description || ''}</p>
							<div class="text-base font-extrabold text-orange-600 mt-2 font-['Outfit']">
								{formatRupiah(selectedMenu.price)}
							</div>
						</div>

						<!-- Modifiers -->
						{#if selectedMenu.modifiers && selectedMenu.modifiers.length > 0}
							{#each selectedMenu.modifiers as mod}
								<div class="border-t border-slate-100 pt-4">
									<div class="flex items-center justify-between mb-2.5">
										<span class="text-sm font-bold text-slate-800">{mod.name}</span>
										<span class="text-[11px] font-semibold px-2 py-0.5 rounded-full {mod.required ? 'bg-orange-100 text-orange-700' : 'bg-slate-100 text-slate-600'}">
											{mod.required ? 'Wajib' : 'Opsional'}
										</span>
									</div>

									<div class="space-y-2">
										{#each mod.options as opt}
											<button
												type="button"
												onclick={() => toggleOption(mod.type, opt)}
												class="w-full flex items-center justify-between p-3 rounded-xl border text-left text-sm transition-all {isOptionSelected(opt.id)
													? 'border-orange-500 bg-orange-50/50 text-orange-950 font-medium'
													: 'border-slate-200 hover:border-slate-300 text-slate-700'}"
											>
												<div class="flex items-center gap-2.5">
													<div class="w-4 h-4 rounded-{mod.type === 'SINGLE' ? 'full' : 'md'} border flex items-center justify-center {isOptionSelected(opt.id) ? 'border-orange-600 bg-orange-600 text-white' : 'border-slate-300'}">
														{#if isOptionSelected(opt.id)}
															<Check class="w-3 h-3 stroke-[3]" />
														{/if}
													</div>
													<span>{opt.name}</span>
												</div>
												{#if opt.additional_price > 0}
													<span class="text-xs font-semibold text-slate-500">
														+{formatRupiah(opt.additional_price)}
													</span>
												{/if}
											</button>
										{/each}
									</div>
								</div>
							{/each}
						{/if}

						<!-- Special Notes -->
						<div class="border-t border-slate-100 pt-4">
							<label for="order-notes" class="text-sm font-bold text-slate-800 block mb-1.5">Catatan Khusus (Opsional)</label>
							<input
								id="order-notes"
								type="text"
								bind:value={itemNotes}
								placeholder="Contoh: jangan pakai bawang merah, sedikit es..."
								class="w-full text-xs px-3.5 py-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
							/>
						</div>
					</div>

					<!-- Modal Footer with Quantity and Add Button -->
					<div class="p-4 bg-slate-50 border-t border-slate-200/80 flex items-center gap-3">
						<div class="flex items-center gap-2 bg-white px-2 py-1.5 rounded-xl border border-slate-200 shadow-xs">
							<button
								type="button"
								onclick={() => itemQuantity > 1 && itemQuantity--}
								class="w-7 h-7 rounded-lg bg-slate-100 text-slate-700 flex items-center justify-center hover:bg-slate-200 active:scale-95"
							>
								<Minus class="w-3.5 h-3.5" />
							</button>
							<span class="w-6 text-center font-bold text-sm">{itemQuantity}</span>
							<button
								type="button"
								onclick={() => itemQuantity++}
								class="w-7 h-7 rounded-lg bg-orange-50 text-orange-600 flex items-center justify-center hover:bg-orange-100 active:scale-95"
							>
								<Plus class="w-3.5 h-3.5" />
							</button>
						</div>

						<button
							type="button"
							onclick={handleAddToCart}
							class="flex-1 bg-orange-600 hover:bg-orange-700 text-white font-bold py-3 px-4 rounded-xl shadow-lg shadow-orange-500/30 flex items-center justify-between text-sm transition-all"
						>
							<span>Tambah ke Keranjang</span>
							<span class="font-['Outfit']">{formatRupiah(getModalItemTotal())}</span>
						</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>
