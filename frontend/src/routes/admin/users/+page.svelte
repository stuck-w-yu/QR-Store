<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/stores/auth.svelte';
	import type { User, Role } from '$lib/types';
	import { Plus, Users, Trash2, Shield, Mail, Key } from '@lucide/svelte';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let showCreateModal = $state(false);

	let newUser = $state({
		name: '',
		email: '',
		password: '',
		role: 'CASHIER' as Role
	});

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
			class="bg-orange-600 hover:bg-orange-700 text-white font-bold text-xs px-4 py-2.5 rounded-xl shadow-md shadow-orange-600/30 flex items-center gap-1.5 transition-colors"
		>
			<Plus class="w-4 h-4" />
			<span>+ Tambah Karyawan</span>
		</button>
	</div>

	<!-- Modal -->
	{#if showCreateModal}
		<div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
			<div class="bg-white rounded-3xl p-6 w-full max-w-md space-y-4 shadow-xl">
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
					<div class="p-4 sm:px-6 flex items-center justify-between gap-4 hover:bg-slate-50/60 transition-colors">
						<div class="flex items-center gap-3.5">
							<div class="w-10 h-10 rounded-xl bg-orange-50 text-orange-600 flex items-center justify-center font-bold text-sm">
								{u.name.charAt(0)}
							</div>
							<div>
								<h3 class="font-bold text-sm text-slate-900">{u.name}</h3>
								<span class="text-xs text-slate-400 font-mono">{u.email}</span>
							</div>
						</div>

						<div class="flex items-center gap-4">
							<span class="px-3 py-1 rounded-full text-[10px] font-bold tracking-wider uppercase bg-slate-100 text-slate-700">
								{u.role}
							</span>

							{#if u.id !== auth.user?.id}
								<button
									type="button"
									onclick={() => handleDeleteUser(u.id)}
									class="p-2 text-slate-400 hover:text-red-500 rounded-xl transition-colors"
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
