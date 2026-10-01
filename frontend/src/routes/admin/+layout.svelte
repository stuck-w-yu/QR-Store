<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { 
		LayoutDashboard, ShoppingCart, Utensils, QrCode, 
		Users, ChefHat, LogOut, Store, ExternalLink, Menu as MenuIcon, X,
		CreditCard, Monitor, ClipboardList, ShieldCheck
	} from '@lucide/svelte';

	let { children } = $props();
	let sidebarOpen = $state(false);

	const navItems = [
		{ href: '/admin/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/admin/cashier', label: 'Kasir POS', icon: CreditCard, roles: ['OWNER', 'ADMIN', 'CASHIER'] },
		{ href: '/admin/orders', label: 'Pesanan Masuk', icon: ShoppingCart },
		{ href: '/admin/menus', label: 'Menu & Kategori', icon: Utensils },
		{ href: '/admin/tables', label: 'Meja & QR Code', icon: QrCode },
		{ href: '/admin/registers', label: 'Mesin Kasir', icon: Monitor, roles: ['OWNER', 'ADMIN'] },
		{ href: '/admin/shifts', label: 'Laporan Shift', icon: ClipboardList, roles: ['OWNER', 'ADMIN'] },
		{ href: '/admin/users', label: 'Kelola Karyawan', icon: Users, roles: ['OWNER', 'ADMIN'] },
		{ href: '/superadmin', label: 'Superadmin HQ', icon: ShieldCheck, roles: ['SUPERADMIN'] }
	];

	onMount(async () => {
		if (!auth.initialized) {
			await auth.init();
		}
		if (!auth.user) {
			// Auto login as admin for demonstration convenience if not logged in
			try {
				await auth.login('admin@resto.com', 'password123');
			} catch (e) {
				goto('/login');
			}
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-slate-100 flex font-['Plus_Jakarta_Sans',sans-serif]">
	<!-- Desktop Sidebar -->
	<aside class="w-64 bg-slate-900 text-white hidden md:flex flex-col shrink-0 border-r border-slate-800">
		<div class="p-6 border-b border-slate-800 flex items-center gap-3">
			<div class="w-10 h-10 rounded-xl bg-orange-600 flex items-center justify-center shadow-lg shadow-orange-600/30">
				<Store class="w-6 h-6 text-white" />
			</div>
			<div>
				<h2 class="font-bold text-sm leading-tight font-['Outfit']">Resto Nusantara</h2>
				<span class="text-[11px] text-orange-400 font-semibold uppercase tracking-wider">
					{auth.user?.role || 'ADMIN'}
				</span>
			</div>
		</div>

		<nav class="flex-1 p-4 space-y-1 overflow-y-auto text-xs font-semibold">
			{#each navItems as item}
				{#if !item.roles || (auth.user && item.roles.includes(auth.user.role))}
					<a
						href={item.href}
						class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl transition-colors {page.url.pathname === item.href
							? 'bg-orange-600 text-white font-bold shadow-md shadow-orange-600/30'
							: 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'}"
					>
						<item.icon class="w-4 h-4" />
						<span>{item.label}</span>
					</a>
				{/if}
			{/each}

			<div class="pt-4 mt-4 border-t border-slate-800 space-y-1">
				<a
					href="/kitchen/display"
					target="_blank"
					class="flex items-center justify-between px-3.5 py-2.5 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors"
				>
					<div class="flex items-center gap-3">
						<ChefHat class="w-4 h-4 text-amber-400" />
						<span>Kitchen Board</span>
					</div>
					<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
				</a>

				<a
					href="/order?token=demo-qr-token-table-01"
					target="_blank"
					class="flex items-center justify-between px-3.5 py-2.5 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60 transition-colors"
				>
					<div class="flex items-center gap-3">
						<QrCode class="w-4 h-4 text-emerald-400" />
						<span>Simulasi Pelanggan</span>
					</div>
					<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
				</a>
			</div>
		</nav>

		<!-- User Footer -->
		<div class="p-4 border-t border-slate-800 flex items-center justify-between">
			<div class="truncate">
				<div class="text-xs font-bold text-slate-200 truncate">{auth.user?.name || 'Staff User'}</div>
				<div class="text-[10px] text-slate-500 truncate">{auth.user?.email || ''}</div>
			</div>
			<button
				type="button"
				onclick={handleLogout}
				class="text-slate-400 hover:text-red-400 p-1.5 rounded-lg transition-colors"
				title="Keluar"
			>
				<LogOut class="w-4 h-4" />
			</button>
		</div>
	</aside>

	<!-- Main Content Area -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- Mobile Top Nav -->
		<header class="md:hidden bg-slate-900 text-white p-4 flex items-center justify-between border-b border-slate-800">
			<div class="flex items-center gap-2.5">
				<div class="w-8 h-8 rounded-lg bg-orange-600 flex items-center justify-center">
					<Store class="w-5 h-5 text-white" />
				</div>
				<span class="font-bold text-sm font-['Outfit']">Resto Nusantara</span>
			</div>

			<button
				type="button"
				onclick={() => (sidebarOpen = !sidebarOpen)}
				class="p-2 text-slate-400 hover:text-white"
			>
				{#if sidebarOpen}
					<X class="w-6 h-6" />
				{:else}
					<MenuIcon class="w-6 h-6" />
				{/if}
			</button>
		</header>

		<!-- Mobile Navigation Dropdown -->
		{#if sidebarOpen}
			<div class="md:hidden bg-slate-900 border-b border-slate-800 p-4 space-y-2 text-xs font-semibold animate-in fade-in slide-in-from-top-2 duration-150">
				{#each navItems as item}
					{#if !item.roles || (auth.user && item.roles.includes(auth.user.role))}
						<a
							href={item.href}
							onclick={() => (sidebarOpen = false)}
							class="flex items-center gap-3 px-3.5 py-2.5 rounded-xl {page.url.pathname === item.href
								? 'bg-orange-600 text-white font-bold'
								: 'text-slate-400 hover:text-white hover:bg-slate-800/60'}"
						>
							<item.icon class="w-4 h-4" />
							<span>{item.label}</span>
						</a>
					{/if}
				{/each}

				<div class="pt-3 mt-3 border-t border-slate-800 space-y-1">
					<a
						href="/kitchen/display"
						target="_blank"
						onclick={() => (sidebarOpen = false)}
						class="flex items-center justify-between px-3.5 py-2.5 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60"
					>
						<div class="flex items-center gap-3">
							<ChefHat class="w-4 h-4 text-amber-400" />
							<span>Kitchen Board</span>
						</div>
						<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
					</a>

					<a
						href="/order?token=demo-qr-token-table-01"
						target="_blank"
						onclick={() => (sidebarOpen = false)}
						class="flex items-center justify-between px-3.5 py-2.5 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800/60"
					>
						<div class="flex items-center gap-3">
							<QrCode class="w-4 h-4 text-emerald-400" />
							<span>Simulasi Pelanggan</span>
						</div>
						<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
					</a>
				</div>

				<div class="pt-3 mt-3 border-t border-slate-800 flex items-center justify-between px-1">
					<div class="truncate mr-3">
						<div class="text-xs font-bold text-slate-200 truncate">{auth.user?.name || 'Staff User'}</div>
						<div class="text-[10px] text-slate-500 truncate">{auth.user?.email || ''}</div>
					</div>
					<button
						type="button"
						onclick={handleLogout}
						class="flex items-center gap-2 px-3 py-2 rounded-xl text-red-400 hover:bg-red-500/10 font-bold shrink-0"
					>
						<LogOut class="w-4 h-4" />
						<span>Keluar</span>
					</button>
				</div>
			</div>
		{/if}

		<main class="flex-1 p-4 sm:p-6 md:p-8 overflow-y-auto">
			{@render children()}
		</main>
	</div>
</div>
