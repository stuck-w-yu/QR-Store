<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { Category, Menu } from '$lib/types';
	import { 
		Plus, Utensils, Trash2, Check, X, 
		Image, ToggleLeft, ToggleRight, FolderPlus 
	} from '@lucide/svelte';

	let categories = $state<Category[]>([]);
	let menus = $state<Menu[]>([]);
	let loading = $state(true);

	// Modal / Form state
	let showAddMenu = $state(false);
	let showAddCategory = $state(false);

	let newCatName = $state('');
	let newMenu = $state({
		name: '',
		category_id: '',
		price: 25000,
		description: '',
		image_url: '',
		available: true
	});

	async function loadData() {
		try {
			loading = true;
			const [cats, mList] = await Promise.all([
				api.get<Category[]>('/categories'),
				api.get<Menu[]>('/menus')
			]);
			categories = cats;
			menus = mList;
			if (cats.length > 0 && !newMenu.category_id) {
				newMenu.category_id = cats[0].id;
			}
		} catch (e) {
			console.error('Failed to load menu data', e);
		} finally {
			loading = false;
		}
	}

	async function handleAddCategory(e: SubmitEvent) {
		e.preventDefault();
		if (!newCatName.trim()) return;
		try {
			await api.post('/categories', { name: newCatName.trim() });
			newCatName = '';
			showAddCategory = false;
			await loadData();
		} catch (e) {
			alert('Gagal menambah kategori');
		}
	}

	async function handleAddMenu(e: SubmitEvent) {
		e.preventDefault();
		try {
			await api.post('/menus', {
				name: newMenu.name,
				category_id: newMenu.category_id || undefined,
				price: Number(newMenu.price),
				description: newMenu.description || undefined,
				image_url: newMenu.image_url || undefined,
				available: true
			});
			showAddMenu = false;
			newMenu = {
				name: '',
				category_id: categories[0]?.id || '',
				price: 25000,
				description: '',
				image_url: '',
				available: true
			};
			await loadData();
		} catch (e) {
			alert('Gagal menambah menu');
		}
	}

	async function handleToggleAvailability(menuItem: Menu) {
		try {
			const updated = !menuItem.available;
			await api.patch(`/menus/${menuItem.id}/availability`, { available: updated });
			menuItem.available = updated;
		} catch (e) {
			alert('Gagal mengubah status ketersediaan');
		}
	}

	async function handleDeleteMenu(id: string) {
		if (!confirm('Hapus menu ini?')) return;
		try {
			await api.delete(`/menus/${id}`);
			await loadData();
		} catch (e) {
			alert('Gagal menghapus menu');
		}
	}

	onMount(() => {
		loadData();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Menu & Kategori</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Kelola daftar hidangan makanan, minuman, dan ketersediaan stok menu
			</p>
		</div>

		<div class="flex flex-wrap items-center gap-2 w-full sm:w-auto">
			<button
				type="button"
				onclick={() => (showAddCategory = true)}
				class="flex-1 sm:flex-initial justify-center bg-white hover:bg-slate-50 text-slate-700 font-bold text-xs px-3.5 py-2.5 rounded-xl border border-slate-200 shadow-xs flex items-center gap-1.5 transition-colors"
			>
				<FolderPlus class="w-4 h-4 text-slate-500" />
				<span>+ Kategori</span>
			</button>

			<button
				type="button"
				onclick={() => (showAddMenu = true)}
				class="flex-1 sm:flex-initial justify-center bg-orange-600 hover:bg-orange-700 text-white font-bold text-xs px-4 py-2.5 rounded-xl shadow-md shadow-orange-600/30 flex items-center gap-1.5 transition-colors"
			>
				<Plus class="w-4 h-4" />
				<span>+ Tambah Menu</span>
			</button>
		</div>
	</div>

	<!-- Add Category Modal -->
	{#if showAddCategory}
		<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-6 w-full max-w-sm space-y-4 shadow-xl max-h-[90vh] overflow-y-auto">
				<h3 class="font-bold text-base text-slate-900">Tambah Kategori Baru</h3>
				<form onsubmit={handleAddCategory} class="space-y-3">
					<input
						type="text"
						bind:value={newCatName}
						placeholder="Nama Kategori (cth: Paket Hemat)..."
						required
						class="w-full text-xs p-3 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 font-medium"
					/>
					<div class="flex justify-end gap-2 pt-2">
						<button
							type="button"
							onclick={() => (showAddCategory = false)}
							class="px-4 py-2 text-xs font-semibold text-slate-600 rounded-xl hover:bg-slate-100"
						>
							Batal
						</button>
						<button
							type="submit"
							class="px-4 py-2 text-xs font-bold bg-orange-600 text-white rounded-xl shadow-md"
						>
							Simpan
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Add Menu Modal -->
	{#if showAddMenu}
		<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-6 w-full max-w-md space-y-4 shadow-xl overflow-y-auto max-h-[90vh]">
				<h3 class="font-bold text-base text-slate-900">Tambah Menu Makanan / Minuman</h3>
				<form onsubmit={handleAddMenu} class="space-y-3.5 text-xs">
					<div>
						<label for="menu-name" class="block font-bold text-slate-700 mb-1">Nama Menu</label>
						<input
							id="menu-name"
							type="text"
							bind:value={newMenu.name}
							required
							placeholder="Cth: Sate Ayam Madura"
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div>
						<label for="menu-category" class="block font-bold text-slate-700 mb-1">Kategori</label>
						<select
							id="menu-category"
							bind:value={newMenu.category_id}
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 bg-white"
						>
							{#each categories as cat}
								<option value={cat.id}>{cat.name}</option>
							{/each}
						</select>
					</div>

					<div>
						<label for="menu-price" class="block font-bold text-slate-700 mb-1">Harga (Rupiah)</label>
						<input
							id="menu-price"
							type="number"
							bind:value={newMenu.price}
							required
							min="0"
							step="1000"
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div>
						<label for="menu-desc" class="block font-bold text-slate-700 mb-1">Deskripsi Singkat</label>
						<textarea
							id="menu-desc"
							bind:value={newMenu.description}
							rows="2"
							placeholder="Bahan utama, rasa, isi porsi..."
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 resize-none"
						></textarea>
					</div>

					<div>
						<label for="menu-image" class="block font-bold text-slate-700 mb-1">URL Gambar (Unsplash / CDN)</label>
						<input
							id="menu-image"
							type="url"
							bind:value={newMenu.image_url}
							placeholder="https://images.unsplash.com/..."
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div class="flex justify-end gap-2 pt-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (showAddMenu = false)}
							class="px-4 py-2 font-semibold text-slate-600 rounded-xl hover:bg-slate-100"
						>
							Batal
						</button>
						<button
							type="submit"
							class="px-5 py-2 font-bold bg-orange-600 text-white rounded-xl shadow-md"
						>
							Tambah Menu
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Menus Table / Cards List -->
	{#if loading}
		<div class="text-center py-20 text-slate-500 text-xs">Memuat katalog menu...</div>
	{:else}
		<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
			<div class="p-5 border-b border-slate-100 flex items-center justify-between">
				<h2 class="font-extrabold text-sm text-slate-900 font-['Outfit']">Daftar Hidangan Restoran</h2>
				<span class="text-xs font-semibold text-slate-500">{menus.length} Item Terdaftar</span>
			</div>

			<div class="divide-y divide-slate-100">
				{#each menus as m}
					<div class="p-4 sm:px-6 flex flex-col sm:flex-row sm:items-center justify-between gap-3 sm:gap-4 hover:bg-slate-50/60 transition-colors">
						<div class="flex items-start sm:items-center gap-3.5 sm:gap-4 min-w-0">
							{#if m.image_url}
								<img
									src={m.image_url}
									alt={m.name}
									class="w-14 h-14 rounded-2xl object-cover shrink-0 bg-slate-100 shadow-2xs"
								/>
							{:else}
								<div class="w-14 h-14 rounded-2xl bg-orange-50 text-orange-500 flex items-center justify-center shrink-0">
									<Utensils class="w-6 h-6" />
								</div>
							{/if}

							<div class="min-w-0 flex-1">
								<div class="flex flex-wrap items-center gap-2">
									<h3 class="font-bold text-sm text-slate-900">{m.name}</h3>
									{#if m.category_name}
										<span class="text-[10px] font-semibold px-2 py-0.5 bg-slate-100 text-slate-600 rounded-md">
											{m.category_name}
										</span>
									{/if}
								</div>
								<div class="font-black text-xs text-orange-600 font-['Outfit'] mt-0.5">
									{formatRupiah(m.price)}
								</div>
								{#if m.description}
									<p class="text-[11px] text-slate-500 line-clamp-2 sm:line-clamp-1 max-w-md mt-0.5">
										{m.description}
									</p>
								{/if}
							</div>
						</div>

						<div class="flex items-center justify-between sm:justify-end gap-3 w-full sm:w-auto pt-2.5 sm:pt-0 border-t border-slate-100 sm:border-0">
							<!-- Availability Toggle -->
							<button
								type="button"
								onclick={() => handleToggleAvailability(m)}
								class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold transition-all {m.available
									? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
									: 'bg-slate-100 text-slate-400 border border-slate-200 line-through'}"
								title="Ubah status ketersediaan"
							>
								<span>{m.available ? 'Tersedia' : 'Habis'}</span>
							</button>

							<button
								type="button"
								onclick={() => handleDeleteMenu(m.id)}
								class="p-2 text-slate-400 hover:text-red-500 rounded-xl transition-colors"
								title="Hapus Menu"
							>
								<Trash2 class="w-4 h-4" />
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>
