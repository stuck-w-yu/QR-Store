import type { 
	PublicTableInfo, CategoryWithMenus, Menu, Order, 
	Table, Category, User, Payment, OrderStatus,
	Register, CashierShift, ShiftTransaction, ShiftClosing, Refund, ShiftSummary, ShiftReportDetail
} from '$lib/types';
import QRCode from 'qrcode';

// Cross-tab Realtime Bus via BroadcastChannel
const realtimeBus = typeof window !== 'undefined' ? new BroadcastChannel('qr_store_mock_realtime') : null;

export function mockPublish(channel: string, event: string, data: any) {
	if (realtimeBus) {
		realtimeBus.postMessage({ channel, event, data });
	}
}

export function mockSubscribe(targetChannel: string, callback: (event: { event: string; data: any }) => void) {
	if (!realtimeBus) return () => {};
	const handler = (ev: MessageEvent) => {
		if (ev.data && (ev.data.channel === targetChannel || targetChannel === '*')) {
			callback(ev.data);
		}
	};
	realtimeBus.addEventListener('message', handler);
	return () => realtimeBus.removeEventListener('message', handler);
}

// Initial Data
const initialRestaurant = {
	id: 'rst_nusantara',
	name: 'Resto Nusantara',
	slug: 'resto-nusantara',
	logo_url: 'https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?w=500&q=80',
	address: 'Jl. Malioboro No. 45, Yogyakarta',
	phone: '081234567890',
	tax_percent: 10,
	service_percent: 5,
	status: 'ACTIVE'
};

const initialTables: Table[] = [
	{ id: 'tbl_01', restaurant_id: 'rst_nusantara', name: 'Meja 01', qr_token: 'demo-qr-token-table-01', status: 'ACTIVE' },
	{ id: 'tbl_02', restaurant_id: 'rst_nusantara', name: 'Meja 02', qr_token: 'demo-qr-token-table-02', status: 'ACTIVE' },
	{ id: 'tbl_03', restaurant_id: 'rst_nusantara', name: 'Meja 03', qr_token: 'demo-qr-token-table-03', status: 'ACTIVE' },
	{ id: 'tbl_04', restaurant_id: 'rst_nusantara', name: 'Meja 04', qr_token: 'demo-qr-token-table-04', status: 'ACTIVE' }
];

const initialCategories: Category[] = [
	{ id: 'cat_food', restaurant_id: 'rst_nusantara', name: 'Makanan Utama', sort_order: 1, status: 'ACTIVE' },
	{ id: 'cat_drink', restaurant_id: 'rst_nusantara', name: 'Minuman Segar', sort_order: 2, status: 'ACTIVE' },
	{ id: 'cat_snack', restaurant_id: 'rst_nusantara', name: 'Camilan & Penutup', sort_order: 3, status: 'ACTIVE' }
];

const initialMenus: Menu[] = [
	{
		id: 'menu_nasgor',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_food',
		category_name: 'Makanan Utama',
		name: 'Nasi Goreng Spesial Nusantara',
		description: 'Nasi goreng bumbu rempah dengan suwiran ayam, telur, acar segar, dan kerupuk udang.',
		price: 28000,
		image_url: 'https://images.unsplash.com/photo-1603133872878-684f208fb84b?w=600&q=80',
		available: true,
		sort_order: 1,
		modifiers: [
			{
				id: 'mod_pedas',
				restaurant_id: 'rst_nusantara',
				name: 'Tingkat Kepedasan',
				type: 'SINGLE',
				required: true,
				options: [
					{ id: 'opt_p0', modifier_id: 'mod_pedas', name: 'Level 0 (Tidak Pedas)', additional_price: 0 },
					{ id: 'opt_p1', modifier_id: 'mod_pedas', name: 'Level 1 (Sedang)', additional_price: 0 },
					{ id: 'opt_p2', modifier_id: 'mod_pedas', name: 'Level 2 (Pedas Banget)', additional_price: 2000 }
				]
			},
			{
				id: 'mod_top',
				restaurant_id: 'rst_nusantara',
				name: 'Extra Topping',
				type: 'MULTIPLE',
				required: false,
				options: [
					{ id: 'opt_egg', modifier_id: 'mod_top', name: 'Telur Mata Sapi', additional_price: 5000 },
					{ id: 'opt_cheese', modifier_id: 'mod_top', name: 'Keju Mozzarella', additional_price: 6000 }
				]
			}
		]
	},
	{
		id: 'menu_ayambakar',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_food',
		category_name: 'Makanan Utama',
		name: 'Ayam Bakar Madu Pedas',
		description: 'Paha ayam bakar bumbu madu karamel dengan sambal terasi khas dan lalapan segar.',
		price: 34000,
		image_url: 'https://images.unsplash.com/photo-1598515214211-89d3c73ae83b?w=600&q=80',
		available: true,
		sort_order: 2
	},
	{
		id: 'menu_miegor',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_food',
		category_name: 'Makanan Utama',
		name: 'Mie Goreng Seafood Jawa',
		description: 'Mie pipih kenyal dengan udang, cumi, bakso ikan, dan sayuran segar.',
		price: 30000,
		image_url: 'https://images.unsplash.com/photo-1585032226651-759b368d7246?w=600&q=80',
		available: true,
		sort_order: 3
	},
	{
		id: 'menu_esteh',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_drink',
		category_name: 'Minuman Segar',
		name: 'Es Teh Manis Melati',
		description: 'Seduhan teh melati harum khas Jawa dengan gula tebu murni.',
		price: 6000,
		image_url: 'https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=600&q=80',
		available: true,
		sort_order: 4
	},
	{
		id: 'menu_alpukat',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_drink',
		category_name: 'Minuman Segar',
		name: 'Jus Alpukat Kocok Cokelat',
		description: 'Alpukat mentega legit dikocok lembut dengan siraman kental manis cokelat.',
		price: 18000,
		image_url: 'https://images.unsplash.com/photo-1600271886742-f049cd451bba?w=600&q=80',
		available: true,
		sort_order: 5
	},
	{
		id: 'menu_pisang',
		restaurant_id: 'rst_nusantara',
		category_id: 'cat_snack',
		category_name: 'Camilan & Penutup',
		name: 'Pisang Goreng Keju Karamel',
		description: 'Pisang raja krispi dengan taburan keju cheddar melimpah dan saus karamel.',
		price: 20000,
		image_url: 'https://images.unsplash.com/photo-1528735602780-2552fd46c7af?w=600&q=80',
		available: true,
		sort_order: 6
	}
];

const initialUsers: User[] = [
	{ id: 'usr_owner', restaurant_id: 'rst_nusantara', name: 'Budi Owner', email: 'owner@resto.com', role: 'OWNER', status: 'ACTIVE' },
	{ id: 'usr_admin', restaurant_id: 'rst_nusantara', name: 'Siti Admin', email: 'admin@resto.com', role: 'ADMIN', status: 'ACTIVE' },
	{ id: 'usr_cashier', restaurant_id: 'rst_nusantara', name: 'Rian Kasir', email: 'cashier@resto.com', role: 'CASHIER', status: 'ACTIVE' },
	{ id: 'usr_kitchen', restaurant_id: 'rst_nusantara', name: 'Chef Joko', email: 'kitchen@resto.com', role: 'KITCHEN', status: 'ACTIVE' }
];

const initialOrders: Order[] = [
	{
		id: 'ord_demo_01',
		restaurant_id: 'rst_nusantara',
		table_id: 'tbl_02',
		table_name: 'Meja 02',
		order_number: 'ORD-DEMO-001',
		status: 'CONFIRMED',
		subtotal: 58000,
		tax: 5800,
		service_charge: 2900,
		discount: 0,
		total: 66700,
		notes: 'Jangan terlalu pedas',
		items: [
			{
				id: 'it_1',
				order_id: 'ord_demo_01',
				menu_name_snapshot: 'Nasi Goreng Spesial Nusantara',
				unit_price: 28000,
				quantity: 2,
				subtotal: 56000,
				notes: 'Level 1'
			}
		],
		created_at: new Date(Date.now() - 6 * 60000).toISOString(),
		updated_at: new Date(Date.now() - 6 * 60000).toISOString()
	},
	{
		id: 'ord_demo_02',
		restaurant_id: 'rst_nusantara',
		table_id: 'tbl_03',
		table_name: 'Meja 03',
		order_number: 'ORD-DEMO-002',
		status: 'PREPARING',
		subtotal: 40000,
		tax: 4000,
		service_charge: 2000,
		discount: 0,
		total: 46000,
		items: [
			{
				id: 'it_2',
				order_id: 'ord_demo_02',
				menu_name_snapshot: 'Ayam Bakar Madu Pedas',
				unit_price: 34000,
				quantity: 1,
				subtotal: 34000
			},
			{
				id: 'it_3',
				order_id: 'ord_demo_02',
				menu_name_snapshot: 'Es Teh Manis Melati',
				unit_price: 6000,
				quantity: 1,
				subtotal: 6000
			}
		],
		created_at: new Date(Date.now() - 14 * 60000).toISOString(),
		updated_at: new Date(Date.now() - 10 * 60000).toISOString()
	}
];

// Local Storage Store Manager
class MockDatabase {
	private get<T>(key: string, fallback: T): T {
		if (typeof window === 'undefined') return fallback;
		try {
			const saved = localStorage.getItem(`mock_db_${key}`);
			return saved ? JSON.parse(saved) : fallback;
		} catch {
			return fallback;
		}
	}

	private set<T>(key: string, data: T) {
		if (typeof window === 'undefined') return;
		try {
			localStorage.setItem(`mock_db_${key}`, JSON.stringify(data));
		} catch (e) {
			console.error(e);
		}
	}

	get tables(): Table[] { return this.get('tables', initialTables); }
	set tables(v: Table[]) { this.set('tables', v); }

	get categories(): Category[] { return this.get('categories', initialCategories); }
	set categories(v: Category[]) { this.set('categories', v); }

	get menus(): Menu[] { return this.get('menus', initialMenus); }
	set menus(v: Menu[]) { this.set('menus', v); }

	get orders(): Order[] { return this.get('orders', initialOrders); }
	set orders(v: Order[]) { this.set('orders', v); }

	get users(): User[] { return this.get('users', initialUsers); }
	set users(v: User[]) { this.set('users', v); }

	get registers(): Register[] {
		return this.get('registers', [
			{ id: 'reg_pos_01', restaurant_id: initialRestaurant.id, name: 'POS 01 - Kasir Utama', status: 'ACTIVE', created_at: new Date().toISOString(), updated_at: new Date().toISOString() },
			{ id: 'reg_pos_02', restaurant_id: initialRestaurant.id, name: 'POS 02 - Kasir Bar', status: 'ACTIVE', created_at: new Date().toISOString(), updated_at: new Date().toISOString() }
		]);
	}
	set registers(v: Register[]) { this.set('registers', v); }

	get shifts(): CashierShift[] { return this.get('shifts', []); }
	set shifts(v: CashierShift[]) { this.set('shifts', v); }

	get shiftTransactions(): ShiftTransaction[] { return this.get('shift_transactions', []); }
	set shiftTransactions(v: ShiftTransaction[]) { this.set('shift_transactions', v); }

	get shiftClosings(): ShiftClosing[] { return this.get('shift_closings', []); }
	set shiftClosings(v: ShiftClosing[]) { this.set('shift_closings', v); }

	get refunds(): Refund[] { return this.get('refunds', []); }
	set refunds(v: Refund[]) { this.set('refunds', v); }
}

export const mockDB = new MockDatabase();

// Mock API Dispatcher
export async function handleMockRequest<T>(endpoint: string, method: string = 'GET', body?: any): Promise<T> {
	// 1. Table Scan
	if (endpoint.startsWith('/public/tables/')) {
		const token = endpoint.replace('/public/tables/', '');
		const table = mockDB.tables.find(t => t.qr_token === token) || mockDB.tables[0];
		return {
			restaurant: initialRestaurant,
			table: table
		} as T;
	}

	// 2. Catalog
	if (endpoint.startsWith('/public/restaurants/') && endpoint.endsWith('/menu')) {
		const cats = mockDB.categories;
		const allMenus = mockDB.menus;
		const result: CategoryWithMenus[] = cats.map(c => ({
			...c,
			menus: allMenus.filter(m => m.category_id === c.id && m.available)
		}));
		return result as T;
	}

	// 3. Create Public Order
	if (endpoint === '/public/orders' && method === 'POST') {
		const token = body.qr_token;
		const targetTable = mockDB.tables.find(t => t.qr_token === token) || mockDB.tables[0];
		
		let subtotal = 0;
		const orderItems = (body.items || []).map((itemReq: any, idx: number) => {
			const m = mockDB.menus.find(m => m.id === itemReq.menu_id) || mockDB.menus[0];
			let unitPrice = m.price;
			const selectedOptions: any[] = [];

			if (m.modifiers) {
				for (const mod of m.modifiers) {
					for (const opt of mod.options) {
						if (itemReq.modifier_option_ids?.includes(opt.id)) {
							unitPrice += opt.additional_price;
							selectedOptions.push({
								modifier_id: mod.id,
								modifier_name: mod.name,
								option_id: opt.id,
								option_name: opt.name,
								additional_price: opt.additional_price
							});
						}
					}
				}
			}

			const itemSubtotal = unitPrice * itemReq.quantity;
			subtotal += itemSubtotal;

			return {
				id: `it_${Date.now()}_${idx}`,
				order_id: '',
				menu_id: m.id,
				menu_name_snapshot: m.name,
				unit_price: unitPrice,
				quantity: itemReq.quantity,
				subtotal: itemSubtotal,
				selected_modifiers: selectedOptions,
				notes: itemReq.notes
			};
		});

		const tax = Math.round((subtotal * initialRestaurant.tax_percent) / 100);
		const service = Math.round((subtotal * initialRestaurant.service_percent) / 100);
		const total = subtotal + tax + service;

		const orderId = `ord_${Date.now()}`;
		const orderNum = `ORD-${new Date().toISOString().slice(2, 10).replace(/-/g, '')}-${Math.floor(1000 + Math.random() * 9000)}`;

		const newOrder: Order = {
			id: orderId,
			restaurant_id: initialRestaurant.id,
			table_id: targetTable.id,
			table_name: targetTable.name,
			order_number: orderNum,
			status: 'WAITING_PAYMENT',
			subtotal,
			tax,
			service_charge: service,
			discount: 0,
			total,
			notes: body.notes,
			items: orderItems,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		};

		const orders = mockDB.orders;
		orders.unshift(newOrder);
		mockDB.orders = orders;

		return newOrder as T;
	}

	// 4. Get Order Detail
	if (endpoint.startsWith('/public/orders/')) {
		const id = endpoint.replace('/public/orders/', '');
		const order = mockDB.orders.find(o => o.id === id);
		if (!order) throw new Error('Order not found');
		return order as T;
	}

	// 5. Payment Session Creation
	if (endpoint.startsWith('/orders/') && endpoint.endsWith('/payment')) {
		const orderId = endpoint.split('/')[2];
		const order = mockDB.orders.find(o => o.id === orderId);
		if (!order) throw new Error('Order not found');

		const qrString = `00020101021226600016ID.CO.QRSTORE.WWW011893600914${order.id}520458125303360540${order.total}5802ID5913RestoNusantara6007JAKARTA`;

		const payment: Payment = {
			id: `pay_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			order_id: orderId,
			provider: 'mock',
			amount: order.total,
			status: 'PENDING',
			qr_string: qrString,
			expired_at: new Date(Date.now() + 15 * 60000).toISOString()
		};
		return payment as T;
	}

	// 6. Simulate Pay
	if (endpoint === '/payments/simulate-pay' && method === 'POST') {
		const orderId = body.order_id;
		const orders = mockDB.orders;
		const idx = orders.findIndex(o => o.id === orderId);
		if (idx >= 0) {
			orders[idx].status = 'CONFIRMED';
			orders[idx].updated_at = new Date().toISOString();
			mockDB.orders = orders;

			// Broadcast realtime event
			mockPublish(`order:${orderId}`, 'ORDER_STATUS_CHANGED', { order: orders[idx] });
			mockPublish(`restaurant:${initialRestaurant.id}:kitchen`, 'NEW_ORDER_CONFIRMED', { order: orders[idx] });

			// Record in active cashier shift if open
			const activeShift = mockDB.shifts.find(s => s.status === 'OPEN');
			if (activeShift) {
				const txs = mockDB.shiftTransactions;
				const newTx: ShiftTransaction = {
					id: `cst_${Date.now()}`,
					shift_id: activeShift.id,
					order_id: orderId,
					type: 'SALE',
					amount: orders[idx].total,
					metadata: { payment_method: 'QRIS' },
					created_at: new Date().toISOString()
				};
				txs.unshift(newTx);
				mockDB.shiftTransactions = txs;
				mockPublish(`shift:${activeShift.id}`, 'SHIFT_TRANSACTION_ADDED', newTx);
			}

			return { success: true, status: 'PAID', order: orders[idx] } as T;
		}
		throw new Error('Order not found');
	}

	// 7. Kitchen Orders
	if (endpoint === '/kitchen/orders') {
		const active = mockDB.orders.filter(o => ['CONFIRMED', 'PREPARING', 'READY'].includes(o.status));
		return active as T;
	}

	// 8. Kitchen Status Updates
	if (endpoint.startsWith('/kitchen/orders/') && endpoint.endsWith('/accept')) {
		const id = endpoint.split('/')[3];
		const orders = mockDB.orders;
		const o = orders.find(x => x.id === id);
		if (o) {
			o.status = 'PREPARING';
			mockDB.orders = orders;
			mockPublish(`order:${id}`, 'ORDER_STATUS_CHANGED', { order: o });
			mockPublish(`restaurant:${initialRestaurant.id}:kitchen`, 'ORDER_STATUS_CHANGED', { order: o });
			return { status: 'PREPARING' } as T;
		}
	}

	if (endpoint.startsWith('/kitchen/orders/') && endpoint.endsWith('/ready')) {
		const id = endpoint.split('/')[3];
		const orders = mockDB.orders;
		const o = orders.find(x => x.id === id);
		if (o) {
			o.status = 'READY';
			mockDB.orders = orders;
			mockPublish(`order:${id}`, 'ORDER_STATUS_CHANGED', { order: o });
			mockPublish(`restaurant:${initialRestaurant.id}:kitchen`, 'ORDER_STATUS_CHANGED', { order: o });
			return { status: 'READY' } as T;
		}
	}

	if (endpoint.startsWith('/kitchen/orders/') && endpoint.endsWith('/complete')) {
		const id = endpoint.split('/')[3];
		const orders = mockDB.orders;
		const o = orders.find(x => x.id === id);
		if (o) {
			o.status = 'COMPLETED';
			mockDB.orders = orders;
			mockPublish(`order:${id}`, 'ORDER_STATUS_CHANGED', { order: o });
			mockPublish(`restaurant:${initialRestaurant.id}:kitchen`, 'ORDER_STATUS_CHANGED', { order: o });
			return { status: 'COMPLETED' } as T;
		}
	}

	// 9. Admin Orders List & Analytics
	if (endpoint.startsWith('/orders/analytics/today')) {
		const orders = mockDB.orders;
		const paidOrders = orders.filter(o => o.status !== 'CANCELLED' && o.status !== 'WAITING_PAYMENT');
		const revenue = paidOrders.reduce((sum, o) => sum + o.total, 0);
		return {
			today_revenue: revenue,
			today_orders: orders.length,
			today_paid: paidOrders.length,
			today_cancelled: orders.filter(o => o.status === 'CANCELLED').length,
			average_order: paidOrders.length > 0 ? Math.round(revenue / paidOrders.length) : 0
		} as T;
	}

	if (endpoint.startsWith('/orders')) {
		const url = new URL(`http://localhost${endpoint}`);
		const status = url.searchParams.get('status');
		let list = mockDB.orders;
		if (status) {
			list = list.filter(o => o.status === status);
		}
		return list as T;
	}

	// 10. Tables CRUD
	if (endpoint === '/tables' && method === 'GET') {
		return mockDB.tables as T;
	}

	if (endpoint === '/tables' && method === 'POST') {
		const tables = mockDB.tables;
		const newT: Table = {
			id: `tbl_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			name: body.name,
			qr_token: `demo-qr-${Date.now().toString(36)}`,
			status: 'ACTIVE'
		};
		tables.push(newT);
		mockDB.tables = tables;
		return newT as T;
	}

	if (endpoint.startsWith('/tables/') && endpoint.endsWith('/qr/regenerate')) {
		const id = endpoint.split('/')[2];
		const tables = mockDB.tables;
		const t = tables.find(x => x.id === id);
		if (t) {
			t.qr_token = `demo-qr-${Date.now().toString(36)}`;
			mockDB.tables = tables;
			return { qr_token: t.qr_token } as T;
		}
	}

	if (endpoint.startsWith('/tables/') && method === 'DELETE') {
		const id = endpoint.split('/')[2];
		mockDB.tables = mockDB.tables.filter(t => t.id !== id);
		return { deleted: true } as T;
	}

	// 11. Categories & Menus
	if (endpoint === '/categories' && method === 'GET') {
		return mockDB.categories as T;
	}
	if (endpoint === '/categories' && method === 'POST') {
		const cats = mockDB.categories;
		const newCat: Category = {
			id: `cat_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			name: body.name,
			sort_order: cats.length + 1,
			status: 'ACTIVE'
		};
		cats.push(newCat);
		mockDB.categories = cats;
		return newCat as T;
	}

	if (endpoint.startsWith('/menus') && method === 'GET') {
		return mockDB.menus as T;
	}
	if (endpoint === '/menus' && method === 'POST') {
		const menus = mockDB.menus;
		const cat = mockDB.categories.find(c => c.id === body.category_id);
		const newM: Menu = {
			id: `menu_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			category_id: body.category_id,
			category_name: cat?.name || 'Makanan',
			name: body.name,
			description: body.description,
			price: body.price,
			image_url: body.image_url || 'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=600&q=80',
			available: true,
			sort_order: menus.length + 1
		};
		menus.push(newM);
		mockDB.menus = menus;
		return newM as T;
	}
	if (endpoint.startsWith('/menus/') && endpoint.endsWith('/availability')) {
		const id = endpoint.split('/')[2];
		const menus = mockDB.menus;
		const m = menus.find(x => x.id === id);
		if (m) {
			m.available = body.available;
			mockDB.menus = menus;
			return { available: m.available } as T;
		}
	}
	if (endpoint.startsWith('/menus/') && method === 'DELETE') {
		const id = endpoint.split('/')[2];
		mockDB.menus = mockDB.menus.filter(m => m.id !== id);
		return { deleted: true } as T;
	}

	// 12. Auth
	if (endpoint === '/auth/login' && method === 'POST') {
		const user = mockDB.users.find(u => u.email === body.email) || mockDB.users[0];
		return {
			user,
			access_token: `mock_jwt_token_${user.role.toLowerCase()}`
		} as T;
	}
	if (endpoint === '/auth/me') {
		return mockDB.users[0] as T;
	}

	// 13. Users
	if (endpoint === '/users' && method === 'GET') {
		return mockDB.users as T;
	}
	if (endpoint === '/users' && method === 'POST') {
		const users = mockDB.users;
		const newU: User = {
			id: `usr_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			name: body.name,
			email: body.email,
			role: body.role,
			status: 'ACTIVE'
		};
		users.push(newU);
		mockDB.users = users;
		return newU as T;
	}
	if (endpoint.startsWith('/users/') && method === 'DELETE') {
		const id = endpoint.split('/')[2];
		mockDB.users = mockDB.users.filter(u => u.id !== id);
		return { deleted: true } as T;
	}

	// 14. Registers
	if (endpoint === '/registers' && method === 'GET') {
		return mockDB.registers as T;
	}
	if (endpoint === '/registers' && method === 'POST') {
		const regs = mockDB.registers;
		const newReg: Register = {
			id: `reg_${Date.now().toString(36)}`,
			restaurant_id: initialRestaurant.id,
			name: body.name,
			status: 'ACTIVE',
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		};
		regs.push(newReg);
		mockDB.registers = regs;
		return newReg as T;
	}
	if (endpoint.startsWith('/registers/') && method === 'GET') {
		const id = endpoint.split('/')[2];
		const reg = mockDB.registers.find(r => r.id === id);
		if (!reg) throw new Error('Register not found');
		return reg as T;
	}
	if (endpoint.startsWith('/registers/') && method === 'PATCH') {
		const id = endpoint.split('/')[2];
		const regs = mockDB.registers;
		const r = regs.find(x => x.id === id);
		if (r) {
			if (body.name !== undefined) r.name = body.name;
			if (body.status !== undefined) r.status = body.status;
			r.updated_at = new Date().toISOString();
			mockDB.registers = regs;
			return r as T;
		}
		throw new Error('Register not found');
	}
	if (endpoint.startsWith('/registers/') && method === 'DELETE') {
		const id = endpoint.split('/')[2];
		mockDB.registers = mockDB.registers.filter(r => r.id !== id);
		return { deleted: true } as T;
	}

	// 15. Cashier Shifts
	if (endpoint === '/cashier/shifts/current' && method === 'GET') {
		const active = mockDB.shifts.find(s => s.status === 'OPEN');
		return (active || null) as T;
	}

	if (endpoint === '/cashier/shifts' && method === 'POST') {
		const shifts = mockDB.shifts;
		// Ensure only 1 active shift per register
		const existingActive = shifts.find(s => s.status === 'OPEN' && s.register_id === body.register_id);
		if (existingActive) {
			throw new Error('Shift aktif sudah terbuka pada mesin kasir ini');
		}

		const reg = mockDB.registers.find(r => r.id === body.register_id);
		const user = mockDB.users[0]; // Active logged-in user

		const newShift: CashierShift = {
			id: `shf_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			register_id: body.register_id,
			register_name: reg?.name || 'POS 01',
			cashier_id: user.id,
			cashier_name: user.name,
			status: 'OPEN',
			opening_balance: body.opening_balance || 0,
			opened_at: new Date().toISOString(),
			expected_total: body.opening_balance || 0,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		};

		shifts.unshift(newShift);
		mockDB.shifts = shifts;
		mockPublish(`restaurant:${initialRestaurant.id}:cashier`, 'CASHIER_SHIFT_OPENED', newShift);
		return newShift as T;
	}

	if (endpoint.startsWith('/cashier/shifts/') && endpoint.endsWith('/transactions') && method === 'GET') {
		const shiftId = endpoint.split('/')[3];
		const txs = mockDB.shiftTransactions.filter(t => t.shift_id === shiftId);
		return {
			transactions: txs,
			total: txs.length,
			page: 1,
			limit: 50
		} as T;
	}

	if (endpoint.startsWith('/cashier/shifts/') && endpoint.endsWith('/summary') && method === 'GET') {
		const shiftId = endpoint.split('/')[3];
		const shift = mockDB.shifts.find(s => s.id === shiftId);
		if (!shift) throw new Error('Shift not found');

		const txs = mockDB.shiftTransactions.filter(t => t.shift_id === shiftId);
		let gross = 0;
		let refund = 0;
		let voidVal = 0;
		let adjustment = 0;
		const paymentMethods: Record<string, number> = { 'QRIS': 0, 'E-Wallet': 0, 'Virtual Account': 0 };

		for (const t of txs) {
			if (t.type === 'SALE') {
				gross += t.amount;
				const m = (t.metadata?.payment_method as string) || 'QRIS';
				paymentMethods[m] = (paymentMethods[m] || 0) + t.amount;
			} else if (t.type === 'REFUND') {
				refund += t.amount;
			} else if (t.type === 'VOID') {
				voidVal += t.amount;
			} else if (t.type === 'ADJUSTMENT') {
				adjustment += t.amount;
			}
		}

		const net = gross - refund - voidVal + adjustment;
		const summary: ShiftSummary = {
			shift_id: shift.id,
			status: shift.status,
			register_id: shift.register_id,
			register_name: shift.register_name || 'POS 01',
			cashier_id: shift.cashier_id,
			cashier_name: shift.cashier_name || 'Kasir',
			opening_balance: shift.opening_balance,
			opened_at: shift.opened_at,
			closed_at: shift.closed_at,
			orders_count: txs.filter(t => t.type === 'SALE').length,
			gross_sales: gross,
			refund_total: refund,
			void_total: voidVal,
			adjustment_total: adjustment,
			net_sales: net,
			expected_amount: net + shift.opening_balance,
			payment_methods: paymentMethods
		};
		return summary as T;
	}

	if (endpoint.startsWith('/cashier/shifts/') && endpoint.endsWith('/close') && method === 'POST') {
		const shiftId = endpoint.split('/')[3];
		const shifts = mockDB.shifts;
		const shift = shifts.find(s => s.id === shiftId);
		if (!shift) throw new Error('Shift not found');
		if (shift.status === 'CLOSED') throw new Error('Shift sudah ditutup sebelumnya');

		const summaryResp = await handleMockRequest<ShiftSummary>(`/cashier/shifts/${shiftId}/summary`, 'GET');

		const now = new Date().toISOString();
		const actualAmount = body.actual_amount !== undefined ? body.actual_amount : summaryResp.expected_amount;
		const diff = actualAmount - summaryResp.expected_amount;

		shift.status = 'CLOSED';
		shift.closed_at = now;
		shift.actual_total = actualAmount;
		shift.difference = diff;
		shift.updated_at = now;
		mockDB.shifts = shifts;

		const closings = mockDB.shiftClosings;
		const newClosing: ShiftClosing = {
			id: `cls_${Date.now()}`,
			shift_id: shiftId,
			gross_sales: summaryResp.gross_sales,
			refund_total: summaryResp.refund_total,
			void_total: summaryResp.void_total,
			adjustment_total: summaryResp.adjustment_total,
			net_sales: summaryResp.net_sales,
			expected_amount: summaryResp.expected_amount,
			actual_amount: actualAmount,
			difference: diff,
			notes: body.notes || '',
			closed_by: shift.cashier_id,
			closed_by_name: shift.cashier_name,
			closed_at: now
		};
		closings.unshift(newClosing);
		mockDB.shiftClosings = closings;

		mockPublish(`restaurant:${initialRestaurant.id}:cashier`, 'CASHIER_SHIFT_CLOSED', newClosing);
		return newClosing as T;
	}

	if (endpoint.startsWith('/cashier/shifts/') && method === 'GET') {
		const id = endpoint.split('/')[3];
		const shift = mockDB.shifts.find(s => s.id === id);
		if (!shift) throw new Error('Shift not found');
		return shift as T;
	}

	// 16. Orders Refund & Void
	if (endpoint.startsWith('/orders/') && endpoint.endsWith('/refund') && method === 'POST') {
		const orderId = endpoint.split('/')[2];
		const o = mockDB.orders.find(x => x.id === orderId);
		if (!o) throw new Error('Order not found');

		const refunds = mockDB.refunds;
		const newRef: Refund = {
			id: `ref_${Date.now()}`,
			restaurant_id: initialRestaurant.id,
			order_id: orderId,
			order_number: o.order_number,
			payment_id: `pay_${orderId}`,
			amount: body.amount || o.total,
			reason: body.reason || 'Customer refund',
			status: 'COMPLETED',
			requested_by: mockDB.users[0].id,
			requested_by_name: mockDB.users[0].name,
			created_at: new Date().toISOString(),
			completed_at: new Date().toISOString()
		};
		refunds.unshift(newRef);
		mockDB.refunds = refunds;

		// Record in active shift
		const activeShift = mockDB.shifts.find(s => s.status === 'OPEN');
		if (activeShift) {
			const txs = mockDB.shiftTransactions;
			const newTx: ShiftTransaction = {
				id: `cst_${Date.now()}`,
				shift_id: activeShift.id,
				order_id: orderId,
				type: 'REFUND',
				amount: newRef.amount,
				metadata: { reason: newRef.reason },
				created_at: new Date().toISOString()
			};
			txs.unshift(newTx);
			mockDB.shiftTransactions = txs;
			mockPublish(`shift:${activeShift.id}`, 'SHIFT_TRANSACTION_ADDED', newTx);
		}

		return newRef as T;
	}

	if (endpoint.startsWith('/orders/') && endpoint.endsWith('/void') && method === 'POST') {
		const orderId = endpoint.split('/')[2];
		const orders = mockDB.orders;
		const o = orders.find(x => x.id === orderId);
		if (!o) throw new Error('Order not found');

		o.status = 'CANCELLED';
		o.updated_at = new Date().toISOString();
		mockDB.orders = orders;

		// Record VOID in active shift
		const activeShift = mockDB.shifts.find(s => s.status === 'OPEN');
		if (activeShift) {
			const txs = mockDB.shiftTransactions;
			const newTx: ShiftTransaction = {
				id: `cst_${Date.now()}`,
				shift_id: activeShift.id,
				order_id: orderId,
				type: 'VOID',
				amount: o.total,
				metadata: { reason: body.reason || 'Voided by cashier' },
				created_at: new Date().toISOString()
			};
			txs.unshift(newTx);
			mockDB.shiftTransactions = txs;
			mockPublish(`shift:${activeShift.id}`, 'SHIFT_TRANSACTION_ADDED', newTx);
		}

		return { voided: true } as T;
	}

	// 17. Reports
	if (endpoint.startsWith('/reports/shifts/') && method === 'GET') {
		const shiftId = endpoint.split('/')[3];
		const shift = mockDB.shifts.find(s => s.id === shiftId);
		if (!shift) throw new Error('Shift not found');
		const summary = await handleMockRequest<ShiftSummary>(`/cashier/shifts/${shiftId}/summary`, 'GET');
		const closing = mockDB.shiftClosings.find(c => c.shift_id === shiftId);
		const txs = mockDB.shiftTransactions.filter(t => t.shift_id === shiftId);

		return {
			shift,
			summary,
			closing,
			transactions: txs
		} as T;
	}

	if (endpoint.startsWith('/reports/shifts') && method === 'GET') {
		return mockDB.shifts as T;
	}

	throw new Error(`Mock endpoint not implemented: ${method} ${endpoint}`);
}

// Generate QR Code as Data URL directly in browser
export async function getTableQRCodeDataURL(qrToken: string): Promise<string> {
	const currentOrigin = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173';
	const url = `${currentOrigin}/order?token=${qrToken}`;
	return await QRCode.toDataURL(url, {
		width: 300,
		margin: 2,
		color: {
			dark: '#1e293b',
			light: '#ffffff'
		}
	});
}
