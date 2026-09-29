<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import type { Register } from '$lib/types';
	import { Monitor, Plus, Edit2, Trash2, CheckCircle2, XCircle, RefreshCw, AlertCircle } from '@lucide/svelte';

	let registers = $state<Register[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingRegister = $state<Register | null>(null);
	let registerName = $state('');
	let registerStatus = $state<'ACTIVE' | 'INACTIVE'>('ACTIVE');
	let errorMsg = $state<string | null>(null);
	let submitting = $state(false);

	async function loadRegisters() {
		loading = true;
		try {
			const data = await api.get<Register[]>('/registers');
			registers = data;
		} catch (e: any) {
			console.error('Failed to load registers:', e);
		} finally {
			loading = false;
		}
	}

	function openCreateModal() {
		editingRegister = null;
		registerName = '';
		registerStatus = 'ACTIVE';
		errorMsg = null;
		showModal = true;
	}

	function openEditModal(reg: Register) {
		editingRegister = reg;
		registerName = reg.name;
		registerStatus = reg.status as 'ACTIVE' | 'INACTIVE';
		errorMsg = null;
		showModal = true;
	}

	async function handleSaveRegister() {
		if (!registerName.trim()) {
			errorMsg = 'Nama mesin kasir wajib diisi';
			return;
		}

		submitting = true;
		errorMsg = null;
		try {
			if (editingRegister) {
				await api.patch(`/registers/${editingRegister.id}`, {
					name: registerName,
					status: registerStatus
				});
			} else {
				await api.post('/registers', {
					name: registerName
				});
			}
			showModal = false;
			await loadRegisters();
		} catch (e: any) {
			errorMsg = e?.message || 'Gagal menyimpan register';
		} finally {
			submitting = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Hapus mesin kasir ini? Pastikan tidak ada shift yang sedang berjalan.')) return;
		try {
			await api.delete(`/registers/${id}`);
			await loadRegisters();
		} catch (e: any) {
			alert(e?.message || 'Gagal menghapus register');
		}
	}

	onMount(() => {
		loadRegisters();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Mesin Kasir (Registers)</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Kelola titik terminal kasir POS restoran untuk pembagian sesi shift kasir
			</p>
		</div>

		<div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto">
			<button
				type="button"
				onclick={loadRegisters}
				class="p-2.5 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-colors shrink-0"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>

			<button
				type="button"
				onclick={openCreateModal}
				class="flex-1 sm:flex-initial justify-center px-4 py-2.5 bg-orange-600 hover:bg-orange-700 text-white rounded-xl shadow-md shadow-orange-600/20 transition-all font-bold text-xs flex items-center gap-2"
			>
				<Plus class="w-4 h-4" />
				<span>Tambah Mesin Kasir</span>
			</button>
		</div>
	</div>

	<!-- Registers Table -->
	<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
		{#if loading}
			<div class="text-center py-20 text-slate-400 text-xs font-semibold">Memuat daftar mesin kasir...</div>
		{:else if registers.length === 0}
			<div class="text-center py-20 text-slate-400 text-xs font-medium">Belum ada mesin kasir yang terdaftar.</div>
		{:else}
			<!-- Mobile Cards View -->
			<div class="md:hidden divide-y divide-slate-100">
				{#each registers as reg}
					<div class="p-4 space-y-3">
						<div class="flex items-start justify-between gap-2">
							<div class="flex items-center gap-2.5">
								<div class="w-8 h-8 rounded-lg bg-orange-50 text-orange-600 flex items-center justify-center shrink-0">
									<Monitor class="w-4 h-4" />
								</div>
								<div>
									<h3 class="font-bold text-sm text-slate-900">{reg.name}</h3>
									<span class="font-mono text-[10px] text-slate-400">ID: {reg.id}</span>
								</div>
							</div>

							{#if reg.status === 'ACTIVE'}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 shrink-0">
									<span class="w-1.5 h-1.5 rounded-full bg-emerald-600"></span>
									Aktif
								</span>
							{:else}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600 border border-slate-200 shrink-0">
									<span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
									Nonaktif
								</span>
							{/if}
						</div>

						<div class="flex items-center justify-between text-xs pt-1 border-t border-slate-100">
							<span class="text-[11px] text-slate-400 font-mono">
								{new Date(reg.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}
							</span>

							<div class="flex items-center gap-2">
								<button
									type="button"
									onclick={() => openEditModal(reg)}
									class="px-3 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs flex items-center gap-1 transition-colors"
								>
									<Edit2 class="w-3.5 h-3.5" />
									<span>Edit</span>
								</button>
								<button
									type="button"
									onclick={() => handleDelete(reg.id)}
									class="p-1.5 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
									title="Hapus Register"
								>
									<Trash2 class="w-3.5 h-3.5" />
								</button>
							</div>
						</div>
					</div>
				{/each}
			</div>

			<!-- Desktop Table View -->
			<div class="hidden md:block overflow-x-auto">
				<table class="w-full text-left text-xs border-collapse">
					<thead>
						<tr class="bg-slate-50/80 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
							<th class="p-4 pl-6">ID Register</th>
							<th class="p-4">Nama Terminal</th>
							<th class="p-4">Status Operasional</th>
							<th class="p-4">Terdaftar Sejak</th>
							<th class="p-4 pr-6 text-right">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 font-medium text-slate-700">
						{#each registers as reg}
							<tr class="hover:bg-slate-50/50 transition-colors">
								<td class="p-4 pl-6 font-mono font-bold text-slate-900">{reg.id}</td>
								<td class="p-4 font-bold text-slate-900 flex items-center gap-2.5">
									<div class="w-8 h-8 rounded-lg bg-orange-50 text-orange-600 flex items-center justify-center shrink-0">
										<Monitor class="w-4 h-4" />
									</div>
									<span>{reg.name}</span>
								</td>
								<td class="p-4">
									{#if reg.status === 'ACTIVE'}
										<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200">
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-600"></span>
											Aktif
										</span>
									{:else}
										<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-bold bg-slate-100 text-slate-600 border border-slate-200">
											<span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span>
											Nonaktif
										</span>
									{/if}
								</td>
								<td class="p-4 text-slate-400 font-mono text-[11px]">
									{new Date(reg.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })}
								</td>
								<td class="p-4 pr-6 text-right space-x-2">
									<button
										type="button"
										onclick={() => openEditModal(reg)}
										class="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 transition-colors"
										title="Ubah Nama/Status"
									>
										<Edit2 class="w-3.5 h-3.5" />
									</button>
									<button
										type="button"
										onclick={() => handleDelete(reg.id)}
										class="p-1.5 rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
										title="Hapus Register"
									>
										<Trash2 class="w-3.5 h-3.5" />
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

	<!-- Modal Form -->
	{#if showModal}
		<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-8 w-full max-w-md space-y-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
				<div class="flex items-center justify-between border-b border-slate-100 pb-4">
					<h3 class="text-lg font-black text-slate-900 font-['Outfit']">
						{editingRegister ? 'Ubah Mesin Kasir' : 'Tambah Mesin Kasir'}
					</h3>
					<button
						type="button"
						onclick={() => (showModal = false)}
						class="text-slate-400 hover:text-slate-600 text-xs font-bold p-1 rounded-lg"
					>
						✕
					</button>
				</div>

				{#if errorMsg}
					<div class="p-3 bg-red-50 text-red-600 text-xs rounded-xl flex items-center gap-2 border border-red-200">
						<AlertCircle class="w-4 h-4 shrink-0" />
						<span>{errorMsg}</span>
					</div>
				{/if}

				<form onsubmit={(e) => { e.preventDefault(); handleSaveRegister(); }} class="space-y-4 text-xs font-medium">
					<div class="space-y-1.5">
						<label for="reg-name-input" class="font-bold text-slate-700">Nama Terminal Register</label>
						<input
							id="reg-name-input"
							type="text"
							bind:value={registerName}
							class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 font-bold focus:outline-none focus:border-orange-500"
							placeholder="Contoh: POS 01 - Kasir Utama"
							required
						/>
					</div>

					{#if editingRegister}
						<div class="space-y-1.5">
							<label for="reg-status-select" class="font-bold text-slate-700">Status</label>
							<select
								id="reg-status-select"
								bind:value={registerStatus}
								class="w-full px-3.5 py-2.5 rounded-xl border border-slate-200 bg-slate-50 font-semibold focus:outline-none focus:border-orange-500"
							>
								<option value="ACTIVE">Aktif</option>
								<option value="INACTIVE">Nonaktif</option>
							</select>
						</div>
					{/if}

					<div class="pt-4 flex items-center justify-end gap-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (showModal = false)}
							class="px-4 py-2.5 text-slate-600 hover:bg-slate-100 rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="submit"
							disabled={submitting}
							class="px-5 py-2.5 bg-orange-600 hover:bg-orange-700 text-white rounded-xl font-bold shadow-md shadow-orange-600/20 disabled:opacity-50 transition-all"
						>
							{submitting ? 'Menyimpan...' : 'Simpan'}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>
