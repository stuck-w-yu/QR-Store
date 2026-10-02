<script lang="ts">
	import { onMount } from 'svelte';
	import { api, formatRupiah } from '$lib/api/client';
	import type { 
		TenantSummary, PlatformStats, PlatformOrder, 
		PlatformOwner, SystemHealth 
	} from '$lib/types';
	import { 
		Store, Users, ShoppingCart, DollarSign, TrendingUp, 
		Plus, RefreshCw, Search, Filter, Eye, Edit2, 
		ShieldAlert, CheckCircle2, XCircle, AlertTriangle, 
		ExternalLink, Trash2, ArrowUpRight, ShieldCheck, 
		Lock, Unlock, ChevronRight, QrCode, Phone, Mail, MapPin, Sparkles,
		Activity, Server, Database, Cpu, Layers, CreditCard, Clock, BadgeCheck,
		FileText, ArrowDownRight, UserCheck, MessageSquare, Zap, Check, HelpCircle,
		Download, Radio, Terminal, Laptop
	} from '@lucide/svelte';

	// Tab navigation types
	type TabType = 'overview' | 'tenants' | 'transactions' | 'owners' | 'subscriptions' | 'system';
	let activeTab = $state<TabType>('overview');

	let loading = $state(true);
	let stats = $state<PlatformStats>({
		total_restaurants: 0,
		active_restaurants: 0,
		suspended_restaurants: 0,
		total_owners: 0,
		total_orders: 0,
		total_revenue: 0,
		total_tables: 0,
		total_menus: 0,
		today_revenue: 0,
		today_orders: 0,
		plan_distribution: { PRO: 1 }
	});

	let tenants = $state<TenantSummary[]>([]);
	let recentOrders = $state<PlatformOrder[]>([]);
	let owners = $state<PlatformOwner[]>([]);
	let systemHealth = $state<SystemHealth | null>(null);

	// Filters & search for Tenants
	let searchQuery = $state('');
	let statusFilter = $state('');
	let planFilter = $state('');
	let sortBy = $state<'newest' | 'revenue' | 'orders' | 'name'>('newest');
	let tenantViewMode = $state<'table' | 'grid'>('table');

	// Filters for Transactions
	let orderSearchQuery = $state('');
	let orderStatusFilter = $state('');
	let orderPaymentFilter = $state('');

	// Filters for Owners
	let ownerSearchQuery = $state('');

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

	// Audit Logs (Simulated platform events)
	let auditLogs = $state([
		{ id: '1', action: 'SYSTEM_STARTUP', detail: 'Multi-Tenant Go Engine v1.24 initialized with PostgreSQL Pool', time: 'Hari ini, 00:00', type: 'info' },
		{ id: '2', action: 'TENANT_VERIFIED', detail: 'Resto Nusantara status ACTIVE di paket PRO', time: 'Hari ini, 00:15', type: 'success' },
		{ id: '3', action: 'SECURITY_CHECK', detail: 'Super Administrator authenticated via JWT HMAC-256', time: 'Hari ini, 00:18', type: 'info' }
	]);

	async function loadSuperadminData() {
		try {
			loading = true;
			const [statsRes, tenantsRes, ordersRes, ownersRes, healthRes] = await Promise.all([
				api.get<PlatformStats>('/superadmin/stats'),
				api.get<TenantSummary[]>('/superadmin/restaurants'),
				api.get<PlatformOrder[]>('/superadmin/orders').catch(() => []),
				api.get<PlatformOwner[]>('/superadmin/owners').catch(() => []),
				api.get<SystemHealth>('/superadmin/health').catch(() => null)
			]);
			stats = statsRes;
			tenants = tenantsRes || [];
			recentOrders = ordersRes || [];
			owners = ownersRes || [];
			if (healthRes) {
				systemHealth = healthRes;
			}
		} catch (e: any) {
			console.error('Failed to load superadmin data:', e);
		} finally {
			loading = false;
		}
	}

	function setTab(tab: TabType) {
		activeTab = tab;
		if (typeof window !== 'undefined') {
			const url = new URL(window.location.href);
			url.searchParams.set('tab', tab);
			window.history.replaceState({}, '', url.toString());
		}
	}

	function handleNameChange(name: string) {
		onboardForm.name = name;
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
			const newRestoName = onboardForm.name;
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
			auditLogs.unshift({
				id: String(Date.now()),
				action: 'TENANT_ONBOARDED',
				detail: `Restoran "${newRestoName}" berhasil didaftarkan ke platform`,
				time: 'Baru saja',
				type: 'success'
			});
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
			auditLogs.unshift({
				id: String(Date.now()),
				action: 'TENANT_UPDATED',
				detail: `Konfigurasi toko "${selectedTenant.name}" berhasil diperbarui`,
				time: 'Baru saja',
				type: 'info'
			});
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
			auditLogs.unshift({
				id: String(Date.now()),
				action: newStatus === 'SUSPENDED' ? 'TENANT_SUSPENDED' : 'TENANT_ACTIVATED',
				detail: `Status toko "${t.name}" diubah menjadi ${newStatus}`,
				time: 'Baru saja',
				type: newStatus === 'SUSPENDED' ? 'danger' : 'success'
			});
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
			auditLogs.unshift({
				id: String(Date.now()),
				action: 'TENANT_DELETED',
				detail: `Toko "${t.name}" dihapus permanen dari platform`,
				time: 'Baru saja',
				type: 'danger'
			});
			await loadSuperadminData();
			if (selectedTenant?.id === t.id) {
				showDetailModal = false;
			}
		} catch (err: any) {
			alert(err?.message || 'Gagal menghapus tenant');
		}
	}

	function formatDate(dateStr?: string) {
		if (!dateStr) return '-';
		try {
			const d = new Date(dateStr);
			return d.toLocaleDateString('id-ID', {
				day: 'numeric',
				month: 'short',
				year: 'numeric',
				hour: '2-digit',
				minute: '2-digit'
			});
		} catch {
			return dateStr;
		}
	}

	function formatWAUrl(phoneStr?: string) {
		if (!phoneStr) return '#';
		let cleaned = phoneStr.replace(/[^0-9]/g, '');
		if (cleaned.startsWith('0')) {
			cleaned = '62' + cleaned.slice(1);
		}
		return `https://wa.me/${cleaned}`;
	}

	// Filter and sort computation for Tenants
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

	// Filter for Recent Orders
	let filteredOrders = $derived(
		recentOrders.filter((o) => {
			const matchSearch = 
				o.order_number.toLowerCase().includes(orderSearchQuery.toLowerCase()) ||
				o.restaurant_name.toLowerCase().includes(orderSearchQuery.toLowerCase()) ||
				o.table_name.toLowerCase().includes(orderSearchQuery.toLowerCase());
			const matchStatus = !orderStatusFilter || o.status === orderStatusFilter;
			const matchPayment = !orderPaymentFilter || o.payment_status === orderPaymentFilter;
			return matchSearch && matchStatus && matchPayment;
		})
	);

	// Filter for Business Owners
	let filteredOwners = $derived(
		owners.filter((ow) => {
			return (
				ow.name.toLowerCase().includes(ownerSearchQuery.toLowerCase()) ||
				ow.email.toLowerCase().includes(ownerSearchQuery.toLowerCase()) ||
				ow.restaurant_name.toLowerCase().includes(ownerSearchQuery.toLowerCase()) ||
				ow.phone.toLowerCase().includes(ownerSearchQuery.toLowerCase())
			);
		})
	);

	// Calculations
	let estimatedCommission = $derived(Math.round(stats.total_revenue * 0.02)); // 2% platform commission
	let averageOrderValue = $derived(stats.total_orders > 0 ? Math.round(stats.total_revenue / stats.total_orders) : 0);
	
	// MRR Estimate (Starter: 99k, Pro: 299k, Enterprise: 799k)
	let estimatedMRR = $derived(
		tenants.reduce((acc, t) => {
			if (t.status !== 'ACTIVE') return acc;
			if (t.plan === 'ENTERPRISE') return acc + 799000;
			if (t.plan === 'STARTER') return acc + 99000;
			return acc + 299000; // PRO default
		}, 0)
	);

	onMount(() => {
		if (typeof window !== 'undefined') {
			const urlTab = new URL(window.location.href).searchParams.get('tab') as TabType;
			if (urlTab && ['overview', 'tenants', 'transactions', 'owners', 'subscriptions', 'system'].includes(urlTab)) {
				activeTab = urlTab;
			}
		}
		loadSuperadminData();
	});
</script>

<div class="space-y-6">
	<!-- Page Header & Action Controls -->
	<div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-slate-900/60 p-5 rounded-3xl border border-slate-800/80 shadow-md backdrop-blur-xs">
		<div>
			<div class="flex items-center gap-2">
				<span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-ping"></span>
				<span class="text-xs font-bold text-indigo-400 uppercase tracking-wider">Multi-Tenant Cloud Control</span>
				<span class="text-[10px] px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 border border-slate-700/80">v2.4 LTS</span>
			</div>
			<h1 class="text-2xl sm:text-3xl font-black text-white font-['Outfit'] mt-1">
				Pusat Monitoring & Manajemen Mitra Restoran
			</h1>
			<p class="text-xs text-slate-400 font-medium mt-0.5 max-w-2xl">
				Monitor performa bisnis seluruh outlet, kelola akun pemilik, pantau transaksi nasional live, dan kendalikan langganan SaaS.
			</p>
		</div>

		<div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full lg:w-auto">
			<button
				type="button"
				onclick={loadSuperadminData}
				class="p-2.5 bg-slate-900 hover:bg-slate-800 text-slate-300 rounded-xl border border-slate-800 shadow-xs transition-colors flex items-center justify-center gap-2 text-xs font-bold"
			>
				<RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
				<span class="inline">Segarkan Data</span>
			</button>

			<button
				type="button"
				onclick={() => (showOnboardModal = true)}
				class="flex-1 sm:flex-initial justify-center px-4 py-2.5 bg-linear-to-r from-indigo-600 via-indigo-500 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl shadow-lg shadow-indigo-600/30 transition-all font-bold text-xs flex items-center gap-2"
			>
				<Plus class="w-4 h-4" />
				<span>+ Daftarkan Toko Baru</span>
			</button>
		</div>
	</div>

	<!-- Navigation Menus (Superadmin Core Tabs) -->
	<nav class="flex items-center gap-1.5 p-1.5 bg-slate-900/90 rounded-2xl border border-slate-800/80 overflow-x-auto scrollbar-none shadow-inner">
		<!-- 1. Monitoring / Overview -->
		<button
			type="button"
			onclick={() => setTab('overview')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'overview'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<Activity class="w-4 h-4" />
			<span>Ringkasan & KPI</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'overview' ? 'bg-white/20 text-white' : 'bg-emerald-500/20 text-emerald-400'}">
				Live
			</span>
		</button>

		<!-- 2. Klien & Restoran (Tenants) -->
		<button
			type="button"
			onclick={() => setTab('tenants')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'tenants'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<Store class="w-4 h-4" />
			<span>Klien Restoran</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'tenants' ? 'bg-white/20 text-white' : 'bg-slate-800 text-slate-300'}">
				{stats.total_restaurants}
			</span>
		</button>

		<!-- 3. Transaksi Nasional -->
		<button
			type="button"
			onclick={() => setTab('transactions')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'transactions'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<ShoppingCart class="w-4 h-4" />
			<span>Transaksi Nasional</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'transactions' ? 'bg-white/20 text-white' : 'bg-slate-800 text-slate-300'}">
				{stats.total_orders}
			</span>
		</button>

		<!-- 4. Business Owners -->
		<button
			type="button"
			onclick={() => setTab('owners')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'owners'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<Users class="w-4 h-4" />
			<span>Business Owners</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'owners' ? 'bg-white/20 text-white' : 'bg-slate-800 text-slate-300'}">
				{stats.total_owners}
			</span>
		</button>

		<!-- 5. Paket SaaS & Billing -->
		<button
			type="button"
			onclick={() => setTab('subscriptions')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'subscriptions'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<Layers class="w-4 h-4" />
			<span>Paket SaaS & Billing</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'subscriptions' ? 'bg-white/20 text-white' : 'bg-indigo-500/20 text-indigo-400'}">
				3 Tier
			</span>
		</button>

		<!-- 6. Diagnostik Server & Audit -->
		<button
			type="button"
			onclick={() => setTab('system')}
			class="flex items-center gap-2 px-4 py-2.5 rounded-xl font-bold text-xs whitespace-nowrap transition-all {activeTab === 'system'
				? 'bg-linear-to-r from-indigo-600 to-purple-600 text-white shadow-md shadow-indigo-600/30'
				: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
		>
			<Cpu class="w-4 h-4" />
			<span>Diagnostik & Audit</span>
			<span class="ml-1 px-1.5 py-0.2 rounded-full text-[10px] {activeTab === 'system' ? 'bg-white/20 text-white' : 'bg-emerald-500/20 text-emerald-400'}">
				OK
			</span>
		</button>
	</nav>

	<!-- TAB 1: RINGKASAN & MONITORING (OVERVIEW) -->
	{#if activeTab === 'overview'}
		<div class="space-y-6">
			<!-- Live Health & Service Status Ribbon -->
			<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-900/60 p-3.5 rounded-2xl border border-slate-800/80 text-xs">
				<div class="flex items-center gap-2.5 px-3 py-1.5 bg-slate-950/40 rounded-xl border border-slate-800/60">
					<div class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse shrink-0"></div>
					<div class="min-w-0">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">Backend Engine</span>
						<span class="font-bold text-slate-200 truncate block">Go 1.24 (Online)</span>
					</div>
				</div>

				<div class="flex items-center gap-2.5 px-3 py-1.5 bg-slate-950/40 rounded-xl border border-slate-800/60">
					<Database class="w-3.5 h-3.5 text-indigo-400 shrink-0" />
					<div class="min-w-0">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">Database Pool</span>
						<span class="font-bold text-slate-200 truncate block">PostgreSQL (Connected)</span>
					</div>
				</div>

				<div class="flex items-center gap-2.5 px-3 py-1.5 bg-slate-950/40 rounded-xl border border-slate-800/60">
					<Radio class="w-3.5 h-3.5 text-purple-400 shrink-0" />
					<div class="min-w-0">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">WebSocket Hub</span>
						<span class="font-bold text-slate-200 truncate block">Broadcast Ready</span>
					</div>
				</div>

				<div class="flex items-center gap-2.5 px-3 py-1.5 bg-slate-950/40 rounded-xl border border-slate-800/60">
					<CreditCard class="w-3.5 h-3.5 text-amber-400 shrink-0" />
					<div class="min-w-0">
						<span class="text-[10px] text-slate-400 uppercase font-bold block">Payment Gateway</span>
						<span class="font-bold text-slate-200 truncate block">Midtrans Mock Sandbox</span>
					</div>
				</div>
			</div>

			<!-- Core Platform Metrics Cards (6 Grid) -->
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-3.5">
				<!-- Total Restaurants -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Mitra Resto</span>
						<div class="w-7 h-7 rounded-lg bg-indigo-500/10 text-indigo-400 flex items-center justify-center border border-indigo-500/20">
							<Store class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-2xl font-black text-white font-['Outfit']">{stats.total_restaurants}</div>
						<span class="text-[11px] text-emerald-400 font-bold block mt-0.5">● {stats.active_restaurants} Gerai Aktif</span>
					</div>
				</div>

				<!-- Business Owners -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Business Owners</span>
						<div class="w-7 h-7 rounded-lg bg-purple-500/10 text-purple-400 flex items-center justify-center border border-purple-500/20">
							<Users class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-2xl font-black text-white font-['Outfit']">{stats.total_owners}</div>
						<span class="text-[11px] text-slate-400 font-medium block mt-0.5">Akun Terverifikasi</span>
					</div>
				</div>

				<!-- Orders Count -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Transaksi Nasional</span>
						<div class="w-7 h-7 rounded-lg bg-blue-500/10 text-blue-400 flex items-center justify-center border border-blue-500/20">
							<ShoppingCart class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-2xl font-black text-white font-['Outfit']">{stats.total_orders}</div>
						<span class="text-[11px] text-blue-400 font-medium block mt-0.5">{stats.today_orders || 0} order hari ini</span>
					</div>
				</div>

				<!-- GMV -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Platform GMV</span>
						<div class="w-7 h-7 rounded-lg bg-emerald-500/10 text-emerald-400 flex items-center justify-center border border-emerald-500/20">
							<TrendingUp class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-xl font-black text-emerald-400 font-['Outfit'] truncate">{formatRupiah(stats.total_revenue)}</div>
						<span class="text-[11px] text-slate-400 font-medium block mt-0.5">Volume Transaksi</span>
					</div>
				</div>

				<!-- Total Meja & Menu -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Meja & Katalog</span>
						<div class="w-7 h-7 rounded-lg bg-amber-500/10 text-amber-400 flex items-center justify-center border border-amber-500/20">
							<QrCode class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-2xl font-black text-white font-['Outfit']">{stats.total_tables || 5} <span class="text-xs font-normal text-slate-400">meja</span></div>
						<span class="text-[11px] text-slate-400 font-medium block mt-0.5">{stats.total_menus || 6} menu katalog</span>
					</div>
				</div>

				<!-- Estimated MRR / Commission -->
				<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 shadow-xs space-y-2 hover:border-slate-700 transition-colors">
					<div class="flex items-center justify-between">
						<span class="text-[11px] font-bold text-slate-400 uppercase tracking-wider">Est. MRR SaaS</span>
						<div class="w-7 h-7 rounded-lg bg-pink-500/10 text-pink-400 flex items-center justify-center border border-pink-500/20">
							<DollarSign class="w-4 h-4" />
						</div>
					</div>
					<div>
						<div class="text-xl font-black text-pink-400 font-['Outfit'] truncate">{formatRupiah(estimatedMRR)}</div>
						<span class="text-[11px] text-slate-400 font-medium block mt-0.5">Berdasarkan paket tier</span>
					</div>
				</div>
			</div>

			<!-- Analytical Deep Dive Cards (2 Columns) -->
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
				<!-- Card 1: Subscription Tier Distribution & Quotas -->
				<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2.5">
							<div class="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center border border-indigo-500/20">
								<Layers class="w-4 h-4" />
							</div>
							<div>
								<h3 class="font-extrabold text-sm text-white font-['Outfit']">Sebaran Paket Langganan Klien</h3>
								<span class="text-[11px] text-slate-400">Distribusi paket aktif mitra restoran</span>
							</div>
						</div>
						<button 
							type="button" 
							onclick={() => setTab('subscriptions')}
							class="text-xs text-indigo-400 hover:text-indigo-300 font-bold flex items-center gap-1"
						>
							<span>Kelola Tier</span>
							<ChevronRight class="w-3.5 h-3.5" />
						</button>
					</div>

					<div class="space-y-3 pt-2">
						<!-- PRO Plan -->
						<div>
							<div class="flex justify-between text-xs font-semibold mb-1">
								<span class="text-indigo-400">PRO Plan (Rp 299k/bln)</span>
								<span class="text-white font-bold">{stats.plan_distribution?.PRO || 1} Gerai (100%)</span>
							</div>
							<div class="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden border border-slate-800">
								<div class="bg-linear-to-r from-indigo-500 to-purple-500 h-full rounded-full" style="width: 100%"></div>
							</div>
						</div>

						<!-- STARTER Plan -->
						<div>
							<div class="flex justify-between text-xs font-semibold mb-1">
								<span class="text-slate-400">STARTER Plan (Rp 99k/bln)</span>
								<span class="text-slate-500 font-bold">{stats.plan_distribution?.STARTER || 0} Gerai (0%)</span>
							</div>
							<div class="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden border border-slate-800">
								<div class="bg-blue-500 h-full rounded-full" style="width: 0%"></div>
							</div>
						</div>

						<!-- ENTERPRISE Plan -->
						<div>
							<div class="flex justify-between text-xs font-semibold mb-1">
								<span class="text-amber-400">ENTERPRISE Plan (Rp 799k/bln)</span>
								<span class="text-slate-500 font-bold">{stats.plan_distribution?.ENTERPRISE || 0} Gerai (0%)</span>
							</div>
							<div class="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden border border-slate-800">
								<div class="bg-amber-500 h-full rounded-full" style="width: 0%"></div>
							</div>
						</div>
					</div>

					<div class="pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
						<span>Rata-rata Nilai Order (AOV): <strong class="text-emerald-400">{formatRupiah(averageOrderValue)}</strong></span>
						<span>Komisi Take-Rate: <strong class="text-indigo-400">~2.0%</strong></span>
					</div>
				</div>

				<!-- Card 2: Financial Snapshot & Take-rate -->
				<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2.5">
							<div class="w-8 h-8 rounded-xl bg-emerald-500/10 text-emerald-400 flex items-center justify-center border border-emerald-500/20">
								<DollarSign class="w-4 h-4" />
							</div>
							<div>
								<h3 class="font-extrabold text-sm text-white font-['Outfit']">Potensi Revenue & Settlement</h3>
								<span class="text-[11px] text-slate-400">Kalkulasi bagi hasil & pendapatan SaaS</span>
							</div>
						</div>
						<button 
							type="button" 
							onclick={() => setTab('transactions')}
							class="text-xs text-indigo-400 hover:text-indigo-300 font-bold flex items-center gap-1"
						>
							<span>Feed Transaksi</span>
							<ChevronRight class="w-3.5 h-3.5" />
						</button>
					</div>

					<div class="grid grid-cols-2 gap-3 pt-1">
						<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
							<span class="text-[10px] text-slate-400 uppercase font-bold block">Gross Platform GMV</span>
							<span class="text-base font-extrabold text-white font-['Outfit'] mt-0.5 block truncate">
								{formatRupiah(stats.total_revenue)}
							</span>
							<span class="text-[10px] text-slate-500 block mt-1">Total seluruh restoran</span>
						</div>

						<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
							<span class="text-[10px] text-indigo-400 uppercase font-bold block">Estimasi Fee Platform</span>
							<span class="text-base font-extrabold text-emerald-400 font-['Outfit'] mt-0.5 block truncate">
								{formatRupiah(estimatedCommission)}
							</span>
							<span class="text-[10px] text-slate-500 block mt-1">Take-rate transaksi 2%</span>
						</div>

						<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
							<span class="text-[10px] text-slate-400 uppercase font-bold block">Proyeksi ARR (Annual)</span>
							<span class="text-base font-extrabold text-purple-400 font-['Outfit'] mt-0.5 block truncate">
								{formatRupiah(estimatedMRR * 12)}
							</span>
							<span class="text-[10px] text-slate-500 block mt-1">12 x Monthly MRR</span>
						</div>

						<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800">
							<span class="text-[10px] text-slate-400 uppercase font-bold block">Tingkat Retensi Mitra</span>
							<span class="text-base font-extrabold text-emerald-400 font-['Outfit'] mt-0.5 block">
								100.0%
							</span>
							<span class="text-[10px] text-slate-500 block mt-1">0 Restoran Churned</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Live Activity Feed & Quick Roster Preview -->
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-5">
				<!-- Recent Orders Feed -->
				<div class="lg:col-span-2 bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<Clock class="w-4 h-4 text-indigo-400" />
							<h3 class="font-bold text-sm text-white">Aktivitas Transaksi Terbaru</h3>
						</div>
						<button 
							type="button" 
							onclick={() => setTab('transactions')}
							class="text-xs text-indigo-400 hover:underline font-bold"
						>
							Lihat Semua ({recentOrders.length})
						</button>
					</div>

					<div class="divide-y divide-slate-800/80">
						{#if recentOrders.length === 0}
							<div class="py-8 text-center text-slate-500 text-xs">Belum ada transaksi tercatat.</div>
						{:else}
							{#each recentOrders.slice(0, 5) as o}
								<div class="py-2.5 flex items-center justify-between gap-3 text-xs">
									<div class="min-w-0">
										<div class="flex items-center gap-2">
											<span class="font-mono font-bold text-white">{o.order_number}</span>
											<span class="text-[10px] px-2 py-0.2 rounded-md bg-slate-800 text-indigo-300 font-bold">
												{o.restaurant_name}
											</span>
											<span class="text-slate-500 text-[11px]">• {o.table_name}</span>
										</div>
										<span class="text-[10px] text-slate-500 block mt-0.5">{formatDate(o.created_at)}</span>
									</div>

									<div class="text-right shrink-0">
										<span class="font-extrabold text-emerald-400 font-['Outfit']">{formatRupiah(o.total)}</span>
										<div class="flex items-center justify-end gap-1.5 mt-0.5">
											<span class="px-1.5 py-0.2 rounded text-[9px] font-bold {o.status === 'COMPLETED' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-amber-500/10 text-amber-400'}">
												{o.status}
											</span>
										</div>
									</div>
								</div>
							{/each}
						{/if}
					</div>
				</div>

				<!-- Platform Audit Trail Ticker -->
				<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-2">
							<ShieldAlert class="w-4 h-4 text-purple-400" />
							<h3 class="font-bold text-sm text-white">Audit Trail Platform</h3>
						</div>
						<span class="text-[10px] text-slate-500 uppercase font-bold">Log Realtime</span>
					</div>

					<div class="space-y-3">
						{#each auditLogs as log}
							<div class="p-3 bg-slate-950/60 rounded-xl border border-slate-800/80 text-xs space-y-1">
								<div class="flex items-center justify-between">
									<span class="font-bold text-[10px] uppercase {log.type === 'danger' ? 'text-red-400' : log.type === 'success' ? 'text-emerald-400' : 'text-indigo-400'}">
										{log.action}
									</span>
									<span class="text-[10px] text-slate-500">{log.time}</span>
								</div>
								<p class="text-slate-300 text-[11px]">{log.detail}</p>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- TAB 2: MANAJEMEN KLIEN RESTORAN (TENANTS) -->
	{#if activeTab === 'tenants'}
		<div class="space-y-6">
			<!-- Filter & Search Toolbar -->
			<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800/80 flex flex-col md:flex-row md:items-center justify-between gap-3">
				<!-- Search -->
				<div class="relative flex-1">
					<Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
					<input
						type="text"
						bind:value={searchQuery}
						placeholder="Cari nama restoran, slug URL, pemilik, atau email..."
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

					<!-- View Mode Switch -->
					<div class="flex items-center bg-slate-950/60 rounded-xl p-0.5 border border-slate-800">
						<button
							type="button"
							onclick={() => (tenantViewMode = 'table')}
							class="px-2.5 py-1.5 rounded-lg text-xs font-bold transition-colors {tenantViewMode === 'table' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}"
							title="Tampilan Tabel"
						>
							Tabel
						</button>
						<button
							type="button"
							onclick={() => (tenantViewMode = 'grid')}
							class="px-2.5 py-1.5 rounded-lg text-xs font-bold transition-colors {tenantViewMode === 'grid' ? 'bg-indigo-600 text-white' : 'text-slate-400 hover:text-white'}"
							title="Tampilan Kartu"
						>
							Kartu
						</button>
					</div>
				</div>
			</div>

			<!-- Tenants Listing -->
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
			{:else if tenantViewMode === 'grid'}
				<!-- Grid View -->
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each filteredTenants as t}
						<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800/80 shadow-md space-y-4 hover:border-slate-700 transition-colors">
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
									<span class="font-mono text-slate-300 truncate max-w-44">{t.owner_email}</span>
								</div>
								{#if t.phone}
									<div class="flex items-center justify-between">
										<span class="text-slate-400 font-medium">Kontak:</span>
										<a href={formatWAUrl(t.phone)} target="_blank" class="font-mono text-emerald-400 hover:underline flex items-center gap-1">
											<span>{t.phone}</span>
											<MessageSquare class="w-3 h-3" />
										</a>
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
			{:else}
				<!-- Table View -->
				<div class="bg-slate-900 rounded-3xl border border-slate-800/80 shadow-md overflow-hidden">
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
												<a href={formatWAUrl(t.phone)} target="_blank" class="text-[10px] text-emerald-400 font-mono hover:underline inline-flex items-center gap-1">
													<span>{t.phone}</span>
													<MessageSquare class="w-2.5 h-2.5" />
												</a>
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
	{/if}

	<!-- TAB 3: MONITORING TRANSAKSI NASIONAL -->
	{#if activeTab === 'transactions'}
		<div class="space-y-6">
			<!-- Header & Transaction Stats Bar -->
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
				<div class="bg-slate-900 rounded-2xl p-4 border border-slate-800">
					<span class="text-xs font-bold text-slate-400 uppercase">Total Pesanan Tercatat</span>
					<div class="text-2xl font-black text-white font-['Outfit'] mt-1">{stats.total_orders} Pesanan</div>
					<span class="text-[11px] text-blue-400">Seluruh outlet aktif</span>
				</div>

				<div class="bg-slate-900 rounded-2xl p-4 border border-slate-800">
					<span class="text-xs font-bold text-slate-400 uppercase">Gross Merchandise Value (GMV)</span>
					<div class="text-2xl font-black text-emerald-400 font-['Outfit'] mt-1">{formatRupiah(stats.total_revenue)}</div>
					<span class="text-[11px] text-slate-400">Volume transaksi berhasil</span>
				</div>

				<div class="bg-slate-900 rounded-2xl p-4 border border-slate-800">
					<span class="text-xs font-bold text-slate-400 uppercase">Rata-Rata per Order (AOV)</span>
					<div class="text-2xl font-black text-purple-400 font-['Outfit'] mt-1">{formatRupiah(averageOrderValue)}</div>
					<span class="text-[11px] text-slate-400">Basket size konsumen</span>
				</div>
			</div>

			<!-- Filter Bar -->
			<div class="bg-slate-900/80 rounded-2xl p-4 border border-slate-800 flex flex-col md:flex-row md:items-center justify-between gap-3">
				<div class="relative flex-1">
					<Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
					<input
						type="text"
						bind:value={orderSearchQuery}
						placeholder="Cari nomor order (ORD-xxx), nama restoran, meja..."
						class="w-full text-xs pl-10 pr-4 py-2.5 bg-slate-950/60 rounded-xl border border-slate-800 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
					/>
				</div>

				<div class="flex items-center gap-2">
					<select
						bind:value={orderStatusFilter}
						class="text-xs font-semibold px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800 text-slate-200 focus:outline-none focus:border-indigo-500"
					>
						<option value="">Semua Status Pesanan</option>
						<option value="WAITING_PAYMENT">Menunggu Pembayaran</option>
						<option value="CONFIRMED">Dikonfirmasi</option>
						<option value="PREPARING">Dimasak</option>
						<option value="READY">Siap Saji</option>
						<option value="COMPLETED">Selesai</option>
						<option value="CANCELLED">Dibatalkan</option>
					</select>

					<select
						bind:value={orderPaymentFilter}
						class="text-xs font-semibold px-3 py-2 rounded-xl bg-slate-950/60 border border-slate-800 text-slate-200 focus:outline-none focus:border-indigo-500"
					>
						<option value="">Semua Status Bayar</option>
						<option value="PAID">PAID (Lunas)</option>
						<option value="UNPAID">UNPAID (Belum Bayar)</option>
					</select>
				</div>
			</div>

			<!-- Orders Table -->
			{#if filteredOrders.length === 0}
				<div class="bg-slate-900/60 rounded-3xl p-14 text-center border border-slate-800 text-slate-400 text-xs font-medium space-y-3">
					<ShoppingCart class="w-10 h-10 mx-auto text-slate-600" />
					<p>Tidak ada transaksi yang sesuai dengan filter.</p>
				</div>
			{:else}
				<div class="bg-slate-900 rounded-3xl border border-slate-800 shadow-md overflow-hidden">
					<div class="overflow-x-auto">
						<table class="w-full text-left text-xs border-collapse">
							<thead>
								<tr class="bg-slate-950/60 border-b border-slate-800 text-slate-400 font-bold uppercase tracking-wider text-[11px]">
									<th class="p-4 pl-6">No. Pesanan</th>
									<th class="p-4">Restoran Mitra</th>
									<th class="p-4">Meja</th>
									<th class="p-4">Total Tagihan</th>
									<th class="p-4">Status Pesanan</th>
									<th class="p-4">Pembayaran</th>
									<th class="p-4 pr-6 text-right">Waktu</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-slate-800/60 font-medium text-slate-300">
								{#each filteredOrders as o}
									<tr class="hover:bg-slate-800/30 transition-colors">
										<td class="p-4 pl-6 font-mono font-bold text-white">
											{o.order_number}
										</td>
										<td class="p-4">
											<span class="font-bold text-slate-200 block">{o.restaurant_name}</span>
											<span class="text-[10px] text-slate-500 font-mono">/{o.restaurant_slug}</span>
										</td>
										<td class="p-4 font-bold text-slate-300">
											{o.table_name}
										</td>
										<td class="p-4 font-extrabold text-emerald-400 font-['Outfit']">
											{formatRupiah(o.total)}
										</td>
										<td class="p-4">
											<span class="px-2 py-0.5 rounded-full text-[10px] font-bold {o.status === 'COMPLETED' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : o.status === 'CONFIRMED' ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'}">
												{o.status}
											</span>
										</td>
										<td class="p-4">
											<div class="flex items-center gap-1.5">
												<span class="px-1.5 py-0.2 rounded text-[10px] font-mono {o.payment_status === 'PAID' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-red-500/20 text-red-300'}">
													{o.payment_status}
												</span>
												<span class="text-[10px] text-slate-400 uppercase">{o.payment_method || '-'}</span>
											</div>
										</td>
										<td class="p-4 pr-6 text-right font-mono text-[11px] text-slate-400">
											{formatDate(o.created_at)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<!-- TAB 4: DATABASE BUSINESS OWNERS -->
	{#if activeTab === 'owners'}
		<div class="space-y-6">
			<!-- Header & Search -->
			<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-900/80 p-4 rounded-2xl border border-slate-800">
				<div class="relative flex-1">
					<Search class="w-4 h-4 text-slate-400 absolute left-3.5 top-3" />
					<input
						type="text"
						bind:value={ownerSearchQuery}
						placeholder="Cari nama pemilik usaha, email login, restoran, atau no. telepon..."
						class="w-full text-xs pl-10 pr-4 py-2 bg-slate-950/60 rounded-xl border border-slate-800 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
					/>
				</div>
				<div class="text-xs text-slate-400 font-semibold px-2">
					Total: <strong class="text-white">{filteredOwners.length}</strong> Akun Pemilik
				</div>
			</div>

			<!-- Owners Grid Cards -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each filteredOwners as ow}
					<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4 hover:border-slate-700 transition-colors">
						<div class="flex items-start justify-between gap-3">
							<div class="flex items-center gap-3">
								<div class="w-11 h-11 rounded-2xl bg-purple-500/10 text-purple-400 border border-purple-500/20 flex items-center justify-center font-bold text-base shrink-0">
									{ow.name.charAt(0)}
								</div>
								<div>
									<h3 class="font-extrabold text-sm text-white">{ow.name}</h3>
									<span class="text-[11px] text-indigo-400 font-mono block">{ow.email}</span>
								</div>
							</div>
							<span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
								{ow.status}
							</span>
						</div>

						<div class="p-3 bg-slate-950/60 rounded-2xl border border-slate-800 text-xs space-y-1.5">
							<div class="flex justify-between items-center">
								<span class="text-slate-400">Restoran Terafiliasi:</span>
								<span class="font-bold text-white">{ow.restaurant_name}</span>
							</div>
							<div class="flex justify-between items-center">
								<span class="text-slate-400">Paket Langganan:</span>
								<span class="font-bold text-amber-400">{ow.restaurant_plan}</span>
							</div>
							<div class="flex justify-between items-center">
								<span class="text-slate-400">Tanggal Terdaftar:</span>
								<span class="font-mono text-slate-300 text-[11px]">{formatDate(ow.created_at)}</span>
							</div>
						</div>

						<!-- Direct Communication Actions -->
						<div class="flex items-center gap-2 pt-2 border-t border-slate-800/80">
							{#if ow.phone && ow.phone !== '-'}
								<a
									href={formatWAUrl(ow.phone)}
									target="_blank"
									class="flex-1 py-2 px-3 rounded-xl bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 font-bold text-xs flex items-center justify-center gap-1.5 transition-colors"
								>
									<MessageSquare class="w-3.5 h-3.5" />
									<span>Chat WhatsApp</span>
								</a>
							{/if}

							<a
								href={`mailto:${ow.email}`}
								class="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
								title="Kirim Email"
							>
								<Mail class="w-4 h-4" />
							</a>

							<a
								href={`/public/restaurants/${ow.restaurant_slug}`}
								target="_blank"
								class="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors"
								title="Buka Toko Klien"
							>
								<ExternalLink class="w-4 h-4" />
							</a>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- TAB 5: PAKET SAAS & BILLING (SUBSCRIPTIONS) -->
	{#if activeTab === 'subscriptions'}
		<div class="space-y-6">
			<!-- Tier Comparison Cards -->
			<div class="grid grid-cols-1 md:grid-cols-3 gap-5">
				<!-- STARTER -->
				<div class="bg-slate-900 rounded-3xl p-6 border border-slate-800 space-y-4 relative overflow-hidden">
					<div class="flex justify-between items-start">
						<div>
							<span class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase bg-slate-800 text-slate-400 border border-slate-700">
								TIER 1
							</span>
							<h3 class="text-xl font-black text-white font-['Outfit'] mt-2">STARTER</h3>
							<p class="text-xs text-slate-400 mt-0.5">Untuk cafe kecil & gerai pemula</p>
						</div>
						<div class="text-right">
							<span class="text-xl font-black text-white font-['Outfit']">Rp 99.000</span>
							<span class="text-[10px] text-slate-500 block">/ bulan</span>
						</div>
					</div>

					<div class="pt-2 border-t border-slate-800 space-y-2 text-xs text-slate-300">
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Maksimal 10 Meja QR</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>1 Mesin Kasir POS</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Katalog Menu Tanpa Batas</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Laporan Penjualan Standar</span>
						</div>
					</div>

					<div class="pt-3">
						<span class="text-[11px] font-bold text-slate-400">
							Terdaftar: <strong class="text-white">{stats.plan_distribution?.STARTER || 0} Restoran</strong>
						</span>
					</div>
				</div>

				<!-- PRO (RECOMMENDED) -->
				<div class="bg-linear-to-b from-indigo-950/60 to-slate-900 rounded-3xl p-6 border-2 border-indigo-500/60 shadow-xl shadow-indigo-600/10 space-y-4 relative overflow-hidden">
					<div class="absolute top-3 right-3 px-2 py-0.5 rounded-full text-[9px] font-black bg-indigo-500 text-white uppercase tracking-wider">
						POPULAR TIER
					</div>

					<div>
						<span class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
							TIER 2 (RECOMMENDED)
						</span>
						<h3 class="text-xl font-black text-white font-['Outfit'] mt-2">PRO RESTO</h3>
						<p class="text-xs text-slate-400 mt-0.5">Untuk restoran dine-in dengan dapur & kasir</p>
					</div>

					<div class="text-left">
						<span class="text-2xl font-black text-indigo-400 font-['Outfit']">Rp 299.000</span>
						<span class="text-[10px] text-slate-400 inline"> / bulan</span>
					</div>

					<div class="pt-2 border-t border-indigo-500/20 space-y-2 text-xs text-slate-200">
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span class="font-bold text-white">Unlimited Meja & QR Code</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span class="font-bold text-white">Kitchen Display System (KDS Live)</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Kasir Multi-Staff & Shift Cashier</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Dynamic QRIS Realtime Instant Payment</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Pajak Restoran PB1 & Service Charge Config</span>
						</div>
					</div>

					<div class="pt-3">
						<span class="text-[11px] font-bold text-indigo-300">
							Terdaftar: <strong class="text-white">{stats.plan_distribution?.PRO || 1} Restoran (Utama)</strong>
						</span>
					</div>
				</div>

				<!-- ENTERPRISE -->
				<div class="bg-slate-900 rounded-3xl p-6 border border-slate-800 space-y-4 relative overflow-hidden">
					<div class="flex justify-between items-start">
						<div>
							<span class="px-2 py-0.5 rounded-md text-[10px] font-extrabold uppercase bg-amber-500/20 text-amber-300 border border-amber-500/30">
								TIER 3
							</span>
							<h3 class="text-xl font-black text-white font-['Outfit'] mt-2">ENTERPRISE</h3>
							<p class="text-xs text-slate-400 mt-0.5">Waralaba besar & jaringan multi-cabang</p>
						</div>
						<div class="text-right">
							<span class="text-xl font-black text-amber-400 font-['Outfit']">Rp 799.000</span>
							<span class="text-[10px] text-slate-500 block">/ bulan</span>
						</div>
					</div>

					<div class="pt-2 border-t border-slate-800 space-y-2 text-xs text-slate-300">
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span class="font-bold text-white">Semua Fitur PRO</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Multi-Branch Management HQ</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Dedicated Database Connection</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>Custom Domain & Brand Whitelabel</span>
						</div>
						<div class="flex items-center gap-2">
							<Check class="w-3.5 h-3.5 text-emerald-400" />
							<span>SLA 99.9% & 24/7 Priority Support</span>
						</div>
					</div>

					<div class="pt-3">
						<span class="text-[11px] font-bold text-slate-400">
							Terdaftar: <strong class="text-white">{stats.plan_distribution?.ENTERPRISE || 0} Restoran</strong>
						</span>
					</div>
				</div>
			</div>

			<!-- Subscription Client Roster -->
			<div class="bg-slate-900 rounded-3xl p-5 border border-slate-800 space-y-4">
				<h3 class="font-bold text-sm text-white">Daftar Paket Mitra Aktif</h3>
				<div class="divide-y divide-slate-800/80">
					{#each tenants as t}
						<div class="py-3 flex items-center justify-between gap-3 text-xs">
							<div class="flex items-center gap-3">
								<div class="w-9 h-9 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center font-bold">
									{t.name.charAt(0)}
								</div>
								<div>
									<span class="font-bold text-white block">{t.name}</span>
									<span class="text-[10px] text-slate-400 font-mono">{t.owner_email}</span>
								</div>
							</div>

							<div class="flex items-center gap-3">
								<span class="px-2.5 py-1 rounded-lg text-[10px] font-extrabold uppercase {t.plan === 'ENTERPRISE' ? 'bg-amber-500/20 text-amber-300' : 'bg-indigo-500/20 text-indigo-300'}">
									{t.plan || 'PRO'}
								</span>
								<button
									type="button"
									onclick={() => openEditModal(t)}
									class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-bold transition-colors"
								>
									Ubah Paket
								</button>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{/if}

	<!-- TAB 6: DIAGNOSTIK SISTEM & AUDIT LOG -->
	{#if activeTab === 'system'}
		<div class="space-y-6">
			<!-- Diagnostics Cards Grid -->
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				<!-- Backend Engine -->
				<div class="bg-slate-900 rounded-2xl p-5 border border-slate-800 space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-slate-400 uppercase">Backend Server</span>
						<Server class="w-4 h-4 text-emerald-400" />
					</div>
					<div class="text-xl font-black text-white font-['Outfit']">Go Engine v1.24</div>
					<div class="text-[11px] text-slate-400 font-mono">
						Goroutines: <strong class="text-indigo-400">{systemHealth?.goroutines || 7}</strong>
					</div>
					<div class="text-[11px] text-slate-400 font-mono">
						Memory Alloc: <strong class="text-purple-400">{(systemHealth?.memory_alloc_mb || 1.8).toFixed(2)} MB</strong>
					</div>
				</div>

				<!-- Database Pool -->
				<div class="bg-slate-900 rounded-2xl p-5 border border-slate-800 space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-slate-400 uppercase">PostgreSQL Pool</span>
						<Database class="w-4 h-4 text-indigo-400" />
					</div>
					<div class="text-xl font-black text-emerald-400 font-['Outfit']">Connected</div>
					<div class="text-[11px] text-slate-400 font-mono">
						Pool Total Conns: <strong class="text-white">{systemHealth?.pool_total_conns || 6}</strong>
					</div>
					<div class="text-[11px] text-slate-400 font-mono">
						Pool Idle: <strong class="text-emerald-400">{systemHealth?.pool_idle_conns || 6}</strong>
					</div>
				</div>

				<!-- WebSocket Hub -->
				<div class="bg-slate-900 rounded-2xl p-5 border border-slate-800 space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-slate-400 uppercase">WebSocket Hub</span>
						<Radio class="w-4 h-4 text-purple-400" />
					</div>
					<div class="text-xl font-black text-white font-['Outfit']">Broadcast Hub</div>
					<div class="text-[11px] text-emerald-400 font-bold">
						● Live POS & Kitchen Sync
					</div>
					<div class="text-[11px] text-slate-400">
						Latency: <strong class="text-indigo-400">&lt; 15 ms</strong>
					</div>
				</div>

				<!-- Security Status -->
				<div class="bg-slate-900 rounded-2xl p-5 border border-slate-800 space-y-2">
					<div class="flex items-center justify-between">
						<span class="text-xs font-bold text-slate-400 uppercase">Security & Auth</span>
						<ShieldCheck class="w-4 h-4 text-emerald-400" />
					</div>
					<div class="text-xl font-black text-white font-['Outfit']">RBAC Active</div>
					<div class="text-[11px] text-slate-400 font-mono">
						SUPERADMIN Role Validated
					</div>
					<div class="text-[11px] text-slate-400">
						JWT Token: <strong class="text-emerald-400">Secured</strong>
					</div>
				</div>
			</div>

			<!-- System Audit Logs Complete -->
			<div class="bg-slate-900 rounded-3xl p-6 border border-slate-800 space-y-4">
				<div class="flex items-center justify-between">
					<div>
						<h3 class="font-extrabold text-sm text-white font-['Outfit']">Audit Trail & Riwayat Operasional</h3>
						<p class="text-xs text-slate-400">Pencatatan tindakan administratif dan security event</p>
					</div>
					<span class="text-[11px] font-mono text-indigo-400">Log Buffer: Active</span>
				</div>

				<div class="space-y-2.5">
					{#each auditLogs as log}
						<div class="p-3.5 bg-slate-950/60 rounded-2xl border border-slate-800 flex items-start justify-between gap-4 text-xs">
							<div class="flex items-start gap-3">
								<div class="w-8 h-8 rounded-xl flex items-center justify-center shrink-0 {log.type === 'danger' ? 'bg-red-500/10 text-red-400 border border-red-500/20' : log.type === 'success' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20'}">
									<Terminal class="w-4 h-4" />
								</div>
								<div>
									<div class="flex items-center gap-2">
										<span class="font-bold text-white">{log.action}</span>
										<span class="px-2 py-0.2 rounded-md text-[9px] font-mono uppercase {log.type === 'danger' ? 'bg-red-500/20 text-red-300' : log.type === 'success' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-indigo-500/20 text-indigo-300'}">
											{log.type}
										</span>
									</div>
									<p class="text-slate-300 text-xs mt-0.5">{log.detail}</p>
								</div>
							</div>
							<span class="font-mono text-[10px] text-slate-500 shrink-0">{log.time}</span>
						</div>
					{/each}
				</div>
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
					{#if selectedTenant.phone}
						<a href={formatWAUrl(selectedTenant.phone)} target="_blank" class="font-mono text-emerald-400 hover:underline flex items-center gap-1">
							<span>{selectedTenant.phone}</span>
							<MessageSquare class="w-3 h-3" />
						</a>
					{:else}
						<span class="text-slate-500">-</span>
					{/if}
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
					<span>Buka Menu Meja Pelanggan</span>
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
