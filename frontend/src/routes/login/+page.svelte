<script lang="ts">
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { Store, Lock, Mail, ArrowRight, ShieldCheck } from '@lucide/svelte';

	let email = $state('admin@resto.com');
	let password = $state('password123');
	let error = $state<string | null>(null);
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		try {
			loading = true;
			error = null;
			const user = await auth.login(email, password);
			if (user.role === 'KITCHEN') {
				goto('/kitchen/display');
			} else {
				goto('/admin/dashboard');
			}
		} catch (err: any) {
			error = err?.message || 'Login gagal. Periksa kembali email dan password.';
		} finally {
			loading = false;
		}
	}

	function quickSelectRole(roleEmail: string) {
		email = roleEmail;
		password = 'password123';
	}
</script>

<div class="min-h-screen bg-slate-100 flex items-center justify-center p-4">
	<div class="bg-white w-full max-w-md rounded-3xl p-8 shadow-xl border border-slate-200/80 space-y-6">
		<div class="text-center space-y-2">
			<div class="w-14 h-14 bg-linear-to-tr from-orange-600 to-amber-500 rounded-2xl flex items-center justify-center mx-auto shadow-lg shadow-orange-500/30 text-white">
				<Store class="w-8 h-8" />
			</div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Staff Portal</h1>
			<p class="text-xs text-slate-500 font-medium">Masuk untuk mengelola operasional Resto Nusantara</p>
		</div>

		{#if error}
			<div class="p-3.5 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs font-medium text-center">
				{error}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="email" class="block text-xs font-bold text-slate-700 mb-1.5">Email Karyawan</label>
				<div class="relative">
					<Mail class="w-4 h-4 text-slate-400 absolute left-3.5 top-3.5" />
					<input
						id="email"
						type="email"
						bind:value={email}
						required
						class="w-full text-xs pl-10 pr-3.5 py-3 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 font-medium"
					/>
				</div>
			</div>

			<div>
				<label for="password" class="block text-xs font-bold text-slate-700 mb-1.5">Kata Sandi</label>
				<div class="relative">
					<Lock class="w-4 h-4 text-slate-400 absolute left-3.5 top-3.5" />
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						class="w-full text-xs pl-10 pr-3.5 py-3 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 font-medium"
					/>
				</div>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="w-full bg-orange-600 hover:bg-orange-700 text-white font-bold py-3.5 px-4 rounded-xl shadow-lg shadow-orange-500/30 text-xs flex items-center justify-center gap-2 transition-all active:scale-[0.98] disabled:opacity-50"
			>
				{#if loading}
					<div class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
					<span>Memverifikasi...</span>
				{:else}
					<span>Masuk Dashboard</span>
					<ArrowRight class="w-4 h-4" />
				{/if}
			</button>
		</form>

		<!-- Quick Demo Role Switcher -->
		<div class="pt-4 border-t border-slate-100">
			<span class="block text-[11px] font-bold text-slate-400 uppercase tracking-wider mb-2.5 text-center">
				Akun Demo Siap Pakai
			</span>
			<div class="grid grid-cols-2 gap-2 text-xs">
				<button
					type="button"
					onclick={() => quickSelectRole('owner@resto.com')}
					class="p-2 rounded-xl border border-slate-200 hover:border-orange-500 text-left transition-colors font-medium bg-slate-50"
				>
					<div class="font-bold text-slate-900">Owner</div>
					<div class="text-[10px] text-slate-500">owner@resto.com</div>
				</button>
				<button
					type="button"
					onclick={() => quickSelectRole('admin@resto.com')}
					class="p-2 rounded-xl border border-slate-200 hover:border-orange-500 text-left transition-colors font-medium bg-slate-50"
				>
					<div class="font-bold text-slate-900">Admin</div>
					<div class="text-[10px] text-slate-500">admin@resto.com</div>
				</button>
				<button
					type="button"
					onclick={() => quickSelectRole('cashier@resto.com')}
					class="p-2 rounded-xl border border-slate-200 hover:border-orange-500 text-left transition-colors font-medium bg-slate-50"
				>
					<div class="font-bold text-slate-900">Kasir</div>
					<div class="text-[10px] text-slate-500">cashier@resto.com</div>
				</button>
				<button
					type="button"
					onclick={() => quickSelectRole('kitchen@resto.com')}
					class="p-2 rounded-xl border border-slate-200 hover:border-orange-500 text-left transition-colors font-medium bg-slate-50"
				>
					<div class="font-bold text-slate-900">Dapur (KDS)</div>
					<div class="text-[10px] text-slate-500">kitchen@resto.com</div>
				</button>
			</div>
		</div>
	</div>
</div>
