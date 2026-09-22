import type { Menu, ModifierOption, PublicTableInfo } from '$lib/types';

export interface CartItem {
	key: string;
	menu: Menu;
	quantity: number;
	selectedOptions: ModifierOption[];
	notes: string;
	unitPrice: number;
	subtotal: number;
}

class CartStore {
	qrToken = $state<string>('');
	tableInfo = $state<PublicTableInfo | null>(null);
	items = $state<CartItem[]>([]);

	constructor() {
		if (typeof window !== 'undefined') {
			this.loadFromStorage();
		}
	}

	private getStorageKey(): string {
		return `qr_cart_${this.qrToken || 'default'}`;
	}

	private saveToStorage() {
		if (typeof window === 'undefined') return;
		try {
			localStorage.setItem(this.getStorageKey(), JSON.stringify(this.items));
		} catch (e) {
			console.error('Failed to save cart to storage', e);
		}
	}

	private loadFromStorage() {
		if (typeof window === 'undefined') return;
		try {
			const saved = localStorage.getItem(this.getStorageKey());
			if (saved) {
				this.items = JSON.parse(saved);
			}
		} catch (e) {
			this.items = [];
		}
	}

	setSession(token: string, info: PublicTableInfo) {
		this.qrToken = token;
		this.tableInfo = info;
		this.loadFromStorage();
	}

	addItem(menu: Menu, quantity: number, options: ModifierOption[], notes: string) {
		const sortedOptIds = [...options].map((o) => o.id).sort().join('_');
		const key = `${menu.id}:${sortedOptIds}`;

		let unitPrice = menu.price;
		for (const opt of options) {
			unitPrice += opt.additional_price;
		}

		const existingIndex = this.items.findIndex((i) => i.key === key);
		if (existingIndex >= 0) {
			this.items[existingIndex].quantity += quantity;
			this.items[existingIndex].subtotal = this.items[existingIndex].quantity * unitPrice;
			if (notes) {
				this.items[existingIndex].notes = notes;
			}
		} else {
			this.items.push({
				key,
				menu,
				quantity,
				selectedOptions: options,
				notes,
				unitPrice,
				subtotal: unitPrice * quantity
			});
		}

		this.saveToStorage();
	}

	updateQuantity(key: string, delta: number) {
		const index = this.items.findIndex((i) => i.key === key);
		if (index < 0) return;

		const newQty = this.items[index].quantity + delta;
		if (newQty <= 0) {
			this.items.splice(index, 1);
		} else {
			this.items[index].quantity = newQty;
			this.items[index].subtotal = newQty * this.items[index].unitPrice;
		}

		this.saveToStorage();
	}

	removeItem(key: string) {
		this.items = this.items.filter((i) => i.key !== key);
		this.saveToStorage();
	}

	clear() {
		this.items = [];
		this.saveToStorage();
	}

	get count(): number {
		return this.items.reduce((sum, item) => sum + item.quantity, 0);
	}

	get subtotal(): number {
		return this.items.reduce((sum, item) => sum + item.subtotal, 0);
	}

	get tax(): number {
		const pct = this.tableInfo?.restaurant?.tax_percent ?? 10;
		return Math.round((this.subtotal * pct) / 100);
	}

	get serviceCharge(): number {
		const pct = this.tableInfo?.restaurant?.service_percent ?? 0;
		return Math.round((this.subtotal * pct) / 100);
	}

	get total(): number {
		return this.subtotal + this.tax + this.serviceCharge;
	}
}

export const cart = new CartStore();
