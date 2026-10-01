<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { 
		ShieldAlert, Store, LogOut, ArrowLeft, 
		Layers, Users, TrendingUp, Sparkles, ExternalLink,
		Menu as MenuIcon, X
	} from '@lucide/svelte';

	let { children } = $props();
	let mobileMenuOpen = $state(false);

	onMount(async () => {
		if (typeof window !== 'undefined') {
			if (!auth.initialized) {
				await auth.init();
			}
			const token = api.getToken();
			if (!token) {
				goto('/login');
				return;
			}
			// Allow only SUPERADMIN (IT)
			if (auth.user && auth.user.role !== 'SUPERADMIN') {
				goto('/admin/dashboard');
			}
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-['Plus_Jakarta_Sans',sans-serif]">
	<!-- Superadmin Global Navbar -->
	<header class="bg-slate-900/90 backdrop-blur-md border-b border-slate-800/80 sticky top-0 z-40 px-4 sm:px-8 py-3.5">
		<div class="max-w-7xl mx-auto flex items-center justify-between gap-4">
			<!-- Brand & Title -->
			<div class="flex items-center gap-3 min-w-0">
				<div class="w-10 h-10 rounded-2xl bg-linear-to-tr from-indigo-600 via-indigo-500 to-purple-500 flex items-center justify-center shadow-lg shadow-indigo-600/30 shrink-0">
					<Sparkles class="w-5 h-5 text-white" />
				</div>
				<div class="min-w-0">
					<div class="flex items-center gap-2">
						<span class="text-base sm:text-lg font-black tracking-tight font-['Outfit'] text-white">QR-Store</span>
						<span class="px-2 py-0.5 rounded-full text-[10px] font-extrabold uppercase bg-indigo-500/20 text-indigo-400 border border-indigo-500/30">
							SUPERADMIN
						</span>
					</div>
					<span class="text-[11px] text-slate-400 font-medium block -mt-0.5 truncate">
						Pusat Monitoring & Manajemen Tenant Restoran
					</span>
				</div>
			</div>

			<!-- Desktop Actions & Profile -->
			<div class="hidden md:flex items-center gap-4">
				<div class="flex items-center gap-2 bg-slate-800/80 px-3 py-1.5 rounded-xl border border-slate-700/60 text-xs">
					<span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
					<span class="text-slate-300 font-semibold">Engine: Multi-Tenant Active</span>
				</div>

				<a
					href="/admin/dashboard"
					class="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-bold bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors border border-slate-700/60"
				>
					<Store class="w-3.5 h-3.5 text-orange-400" />
					<span>Outlet Admin</span>
				</a>

				<div class="w-px h-6 bg-slate-800"></div>

				<div class="flex items-center gap-3">
					<div class="text-right">
						<div class="text-xs font-bold text-slate-200">{auth.user?.name || 'Super Administrator'}</div>
						<div class="text-[10px] text-indigo-400 font-mono">{auth.user?.email || 'superadmin@qrstore.id'}</div>
					</div>
					<button
						type="button"
						onclick={handleLogout}
						class="p-2 rounded-xl text-slate-400 hover:text-red-400 hover:bg-red-500/10 transition-colors"
						title="Keluar"
					>
						<LogOut class="w-4 h-4" />
					</button>
				</div>
			</div>

			<!-- Mobile Hamburger Toggle -->
			<button
				type="button"
				onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
				class="md:hidden p-2 rounded-xl text-slate-400 hover:text-white bg-slate-800/80 border border-slate-700"
			>
				{#if mobileMenuOpen}
					<X class="w-5 h-5" />
				{:else}
					<MenuIcon class="w-5 h-5" />
				{/if}
			</button>
		</div>

		<!-- Mobile Menu Dropdown -->
		{#if mobileMenuOpen}
			<div class="md:hidden pt-4 pb-2 border-t border-slate-800 mt-3 space-y-3 animate-in fade-in slide-in-from-top-2 duration-150 text-xs font-semibold">
				<div class="flex items-center gap-2 px-2 text-slate-300">
					<span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
					<span>Status Sistem: Multi-Tenant Engine Online</span>
				</div>

				<a
					href="/admin/dashboard"
					onclick={() => (mobileMenuOpen = false)}
					class="w-full flex items-center gap-2.5 px-3 py-2.5 rounded-xl bg-slate-800 text-slate-200"
				>
					<Store class="w-4 h-4 text-orange-400" />
					<span>Buka Portal Resto Nusantara (Outlet Admin)</span>
				</a>

				<div class="pt-3 border-t border-slate-800 flex items-center justify-between px-2">
					<div class="truncate mr-2">
						<div class="font-bold text-slate-200 truncate">{auth.user?.name || 'Super Admin'}</div>
						<div class="text-[10px] text-indigo-400 truncate">{auth.user?.email || 'superadmin@qrstore.id'}</div>
					</div>
					<button
						type="button"
						onclick={handleLogout}
						class="px-3 py-1.5 rounded-xl bg-red-500/10 text-red-400 font-bold flex items-center gap-1.5 shrink-0"
					>
						<LogOut class="w-3.5 h-3.5" />
						<span>Keluar</span>
					</button>
				</div>
			</div>
		{/if}
	</header>

	<!-- Main Body Stage -->
	<main class="flex-1 p-4 sm:p-8 max-w-7xl mx-auto w-full">
		{@render children()}
	</main>
</div>
