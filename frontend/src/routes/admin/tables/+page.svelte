<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { getTableQRCodeDataURL } from '$lib/utils/qr';
	import type { Table } from '$lib/types';
	import { 
		Plus, QrCode, Download, RefreshCw, Trash2, 
		ExternalLink, Check, X, AlertCircle 
	} from '@lucide/svelte';

	let tables = $state<Table[]>([]);
	let loading = $state(true);
	let newTableName = $state('');
	let creating = $state(false);
	let activeQRTable = $state<Table | null>(null);
	let qrDataUrls = $state<Record<string, string>>({});

	async function loadTables() {
		try {
			loading = true;
			const data = await api.get<Table[]>('/tables');
			tables = data;

			// Generate QR data URLs locally
			for (const t of data) {
				getTableQRCodeDataURL(t.qr_token).then((url) => {
					qrDataUrls[t.id] = url;
				});
			}
		} catch (e) {
			console.error('Failed to load tables', e);
		} finally {
			loading = false;
		}
	}

	async function handleCreateTable(e: SubmitEvent) {
		e.preventDefault();
		if (!newTableName.trim()) return;

		try {
			creating = true;
			await api.post('/tables', { name: newTableName.trim() });
			newTableName = '';
			await loadTables();
		} catch (e) {
			alert('Gagal menambah meja');
		} finally {
			creating = false;
		}
	}

	async function handleRegenerateQR(tableId: string) {
		if (!confirm('Apakah Anda yakin ingin memperbarui QR Code meja ini? QR Code lama tidak akan dapat digunakan lagi.')) {
			return;
		}

		try {
			await api.post(`/tables/${tableId}/qr/regenerate`);
			await loadTables();
			if (activeQRTable && activeQRTable.id === tableId) {
				activeQRTable = tables.find(t => t.id === tableId) || null;
			}
		} catch (e) {
			alert('Gagal meregenerasi token');
		}
	}

	async function handleDeleteTable(tableId: string) {
		if (!confirm('Hapus meja ini dari sistem?')) return;
		try {
			await api.delete(`/tables/${tableId}`);
			await loadTables();
			if (activeQRTable?.id === tableId) activeQRTable = null;
		} catch (e) {
			alert('Gagal menghapus meja');
		}
	}

	onMount(() => {
		loadTables();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Manajemen Meja & QR Code</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Setiap meja memiliki token QR Code unik untuk pemesanan mandiri oleh pelanggan
			</p>
		</div>

		<!-- Add Table Form -->
		<form onsubmit={handleCreateTable} class="flex items-center gap-2 w-full sm:w-auto">
			<input
				type="text"
				bind:value={newTableName}
				placeholder="Nama/Nomor Meja (cth: Meja 06)..."
				required
				class="text-xs px-3.5 py-2.5 rounded-xl border border-slate-200 bg-white focus:outline-none focus:border-orange-500 flex-1 sm:w-60"
			/>
			<button
				type="submit"
				disabled={creating}
				class="bg-orange-600 hover:bg-orange-700 text-white font-bold text-xs px-4 py-2.5 rounded-xl shadow-md shadow-orange-600/30 flex items-center justify-center gap-1.5 disabled:opacity-50 transition-all active:scale-95 shrink-0"
			>
				<Plus class="w-4 h-4" />
				<span>Tambah</span>
			</button>
		</form>
	</div>

	<!-- Tables Grid -->
	{#if loading}
		<div class="text-center py-20 text-slate-500 text-xs">Memuat daftar meja...</div>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
			{#each tables as table}
				<div class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs flex flex-col justify-between space-y-4 hover:shadow-md transition-shadow">
					<div class="flex items-start justify-between">
						<div>
							<h3 class="font-bold text-base text-slate-900">{table.name}</h3>
							<span class="text-[11px] font-mono text-slate-400 block mt-0.5 truncate max-w-35">
								{table.qr_token}
							</span>
						</div>
						<span class="px-2 py-0.5 rounded-full text-[10px] font-bold {table.status === 'ACTIVE' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-slate-100 text-slate-500'}">
							{table.status}
						</span>
					</div>

					<!-- QR Preview Box -->
					<div class="bg-slate-50 border border-slate-200/80 rounded-2xl p-4 flex flex-col items-center justify-center space-y-2">
						{#if qrDataUrls[table.id]}
							<img
								src={qrDataUrls[table.id]}
								alt={`QR ${table.name}`}
								class="w-32 h-32 object-contain bg-white p-2 rounded-xl shadow-xs"
							/>
						{:else}
							<div class="w-32 h-32 bg-white rounded-xl flex items-center justify-center text-slate-300">
								<QrCode class="w-12 h-12" />
							</div>
						{/if}
						<span class="text-[10px] text-slate-400 font-semibold uppercase tracking-wider">Pindai Meja Ini</span>
					</div>

					<!-- Action Buttons -->
					<div class="space-y-2 pt-2 border-t border-slate-100">
						<div class="flex items-center gap-2">
							<a
								href={qrDataUrls[table.id] || '#'}
								download={`QR-${table.name}.png`}
								class="flex-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs py-2 px-3 rounded-xl flex items-center justify-center gap-1.5 transition-colors"
							>
								<Download class="w-3.5 h-3.5" />
								<span>Unduh PNG</span>
							</a>

							<a
								href={`/order?token=${table.qr_token}`}
								target="_blank"
								class="p-2 bg-orange-50 hover:bg-orange-100 text-orange-600 rounded-xl transition-colors"
								title="Buka link pelanggan"
							>
								<ExternalLink class="w-4 h-4" />
							</a>
						</div>

						<div class="flex items-center justify-between text-[11px]">
							<button
								type="button"
								onclick={() => handleRegenerateQR(table.id)}
								class="text-slate-500 hover:text-slate-800 font-medium flex items-center gap-1 transition-colors"
							>
								<RefreshCw class="w-3 h-3" />
								<span>Reset Token</span>
							</button>

							<button
								type="button"
								onclick={() => handleDeleteTable(table.id)}
								class="text-red-500 hover:text-red-700 font-medium flex items-center gap-1 transition-colors"
							>
								<Trash2 class="w-3 h-3" />
								<span>Hapus</span>
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
