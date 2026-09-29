<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { CashierShift } from '$lib/types';
	import { ClipboardList, Eye, RefreshCw, Filter, Clock, User, Monitor } from '@lucide/svelte';

	let shifts = $state<CashierShift[]>([]);
	let loading = $state(true);
	let selectedStatus = $state<string>('');

	async function loadShifts() {
		loading = true;
		try {
			const query = selectedStatus ? `?status=${selectedStatus}` : '';
			const data = await api.get<CashierShift[]>(`/reports/shifts${query}`);
			shifts = data || [];
		} catch (e) {
			console.error('Failed to load shift reports:', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadShifts();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Laporan & Riwayat Shift</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Audit riwayat pembukaan, penutupan, dan rekonsiliasi kasir restoran
			</p>
		</div>

		<div class="flex items-center gap-2 sm:gap-3 w-full sm:w-auto">
			<select
				bind:value={selectedStatus}
				onchange={loadShifts}
				class="text-xs font-semibold px-3 py-2 rounded-xl border border-slate-200 bg-white focus:outline-none focus:border-orange-500 flex-1 sm:flex-initial"
			>
				<option value="">Semua Status</option>
				<option value="OPEN">Sedang Aktif (OPEN)</option>
				<option value="CLOSED">Sudah Ditutup (CLOSED)</option>
			</select>

			<button
				type="button"
				onclick={loadShifts}
				class="p-2 bg-white hover:bg-slate-50 text-slate-700 rounded-xl border border-slate-200 shadow-xs transition-colors shrink-0"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
			</button>
		</div>
	</div>

	<!-- Shifts Table -->
	<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
		{#if loading}
			<div class="text-center py-20 text-slate-400 text-xs font-semibold">Memuat riwayat shift kasir...</div>
		{:else if shifts.length === 0}
			<div class="text-center py-20 text-slate-400 text-xs font-medium">Belum ada riwayat shift pada filter ini.</div>
		{:else}
			<!-- Mobile Card List View -->
			<div class="md:hidden divide-y divide-slate-100">
				{#each shifts as s}
					<div class="p-4 space-y-3">
						<div class="flex items-start justify-between gap-2">
							<div>
								<span class="font-mono font-bold text-xs text-slate-900 block">#{s.id.slice(-6).toUpperCase()}</span>
								<div class="font-bold text-sm text-slate-900 flex items-center gap-1.5 mt-0.5">
									<Monitor class="w-3.5 h-3.5 text-slate-400" />
									<span>{s.register_name || 'POS Utama'}</span>
								</div>
							</div>

							{#if s.status === 'OPEN'}
								<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 animate-pulse shrink-0">
									● OPEN
								</span>
							{:else if s.status === 'CLOSED'}
								<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600 border border-slate-200 shrink-0">
									CLOSED
								</span>
							{:else}
								<span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200 shrink-0">
									{s.status}
								</span>
							{/if}
						</div>

						<div class="flex items-center justify-between text-xs pt-1 border-t border-slate-100">
							<div class="flex items-center gap-1.5 text-slate-700 font-semibold">
								<User class="w-3.5 h-3.5 text-slate-400" />
								<span>{s.cashier_name || 'Kasir'}</span>
							</div>

							<div class="text-right">
								<span class="font-black text-sm text-slate-900 font-['Outfit']">{formatRupiah(s.expected_total)}</span>
								{#if s.difference !== undefined && s.difference !== null && s.difference !== 0}
									<span class="block text-[10px] font-semibold {s.difference > 0 ? 'text-emerald-600' : 'text-red-600'}">
										Diff: {s.difference > 0 ? '+' : ''}{formatRupiah(s.difference)}
									</span>
								{/if}
							</div>
						</div>

						<div class="flex items-center justify-between pt-2 border-t border-slate-100 text-[11px] text-slate-400 font-mono">
							<div class="truncate mr-2">
								{new Date(s.opened_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })} {new Date(s.opened_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}
							</div>

							<a
								href={`/admin/shifts/${s.id}`}
								class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl font-bold transition-colors text-[11px] shrink-0"
							>
								<Eye class="w-3.5 h-3.5" />
								<span>Rincian</span>
							</a>
						</div>
					</div>
				{/each}
			</div>

			<!-- Desktop Table View -->
			<div class="hidden md:block overflow-x-auto">
				<table class="w-full text-left text-xs border-collapse">
					<thead>
						<tr class="bg-slate-50/80 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
							<th class="p-4 pl-6">ID Shift</th>
							<th class="p-4">Mesin Kasir</th>
							<th class="p-4">Kasir Bertugas</th>
							<th class="p-4">Waktu Buka / Tutup</th>
							<th class="p-4">Status</th>
							<th class="p-4">Expected Total</th>
							<th class="p-4 pr-6 text-right">Rincian</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-100 font-medium text-slate-700">
						{#each shifts as s}
							<tr class="hover:bg-slate-50/50 transition-colors">
								<td class="p-4 pl-6 font-mono font-bold text-slate-900">#{s.id.slice(-6).toUpperCase()}</td>
								<td class="p-4 font-bold text-slate-900 flex items-center gap-2">
									<Monitor class="w-3.5 h-3.5 text-slate-400" />
									<span>{s.register_name || 'POS Utama'}</span>
								</td>
								<td class="p-4 text-slate-700">
									<div class="flex items-center gap-1.5 font-semibold">
										<User class="w-3.5 h-3.5 text-slate-400" />
										<span>{s.cashier_name || 'Kasir'}</span>
									</div>
								</td>
								<td class="p-4 text-slate-500 font-mono text-[11px]">
									<div>Buka: {new Date(s.opened_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short' })} {new Date(s.opened_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</div>
									{#if s.closed_at}
										<div class="text-slate-400">Tutup: {new Date(s.closed_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</div>
									{/if}
								</td>
								<td class="p-4">
									{#if s.status === 'OPEN'}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 animate-pulse">
											● OPEN
										</span>
									{:else if s.status === 'CLOSED'}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-slate-100 text-slate-600 border border-slate-200">
											CLOSED
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200">
											{s.status}
										</span>
									{/if}
								</td>
								<td class="p-4 font-black text-slate-900 font-['Outfit']">
									{formatRupiah(s.expected_total)}
									{#if s.difference !== undefined && s.difference !== null && s.difference !== 0}
										<span class="block text-[10px] font-semibold {s.difference > 0 ? 'text-emerald-600' : 'text-red-600'}">
											Diff: {s.difference > 0 ? '+' : ''}{formatRupiah(s.difference)}
										</span>
									{/if}
								</td>
								<td class="p-4 pr-6 text-right">
									<a
										href={`/admin/shifts/${s.id}`}
										class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-xl font-bold transition-colors text-[11px]"
									>
										<Eye class="w-3.5 h-3.5" />
										<span>Rincian</span>
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
