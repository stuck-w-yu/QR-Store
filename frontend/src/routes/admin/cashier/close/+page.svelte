<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { formatRupiah } from '$lib/api/client';
	import { shiftStore } from '$lib/stores/shift.svelte';
	import { Lock, ArrowLeft, AlertCircle, AlertTriangle, CheckCircle2 } from '@lucide/svelte';

	let loading = $state(true);
	let submitting = $state(false);
	let error = $state<string | null>(null);
	let actualAmount = $state<number>(0);
	let closingNotes = $state('');

	let difference = $derived(
		actualAmount - (shiftStore.summary?.expected_amount || 0)
	);

	async function init() {
		loading = true;
		try {
			await shiftStore.loadCurrentShift();
			if (!shiftStore.currentShift) {
				goto('/admin/cashier');
				return;
			}
			actualAmount = shiftStore.summary?.expected_amount || 0;
		} catch (e: any) {
			error = e?.message || 'Gagal memuat status shift';
		} finally {
			loading = false;
		}
	}

	async function handleSubmit() {
		submitting = true;
		error = null;
		try {
			await shiftStore.closeShift(actualAmount, closingNotes);
			goto('/admin/cashier');
		} catch (e: any) {
			error = e?.message || 'Gagal menutup shift kasir';
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
			<div class="w-12 h-12 rounded-2xl bg-red-50 text-red-600 flex items-center justify-center mb-3">
				<Lock class="w-6 h-6" />
			</div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Tutup Shift Kasir (Closing)</h1>
			<p class="text-xs text-slate-500 font-medium">
				Pastikan perhitungan rekonsiliasi penjualan dan penerimaan telah sesuai sebelum menutup shift
			</p>
		</div>

		{#if error}
			<div class="p-3.5 bg-red-50 text-red-600 text-xs rounded-xl flex items-center gap-2.5 border border-red-200">
				<AlertCircle class="w-4 h-4 shrink-0" />
				<span>{error}</span>
			</div>
		{/if}

		{#if loading}
			<div class="text-center py-10 text-slate-400 text-xs">Menghitung rekonsiliasi kasir...</div>
		{:else if shiftStore.currentShift}
			<!-- Reconciliation Summary -->
			<div class="p-5 bg-slate-50 rounded-2xl border border-slate-100 space-y-3 text-xs">
				<div class="flex justify-between text-slate-500">
					<span>Gross Penjualan</span>
					<span class="font-bold text-slate-900">{formatRupiah(shiftStore.summary?.gross_sales || 0)}</span>
				</div>
				<div class="flex justify-between text-slate-500">
					<span>Total Refund</span>
					<span class="font-bold text-red-600">-{formatRupiah(shiftStore.summary?.refund_total || 0)}</span>
				</div>
				<div class="flex justify-between text-slate-500">
					<span>Total Void</span>
					<span class="font-bold text-amber-600">-{formatRupiah(shiftStore.summary?.void_total || 0)}</span>
				</div>
				<div class="border-t border-slate-200 pt-3 flex justify-between font-extrabold text-slate-900 text-sm">
					<span>Net Penjualan</span>
					<span class="text-orange-600 font-['Outfit']">{formatRupiah(shiftStore.summary?.net_sales || 0)}</span>
				</div>
				<div class="flex justify-between text-slate-500 pt-1 border-t border-slate-200/60">
					<span>Expected Total (Net + Saldo Awal)</span>
					<span class="font-bold text-slate-900">{formatRupiah(shiftStore.summary?.expected_amount || 0)}</span>
				</div>
			</div>

			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="space-y-5 text-xs font-medium">
				<div class="space-y-2">
					<label for="close-actual-page" class="font-bold text-slate-700">Nominal Aktual Tercatat</label>
					<div class="relative">
						<span class="absolute left-4 top-3 font-bold text-slate-400">Rp</span>
						<input
							id="close-actual-page"
							type="number"
							min="0"
							bind:value={actualAmount}
							class="w-full pl-11 pr-4 py-3 rounded-xl border border-slate-200 font-bold text-sm focus:outline-none focus:border-orange-500"
							placeholder="0"
						/>
					</div>
					{#if difference !== 0}
						<div class="text-[11px] font-bold {difference > 0 ? 'text-emerald-600' : 'text-red-600'}">
							Selisih (Variance): {difference > 0 ? '+' : ''}{formatRupiah(difference)}
						</div>
					{/if}
				</div>

				<div class="space-y-2">
					<label for="close-notes-page" class="font-bold text-slate-700">Catatan Penutupan Shift</label>
					<textarea
						id="close-notes-page"
						bind:value={closingNotes}
						rows="3"
						class="w-full px-4 py-3 rounded-xl border border-slate-200 font-normal focus:outline-none focus:border-orange-500"
						placeholder="Catatan penutupan shift..."
					></textarea>
				</div>

				<div class="p-4 bg-amber-50 rounded-2xl border border-amber-200 text-amber-900 text-xs flex gap-2.5 items-start">
					<AlertTriangle class="w-4 h-4 shrink-0 text-amber-600 mt-0.5" />
					<p class="text-[11px] leading-relaxed">
						Shift yang ditutup berstatus final (CLOSED) dan tercatat dalam audit log.
					</p>
				</div>

				<div class="pt-3">
					<button
						type="submit"
						disabled={submitting}
						class="w-full py-3 bg-red-600 hover:bg-red-700 text-white rounded-2xl font-bold shadow-lg shadow-red-600/20 transition-all text-xs disabled:opacity-50"
					>
						{submitting ? 'Menutup Shift...' : 'Konfirmasi Tutup Kasir Sekarang'}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
