import { api } from '$lib/api/client';
import type { CashierShift, ShiftSummary, ShiftClosing } from '$lib/types';

class ShiftStore {
	currentShift = $state<CashierShift | null>(null);
	summary = $state<ShiftSummary | null>(null);
	loading = $state<boolean>(false);
	initialized = $state<boolean>(false);
	error = $state<string | null>(null);

	async init() {
		await this.loadCurrentShift();
	}

	async loadCurrentShift(): Promise<CashierShift | null> {
		this.loading = true;
		this.error = null;
		try {
			const shift = await api.get<CashierShift | null>('/cashier/shifts/current');
			this.currentShift = shift;
			if (shift && shift.status === 'OPEN') {
				await this.refreshSummary();
			} else {
				this.summary = null;
			}
			return shift;
		} catch (e: any) {
			console.error('Failed to load current shift:', e);
			this.error = e?.message || 'Gagal memuat shift kasir';
			this.currentShift = null;
			this.summary = null;
			return null;
		} finally {
			this.loading = false;
			this.initialized = true;
		}
	}

	async refreshSummary(): Promise<ShiftSummary | null> {
		if (!this.currentShift) return null;
		try {
			const summary = await api.get<ShiftSummary>(`/cashier/shifts/${this.currentShift.id}/summary`);
			this.summary = summary;
			return summary;
		} catch (e: any) {
			console.error('Failed to load shift summary:', e);
			return null;
		}
	}

	async openShift(registerId: string, openingBalance: number = 0): Promise<CashierShift> {
		this.loading = true;
		this.error = null;
		try {
			const shift = await api.post<CashierShift>('/cashier/shifts', {
				register_id: registerId,
				opening_balance: openingBalance
			});
			this.currentShift = shift;
			await this.refreshSummary();
			return shift;
		} catch (e: any) {
			this.error = e?.message || 'Gagal membuka shift kasir';
			throw e;
		} finally {
			this.loading = false;
		}
	}

	async closeShift(actualAmount?: number, notes?: string): Promise<ShiftClosing> {
		if (!this.currentShift) {
			throw new Error('Tidak ada shift aktif yang dapat ditutup');
		}

		this.loading = true;
		this.error = null;
		try {
			const closing = await api.post<ShiftClosing>(`/cashier/shifts/${this.currentShift.id}/close`, {
				actual_amount: actualAmount,
				notes: notes
			});
			this.currentShift = null;
			this.summary = null;
			return closing;
		} catch (e: any) {
			this.error = e?.message || 'Gagal menutup shift kasir';
			throw e;
		} finally {
			this.loading = false;
		}
	}
}

export const shiftStore = new ShiftStore();
