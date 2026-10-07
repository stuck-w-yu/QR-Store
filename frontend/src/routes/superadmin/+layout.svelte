<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { 
		ShieldAlert, Store, LogOut, ArrowLeft, 
		Layers, Users, TrendingUp, Sparkles, ExternalLink,
		Menu as MenuIcon, X, KeyRound, UserCog, Mail, Lock,
		Eye, EyeOff, CheckCircle2, AlertCircle, Loader2, ShieldCheck
	} from '@lucide/svelte';

	let { children } = $props();
	let mobileMenuOpen = $state(false);

	// Modal Manajemen Akun Superadmin
	let showAccountModal = $state(false);
	let accountForm = $state({
		id: '',
		name: '',
		email: '',
		current_password: '',
		new_password: '',
		confirm_password: ''
	});
	let showCurrentPassword = $state(false);
	let showNewPassword = $state(false);
	let accountLoading = $state(false);
	let accountError = $state<string | null>(null);
	let accountSuccess = $state<string | null>(null);

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

	function openAccountModal() {
		accountForm = {
			id: auth.user?.id || 'usr_superadmin',
			name: auth.user?.name || 'Super Administrator',
			email: auth.user?.email || 'superadmin@qrstore.id',
			current_password: '',
			new_password: '',
			confirm_password: ''
		};
		showCurrentPassword = false;
		showNewPassword = false;
		accountError = null;
		accountSuccess = null;
		showAccountModal = true;
	}

	async function handleSaveAccount(e: SubmitEvent) {
		e.preventDefault();
		accountError = null;
		accountSuccess = null;

		if (!accountForm.current_password) {
			accountError = 'Kata sandi saat ini wajib diisi untuk verifikasi keamanan.';
			return;
		}

		if (accountForm.new_password) {
			if (accountForm.new_password.length < 6) {
				accountError = 'Kata sandi baru minimal 6 karakter.';
				return;
			}
			if (accountForm.new_password !== accountForm.confirm_password) {
				accountError = 'Konfirmasi kata sandi baru tidak cocok.';
				return;
			}
		}

		try {
			accountLoading = true;
			const payload: any = {
				current_password: accountForm.current_password
			};
			if (accountForm.name && accountForm.name !== auth.user?.name) {
				payload.name = accountForm.name;
			}
			if (accountForm.email && accountForm.email !== auth.user?.email) {
				payload.email = accountForm.email;
			}
			if (accountForm.id && accountForm.id !== auth.user?.id) {
				payload.new_id = accountForm.id;
			}
			if (accountForm.new_password) {
				payload.new_password = accountForm.new_password;
			}

			const res = await api.put<{ user: any; access_token: string }>('/auth/account', payload);

			if (res && res.user) {
				auth.updateUser(res.user, res.access_token);
				accountSuccess = 'Identitas akun Superadmin berhasil diperbarui!';
				setTimeout(() => {
					showAccountModal = false;
					accountSuccess = null;
				}, 1200);
			}
		} catch (err: any) {
			accountError = err?.message || 'Gagal memperbarui akun Superadmin.';
		} finally {
			accountLoading = false;
		}
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

				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={openAccountModal}
						class="text-right hover:opacity-85 transition-all p-1.5 px-2.5 rounded-xl hover:bg-slate-800/80 cursor-pointer group flex items-center gap-2.5 border border-transparent hover:border-slate-700/60"
						title="Kelola ID, Email, & Password Superadmin"
					>
						<div class="text-right">
							<div class="text-xs font-bold text-slate-200 group-hover:text-indigo-300 transition-colors">
								{auth.user?.name || 'Super Administrator'}
							</div>
							<div class="text-[10px] text-indigo-400 font-mono">
								{auth.user?.email || 'superadmin@qrstore.id'}
							</div>
						</div>
						<div class="w-8 h-8 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 flex items-center justify-center shrink-0 group-hover:bg-indigo-500/20 transition-colors">
							<UserCog class="w-4 h-4" />
						</div>
					</button>

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
					<button
						type="button"
						onclick={() => { mobileMenuOpen = false; openAccountModal(); }}
						class="truncate mr-2 text-left hover:opacity-80 transition-opacity"
					>
						<div class="font-bold text-slate-200 truncate">{auth.user?.name || 'Super Admin'}</div>
						<div class="text-[10px] text-indigo-400 truncate font-mono">{auth.user?.email || 'superadmin@qrstore.id'}</div>
					</button>
					<div class="flex items-center gap-1.5 shrink-0">
						<button
							type="button"
							onclick={() => { mobileMenuOpen = false; openAccountModal(); }}
							class="px-2.5 py-1.5 rounded-xl bg-indigo-500/15 text-indigo-300 hover:bg-indigo-500/25 font-bold text-xs flex items-center gap-1 border border-indigo-500/30"
							title="Kelola Akun Superadmin"
						>
							<UserCog class="w-3.5 h-3.5" />
							<span>Akun</span>
						</button>
						<button
							type="button"
							onclick={handleLogout}
							class="px-2.5 py-1.5 rounded-xl bg-red-500/10 text-red-400 font-bold flex items-center gap-1 shrink-0"
						>
							<LogOut class="w-3.5 h-3.5" />
							<span>Keluar</span>
						</button>
					</div>
				</div>
			</div>
		{/if}
	</header>

	<!-- Modal: Manajemen Akun Superadmin -->
	{#if showAccountModal}
		<div class="fixed inset-0 bg-black/75 backdrop-blur-xs z-50 flex items-center justify-center p-4 animate-in fade-in duration-150">
			<div class="bg-slate-900 text-white rounded-3xl p-6 sm:p-7 w-full max-w-md space-y-5 shadow-2xl border border-slate-800 animate-in zoom-in-95 duration-150 max-h-[90vh] overflow-y-auto">
				<div class="flex items-start justify-between border-b border-slate-800 pb-4">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-2xl bg-indigo-500/15 text-indigo-400 border border-indigo-500/30 flex items-center justify-center shrink-0">
							<ShieldCheck class="w-5 h-5" />
						</div>
						<div>
							<h3 class="text-base sm:text-lg font-black font-['Outfit'] text-white">Manajemen Akun Superadmin</h3>
							<p class="text-[11px] text-slate-400">Ubah ID, alamat email login, dan kata sandi</p>
						</div>
					</div>
					<button
						type="button"
						onclick={() => (showAccountModal = false)}
						class="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
						aria-label="Tutup"
					>
						<X class="w-5 h-5" />
					</button>
				</div>

				{#if accountError}
					<div class="p-3 bg-red-500/15 border border-red-500/30 text-red-300 rounded-xl text-xs flex items-center gap-2">
						<AlertCircle class="w-4 h-4 shrink-0" />
						<span>{accountError}</span>
					</div>
				{/if}

				{#if accountSuccess}
					<div class="p-3 bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 rounded-xl text-xs flex items-center gap-2">
						<CheckCircle2 class="w-4 h-4 shrink-0" />
						<span>{accountSuccess}</span>
					</div>
				{/if}

				<form onsubmit={handleSaveAccount} class="space-y-4 text-xs font-medium">
					<div>
						<label for="acc-id" class="block font-bold text-slate-300 mb-1">ID Pengguna (User ID)</label>
						<div class="relative">
							<input
								id="acc-id"
								type="text"
								bind:value={accountForm.id}
								required
								placeholder="usr_superadmin"
								class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-indigo-300 font-mono focus:outline-none focus:border-indigo-500 font-bold"
							/>
							<span class="absolute right-3 top-2.5 text-[10px] text-slate-500 uppercase font-mono font-bold">Unik</span>
						</div>
						<p class="text-[10px] text-slate-500 mt-1">ID identitas akun Super Administrator di database.</p>
					</div>

					<div>
						<label for="acc-name" class="block font-bold text-slate-300 mb-1">Nama Superadmin</label>
						<input
							id="acc-name"
							type="text"
							bind:value={accountForm.name}
							required
							placeholder="Super Administrator"
							class="w-full px-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500 font-bold"
						/>
					</div>

					<div>
						<label for="acc-email" class="block font-bold text-slate-300 mb-1">ID Email Login</label>
						<div class="relative">
							<input
								id="acc-email"
								type="email"
								bind:value={accountForm.email}
								required
								placeholder="superadmin@qrstore.id"
								class="w-full pl-9 pr-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono focus:outline-none focus:border-indigo-500"
							/>
							<Mail class="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
						</div>
						<p class="text-[10px] text-slate-500 mt-1">Email ini digunakan untuk login / masuk ke sistem Superadmin.</p>
					</div>

					<!-- Section: Ganti Kata Sandi Baru (Opsional) -->
					<div class="pt-3 border-t border-slate-800 space-y-3">
						<div class="flex items-center justify-between">
							<span class="text-[11px] font-bold text-slate-300">Ganti Kata Sandi (Opsional)</span>
							<span class="text-[10px] text-slate-500">Kosongkan jika tidak ingin ganti</span>
						</div>

						<div>
							<label for="acc-new-pass" class="block text-slate-400 mb-1 font-semibold">Kata Sandi Baru</label>
							<div class="relative">
								<input
									id="acc-new-pass"
									type={showNewPassword ? 'text' : 'password'}
									bind:value={accountForm.new_password}
									placeholder="Minimal 6 karakter"
									minlength="6"
									class="w-full pl-9 pr-10 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
								/>
								<Lock class="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
								<button
									type="button"
									onclick={() => (showNewPassword = !showNewPassword)}
									class="absolute right-3 top-2.5 text-slate-500 hover:text-slate-300"
								>
									{#if showNewPassword}
										<EyeOff class="w-4 h-4" />
									{:else}
										<Eye class="w-4 h-4" />
									{/if}
								</button>
							</div>
						</div>

						{#if accountForm.new_password}
							<div>
								<label for="acc-conf-pass" class="block text-slate-400 mb-1 font-semibold">Konfirmasi Kata Sandi Baru</label>
								<div class="relative">
									<input
										id="acc-conf-pass"
										type={showNewPassword ? 'text' : 'password'}
										bind:value={accountForm.confirm_password}
										placeholder="Ulangi kata sandi baru"
										required
										class="w-full pl-9 pr-3.5 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
									/>
									<Lock class="w-4 h-4 text-slate-500 absolute left-3 top-2.5" />
								</div>
							</div>
						{/if}
					</div>

					<!-- Section: Verifikasi Keamanan (Kata Sandi Saat Ini) -->
					<div class="pt-3 border-t border-slate-800">
						<label for="acc-curr-pass" class="block font-bold text-amber-400 mb-1 flex items-center gap-1.5">
							<KeyRound class="w-3.5 h-3.5 text-amber-400" />
							<span>Kata Sandi Saat Ini (Verifikasi Keamanan)</span>
						</label>
						<div class="relative">
							<input
								id="acc-curr-pass"
								type={showCurrentPassword ? 'text' : 'password'}
								bind:value={accountForm.current_password}
								required
								placeholder="Masukkan kata sandi saat ini"
								class="w-full pr-10 pl-3.5 py-2.5 rounded-xl bg-slate-950 border border-amber-500/40 text-white focus:outline-none focus:border-amber-400"
							/>
							<button
								type="button"
								onclick={() => (showCurrentPassword = !showCurrentPassword)}
								class="absolute right-3 top-2.5 text-slate-500 hover:text-slate-300"
							>
								{#if showCurrentPassword}
									<EyeOff class="w-4 h-4" />
								{:else}
									<Eye class="w-4 h-4" />
								{/if}
							</button>
						</div>
						<p class="text-[10px] text-slate-500 mt-1">Wajib dimasukkan untuk memvalidasi perubahan akun Superadmin.</p>
					</div>

					<div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-800">
						<button
							type="button"
							onclick={() => (showAccountModal = false)}
							class="px-4 py-2.5 text-slate-400 hover:text-white rounded-xl font-bold transition-colors"
						>
							Batal
						</button>
						<button
							type="submit"
							disabled={accountLoading}
							class="px-5 py-2.5 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl font-bold shadow-lg shadow-indigo-600/30 disabled:opacity-50 transition-all flex items-center gap-2"
						>
							{#if accountLoading}
								<Loader2 class="w-4 h-4 animate-spin text-white" />
								<span>Menyimpan...</span>
							{:else}
								<span>Simpan Perubahan</span>
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Main Body Stage -->
	<main class="flex-1 p-4 sm:p-8 max-w-7xl mx-auto w-full">
		{@render children()}
	</main>
</div>
