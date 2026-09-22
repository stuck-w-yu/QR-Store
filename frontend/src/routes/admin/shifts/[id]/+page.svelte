<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api, formatRupiah } from '$lib/api/client';
	import type { ShiftReportDetail } from '$lib/types';
	import { ArrowLeft, Clock, User, Monitor, Receipt, Wallet, AlertTriangle, CheckCircle2 } from '@lucide/svelte';

	let shiftId = $derived(page.params.id);
	let report = $state<ShiftReportDetail | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	async function loadDetail() {
		loading = true;
		error = null;
		try {
			const data = await api.get<ShiftReportDetail>(`/reports/shifts/${shiftId}`);
			report = data;
		} catch (e: any) {
			error = e?.message || 'Gagal memuat detail laporan shift';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDetail();
	});
</script>

<div class="space-y-8 max-w-5xl mx-auto">
	<div class="flex items-center justify-between">
		<a
			href="/admin/shifts"
			class="inline-flex items-center gap-2 text-xs font-bold text-slate-500 hover:text-slate-900 bg-white border border-slate-200 px-3.5 py-2 rounded-xl shadow-xs transition-colors"
		>
			<ArrowLeft class="w-4 h-4" />
			<span>Kembali ke Riwayat Shift</span>
		</a>

		{#if report}
			<span class="text-xs font-mono font-bold text-slate-400">
				REF: #{report.shift.id}
			</span>
		{/if}
	</div>

	{#if loading}
		<div class="bg-white rounded-3xl p-16 text-center border border-slate-200 text-slate-400 text-xs font-semibold">
			Memuat laporan rincian shift...
		</div>
	{:else if error || !report}
		<div class="bg-white rounded-3xl p-10 text-center border border-red-200 text-red-600 text-xs font-bold space-y-2">
			<AlertTriangle class="w-8 h-8 mx-auto text-red-500" />
			<p>{error || 'Laporan shift tidak ditemukan'}</p>
		</div>
	{:else}
		<!-- Header Card -->
		<div class="bg-white rounded-3xl p-6 md:p-8 border border-slate-200/80 shadow-xs space-y-4">
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-100 pb-4">
				<div>
					<div class="flex items-center gap-2">
						<span class="px-2.5 py-1 rounded-full text-[10px] font-bold {report.shift.status === 'OPEN' ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-slate-100 text-slate-700 border border-slate-200'}">
							STATUS: {report.shift.status}
						</span>
					</div>
					<h1 class="text-2xl font-black text-slate-900 font-['Outfit'] mt-1">
						Laporan Rekonsiliasi Kasir
					</h1>
				</div>

				<div class="text-right">
					<span class="text-xs font-bold text-slate-400 uppercase tracking-wider block">Net Sales Final</span>
					<span class="text-2xl font-black text-orange-600 font-['Outfit']">
						{formatRupiah(report.summary.net_sales)}
					</span>
				</div>
			</div>

			<div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs">
				<div class="space-y-0.5">
					<span class="text-slate-400 font-medium">Terminal Register</span>
					<div class="font-bold text-slate-900 flex items-center gap-1.5">
						<Monitor class="w-3.5 h-3.5 text-slate-400" />
						<span>{report.shift.register_name || 'POS Utama'}</span>
					</div>
				</div>

				<div class="space-y-0.5">
					<span class="text-slate-400 font-medium">Kasir Bertugas</span>
					<div class="font-bold text-slate-900 flex items-center gap-1.5">
						<User class="w-3.5 h-3.5 text-slate-400" />
						<span>{report.shift.cashier_name || 'Kasir'}</span>
					</div>
				</div>

				<div class="space-y-0.5">
					<span class="text-slate-400 font-medium">Waktu Buka Shift</span>
					<div class="font-bold text-slate-900 flex items-center gap-1.5">
						<Clock class="w-3.5 h-3.5 text-slate-400" />
						<span>{new Date(report.shift.opened_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })} WIB</span>
					</div>
				</div>

				<div class="space-y-0.5">
					<span class="text-slate-400 font-medium">Waktu Tutup Shift</span>
					<div class="font-bold text-slate-900 flex items-center gap-1.5">
						<Clock class="w-3.5 h-3.5 text-slate-400" />
						<span>{report.shift.closed_at ? new Date(report.shift.closed_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) + ' WIB' : 'Masih Berjalan'}</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Financial Reconciliation Detail -->
		<div class="bg-white rounded-3xl p-6 md:p-8 border border-slate-200/80 shadow-xs space-y-4">
			<h2 class="text-sm font-extrabold text-slate-900 flex items-center gap-2">
				<Receipt class="w-4 h-4 text-orange-600" />
				Rincian Formula Rekonsiliasi Finansial
			</h2>

			<div class="bg-slate-50 rounded-2xl p-5 border border-slate-100 space-y-3 text-xs">
				<div class="flex justify-between text-slate-600">
					<span>Gross Penjualan Pesanan</span>
					<span class="font-bold text-slate-900">{formatRupiah(report.summary.gross_sales)}</span>
				</div>
				<div class="flex justify-between text-slate-600">
					<span>Total Pengembalian (Refund)</span>
					<span class="font-bold text-red-600">-{formatRupiah(report.summary.refund_total)}</span>
				</div>
				<div class="flex justify-between text-slate-600">
					<span>Total Pembatalan (Void)</span>
					<span class="font-bold text-amber-600">-{formatRupiah(report.summary.void_total)}</span>
				</div>
				<div class="border-t border-slate-200 pt-3 flex justify-between font-extrabold text-sm text-slate-900">
					<span>Net Sales (Gross - Refund - Void)</span>
					<span class="text-orange-600 font-['Outfit']">{formatRupiah(report.summary.net_sales)}</span>
				</div>

				<div class="border-t border-slate-200 pt-3 space-y-2">
					<div class="flex justify-between text-slate-500">
						<span>Saldo Awal Kasir (Cash Float)</span>
						<span class="font-bold text-slate-800">{formatRupiah(report.shift.opening_balance)}</span>
					</div>
					<div class="flex justify-between font-bold text-slate-900">
						<span>Expected Amount (Net + Saldo Awal)</span>
						<span>{formatRupiah(report.summary.expected_amount)}</span>
					</div>
					{#if report.closing?.actual_amount !== undefined && report.closing?.actual_amount !== null}
						<div class="flex justify-between font-bold text-slate-900">
							<span>Actual Amount (Tercatat Saat Closing)</span>
							<span>{formatRupiah(report.closing.actual_amount)}</span>
						</div>
						<div class="flex justify-between font-extrabold {report.closing.difference === 0 ? 'text-emerald-600' : 'text-red-600'}">
							<span>Selisih Rekonsiliasi (Variance)</span>
							<span>{report.closing.difference && report.closing.difference > 0 ? '+' : ''}{formatRupiah(report.closing.difference || 0)}</span>
						</div>
					{/if}
				</div>
			</div>

			{#if report.closing?.notes}
				<div class="p-4 bg-orange-50/50 border border-orange-100 rounded-2xl text-xs space-y-1">
					<span class="font-bold text-orange-900 block">Catatan Penutupan Kasir:</span>
					<p class="text-slate-600 italic">"{report.closing.notes}"</p>
				</div>
			{/if}
		</div>

		<!-- Transactions Table -->
		<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
			<div class="p-5 border-b border-slate-100">
				<h3 class="text-sm font-extrabold text-slate-900">
					Daftar Transaksi Kasir ({report.transactions.length})
				</h3>
			</div>

			{#if report.transactions.length === 0}
				<div class="text-center py-10 text-slate-400 text-xs">Tidak ada transaksi dalam shift ini.</div>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-left text-xs border-collapse">
						<thead>
							<tr class="bg-slate-50/80 border-b border-slate-200/80 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
								<th class="p-4 pl-6">Waktu</th>
								<th class="p-4">Tipe</th>
								<th class="p-4">ID Pesanan</th>
								<th class="p-4">Keterangan</th>
								<th class="p-4 pr-6 text-right">Nominal</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 font-medium text-slate-700">
							{#each report.transactions as tx}
								<tr class="hover:bg-slate-50/50 transition-colors">
									<td class="p-4 pl-6 font-mono text-slate-400 text-[11px]">
										{new Date(tx.created_at).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })} WIB
									</td>
									<td class="p-4 font-bold">
										<span class="px-2 py-0.5 rounded-full text-[10px] {tx.type === 'SALE' ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'}">
											{tx.type}
										</span>
									</td>
									<td class="p-4 font-mono font-bold text-slate-900">
										{tx.order_id ? tx.order_id.slice(-6) : '-'}
									</td>
									<td class="p-4 text-slate-500 text-[11px]">
										{tx.metadata?.payment_method ? 'via ' + tx.metadata.payment_method : tx.metadata?.reason || '-'}
									</td>
									<td class="p-4 pr-6 text-right font-black font-['Outfit'] {tx.type === 'SALE' ? 'text-emerald-600' : 'text-red-600'}">
										{tx.type === 'SALE' ? '+' : '-'}{formatRupiah(tx.amount)}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</div>
	{/if}
</div>
