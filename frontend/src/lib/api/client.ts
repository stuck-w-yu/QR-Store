import type { APIResponse } from '$lib/types';

function getBaseURL(): string {
	if (import.meta.env.VITE_API_URL) {
		return import.meta.env.VITE_API_URL;
	}
	// Check Vercel service binding when executing in server-side functions / SSR
	if (typeof process !== 'undefined' && process.env?.BACKEND_URL) {
		return `${process.env.BACKEND_URL.replace(/\/$/, '')}/api/v1`;
	}
	// In the browser, use relative path to route through Vercel's rewrite on the same domain
	if (typeof window !== 'undefined') {
		return '/api/v1';
	}
	return 'http://localhost:8080/api/v1';
}

const API_BASE = getBaseURL();

class APIClient {
	private token: string | null = null;

	constructor() {
		if (typeof window !== 'undefined') {
			this.token = localStorage.getItem('qr_store_token');
		}
	}

	setToken(token: string | null) {
		this.token = token;
		if (typeof window !== 'undefined') {
			if (token) {
				localStorage.setItem('qr_store_token', token);
			} else {
				localStorage.removeItem('qr_store_token');
			}
		}
	}

	getToken(): string | null {
		if (!this.token && typeof window !== 'undefined') {
			this.token = localStorage.getItem('qr_store_token');
		}
		return this.token;
	}

	private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
		const url = `${API_BASE}${endpoint}`;
		const headers = new Headers(options.headers || {});

		if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
			headers.set('Content-Type', 'application/json');
		}

		const token = this.getToken();
		if (token) {
			headers.set('Authorization', `Bearer ${token}`);
		}

		const response = await fetch(url, {
			...options,
			headers
		});

		let json: any = null;
		const contentType = response.headers.get('content-type') || '';
		if (contentType.includes('application/json')) {
			try {
				json = await response.json();
			} catch (err) {
				json = null;
			}
		}

		if (!response.ok || !json || !json.success) {
			let errMsg = json?.error?.message || json?.message;
			if (!errMsg) {
				const rawText = !json ? await response.text().catch(() => '') : '';
				errMsg = rawText || `HTTP ${response.status}: ${response.statusText || 'Terjadi kesalahan pada sistem'}`;
			}
			const errCode = json?.error?.code || (response.status ? `HTTP_${response.status}` : 'UNKNOWN_ERROR');
			const error = new Error(errMsg) as any;
			error.code = errCode;
			throw error;
		}

		return json.data as T;
	}

	get<T>(endpoint: string): Promise<T> {
		return this.request<T>(endpoint, { method: 'GET' });
	}

	post<T>(endpoint: string, body?: any): Promise<T> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	patch<T>(endpoint: string, body?: any): Promise<T> {
		return this.request<T>(endpoint, {
			method: 'PATCH',
			body: body ? JSON.stringify(body) : undefined
		});
	}

	delete<T>(endpoint: string): Promise<T> {
		return this.request<T>(endpoint, { method: 'DELETE' });
	}
}

export const api = new APIClient();

export function formatRupiah(amount: number): string {
	return new Intl.NumberFormat('id-ID', {
		style: 'currency',
		currency: 'IDR',
		minimumFractionDigits: 0,
		maximumFractionDigits: 0
	}).format(amount);
}
