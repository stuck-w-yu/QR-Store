import type { APIResponse } from '$lib/types';
import { handleMockRequest } from './mock';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
// Default to mock mode for frontend-only demo unless explicitly set to 'false'
const FORCE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false';

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
		// 1. If mock mode is active, directly dispatch to mock DB
		if (FORCE_MOCK) {
			const method = options.method || 'GET';
			const body = options.body ? JSON.parse(options.body as string) : undefined;
			// Simulated network delay (80ms) for smooth feel
			await new Promise((r) => setTimeout(r, 80));
			return handleMockRequest<T>(endpoint, method, body);
		}

		// 2. Real Backend Request with automatic fallback to mock if backend is down
		const url = `${API_BASE}${endpoint}`;
		const headers = new Headers(options.headers || {});

		if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
			headers.set('Content-Type', 'application/json');
		}

		const token = this.getToken();
		if (token) {
			headers.set('Authorization', `Bearer ${token}`);
		}

		try {
			const response = await fetch(url, {
				...options,
				headers
			});

			const json = await response.json();

			if (!response.ok || !json.success) {
				const errMsg = json?.error?.message || json?.message || 'Terjadi kesalahan pada sistem';
				const errCode = json?.error?.code || 'UNKNOWN_ERROR';
				const error = new Error(errMsg) as any;
				error.code = errCode;
				throw error;
			}

			return json.data as T;
		} catch (err: any) {
			// Fallback to mock on connection refused
			if (err instanceof TypeError || err.message?.includes('fetch')) {
				console.warn(`[API] Backend unreachable at ${url}, falling back to local Mock Service.`);
				const method = options.method || 'GET';
				const body = options.body ? JSON.parse(options.body as string) : undefined;
				return handleMockRequest<T>(endpoint, method, body);
			}
			throw err;
		}
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
