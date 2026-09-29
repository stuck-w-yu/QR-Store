<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { TenantSummary, PlatformStats } from '$lib/types';
	import { 
		Store, Users, ShoppingCart, DollarSign, TrendingUp, 
		Plus, RefreshCw, Search, Filter, Eye, Edit2, 
		ShieldAlert, CheckCircle2, XCircle, AlertTriangle, 
		ExternalLink, Trash2, ArrowUpRight, ShieldCheck, 
		Lock, Unlock, ChevronRight, QrCode, Phone, Mail, MapPin, Sparkles
	} from '@lucide/svelte';

	let loading = $state(true);
	let stats = $state<PlatformStats>({
		total_restaurants: 0,
		active_restaurants: 0,
		suspended_restaurants: 0,
		total_owners: 0,
		total_orders: 0,
		total_revenue: 0
	});

	let tenants = $state<TenantSummary[]>([]);
	let searchQuery = $state('');
	let statusFilter = $state('');
	let planFilter = $state('');
	let sortBy = $state<'newest' | 'revenue' | 'orders' | 'name'>('newest');

	// Modal States
	let showOnboardModal = $state(false);
	let showEditModal = $state(false);
	let showDetailModal = $state(false);
	let selectedTenant = $state<TenantSummary | null>(null);

	// Onboarding Form State
	let onboardForm = $state({
		name: '',
		slug: '',
		owner_name: '',
		owner_email: '',
		owner_password: '',
		phone: '',
		address: '',
		plan: 'PRO',
		tax_percent: 10,
		service_percent: 5
	});
	let onboardSubmitting = $state(false);
	let onboardError = $state<string | null>(null);

	// Edit Form State
	let editForm = $state({
		name: '',
		phone: '',
		address: '',
		plan: 'PRO',
		status: 'ACTIVE',
		tax_percent: 10,
		service_percent: 5
	});
	let editSubmitting = $state(false);

	async function loadSuperadminData() {
		try {
			loading = true;
			const [statsRes, tenantsRes] = await Promise.all([
				api.get<PlatformStats>('/superadmin/stats'),
				api.get<TenantSummary[]>('/superadmin/restaurants')
			]);
			stats = statsRes;
			tenants = tenantsRes || [];
		} catch (e: any) {
			console.error('Failed to load superadmin data:', e);
		} finally {
			loading = false;
		}
	}

	function handleNameChange(name: string) {
		onboardForm.name = name;
		// Auto slug generator if not manually edited
		onboardForm.slug = name
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, '-')
			.replace(/(^-|-$)/g, '');
	}

	async function handleOnboard(e: SubmitEvent) {
		e.preventDefault();
		onboardSubmitting = true;
		onboardError = null;
		try {
			await api.post('/superadmin/restaurants', onboardForm);
			showOnboardModal = false;
			onboardForm = {
				name: '',
				slug: '',
				owner_name: '',
				owner_email: '',
				owner_password: '',
				phone: '',
				address: '',
				plan: 'PRO',
				tax_percent: 10,
				service_percent: 5
			};
			await loadSuperadminData();
			alert('Tenant berhasil didaftarkan dan siap beroperasi!');
		} catch (err: any) {
			onboardError = err?.message || 'Gagal mendaftarkan tenant';
		} finally {
			onboardSubmitting = false;
		}
	}

	function openEditModal(t: TenantSummary) {
		selectedTenant = t;
		editForm = {
			name: t.name,
			phone: t.phone || '',
			address: t.address || '',
			plan: t.plan || 'PRO',
			status: t.status,
			tax_percent: t.tax_percent,
			service_percent: t.service_percent
		};
		showEditModal = true;
	}

	async function handleSaveEdit(e: SubmitEvent) {
		e.preventDefault();
		if (!selectedTenant) return;
		editSubmitting = true;
		try {
			await api.patch(`/superadmin/restaurants/${selectedTenant.id}`, editForm);
			showEditModal = false;
			await loadSuperadminData();
		} catch (err: any) {
			alert(err?.message || 'Gagal memperbarui data tenant');
		} finally {
			editSubmitting = false;
		}
	}

	async function handleToggleStatus(t: TenantSummary) {
		const newStatus = t.status === 'ACTIVE' ? 'SUSPENDED' : 'ACTIVE';
		const confirmMsg = newStatus === 'SUSPENDED' 
			? `Bekukan akun toko "${t.name}"? Toko tidak akan dapat memproses pesanan dan login kasir dinonaktifkan.`
			: `Aktifkan kembali akun toko "${t.name}"?`;

		if (!confirm(confirmMsg)) return;

		try {
			await api.patch(`/superadmin/restaurants/${t.id}`, { status: newStatus });
			await loadSuperadminData();
		} catch (err: any) {
			alert(err?.message || 'Gagal mengubah status tenant');
		}
	}

	async function handleDeleteTenant(t: TenantSummary) {
		if (!confirm(`Hapus permanen tenant "${t.name}" dan semua data terkait? Tindakan ini tidak dapat dibatalkan.`)) {
			return;
		}
		try {
			await api.delete(`/superadmin/restaurants/${t.id}`);
			await loadSuperadminData();
			if (selectedTenant?.id === t.id) {
				showDetailModal = false;
			}
		} catch (err: any) {
			alert(err?.message || 'Gagal menghapus tenant');
		}
	}

	// Filter and sort computation
	let filteredTenants = $derived(
		tenants
			.filter((t) => {
				const matchesSearch = 
					t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
					t.slug.toLowerCase().includes(searchQuery.toLowerCase()) ||
					t.owner_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
					t.owner_email.toLowerCase().includes(searchQuery.toLowerCase());
				
				const matchesStatus = !statusFilter || t.status === statusFilter;
				const matchesPlan = !planFilter || t.plan === planFilter;

				return matchesSearch && matchesStatus && matchesPlan;
			})
			.sort((a, b) => {
				if (sortBy === 'revenue') return b.total_revenue - a.total_revenue;
				if (sortBy === 'orders') return b.total_orders - a.total_orders;
				if (sortBy === 'name') return a.name.localeCompare(b.name);
				return new Date(b.created_at || '').getTime() - new Date(a.created_at || '').getTime();
			})
	);

	onMount(() => {
		loadSuperadminData();
	});
</script>

<div class="space-y-8">
	<!-- Page Header & Onboard Trigger -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<div class="flex items-center gap-2">
				<span class="w-2.5 h-2.5 rounded-full bg-indigo-500 animate-ping"></span>
				<span class="text-xs font-bold text-indigo-400 uppercase tracking-wider">Multi-Tenant Oversight</span>
			</div>
			<h1 class="text-2xl sm:text-3xl font-black text-white font-['Outfit'] mt-1">
				Manajemen Business Owner & Toko
			</h1>
			<p class="text-xs text-slate-400 font-medium mt-0.5 max-w-2xl">
				Monitor performa seluruh gerai kuliner, aktifkan/bekukan akses layanan, kelola paket langganan, dan daftarkan mitra restoran baru.
			</p>
		</div>

		<div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto">
			<button
				type="button"
				onclick={loadSuperadminData}
				class="p-2.5 bg-slate-900 hover:bg-slate-800 text-slate-300 rounded-xl border border-slate-800 shadow-xs transition-colors flex items-center justify-center gap-2 text-xs font-bold"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
				<span class="inline">Segarkan</span>
			</button>

			<button
				type="button"
				onclick={() => (showOnboardModal = true)}
				class="flex-1 sm:flex-initial justify-center px-4 py-2.5 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl shadow-lg shadow-indigo-600/30 transition-all font-bold text-xs flex items-center gap-2"
			>
				<Plus class="w-4 h-4" />
				<span>+ Daftarkan Toko Baru</span>
			</button>
		</div>
	</div>

	<!-- Platform Metrics Grid -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
		<!-- Total Restaurants -->
		<div class="bg-slate-900/80 rounded-3xl p-5 border border-slate-800/80 shadow-xs space-y-3 relative overflow-hidden group hover:border-slate-700 transition-colors">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-400 uppercase tracking-wider">Mitra Restoran</span>
				<div class="w-9 h-9 rounded-xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center border border-indigo-500/20">
					<Store class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-3xl font-black text-white font-['Outfit']">
					{stats.total_restaurants}
				</div>
				<div class="flex items-center gap-2 text-[11px] text-slate-400 font-medium mt-1">
					<span class="text-emerald-400 font-bold">● {stats.active_restaurants} Aktif</span>
					{#if stats.suspended_restaurants > 0}
						<span class="text-red-400 font-bold">• {stats.suspended_restaurants} Ditangguhkan</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Business Owners -->
		<div class="bg-slate-900/80 rounded-3xl p-5 border border-slate-800/80 shadow-xs space-y-3 relative overflow-hidden group hover:border-slate-700 transition-colors">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-400 uppercase tracking-wider">Business Owners</span>
				<div class="w-9 h-9 rounded-xl bg-purple-500/10 text-purple-400 flex items-center justify-center border border-purple-500/20">
					<Users class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-3xl font-black text-white font-['Outfit']">
					{stats.total_owners}
				</div>
				<span class="text-[11px] text-slate-400 font-medium block mt-1">
					Akun pemilik usaha terverifikasi
				</span>
			</div>
		</div>

		<!-- Platform Orders -->
		<div class="bg-slate-900/80 rounded-3xl p-5 border border-slate-800/80 shadow-xs space-y-3 relative overflow-hidden group hover:border-slate-700 transition-colors">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-400 uppercase tracking-wider">Total Transaksi Nasional</span>
				<div class="w-9 h-9 rounded-xl bg-blue-500/10 text-blue-400 flex items-center justify-center border border-blue-500/20">
					<ShoppingCart class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-3xl font-black text-white font-['Outfit']">
					{stats.total_orders}
				</div>
				<span class="text-[11px] text-slate-400 font-medium block mt-1">
					Seluruh pesanan meja & kasir
				</span>
			</div>
		</div>

		<!-- Total Platform GMV -->
		<div class="bg-slate-900/80 rounded-3xl p-5 border border-slate-800/80 shadow-xs space-y-3 relative overflow-hidden group hover:border-slate-700 transition-colors">
			<div class="flex items-center justify-between">
				<span class="text-xs font-bold text-slate-400 uppercase tracking-wider">Gross Merchandise Value (GMV)</span>
				<div class="w-9 h-9 rounded-xl bg-emerald-500/10 text-emerald-400 flex items-center justify-center border border-emerald-500/20">
					<TrendingUp class="w-5 h-5" />
				</div>
			</div>
			<div>
				<div class="text-2xl sm:text-3xl font-black text-emerald-400 font-['Outfit'] truncate">
					{formatRupiah(stats.total_revenue)}
				</div>
				<span class="text-[11px] text-slate-400 font-medium block mt-1">
					Volume transaksi sukses diproses
				</span>
			</div>
		</div>
	</div>

	<!-- Filter & Search Toolbar -->
	<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800/80 flex flex-col md:flex-row md:items-center justify-between gap-3">
		<!-- Search -->
		<div class="relative flex-1">
			<Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
			<input
				type="text"
				bind:value={searchQuery}
				placeholder="Cari nama restoran, slug, pemilik, atau email..."
				class="w-full text-xs pl-10 pr-4 py-2.5 bg-slate-950/60 rounded-xl border border-slate-800 focus:outline-none focus:border-indigo-500 text-white placeholder-slate-500 font-medium"
			/>
		</div>

		<!-- Filters & Sort -->
		<div class="flex flex-wrap items-center gap-2 sm:gap-3">
			<!-- Status Filter -->
			<select
				bind:value={statusFilter}
				class="text-xs font-semibold px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800 text-slate-200 focus:outline-none focus:border-indigo-500"
			>
				<option value="">Semua Status</option>
				<option value="ACTIVE">Aktif (ACTIVE)</option>
				<option value="SUSPENDED">Ditangguhkan (SUSPENDED)</option>
			</select>

			<!-- Plan Filter -->
			<select
				bind:value={planFilter}
				class="text-xs font-semibold px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800 text-slate-200 focus:outline-none focus:border-indigo-500"
			>
				<option value="">Semua Paket</option>
				<option value="STARTER">Starter</option>
				<option value="PRO">Pro</option>
				<option value="ENTERPRISE">Enterprise</option>
			</select>

			<!-- Sort -->
			<select
				bind:value={sortBy}
				class="text-xs font-semibold px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800 text-slate-200 focus:outline-none focus:border-indigo-500"
			>
				<option value="newest">Terbaru Bergabung</option>
				<option value="revenue">Omset Tertinggi</option>
				<option value="orders">Pesanan Terbanyak</option>
				<option value="name">Nama Restoran (A-Z)</option>
			</select>
		</div>
	</div>

	<!-- Tenants Roster -->
	{#if loading}
		<div class="bg-slate-900/60 rounded-3xl p-16 text-center border border-slate-800 text-slate-400 text-xs font-semibold">
			<RefreshCw class="w-6 h-6 animate-spin mx-auto mb-3 text-indigo-500" />
			Memuat database tenant & mitra restoran...
		</div>
	{:else if filteredTenants.length === 0}
		<div class="bg-slate-900/60 rounded-3xl p-14 text-center border border-slate-800 text-slate-400 text-xs font-medium space-y-3">
			<Store class="w-10 h-10 mx-auto text-slate-600" />
			<p>Tidak ada tenant yang cocok dengan kriteria pencarian.</p>
			{#if searchQuery || statusFilter || planFilter}
				<button
					type="button"
					onclick={() => { searchQuery = ''; statusFilter = ''; planFilter = ''; }}
					class="text-indigo-400 hover:underline font-bold text-xs"
				>
					Reset Semua Filter
				</button>
			{/if}
		</div>
	{:else}
		<!-- Mobile Card View -->
		<div class="md:hidden space-y-4">
			{#each filteredTenants as t}
				<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800/80 shadow-md space-y-4">
					<div class="flex items-start justify-between gap-3">
						<div class="flex items-center gap-3 min-w-0">
							{#if t.logo_url}
								<img src={t.logo_url} alt={t.name} class="w-11 h-11 rounded-2xl object-cover shrink-0 bg-slate-800" />
							{:else}
								<div class="w-11 h-11 rounded-2xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center font-bold text-base shrink-0">
									{t.name.charAt(0)}
								</div>
							{/if}
							<div class="min-w-0">
								<h3 class="font-extrabold text-sm text-white truncate">{t.name}</h3>
								<span class="text-[11px] text-slate-400 font-mono block">/{t.slug}</span>
							</div>
						</div>

						<div class="flex flex-col items-end gap-1 shrink-0">
							{#if t.status === 'ACTIVE'}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
									<span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
									AKTIF
								</span>
							{:else}
								<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20">
									DITANGGUHKAN
								</span>
							{/if}
							<span class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase {t.plan === 'ENTERPRISE' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30' : t.plan === 'PRO' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' : 'bg-slate-800 text-slate-400'}">
								{t.plan || 'PRO'}
							</span>
						</div>
					</div>

					<!-- Owner Info Box -->
					<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800/80 text-xs space-y-1">
						<div class="flex items-center justify-between">
							<span class="text-slate-400 font-medium">Pemilik:</span>
							<span class="font-bold text-white">{t.owner_name}</span>
						</div>
						<div class="flex items-center justify-between">
							<span class="text-slate-400 font-medium">Email:</span>
							<span class="font-mono text-slate-300 truncate max-w-50">{t.owner_email}</span>
						</div>
						{#if t.phone}
							<div class="flex items-center justify-between">
								<span class="text-slate-400 font-medium">Kontak:</span>
								<span class="font-mono text-slate-300">{t.phone}</span>
							</div>
						{/if}
					</div>

					<!-- Metrics Ribbon -->
					<div class="grid grid-cols-3 gap-2 text-center text-xs py-1">
						<div class="bg-slate-950/40 p-2 rounded-xl border border-slate-800/40">
							<span class="text-[10px] text-slate-400 block uppercase">Meja</span>
							<span class="font-bold text-white text-sm">{t.total_tables}</span>
						</div>
						<div class="bg-slate-950/40 p-2 rounded-xl border border-slate-800/40">
							<span class="text-[10px] text-slate-400 block uppercase">Pesanan</span>
							<span class="font-bold text-white text-sm">{t.total_orders}</span>
						</div>
						<div class="bg-slate-950/40 p-2 rounded-xl border border-slate-800/40">
							<span class="text-[10px] text-slate-400 block uppercase">Omset</span>
							<span class="font-bold text-emerald-400 text-sm truncate block font-['Outfit']">
								{formatRupiah(t.total_revenue)}
							</span>
						</div>
					</div>

					<!-- Actions Bar -->
					<div class="flex items-center justify-between gap-2 pt-2 border-t border-slate-800/80 text-xs">
						<button
							type="button"
							onclick={() => { selectedTenant = t; showDetailModal = true; }}
							class="flex-1 py-2 px-3 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold flex items-center justify-center gap-1.5 transition-colors"
						>
							<Eye class="w-3.5 h-3.5" />
							<span>Rincian</span>
						</button>

						<button
							type="button"
							onclick={() => handleToggleStatus(t)}
							class="py-2 px-3 rounded-xl font-bold flex items-center justify-center gap-1.5 transition-colors {t.status === 'ACTIVE' ? 'bg-amber-500/10 hover:bg-amber-500/20 text-amber-400' : 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400'}"
							title={t.status === 'ACTIVE' ? 'Bekukan Toko' : 'Aktifkan Toko'}
						>
							{#if t.status === 'ACTIVE'}
								<Lock class="w-3.5 h-3.5" />
								<span>Bekukan</span>
							{:else}
								<Unlock class="w-3.5 h-3.5" />
								<span>Buka</span>
							{/if}
						</button>

						<button
							type="button"
							onclick={() => openEditModal(t)}
							class="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
							title="Ubah Konfigurasi"
						>
							<Edit2 class="w-3.5 h-3.5" />
						</button>

						<button
							type="button"
							onclick={() => handleDeleteTenant(t)}
							class="p-2 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-400 transition-colors"
							title="Hapus Tenant"
						>
							<Trash2 class="w-3.5 h-3.5" />
						</button>
					</div>
				</div>
			{/each}
		</div>

		<!-- Desktop Table View -->
		<div class="hidden md:block bg-slate-900 rounded-3xl border border-slate-800/80 shadow-md overflow-hidden">
			<div class="overflow-x-auto">
				<table class="w-full text-left text-xs border-collapse">
					<thead>
						<tr class="bg-slate-950/60 border-b border-slate-800 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
							<th class="p-4 pl-6">Restoran & Slug</th>
							<th class="p-4">Business Owner</th>
							<th class="p-4">Paket</th>
							<th class="p-4">Status</th>
							<th class="p-4">Katalog & Meja</th>
							<th class="p-4">Total Omset</th>
							<th class="p-4 pr-6 text-right">Aksi Manajemen</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-slate-800/60 font-medium text-slate-300">
						{#each filteredTenants as t}
							<tr class="hover:bg-slate-800/30 transition-colors">
								<!-- Restaurant info -->
								<td class="p-4 pl-6">
									<div class="flex items-center gap-3">
										{#if t.logo_url}
											<img src={t.logo_url} alt={t.name} class="w-10 h-10 rounded-xl object-cover bg-slate-800 shrink-0" />
										{:else}
											<div class="w-10 h-10 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center font-bold text-sm shrink-0">
												{t.name.charAt(0)}
											</div>
										{/if}
										<div>
											<span class="font-extrabold text-sm text-white block">{t.name}</span>
											<a 
												href={`/public/restaurants/${t.slug}`} 
												target="_blank"
												class="text-[11px] text-indigo-400 font-mono hover:underline flex items-center gap-1"
											>
												<span>/{t.slug}</span>
												<ExternalLink class="w-3 h-3" />
											</a>
										</div>
									</div>
								</td>

								<!-- Owner info -->
								<td class="p-4">
									<div class="font-bold text-slate-200">{t.owner_name}</div>
									<div class="text-[11px] text-slate-400 font-mono">{t.owner_email}</div>
									{#if t.phone}
										<div class="text-[10px] text-slate-500">{t.phone}</div>
									{/if}
								</td>

								<!-- Plan -->
								<td class="p-4">
									<span class="px-2.5 py-1 rounded-lg text-[10px] font-extrabold uppercase {t.plan === 'ENTERPRISE' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30' : t.plan === 'PRO' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' : 'bg-slate-800 text-slate-400 border border-slate-700'}">
										{t.plan || 'PRO'}
									</span>
								</td>

								<!-- Status -->
								<td class="p-4">
									{#if t.status === 'ACTIVE'}
										<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
											Aktif
										</span>
									{:else}
										<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold bg-red-500/10 text-red-400 border border-red-500/20">
											Ditangguhkan
										</span>
									{/if}
								</td>

								<!-- Catalog & Tables -->
								<td class="p-4 text-[11px] text-slate-300 font-mono">
									<div>{t.total_tables} Meja Aktif</div>
									<div class="text-slate-500">{t.total_menus} Menu Katalog</div>
								</td>

								<!-- Orders & Revenue -->
								<td class="p-4">
									<div class="font-extrabold text-sm text-emerald-400 font-['Outfit']">
										{formatRupiah(t.total_revenue)}
									</div>
									<span class="text-[10px] text-slate-500 block">
										{t.total_orders} total pesanan
									</span>
								</td>

								<!-- Actions -->
								<td class="p-4 pr-6 text-right space-x-1.5">
									<button
										type="button"
										onclick={() => { selectedTenant = t; showDetailModal = true; }}
										class="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
										title="Lihat Rincian Tenant"
									>
										<Eye class="w-3.5 h-3.5" />
									</button>

									<button
										type="button"
										onclick={() => handleToggleStatus(t)}
										class="p-2 rounded-xl transition-colors {t.status === 'ACTIVE' ? 'bg-amber-500/10 hover:bg-amber-500/20 text-amber-400' : 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400'}"
										title={t.status === 'ACTIVE' ? 'Bekukan Toko' : 'Aktifkan Toko'}
									>
										{#if t.status === 'ACTIVE'}
											<Lock class="w-3.5 h-3.5" />
										{:else}
											<Unlock class="w-3.5 h-3.5" />
										{/if}
									</button>

									<button
										type="button"
										onclick={() => openEditModal(t)}
										class="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
										title="Ubah Konfigurasi"
									>
										<Edit2 class="w-3.5 h-3.5" />
									</button>

									<button
										type="button"
										onclick={() => handleDeleteTenant(t)}
										class="p-2 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-400 transition-colors"
										title="Hapus Tenant"
									>
										<Trash2 class="w-3.5 h-3.5" />
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>

<!-- Modal 1: Onboard Mitra Tenant Baru -->
{#if showOnboardModal}
	<div class="fixed inset-0 bg-black/70 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-slate-900 text-white rounded-3xl p-5 sm:p-7 w-full max-w-lg space-y-5 shadow-2xl border border-slate-800 animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
			<div class="flex items-center justify-between border-b border-slate-800 pb-3">
				<div>
					<h3 class="text-lg font-black font-['Outfit'] text-white">Daftarkan Mitra Usaha Baru</h3>
					<p class="text-xs text-slate-400">Buat restoran dan kredensial akun Business Owner</p>
				</div>
				<button
					type="button"
					onclick={() => (showOnboardModal = false)}
					class="text-slate-400 hover:text-white p-1 rounded-lg"
				>
					✕
				</button>
			</div>

			{#if onboardError}
				<div class="p-3 bg-red-500/10 border border-red-500/30 text-red-400 rounded-xl text-xs font-semibold flex items-center gap-2">
					<AlertTriangle class="w-4 h-4 shrink-0" />
					<span>{onboardError}</span>
				</div>
			{/if}

			<form onsubmit={handleOnboard} class="space-y-4 text-xs font-medium">
				<!-- Restaurant Details -->
				<div class="space-y-3">
					<span class="text-[11px] font-bold text-indigo-400 uppercase tracking-wider block">
						1. Data Restoran / Gerai
					</span>
					
					<div>
						<label for="ob-name" class="block font-bold text-slate-300 mb-1">Nama Restoran / Outlet</label>
						<input
							id="ob-name"
							type="text"
							value={onboardForm.name}
							oninput={(e) => handleNameChange((e.target as HTMLInputElement).value)}
							required
							placeholder="Cth: Kopi Kenangan Senja"
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
						/>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div>
							<label for="ob-slug" class="block font-bold text-slate-300 mb-1">Slug URL Toko</label>
							<input
								id="ob-slug"
								type="text"
								bind:value={onboardForm.slug}
								required
								placeholder="kopi-kenangan-senja"
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-indigo-300 font-mono focus:outline-none focus:border-indigo-500"
							/>
						</div>

						<div>
							<label for="ob-plan" class="block font-bold text-slate-300 mb-1">Paket Langganan</label>
							<select
								id="ob-plan"
								bind:value={onboardForm.plan}
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
							>
								<option value="STARTER">Starter (Maks 10 Meja)</option>
								<option value="PRO">Pro (Unlimited Meja + KDS)</option>
								<option value="ENTERPRISE">Enterprise (Multi-Outlet + API)</option>
							</select>
						</div>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div>
							<label for="ob-tax" class="block font-bold text-slate-300 mb-1">Pajak Resto PB1 (%)</label>
							<input
								id="ob-tax"
								type="number"
								step="0.5"
								bind:value={onboardForm.tax_percent}
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
							/>
						</div>
						<div>
							<label for="ob-srv" class="block font-bold text-slate-300 mb-1">Biaya Layanan Service (%)</label>
							<input
								id="ob-srv"
								type="number"
								step="0.5"
								bind:value={onboardForm.service_percent}
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
							/>
						</div>
					</div>

					<div>
						<label for="ob-addr" class="block font-bold text-slate-300 mb-1">Alamat Outlet</label>
						<input
							id="ob-addr"
							type="text"
							bind:value={onboardForm.address}
							placeholder="Jl. Gatot Subroto No. 45, Jakarta"
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
						/>
					</div>
				</div>

				<!-- Owner Details -->
				<div class="space-y-3 pt-3 border-t border-slate-800">
					<span class="text-[11px] font-bold text-indigo-400 uppercase tracking-wider block">
						2. Akun Business Owner (Super Pengelola)
					</span>

					<div>
						<label for="ob-own-name" class="block font-bold text-slate-300 mb-1">Nama Lengkap Pemilik</label>
						<input
							id="ob-own-name"
							type="text"
							bind:value={onboardForm.owner_name}
							required
							placeholder="Cth: Budi Santoso"
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
						/>
					</div>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<div>
							<label for="ob-own-email" class="block font-bold text-slate-300 mb-1">Email Pemilik (Login)</label>
							<input
								id="ob-own-email"
								type="email"
								bind:value={onboardForm.owner_email}
								required
								placeholder="owner@resto.com"
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-mono"
							/>
						</div>

						<div>
							<label for="ob-own-pass" class="block font-bold text-slate-300 mb-1">Kata Sandi Awal</label>
							<input
								id="ob-own-pass"
								type="password"
								bind:value={onboardForm.owner_password}
								required
								placeholder="Minimal 6 karakter"
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
							/>
						</div>
					</div>

					<div>
						<label for="ob-phone" class="block font-bold text-slate-300 mb-1">No. WhatsApp / Telepon</label>
						<input
							id="ob-phone"
							type="tel"
							bind:value={onboardForm.phone}
							placeholder="081234567890"
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-mono"
						/>
					</div>
				</div>

				<div class="pt-4 flex items-center justify-end gap-3 border-t border-slate-800">
					<button
						type="button"
						onclick={() => (showOnboardModal = false)}
						class="px-4 py-2.5 text-slate-400 hover:text-white rounded-xl font-bold transition-colors"
					>
						Batal
					</button>
					<button
						type="submit"
						disabled={onboardSubmitting}
						class="px-5 py-2.5 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl font-bold shadow-lg shadow-indigo-600/30 disabled:opacity-50 transition-all"
					>
						{onboardSubmitting ? 'Mendaftarkan...' : 'Konfirmasi & Daftarkan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Modal 2: Edit Tenant Konfigurasi -->
{#if showEditModal && selectedTenant}
	<div class="fixed inset-0 bg-black/70 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-slate-900 text-white rounded-3xl p-5 sm:p-7 w-full max-w-md space-y-5 shadow-2xl border border-slate-800 animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
			<div class="flex items-center justify-between border-b border-slate-800 pb-3">
				<div>
					<h3 class="text-lg font-black font-['Outfit'] text-white">Ubah Konfigurasi Tenant</h3>
					<span class="text-xs font-mono text-indigo-400">ID: {selectedTenant.id}</span>
				</div>
				<button
					type="button"
					onclick={() => (showEditModal = false)}
					class="text-slate-400 hover:text-white p-1 rounded-lg"
				>
					✕
				</button>
			</div>

			<form onsubmit={handleSaveEdit} class="space-y-3.5 text-xs font-medium">
				<div>
					<label for="edit-name" class="block font-bold text-slate-300 mb-1">Nama Toko</label>
					<input
						id="edit-name"
						type="text"
						bind:value={editForm.name}
						required
						class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
					/>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-plan" class="block font-bold text-slate-300 mb-1">Paket</label>
						<select
							id="edit-plan"
							bind:value={editForm.plan}
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
						>
							<option value="STARTER">Starter</option>
							<option value="PRO">Pro</option>
							<option value="ENTERPRISE">Enterprise</option>
						</select>
					</div>

					<div>
						<label for="edit-status" class="block font-bold text-slate-300 mb-1">Status Toko</label>
						<select
							id="edit-status"
							bind:value={editForm.status}
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
						>
							<option value="ACTIVE">ACTIVE (Aktif)</option>
							<option value="SUSPENDED">SUSPENDED (Dibekukan)</option>
						</select>
					</div>
				</div>

				<div class="grid grid-cols-2 gap-3">
					<div>
						<label for="edit-tax" class="block font-bold text-slate-300 mb-1">Pajak PB1 (%)</label>
						<input
							id="edit-tax"
							type="number"
							step="0.5"
							bind:value={editForm.tax_percent}
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
						/>
					</div>

					<div>
						<label for="edit-srv" class="block font-bold text-slate-300 mb-1">Service (%)</label>
						<input
							id="edit-srv"
							type="number"
							step="0.5"
							bind:value={editForm.service_percent}
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
						/>
					</div>
				</div>

				<div>
					<label for="edit-phone" class="block font-bold text-slate-300 mb-1">No. Kontak Toko</label>
					<input
						id="edit-phone"
						type="tel"
						bind:value={editForm.phone}
						class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-mono"
					/>
				</div>

				<div>
					<label for="edit-addr" class="block font-bold text-slate-300 mb-1">Alamat Outlet</label>
					<input
						id="edit-addr"
						type="text"
						bind:value={editForm.address}
						class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
					/>
				</div>

				<div class="pt-4 flex items-center justify-end gap-3 border-t border-slate-800">
					<button
						type="button"
						onclick={() => (showEditModal = false)}
						class="px-4 py-2.5 text-slate-400 hover:text-white rounded-xl font-bold transition-colors"
					>
						Batal
					</button>
					<button
						type="submit"
						disabled={editSubmitting}
						class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl font-bold shadow-lg shadow-indigo-600/30 disabled:opacity-50 transition-all"
					>
						{editSubmitting ? 'Menyimpan...' : 'Simpan Perubahan'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Modal 3: Rincian Snapshot Tenant -->
{#if showDetailModal && selectedTenant}
	<div class="fixed inset-0 bg-black/70 backdrop-blur-xs z-50 flex items-center justify-center p-4">
		<div class="bg-slate-900 text-white rounded-3xl p-6 sm:p-7 w-full max-w-lg space-y-6 shadow-2xl border border-slate-800 animate-in fade-in zoom-in-95 duration-200 max-h-[90vh] overflow-y-auto">
			<div class="flex items-start justify-between border-b border-slate-800 pb-4">
				<div class="flex items-center gap-3">
					<div class="w-12 h-12 rounded-2xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center font-bold text-lg shrink-0">
						{selectedTenant.name.charAt(0)}
					</div>
					<div>
						<h3 class="text-lg font-black font-['Outfit'] text-white">{selectedTenant.name}</h3>
						<span class="text-xs font-mono text-slate-400">/{selectedTenant.slug}</span>
					</div>
				</div>
				<button
					type="button"
					onclick={() => (showDetailModal = false)}
					class="text-slate-400 hover:text-white p-1 rounded-lg"
				>
					✕
				</button>
			</div>

			<!-- Quick Key Metrics -->
			<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-center text-xs">
				<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
					<span class="text-[10px] text-slate-400 uppercase block font-bold">Total Meja</span>
					<span class="text-lg font-black text-white">{selectedTenant.total_tables}</span>
				</div>
				<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
					<span class="text-[10px] text-slate-400 uppercase block font-bold">Menu Hidangan</span>
					<span class="text-lg font-black text-white">{selectedTenant.total_menus}</span>
				</div>
				<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
					<span class="text-[10px] text-slate-400 uppercase block font-bold">Total Order</span>
					<span class="text-lg font-black text-white">{selectedTenant.total_orders}</span>
				</div>
				<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
					<span class="text-[10px] text-slate-400 uppercase block font-bold">Omset Toko</span>
					<span class="text-sm font-black text-emerald-400 font-['Outfit'] block truncate">
						{formatRupiah(selectedTenant.total_revenue)}
					</span>
				</div>
			</div>

			<!-- Comprehensive Profile List -->
			<div class="divide-y divide-slate-800 text-xs text-slate-300">
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">ID Tenant</span>
					<span class="font-mono text-indigo-400 font-bold">{selectedTenant.id}</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Business Owner</span>
					<span class="font-bold text-white">{selectedTenant.owner_name}</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Email Pemilik</span>
					<span class="font-mono text-white">{selectedTenant.owner_email}</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Nomor Telepon</span>
					<span class="font-mono text-white">{selectedTenant.phone || '-'}</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Paket Layanan</span>
					<span class="font-bold text-amber-400">{selectedTenant.plan || 'PRO'}</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Tarif Pajak & Service</span>
					<span class="font-bold text-white">PB1 {selectedTenant.tax_percent}% | Service {selectedTenant.service_percent}%</span>
				</div>
				<div class="py-2.5 flex justify-between">
					<span class="text-slate-400">Alamat Resto</span>
					<span class="text-white text-right max-w-xs">{selectedTenant.address || '-'}</span>
				</div>
			</div>

			<!-- Quick Simulation Link -->
			<div class="pt-3 border-t border-slate-800 flex items-center justify-between">
				<a
					href={`/order?token=demo-qr-token-table-01`}
					target="_blank"
					class="inline-flex items-center gap-2 text-xs font-bold text-indigo-400 hover:text-indigo-300"
				>
					<QrCode class="w-4 h-4" />
					<span>Buka Simulasi QR Meja</span>
					<ExternalLink class="w-3 h-3" />
				</a>

				<button
					type="button"
					onclick={() => (showDetailModal = false)}
					class="px-5 py-2 bg-slate-800 hover:bg-slate-700 text-white rounded-xl font-bold text-xs"
				>
					Tutup
				</button>
			</div>
		</div>
	</div>
{/if}
