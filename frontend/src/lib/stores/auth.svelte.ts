import { api } from '$lib/api/client';
import type { User, Role } from '$lib/types';

class AuthStore {
	user = $state<User | null>(null);
	loading = $state<boolean>(false);
	initialized = $state<boolean>(false);

	constructor() {
		if (typeof window !== 'undefined') {
			this.init();
		}
	}

	async init() {
		const token = api.getToken();
		if (!token) {
			this.initialized = true;
			return;
		}

		try {
			this.loading = true;
			const user = await api.get<User>('/auth/me');
			this.user = user;
		} catch (e) {
			api.setToken(null);
			this.user = null;
		} finally {
			this.loading = false;
			this.initialized = true;
		}
	}

	async login(email: string, password: string): Promise<User> {
		this.loading = true;
		try {
			const res = await api.post<{ user: User; access_token: string }>('/auth/login', {
				email,
				password
			});
			api.setToken(res.access_token);
			this.user = res.user;
			return res.user;
		} finally {
			this.loading = false;
		}
	}

	logout() {
		api.setToken(null);
		this.user = null;
	}

	updateUser(user: User, token?: string) {
		if (token) {
			api.setToken(token);
		}
		this.user = user;
	}

	hasRole(...roles: Role[]): boolean {
		if (!this.user) return false;
		return roles.includes(this.user.role);
	}
}

export const auth = new AuthStore();
