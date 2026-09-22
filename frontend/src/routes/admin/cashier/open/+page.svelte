<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { api, formatRupiah } from '$lib/api/client';
	import { shiftStore } from '$lib/stores/shift.svelte';
	import type { Register } from '$lib/types';
	import { Store, ArrowLeft, AlertCircle, CheckCircle2 } from '@lucide/svelte';

	let registers = $state<Register[]>([]);
	let selectedRegisterId = $state('');
	let openingBalance = $state(0);
	let loading = $state(true);
	let submitting = $state(false);
	let error = $state<string | null>(null);

	async function init() {
		loading = true;
		try {
			await shiftStore.loadCurrentShift();
			if (shiftStore.currentShift && shiftStore.currentShift.status === 'OPEN') {
				goto('/admin/cashier');
				return;
			}

			const regs = await api.get<Register[]>('/registers');
			registers = regs.filter(r => r.status === 'ACTIVE');
			if (registers.length > 0) {
				selectedRegisterId = registers[0].id;
			}
		} catch (e: any) {
			error = e?.message || 'Gagal memuat terminal kasir';
		} finally {
			loading = false;
		}
	}

	async function handleSubmit() {
		if (!selectedRegisterId) {
			error = 'Pilih terminal register terlebih dahulu';
			return;
		}
		if (openingBalance < 0) {
			error = 'Saldo awal tidak boleh bernilai negatif';
			return;
		}

		submitting = true;
		error = null;
		try {
			await shiftStore.openShift(selectedRegisterId, openingBalance);
			goto('/admin/cashier');
		} catch (e: any) {
			error = e?.message || 'Gagal membuka shift kasir';
		} finally {
			submitting = false;
		}
	}

	onMount(() => {
		init();
	});
</script>

<div class="max-w-xl mx-auto space-y-6">
	<a
		href="/admin/cashier"
		class="inline-flex items-center gap-2 text-xs font-bold text-slate-500 hover:text-slate-900 bg-white border border-slate-200 px-3.5 py-2 rounded-xl shadow-xs transition-colors"
	>
		<ArrowLeft class="w-4 h-4" />
		<span>Kembali ke Kasir POS</span>
	</a>

	<div class="bg-white rounded-3xl p-8 border border-slate-200/80 shadow-xs space-y-6">
		<div class="space-y-1">
			<div class="w-12 h-12 rounded-2xl bg-orange-50 text-orange-600 flex items-center justify-center mb-3">
				<Store class="w-6 h-6" />
			</div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Buka Shift Kasir</h1>
			<p class="text-xs text-slate-500 font-medium">
				Konfigurasikan terminal dan saldo awal untuk memulai shift kerja kasir baru
			</p>
		</div>

		{#if error}
			<div class="p-3.5 bg-red-50 text-red-600 text-xs rounded-xl flex items-center gap-2.5 border border-red-200">
				<AlertCircle class="w-4 h-4 shrink-0" />
				<span>{error}</span>
			</div>
		{/if}

		{#if loading}
			<div class="text-center py-10 text-slate-400 text-xs">Memuat data register...</div>
		{:else}
			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-5 text-xs font-medium">
				<div class="space-y-2">
					<label for="reg-select-page" class="font-bold text-slate-700">Pilih Mesin Kasir (Register)</label>
					<select
						id="reg-select-page"
						bind:value={selectedRegisterId}
						class="w-full px-4 py-3 rounded-xl border border-slate-200 bg-slate-50 font-bold focus:outline-none focus:border-orange-500"
						required
					>
						{#each registers as reg}
							<option value={reg.id}>{reg.name}</option>
						{/each}
					</select>
				</div>

				<div class="space-y-2">
					<label for="bal-input-page" class="font-bold text-slate-700">Saldo Awal Kasir (Cash Float)</label>
					<div class="relative">
						<span class="absolute left-4 top-3 font-bold text-slate-400">Rp</span>
						<input
							id="bal-input-page"
							type="number"
							min="0"
							bind:value={openingBalance}
							class="w-full pl-11 pr-4 py-3 rounded-xl border border-slate-200 font-bold text-sm focus:outline-none focus:border-orange-500"
							placeholder="0"
						/>
					</div>
					<p class="text-[11px] text-slate-400">
						Saldo modal tunai di laci kasir saat shift dibuka (Rp0 untuk sistem non-tunai/cashless).
					</p>
				</div>

				<div class="p-4 bg-emerald-50 rounded-2xl border border-emerald-200 text-emerald-800 space-y-1 text-xs">
					<div class="font-bold flex items-center gap-1.5">
						<CheckCircle2 class="w-4 h-4 text-emerald-600" />
						Prinsip Self-Order Terpisah
					</div>
					<p class="text-[11px] text-emerald-700">
						Pelanggan di meja tetap dapat melakukan self-order dan bayar via QRIS tanpa terganggu status buka/tutup kasir.
					</p>
				</div>

				<div class="pt-3">
					<button
						type="submit"
						disabled={submitting}
						class="w-full py-3 bg-orange-600 hover:bg-orange-700 text-white rounded-2xl font-bold shadow-lg shadow-orange-600/30 transition-all text-xs disabled:opacity-50"
					>
						{submitting ? 'Membuka Shift...' : 'Buka Shift Kasir Sekarang'}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
