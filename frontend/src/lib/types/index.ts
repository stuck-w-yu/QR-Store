export type Role = 'SUPERADMIN' | 'OWNER' | 'ADMIN' | 'CASHIER' | 'KITCHEN';

export interface User {
	id: string;
	restaurant_id: string;
	name: string;
	email: string;
	role: Role;
	status: string;
}

export interface Restaurant {
	id: string;
	name: string;
	slug: string;
	logo_url?: string;
	address?: string;
	phone?: string;
	tax_percent: number;
	service_percent: number;
	status: string;
	plan?: string;
	created_at?: string;
	updated_at?: string;
}

export interface TenantSummary extends Restaurant {
	plan: string;
	owner_name: string;
	owner_email: string;
	owner_phone: string;
	total_tables: number;
	total_menus: number;
	total_orders: number;
	total_revenue: number;
}

export interface PlatformStats {
	total_restaurants: number;
	active_restaurants: number;
	suspended_restaurants: number;
	total_owners: number;
	total_orders: number;
	total_revenue: number;
	total_tables?: number;
	total_menus?: number;
	today_revenue?: number;
	today_orders?: number;
	plan_distribution?: Record<string, number>;
}

export interface PlatformOrder {
	id: string;
	restaurant_id: string;
	restaurant_name: string;
	restaurant_slug: string;
	table_name: string;
	order_number: string;
	status: string;
	payment_status: string;
	payment_method: string;
	total: number;
	created_at: string;
}

export interface PlatformOwner {
	id: string;
	name: string;
	email: string;
	status: string;
	created_at: string;
	restaurant_id: string;
	restaurant_name: string;
	restaurant_slug: string;
	restaurant_plan: string;
	phone: string;
}

export interface SystemHealth {
	status: string;
	database: string;
	pool_total_conns: number;
	pool_idle_conns: number;
	goroutines: number;
	memory_alloc_mb: number;
	uptime_seconds: number;
	timestamp: string;
}

export interface Table {
	id: string;
	restaurant_id: string;
	name: string;
	qr_token: string;
	status: string;
}

export interface PublicTableInfo {
	restaurant: {
		id: string;
		name: string;
		slug: string;
		logo_url?: string;
		tax_percent: number;
		service_percent: number;
	};
	table: {
		id: string;
		name: string;
		qr_token: string;
		status: string;
	};
}

export interface ModifierOption {
	id: string;
	modifier_id: string;
	name: string;
	additional_price: number;
}

export interface Modifier {
	id: string;
	restaurant_id: string;
	menu_id?: string;
	name: string;
	type: 'SINGLE' | 'MULTIPLE';
	required: boolean;
	options: ModifierOption[];
}

export interface Menu {
	id: string;
	restaurant_id: string;
	category_id?: string;
	category_name?: string;
	name: string;
	description?: string;
	price: number;
	image_url?: string;
	available: boolean;
	sort_order: number;
	modifiers?: Modifier[];
}

export interface Category {
	id: string;
	restaurant_id: string;
	name: string;
	sort_order: number;
	status: string;
}

export interface CategoryWithMenus extends Category {
	menus: Menu[];
}

export type OrderStatus =
	| 'WAITING_PAYMENT'
	| 'CONFIRMED'
	| 'PREPARING'
	| 'READY'
	| 'COMPLETED'
	| 'CANCELLED';

export interface SelectedModifierOption {
	modifier_id: string;
	modifier_name: string;
	option_id: string;
	option_name: string;
	additional_price: number;
}

export interface OrderItem {
	id: string;
	order_id: string;
	menu_id?: string;
	menu_name_snapshot: string;
	unit_price: number;
	quantity: number;
	subtotal: number;
	selected_modifiers?: SelectedModifierOption[];
	notes?: string;
}

export interface Order {
	id: string;
	restaurant_id: string;
	table_id: string;
	table_name?: string;
	order_number: string;
	status: OrderStatus;
	payment_status?: string;
	payment_method?: string;
	subtotal: number;
	tax: number;
	service_charge: number;
	discount: number;
	total: number;
	notes?: string;
	items?: OrderItem[];
	created_at: string;
	updated_at: string;
}

export interface Payment {
	id: string;
	restaurant_id: string;
	order_id: string;
	provider: string;
	provider_transaction_id?: string;
	payment_method?: string;
	amount: number;
	status: 'PENDING' | 'PAID' | 'FAILED' | 'EXPIRED' | 'CANCELLED';
	payment_url?: string;
	qr_string?: string;
	expired_at?: string;
	paid_at?: string;
}

export interface APIResponse<T> {
	success: boolean;
	data: T;
	message?: string;
}

export interface APIErrorResponse {
	success: false;
	error: {
		code: string;
		message: string;
	};
}

// Cashier Shift Management Types
export interface Register {
	id: string;
	restaurant_id: string;
	outlet_id?: string;
	name: string;
	status: 'ACTIVE' | 'INACTIVE';
	created_at: string;
	updated_at: string;
}

export type ShiftStatus = 'OPEN' | 'CLOSING' | 'CLOSED' | 'CANCELLED';
export type TransactionType = 'SALE' | 'REFUND' | 'VOID' | 'ADJUSTMENT';

export interface CashierShift {
	id: string;
	restaurant_id: string;
	outlet_id?: string;
	register_id: string;
	register_name?: string;
	cashier_id: string;
	cashier_name?: string;
	status: ShiftStatus;
	opening_balance: number;
	opened_at: string;
	closed_at?: string;
	expected_total: number;
	actual_total?: number;
	difference?: number;
	created_at: string;
	updated_at: string;
}

export interface ShiftTransaction {
	id: string;
	shift_id: string;
	order_id?: string;
	payment_id?: string;
	type: TransactionType;
	amount: number;
	metadata?: Record<string, any>;
	created_at: string;
}

export interface ShiftClosing {
	id: string;
	shift_id: string;
	gross_sales: number;
	refund_total: number;
	void_total: number;
	adjustment_total: number;
	net_sales: number;
	expected_amount: number;
	actual_amount?: number;
	difference?: number;
	notes?: string;
	closed_by: string;
	closed_by_name?: string;
	closed_at: string;
}

export interface Refund {
	id: string;
	restaurant_id: string;
	order_id: string;
	order_number?: string;
	payment_id: string;
	amount: number;
	reason: string;
	status: 'PENDING' | 'COMPLETED' | 'REJECTED';
	requested_by: string;
	requested_by_name?: string;
	approved_by?: string;
	created_at: string;
	completed_at?: string;
}

export interface ShiftSummary {
	shift_id: string;
	status: ShiftStatus;
	register_id: string;
	register_name: string;
	cashier_id: string;
	cashier_name: string;
	opening_balance: number;
	opened_at: string;
	closed_at?: string;
	orders_count: number;
	gross_sales: number;
	refund_total: number;
	void_total: number;
	adjustment_total: number;
	net_sales: number;
	expected_amount: number;
	payment_methods: Record<string, number>;
}

export interface ShiftReportDetail {
	shift: CashierShift;
	summary: ShiftSummary;
	closing?: ShiftClosing;
	transactions: ShiftTransaction[];
}

