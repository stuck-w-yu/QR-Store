<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { 
		LayoutDashboard, ShoppingCart, Utensils, QrCode, 
		Users, ChefHat, LogOut, Store, ExternalLink, Menu as MenuIcon, X,
		CreditCard, Monitor, ClipboardList, ShieldCheck, Wallet,
		KeyRound, Eye, EyeOff, Lock, CheckCircle2, AlertCircle
	} from '@lucide/svelte';

	let { children } = $props();
	let sidebarOpen = $state(false);

	// Self-service change password state
	let showPasswordModal = $state(false);
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmNewPassword = $state('');
	let passwordError = $state('');
	let passwordSuccess = $state('');
	let passwordLoading = $state(false);
	let showCurrentPassword = $state(false);
	let showNewPassword = $state(false);

	const navItems = [
		{ href: '/admin/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/admin/payments', label: 'Konfirmasi Kasir & Tunai', icon: Wallet, roles: ['SUPERADMIN', 'OWNER', 'ADMIN', 'CASHIER'] },
		{ href: '/admin/cashier', label: 'Kasir POS', icon: CreditCard, roles: ['SUPERADMIN', 'OWNER', 'ADMIN', 'CASHIER'] },
		{ href: '/admin/orders', label: 'Pesanan Masuk', icon: ShoppingCart, roles: ['SUPERADMIN', 'OWNER', 'ADMIN', 'CASHIER'] },
		{ href: '/admin/menus', label: 'Menu & Kategori', icon: Utensils, roles: ['SUPERADMIN', 'OWNER', 'ADMIN'] },
		{ href: '/admin/tables', label: 'Meja & QR Code', icon: QrCode, roles: ['SUPERADMIN', 'OWNER', 'ADMIN'] },
		{ href: '/admin/registers', label: 'Mesin Kasir', icon: Monitor, roles: ['SUPERADMIN', 'OWNER', 'ADMIN', 'CASHIER'] },
		{ href: '/admin/shifts', label: 'Laporan Shift', icon: ClipboardList, roles: ['SUPERADMIN', 'OWNER', 'ADMIN'] },
		{ href: '/admin/users', label: 'Kelola Karyawan', icon: Users, roles: ['SUPERADMIN', 'OWNER', 'ADMIN'] },
		{ href: '/superadmin', label: 'Superadmin HQ', icon: ShieldCheck, roles: ['SUPERADMIN'] }
	];

	onMount(async () => {
		if (!auth.initialized) {
			await auth.init();
		}
		if (!auth.user) {
			goto('/login');
		}
	});

	function handleLogout() {
		auth.logout();
		goto('/login');
	}

	function resetPasswordForm() {
		currentPassword = '';
		newPassword = '';
		confirmNewPassword = '';
		passwordError = '';
		passwordSuccess = '';
		showCurrentPassword = false;
		showNewPassword = false;
	}

	async function handleSelfPasswordChange(e: SubmitEvent) {
		e.preventDefault();
		passwordError = '';
		passwordSuccess = '';

		if (newPassword.length < 6) {
			passwordError = 'Kata sandi baru minimal 6 karakter.';
			return;
		}

		if (newPassword !== confirmNewPassword) {
			passwordError = 'Konfirmasi kata sandi baru tidak cocok.';
			return;
		}

		try {
			passwordLoading = true;
			await api.put('/auth/password', {
				current_password: currentPassword,
				new_password: newPassword
			});
			passwordSuccess = 'Kata sandi Anda berhasil diperbarui!';
			currentPassword = '';
			newPassword = '';
			confirmNewPassword = '';
			setTimeout(() => {
				showPasswordModal = false;
				passwordSuccess = '';
			}, 1500);
		} catch (err: any) {
			passwordError = err?.message || 'Gagal mengubah kata sandi.';
		} finally {
			passwordLoading = false;
		}
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
						<span>Menu Pelanggan</span>
					</div>
					<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
				</a>
			</div>
		</nav>

		<!-- User Footer -->
		<div class="p-4 border-t border-slate-800 flex items-center justify-between">
			<button
				type="button"
				onclick={() => { resetPasswordForm(); showPasswordModal = true; }}
				class="truncate flex-1 text-left hover:opacity-85 transition-opacity cursor-pointer group pr-2"
				title="Klik untuk Kelola Kata Sandi Akun"
			>
				<div class="text-xs font-bold text-slate-200 group-hover:text-orange-400 transition-colors truncate">
					{auth.user?.name || 'Staff User'}
				</div>
				<div class="text-[10px] text-slate-500 truncate">{auth.user?.email || ''}</div>
			</button>
			<div class="flex items-center gap-1 shrink-0">
				<button
					type="button"
					onclick={() => { resetPasswordForm(); showPasswordModal = true; }}
					class="text-slate-400 hover:text-orange-400 hover:bg-slate-800/80 p-1.5 rounded-lg transition-colors cursor-pointer"
					title="Ganti Kata Sandi Akun"
				>
					<KeyRound class="w-4 h-4" />
				</button>
				<button
					type="button"
					onclick={handleLogout}
					class="text-slate-400 hover:text-red-400 hover:bg-slate-800/80 p-1.5 rounded-lg transition-colors cursor-pointer"
					title="Keluar"
				>
					<LogOut class="w-4 h-4" />
				</button>
			</div>
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
							<span>Menu Pelanggan</span>
						</div>
						<ExternalLink class="w-3.5 h-3.5 text-slate-500" />
					</a>
				</div>

				<div class="pt-3 mt-3 border-t border-slate-800 flex items-center justify-between px-1">
					<button
						type="button"
						onclick={() => { sidebarOpen = false; resetPasswordForm(); showPasswordModal = true; }}
						class="truncate mr-3 text-left"
					>
						<div class="text-xs font-bold text-slate-200 truncate">{auth.user?.name || 'Staff User'}</div>
						<div class="text-[10px] text-slate-500 truncate">{auth.user?.email || ''}</div>
					</button>
					<div class="flex items-center gap-1 shrink-0">
						<button
							type="button"
							onclick={() => { sidebarOpen = false; resetPasswordForm(); showPasswordModal = true; }}
							class="flex items-center gap-1 px-2.5 py-1.5 rounded-xl text-orange-400 hover:bg-orange-500/10 font-bold text-xs"
							title="Ganti Kata Sandi"
						>
							<KeyRound class="w-3.5 h-3.5" />
							<span>Sandi</span>
						</button>
						<button
							type="button"
							onclick={handleLogout}
							class="flex items-center gap-1 px-2.5 py-1.5 rounded-xl text-red-400 hover:bg-red-500/10 font-bold text-xs"
						>
							<LogOut class="w-3.5 h-3.5" />
							<span>Keluar</span>
						</button>
					</div>
				</div>
			</div>
		{/if}

		<main class="flex-1 p-4 sm:p-6 md:p-8 overflow-y-auto">
			{@render children()}
		</main>
	</div>
</div>

<!-- Modal Ganti Kata Sandi Akun Sendiri (Semua Role) -->
{#if showPasswordModal}
	<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4 animate-in fade-in duration-150">
		<div class="bg-white rounded-3xl p-6 sm:p-7 w-full max-w-md space-y-5 shadow-2xl border border-slate-100 animate-in zoom-in-95 duration-150">
			<div class="flex items-center justify-between border-b border-slate-100 pb-3">
				<div class="flex items-center gap-2.5">
					<div class="w-9 h-9 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center">
						<KeyRound class="w-5 h-5" />
					</div>
					<div>
						<h3 class="font-extrabold text-base text-slate-900 font-['Outfit']">Ganti Kata Sandi</h3>
						<p class="text-[11px] text-slate-500 font-medium">Perbarui kata sandi akun Anda ({auth.user?.role})</p>
					</div>
				</div>
				<button
					type="button"
					onclick={() => (showPasswordModal = false)}
					class="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			{#if passwordError}
				<div class="p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center gap-2">
					<AlertCircle class="w-4 h-4 shrink-0" />
					<span>{passwordError}</span>
				</div>
			{/if}

			{#if passwordSuccess}
				<div class="p-3 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-700 text-xs flex items-center gap-2">
					<CheckCircle2 class="w-4 h-4 shrink-0" />
					<span>{passwordSuccess}</span>
				</div>
			{/if}

			<form onsubmit={handleSelfPasswordChange} class="space-y-4 text-xs">
				<!-- Kata Sandi Lama -->
				<div>
					<label for="self-current-password" class="block font-bold text-slate-700 mb-1">
						Kata Sandi Saat Ini <span class="text-rose-500">*</span>
					</label>
					<div class="relative">
						<input
							id="self-current-password"
							type={showCurrentPassword ? 'text' : 'password'}
							bind:value={currentPassword}
							required
							placeholder="Masukkan kata sandi saat ini"
							class="w-full p-2.5 pr-10 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 transition-colors"
						/>
						<button
							type="button"
							onclick={() => (showCurrentPassword = !showCurrentPassword)}
							class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
							tabindex="-1"
						>
							{#if showCurrentPassword}
								<EyeOff class="w-4 h-4" />
							{:else}
								<Eye class="w-4 h-4" />
							{/if}
						</button>
					</div>
				</div>

				<!-- Kata Sandi Baru -->
				<div>
					<label for="self-new-password" class="block font-bold text-slate-700 mb-1">
						Kata Sandi Baru <span class="text-rose-500">*</span>
					</label>
					<div class="relative">
						<input
							id="self-new-password"
							type={showNewPassword ? 'text' : 'password'}
							bind:value={newPassword}
							required
							minlength="6"
							placeholder="Minimal 6 karakter"
							class="w-full p-2.5 pr-10 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 transition-colors"
						/>
						<button
							type="button"
							onclick={() => (showNewPassword = !showNewPassword)}
							class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
							tabindex="-1"
						>
							{#if showNewPassword}
								<EyeOff class="w-4 h-4" />
							{:else}
								<Eye class="w-4 h-4" />
							{/if}
						</button>
					</div>
				</div>

				<!-- Konfirmasi Kata Sandi Baru -->
				<div>
					<label for="self-confirm-password" class="block font-bold text-slate-700 mb-1">
						Konfirmasi Kata Sandi Baru <span class="text-rose-500">*</span>
					</label>
					<input
						id="self-confirm-password"
						type={showNewPassword ? 'text' : 'password'}
						bind:value={confirmNewPassword}
						required
						minlength="6"
						placeholder="Ulangi kata sandi baru"
						class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 transition-colors"
					/>
				</div>

				<div class="flex justify-end gap-2.5 pt-3 border-t border-slate-100">
					<button
						type="button"
						onclick={() => (showPasswordModal = false)}
						class="px-4 py-2.5 font-bold text-slate-600 rounded-xl hover:bg-slate-100 transition-colors cursor-pointer"
					>
						Batal
					</button>
					<button
						type="submit"
						disabled={passwordLoading}
						class="px-5 py-2.5 font-bold bg-orange-600 hover:bg-orange-700 disabled:opacity-50 text-white rounded-xl shadow-md shadow-orange-600/30 transition-colors flex items-center gap-1.5 cursor-pointer"
					>
						{#if passwordLoading}
							<span>Memperbarui...</span>
						{:else}
							<Lock class="w-3.5 h-3.5" />
							<span>Simpan Kata Sandi</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
