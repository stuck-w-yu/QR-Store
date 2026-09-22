<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { ChefHat, LogOut, Bell, Clock } from '@lucide/svelte';

	let { children } = $props();
	let currentTime = $state('');

	onMount(() => {
		const updateTime = () => {
			currentTime = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
		};
		updateTime();
		const interval = setInterval(updateTime, 1000);
		return () => clearInterval(interval);
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}
</script>

<div class="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-['Plus_Jakarta_Sans',sans-serif]">
	<!-- Kitchen Top Bar -->
	<header class="bg-slate-900 border-b border-slate-800 px-6 py-3.5 flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="w-10 h-10 rounded-xl bg-orange-600 flex items-center justify-center shadow-lg shadow-orange-600/30">
				<ChefHat class="w-6 h-6 text-white" />
			</div>
			<div>
				<h1 class="text-base font-black tracking-wide font-['Outfit'] text-white">Kitchen Display System (KDS)</h1>
				<span class="text-xs text-orange-400 font-semibold tracking-wider uppercase">Resto Nusantara</span>
			</div>
		</div>

		<div class="flex items-center gap-6">
			<div class="flex items-center gap-2 bg-slate-800/80 px-4 py-1.5 rounded-xl border border-slate-700/60">
				<Clock class="w-4 h-4 text-orange-400" />
				<span class="font-mono font-bold text-sm tracking-wider">{currentTime}</span>
			</div>

			<button
				type="button"
				onclick={handleLogout}
				class="text-xs font-semibold text-slate-400 hover:text-red-400 flex items-center gap-1.5 transition-colors"
			>
				<LogOut class="w-4 h-4" />
				<span>Keluar</span>
			</button>
		</div>
	</header>

	<!-- Main Kitchen Stage -->
	<div class="flex-1 p-6">
		{@render children()}
	</div>
</div>
