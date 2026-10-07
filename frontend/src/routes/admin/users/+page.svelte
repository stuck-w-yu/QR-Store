<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { User, Role } from '$lib/types';
	import { 
		Plus, Users, Trash2, Shield, Mail, Key, KeyRound, 
		Eye, EyeOff, Lock, X, AlertCircle, CheckCircle2 
	} from '@lucide/svelte';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let showCreateModal = $state(false);

	let newUser = $state({
		name: '',
		email: '',
		password: '',
		role: 'CASHIER' as Role
	});

	// Reset password by Owner / Superadmin state
	let showResetPasswordModal = $state(false);
	let targetUserForReset = $state<User | null>(null);
	let resetNewPassword = $state('');
	let resetConfirmPassword = $state('');
	let resetError = $state('');
	let resetSuccess = $state('');
	let resetLoading = $state(false);
	let showResetPassword = $state(false);

	function canOwnerReset(u: User): boolean {
		if (!auth.user) return false;
		if (u.id === auth.user.id) return false;
		if (auth.user.role === 'SUPERADMIN') return true;
		if (auth.user.role === 'OWNER') {
			// Owner can only reset subordinate roles (ADMIN, CASHIER, KITCHEN)
			return u.role !== 'OWNER' && u.role !== 'SUPERADMIN';
		}
		return false;
	}

	function openResetModal(u: User) {
		targetUserForReset = u;
		resetNewPassword = '';
		resetConfirmPassword = '';
		resetError = '';
		resetSuccess = '';
		showResetPassword = false;
		showResetPasswordModal = true;
	}

	async function handleResetPasswordSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!targetUserForReset) return;
		resetError = '';
		resetSuccess = '';

		if (resetNewPassword.length < 6) {
			resetError = 'Kata sandi baru minimal 6 karakter.';
			return;
		}

		if (resetNewPassword !== resetConfirmPassword) {
			resetError = 'Konfirmasi kata sandi tidak cocok.';
			return;
		}

		try {
			resetLoading = true;
			await api.put(`/users/${targetUserForReset.id}/password`, {
				new_password: resetNewPassword
			});
			resetSuccess = `Kata sandi untuk ${targetUserForReset.name} berhasil diperbarui!`;
			setTimeout(() => {
				showResetPasswordModal = false;
				resetSuccess = '';
				targetUserForReset = null;
			}, 1500);
		} catch (err: any) {
			resetError = err?.message || 'Gagal mengubah kata sandi karyawan.';
		} finally {
			resetLoading = false;
		}
	}

	async function loadUsers() {
		try {
			loading = true;
			const data = await api.get<User[]>('/users');
			users = data;
		} catch (e) {
			console.error('Failed to load users', e);
		} finally {
			loading = false;
		}
	}

	async function handleCreateUser(e: SubmitEvent) {
		e.preventDefault();
		try {
			await api.post('/users', newUser);
			showCreateModal = false;
			newUser = { name: '', email: '', password: '', role: 'CASHIER' };
			await loadUsers();
		} catch (e: any) {
			alert(e?.message || 'Gagal menambah karyawan');
		}
	}

	async function handleDeleteUser(userId: string) {
		if (!confirm('Hapus karyawan ini dari sistem?')) return;
		try {
			await api.delete(`/users/${userId}`);
			await loadUsers();
		} catch (e) {
			alert('Gagal menghapus karyawan');
		}
	}

	onMount(() => {
		loadUsers();
	});
</script>

<div class="space-y-8 max-w-6xl mx-auto">
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-black text-slate-900 font-['Outfit']">Kelola Karyawan & Hak Akses</h1>
			<p class="text-xs text-slate-500 font-medium mt-0.5">
				Atur akun staf kasir, koki dapur, admin, dan owner restoran
			</p>
		</div>

		<button
			type="button"
			onclick={() => (showCreateModal = true)}
			class="w-full sm:w-auto justify-center bg-orange-600 hover:bg-orange-700 text-white font-bold text-xs px-4 py-2.5 rounded-xl shadow-md shadow-orange-600/30 flex items-center gap-1.5 transition-colors"
		>
			<Plus class="w-4 h-4" />
			<span>+ Tambah Karyawan</span>
		</button>
	</div>

	<!-- Modal -->
	{#if showCreateModal}
		<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-5 sm:p-6 w-full max-w-md space-y-4 shadow-xl max-h-[90vh] overflow-y-auto">
				<h3 class="font-bold text-base text-slate-900">Tambah Akun Karyawan</h3>
				<form onsubmit={handleCreateUser} class="space-y-3.5 text-xs">
					<div>
						<label for="user-name" class="block font-bold text-slate-700 mb-1">Nama Lengkap</label>
						<input
							id="user-name"
							type="text"
							bind:value={newUser.name}
							required
							placeholder="Nama staf..."
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div>
						<label for="user-email" class="block font-bold text-slate-700 mb-1">Email</label>
						<input
							id="user-email"
							type="email"
							bind:value={newUser.email}
							required
							placeholder="email@resto.com"
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div>
						<label for="user-password" class="block font-bold text-slate-700 mb-1">Kata Sandi</label>
						<input
							id="user-password"
							type="password"
							bind:value={newUser.password}
							required
							minlength="6"
							placeholder="Minimal 6 karakter"
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500"
						/>
					</div>

					<div>
						<label for="user-role" class="block font-bold text-slate-700 mb-1">Peran / Role</label>
						<select
							id="user-role"
							bind:value={newUser.role}
							class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 bg-white"
						>
							<option value="CASHIER">CASHIER (Kasir)</option>
							<option value="KITCHEN">KITCHEN (Dapur)</option>
							<option value="ADMIN">ADMIN (Operasional)</option>
							{#if auth.user?.role === 'OWNER'}
								<option value="OWNER">OWNER (Pemilik)</option>
							{/if}
						</select>
					</div>

					<div class="flex justify-end gap-2 pt-3 border-t border-slate-100">
						<button
							type="button"
							onclick={() => (showCreateModal = false)}
							class="px-4 py-2 font-semibold text-slate-600 rounded-xl hover:bg-slate-100"
						>
							Batal
						</button>
						<button
							type="submit"
							class="px-5 py-2 font-bold bg-orange-600 text-white rounded-xl shadow-md"
						>
							Simpan Karyawan
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Users List -->
	<div class="bg-white rounded-3xl border border-slate-200/80 shadow-xs overflow-hidden">
		{#if loading}
			<div class="text-center py-20 text-slate-500 text-xs">Memuat daftar staf...</div>
		{:else}
			<div class="divide-y divide-slate-100">
				{#each users as u}
					<div class="p-4 sm:px-6 flex items-center justify-between gap-3 sm:gap-4 hover:bg-slate-50/60 transition-colors">
						<div class="flex items-center gap-3 sm:gap-3.5 min-w-0 flex-1">
							<div class="w-10 h-10 rounded-xl bg-orange-50 text-orange-600 flex items-center justify-center font-bold text-sm shrink-0">
								{u.name.charAt(0)}
							</div>
							<div class="min-w-0 flex-1">
								<h3 class="font-bold text-sm text-slate-900 truncate">{u.name}</h3>
								<span class="text-xs text-slate-400 font-mono truncate block">{u.email}</span>
							</div>
						</div>

						<div class="flex items-center gap-1.5 sm:gap-2.5 shrink-0">
							<span class="px-2.5 sm:px-3 py-1 rounded-full text-[10px] font-bold tracking-wider uppercase bg-slate-100 text-slate-700">
								{u.role}
							</span>

							{#if canOwnerReset(u)}
								<button
									type="button"
									onclick={() => openResetModal(u)}
									class="p-2 text-slate-400 hover:text-orange-600 hover:bg-orange-50 rounded-xl transition-colors cursor-pointer"
									title="Ganti Kata Sandi Karyawan"
								>
									<KeyRound class="w-4 h-4" />
								</button>
							{/if}

							{#if u.id !== auth.user?.id}
								<button
									type="button"
									onclick={() => handleDeleteUser(u.id)}
									class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors cursor-pointer"
									title="Hapus Karyawan"
								>
									<Trash2 class="w-4 h-4" />
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>

<!-- Modal Ganti Password Karyawan oleh Owner -->
{#if showResetPasswordModal && targetUserForReset}
	<div class="fixed inset-0 bg-black/60 backdrop-blur-xs z-50 flex items-center justify-center p-4 animate-in fade-in duration-150">
		<div class="bg-white rounded-3xl p-6 sm:p-7 w-full max-w-md space-y-5 shadow-2xl border border-slate-100 animate-in zoom-in-95 duration-150">
			<div class="flex items-center justify-between border-b border-slate-100 pb-3">
				<div class="flex items-center gap-2.5">
					<div class="w-9 h-9 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center">
						<KeyRound class="w-5 h-5" />
					</div>
					<div>
						<h3 class="font-extrabold text-base text-slate-900 font-['Outfit']">Ganti Password Karyawan</h3>
						<p class="text-[11px] text-slate-500 font-medium">Atur ulang kata sandi oleh Owner restoran</p>
					</div>
				</div>
				<button
					type="button"
					onclick={() => (showResetPasswordModal = false)}
					class="p-1.5 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer"
				>
					<X class="w-5 h-5" />
				</button>
			</div>

			<!-- Target User Info Card -->
			<div class="p-3.5 rounded-2xl bg-slate-50 border border-slate-100 flex items-center gap-3">
				<div class="w-9 h-9 rounded-xl bg-orange-100 text-orange-600 flex items-center justify-center font-bold text-xs shrink-0">
					{targetUserForReset.name.charAt(0)}
				</div>
				<div class="min-w-0 flex-1 text-xs">
					<div class="flex items-center gap-2">
						<span class="font-bold text-slate-900 truncate">{targetUserForReset.name}</span>
						<span class="px-2 py-0.5 rounded-full text-[9px] font-bold uppercase tracking-wider bg-white text-slate-700 border border-slate-200">
							{targetUserForReset.role}
						</span>
					</div>
					<div class="text-[11px] text-slate-400 font-mono truncate">{targetUserForReset.email}</div>
				</div>
			</div>

			{#if resetError}
				<div class="p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center gap-2">
					<AlertCircle class="w-4 h-4 shrink-0" />
					<span>{resetError}</span>
				</div>
			{/if}

			{#if resetSuccess}
				<div class="p-3 rounded-xl bg-emerald-50 border border-emerald-200 text-emerald-700 text-xs flex items-center gap-2">
					<CheckCircle2 class="w-4 h-4 shrink-0" />
					<span>{resetSuccess}</span>
				</div>
			{/if}

			<form onsubmit={handleResetPasswordSubmit} class="space-y-4 text-xs">
				<!-- Kata Sandi Baru -->
				<div>
					<label for="reset-new-password" class="block font-bold text-slate-700 mb-1">
						Kata Sandi Baru <span class="text-rose-500">*</span>
					</label>
					<div class="relative">
						<input
							id="reset-new-password"
							type={showResetPassword ? 'text' : 'password'}
							bind:value={resetNewPassword}
							required
							minlength="6"
							placeholder="Minimal 6 karakter"
							class="w-full p-2.5 pr-10 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 transition-colors"
						/>
						<button
							type="button"
							onclick={() => (showResetPassword = !showResetPassword)}
							class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 cursor-pointer"
							tabindex="-1"
						>
							{#if showResetPassword}
								<EyeOff class="w-4 h-4" />
							{:else}
								<Eye class="w-4 h-4" />
							{/if}
						</button>
					</div>
				</div>

				<!-- Konfirmasi Kata Sandi Baru -->
				<div>
					<label for="reset-confirm-password" class="block font-bold text-slate-700 mb-1">
						Konfirmasi Kata Sandi Baru <span class="text-rose-500">*</span>
					</label>
					<input
						id="reset-confirm-password"
						type={showResetPassword ? 'text' : 'password'}
						bind:value={resetConfirmPassword}
						required
						minlength="6"
						placeholder="Ulangi kata sandi baru"
						class="w-full p-2.5 rounded-xl border border-slate-200 focus:outline-none focus:border-orange-500 transition-colors"
					/>
				</div>

				<div class="flex justify-end gap-2.5 pt-3 border-t border-slate-100">
					<button
						type="button"
						onclick={() => (showResetPasswordModal = false)}
						class="px-4 py-2.5 font-bold text-slate-600 rounded-xl hover:bg-slate-100 transition-colors cursor-pointer"
					>
						Batal
					</button>
					<button
						type="submit"
						disabled={resetLoading}
						class="px-5 py-2.5 font-bold bg-orange-600 hover:bg-orange-700 disabled:opacity-50 text-white rounded-xl shadow-md shadow-orange-600/30 transition-colors flex items-center gap-1.5 cursor-pointer"
					>
						{#if resetLoading}
							<span>Menyimpan...</span>
						{:else}
							<Lock class="w-3.5 h-3.5" />
							<span>Simpan Password Karyawan</span>
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
