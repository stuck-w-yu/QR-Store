package restaurant

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
	serverStartTime       = time.Now()
)

type Restaurant struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	LogoURL        *string   `json:"logo_url"`
	QRISImageURL   *string   `json:"qris_image_url"`
	Address        *string   `json:"address"`
	Phone          *string   `json:"phone"`
	TaxPercent     float64   `json:"tax_percent"`
	ServicePercent float64   `json:"service_percent"`
	Status         string    `json:"status"`
	Plan           string    `json:"plan"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TenantSummary struct {
	Restaurant
	OwnerName    string `json:"owner_name"`
	OwnerEmail   string `json:"owner_email"`
	OwnerPhone   string `json:"owner_phone"`
	TotalTables  int    `json:"total_tables"`
	TotalMenus   int    `json:"total_menus"`
	TotalOrders  int    `json:"total_orders"`
	TotalRevenue int64  `json:"total_revenue"`
}

type PlatformStats struct {
	TotalRestaurants     int            `json:"total_restaurants"`
	ActiveRestaurants    int            `json:"active_restaurants"`
	SuspendedRestaurants int            `json:"suspended_restaurants"`
	TotalOwners          int            `json:"total_owners"`
	TotalOrders          int            `json:"total_orders"`
	TotalRevenue         int64          `json:"total_revenue"`
	TotalTables          int            `json:"total_tables"`
	TotalMenus           int            `json:"total_menus"`
	TodayRevenue         int64          `json:"today_revenue"`
	TodayOrders          int            `json:"today_orders"`
	PlanDistribution     map[string]int `json:"plan_distribution"`
}

type PlatformOrder struct {
	ID             string    `json:"id"`
	RestaurantID   string    `json:"restaurant_id"`
	RestaurantName string    `json:"restaurant_name"`
	RestaurantSlug string    `json:"restaurant_slug"`
	TableName      string    `json:"table_name"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	PaymentStatus  string    `json:"payment_status"`
	PaymentMethod  string    `json:"payment_method"`
	Total          int64     `json:"total"`
	CreatedAt      time.Time `json:"created_at"`
}

type PlatformOwner struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	RestaurantID   string    `json:"restaurant_id"`
	RestaurantName string    `json:"restaurant_name"`
	RestaurantSlug string    `json:"restaurant_slug"`
	RestaurantPlan string    `json:"restaurant_plan"`
	Phone          string    `json:"phone"`
}

type SystemHealth struct {
	Status         string    `json:"status"`
	Database       string    `json:"database"`
	PoolTotalConns int32     `json:"pool_total_conns"`
	PoolIdleConns  int32     `json:"pool_idle_conns"`
	Goroutines     int       `json:"goroutines"`
	MemoryAllocMB  float64   `json:"memory_alloc_mb"`
	UptimeSeconds  int64     `json:"uptime_seconds"`
	Timestamp      time.Time `json:"timestamp"`
}

type OnboardTenantRequest struct {
	Name           string  `json:"name" binding:"required"`
	Slug           string  `json:"slug" binding:"required"`
	LogoURL        *string `json:"logo_url"`
	QRISImageURL   *string `json:"qris_image_url"`
	Address        *string `json:"address"`
	Phone          *string `json:"phone"`
	TaxPercent     float64 `json:"tax_percent"`
	ServicePercent float64 `json:"service_percent"`
	Plan           string  `json:"plan"`
	OwnerName      string  `json:"owner_name" binding:"required"`
	OwnerEmail     string  `json:"owner_email" binding:"required"`
	OwnerPassword  string  `json:"owner_password" binding:"required"`
}

type UpdateRestaurantRequest struct {
	Name           *string  `json:"name"`
	LogoURL        *string  `json:"logo_url"`
	QRISImageURL   *string  `json:"qris_image_url"`
	Address        *string  `json:"address"`
	Phone          *string  `json:"phone"`
	TaxPercent     *float64 `json:"tax_percent"`
	ServicePercent *float64 `json:"service_percent"`
	Status         *string  `json:"status"`
	Plan           *string  `json:"plan"`
}

type Repository interface {
	GetByID(ctx context.Context, id string) (*Restaurant, error)
	GetBySlug(ctx context.Context, slug string) (*Restaurant, error)
	Update(ctx context.Context, id string, req UpdateRestaurantRequest) (*Restaurant, error)
	Create(ctx context.Context, r *Restaurant) error
	ListTenants(ctx context.Context) ([]TenantSummary, error)
	GetPlatformStats(ctx context.Context) (*PlatformStats, error)
	OnboardTenant(ctx context.Context, req OnboardTenantRequest, passwordHash string) (*TenantSummary, error)
	Delete(ctx context.Context, id string) error
	ListRecentOrders(ctx context.Context, limit int) ([]PlatformOrder, error)
	ListOwners(ctx context.Context) ([]PlatformOwner, error)
	GetSystemHealth(ctx context.Context) (*SystemHealth, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Restaurant, error) {
	query := `
		SELECT id, name, slug, logo_url, qris_image_url, address, phone, tax_percent, service_percent, status, COALESCE(plan, 'PRO'), created_at, updated_at
		FROM restaurants
		WHERE id = $1
	`
	var res Restaurant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.QRISImageURL, &res.Address, &res.Phone,
		&res.TaxPercent, &res.ServicePercent, &res.Status, &res.Plan, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*Restaurant, error) {
	query := `
		SELECT id, name, slug, logo_url, qris_image_url, address, phone, tax_percent, service_percent, status, COALESCE(plan, 'PRO'), created_at, updated_at
		FROM restaurants
		WHERE slug = $1
	`
	var res Restaurant
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.QRISImageURL, &res.Address, &res.Phone,
		&res.TaxPercent, &res.ServicePercent, &res.Status, &res.Plan, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *repository) Create(ctx context.Context, res *Restaurant) error {
	plan := res.Plan
	if plan == "" {
		plan = "PRO"
	}
	query := `
		INSERT INTO restaurants (id, name, slug, logo_url, qris_image_url, address, phone, tax_percent, service_percent, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.pool.Exec(ctx, query,
		res.ID, res.Name, res.Slug, res.LogoURL, res.QRISImageURL, res.Address, res.Phone,
		res.TaxPercent, res.ServicePercent, res.Status, plan, res.CreatedAt, res.UpdatedAt,
	)
	return err
}

func (r *repository) Update(ctx context.Context, id string, req UpdateRestaurantRequest) (*Restaurant, error) {
	current, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		current.Name = *req.Name
	}
	if req.LogoURL != nil {
		current.LogoURL = req.LogoURL
	}
	if req.QRISImageURL != nil {
		current.QRISImageURL = req.QRISImageURL
	}
	if req.Address != nil {
		current.Address = req.Address
	}
	if req.Phone != nil {
		current.Phone = req.Phone
	}
	if req.TaxPercent != nil {
		current.TaxPercent = *req.TaxPercent
	}
	if req.ServicePercent != nil {
		current.ServicePercent = *req.ServicePercent
	}
	if req.Status != nil {
		current.Status = *req.Status
	}
	if req.Plan != nil {
		current.Plan = *req.Plan
	}
	current.UpdatedAt = time.Now()

	query := `
		UPDATE restaurants
		SET name = $1, logo_url = $2, address = $3, phone = $4, tax_percent = $5, service_percent = $6, status = $7, plan = $8, updated_at = $9, qris_image_url = $10
		WHERE id = $11
	`
	_, err = r.pool.Exec(ctx, query,
		current.Name, current.LogoURL, current.Address, current.Phone,
		current.TaxPercent, current.ServicePercent, current.Status, current.Plan, current.UpdatedAt, current.QRISImageURL, id,
	)
	if err != nil {
		return nil, err
	}
	return current, nil
}

func (r *repository) ListTenants(ctx context.Context) ([]TenantSummary, error) {
	query := `
		SELECT 
			r.id, r.name, r.slug, r.logo_url, r.qris_image_url, r.address, r.phone, r.tax_percent, r.service_percent, r.status, 
			COALESCE(r.plan, 'PRO') as plan, r.created_at, r.updated_at,
			COALESCE(u.name, 'Belum Diatur') as owner_name,
			COALESCE(u.email, '-') as owner_email,
			COALESCE(r.phone, '-') as owner_phone,
			COALESCE((SELECT COUNT(*) FROM tables t WHERE t.restaurant_id = r.id), 0) as total_tables,
			COALESCE((SELECT COUNT(*) FROM menus m WHERE m.restaurant_id = r.id), 0) as total_menus,
			COALESCE((SELECT COUNT(*) FROM orders o WHERE o.restaurant_id = r.id), 0) as total_orders,
			COALESCE((SELECT SUM(o.total) FROM orders o WHERE o.restaurant_id = r.id AND o.status IN ('CONFIRMED','PREPARING','READY','COMPLETED')), 0) as total_revenue
		FROM restaurants r
		LEFT JOIN LATERAL (
			SELECT name, email FROM users WHERE restaurant_id = r.id AND role = 'OWNER' ORDER BY created_at ASC LIMIT 1
		) u ON true
		ORDER BY r.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []TenantSummary
	for rows.Next() {
		var t TenantSummary
		err := rows.Scan(
			&t.ID, &t.Name, &t.Slug, &t.LogoURL, &t.QRISImageURL, &t.Address, &t.Phone,
			&t.TaxPercent, &t.ServicePercent, &t.Status, &t.Plan, &t.CreatedAt, &t.UpdatedAt,
			&t.OwnerName, &t.OwnerEmail, &t.OwnerPhone,
			&t.TotalTables, &t.TotalMenus, &t.TotalOrders, &t.TotalRevenue,
		)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

func (r *repository) GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
	query := `
		SELECT
			COALESCE(COUNT(*), 0) as total_restaurants,
			COALESCE(COUNT(*) FILTER (WHERE status = 'ACTIVE'), 0) as active_restaurants,
			COALESCE(COUNT(*) FILTER (WHERE status != 'ACTIVE'), 0) as suspended_restaurants,
			COALESCE((SELECT COUNT(*) FROM users WHERE role = 'OWNER'), 0) as total_owners,
			COALESCE((SELECT COUNT(*) FROM orders), 0) as total_orders,
			COALESCE((SELECT SUM(total) FROM orders WHERE status IN ('CONFIRMED','PREPARING','READY','COMPLETED')), 0) as total_revenue,
			COALESCE((SELECT COUNT(*) FROM tables), 0) as total_tables,
			COALESCE((SELECT COUNT(*) FROM menus), 0) as total_menus,
			COALESCE((SELECT SUM(total) FROM orders WHERE status IN ('CONFIRMED','PREPARING','READY','COMPLETED') AND created_at >= CURRENT_DATE), 0) as today_revenue,
			COALESCE((SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE), 0) as today_orders
		FROM restaurants
	`
	var s PlatformStats
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.TotalRestaurants,
		&s.ActiveRestaurants,
		&s.SuspendedRestaurants,
		&s.TotalOwners,
		&s.TotalOrders,
		&s.TotalRevenue,
		&s.TotalTables,
		&s.TotalMenus,
		&s.TodayRevenue,
		&s.TodayOrders,
	)
	if err != nil {
		return nil, err
	}

	s.PlanDistribution = make(map[string]int)
	planRows, err := r.pool.Query(ctx, "SELECT COALESCE(plan, 'PRO'), COUNT(*) FROM restaurants GROUP BY plan")
	if err == nil {
		defer planRows.Close()
		for planRows.Next() {
			var plan string
			var count int
			if err := planRows.Scan(&plan, &count); err == nil {
				s.PlanDistribution[plan] = count
			}
		}
	}

	return &s, nil
}

func (r *repository) OnboardTenant(ctx context.Context, req OnboardTenantRequest, passwordHash string) (*TenantSummary, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	restoID := fmt.Sprintf("rst_%d", now.UnixNano()%1000000)
	ownerID := fmt.Sprintf("usr_own_%d", now.UnixNano()%1000000)

	plan := req.Plan
	if plan == "" {
		plan = "PRO"
	}
	if req.TaxPercent == 0 {
		req.TaxPercent = 10.0
	}

	// 1. Insert Restaurant
	restoQuery := `
		INSERT INTO restaurants (id, name, slug, logo_url, qris_image_url, address, phone, tax_percent, service_percent, status, plan, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, restoQuery,
		restoID, req.Name, req.Slug, req.LogoURL, req.QRISImageURL, req.Address, req.Phone,
		req.TaxPercent, req.ServicePercent, "ACTIVE", plan, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert restaurant: %w", err)
	}

	// 2. Insert Owner User
	userQuery := `
		INSERT INTO users (id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.Exec(ctx, userQuery,
		ownerID, restoID, req.OwnerName, req.OwnerEmail, passwordHash, "OWNER", "ACTIVE", now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert owner user: %w", err)
	}

	// 3. Create initial tables
	tableQuery := `
		INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
		VALUES 
			($1, $2, 'Meja 01', $3, 'ACTIVE', $4, $4),
			($5, $2, 'Meja 02', $6, 'ACTIVE', $4, $4)
	`
	tbl1ID := fmt.Sprintf("tbl_%d_1", now.UnixNano()%100000)
	tbl2ID := fmt.Sprintf("tbl_%d_2", now.UnixNano()%100000)
	qrToken1 := fmt.Sprintf("%s-table-01", req.Slug)
	qrToken2 := fmt.Sprintf("%s-table-02", req.Slug)

	_, err = tx.Exec(ctx, tableQuery, tbl1ID, restoID, qrToken1, now, tbl2ID, qrToken2)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial tables: %w", err)
	}

	// 4. Create initial categories
	catQuery := `
		INSERT INTO categories (id, restaurant_id, name, sort_order, status, created_at, updated_at)
		VALUES 
			($1, $2, 'Makanan', 1, 'ACTIVE', $3, $3),
			($4, $2, 'Minuman', 2, 'ACTIVE', $3, $3)
	`
	cat1ID := fmt.Sprintf("cat_%d_1", now.UnixNano()%100000)
	cat2ID := fmt.Sprintf("cat_%d_2", now.UnixNano()%100000)
	_, _ = tx.Exec(ctx, catQuery, cat1ID, restoID, now, cat2ID)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	phoneStr := ""
	if req.Phone != nil {
		phoneStr = *req.Phone
	}

	return &TenantSummary{
		Restaurant: Restaurant{
			ID:             restoID,
			Name:           req.Name,
			Slug:           req.Slug,
			LogoURL:        req.LogoURL,
			QRISImageURL:   req.QRISImageURL,
			Address:        req.Address,
			Phone:          req.Phone,
			TaxPercent:     req.TaxPercent,
			ServicePercent: req.ServicePercent,
			Status:         "ACTIVE",
			Plan:           plan,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		OwnerName:    req.OwnerName,
		OwnerEmail:   req.OwnerEmail,
		OwnerPhone:   phoneStr,
		TotalTables:  2,
		TotalMenus:   0,
		TotalOrders:  0,
		TotalRevenue: 0,
	}, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM restaurants WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *repository) ListRecentOrders(ctx context.Context, limit int) ([]PlatformOrder, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT 
			o.id, o.restaurant_id, r.name, r.slug, COALESCE(t.name, 'Meja -'),
			o.order_number, o.status, COALESCE(o.payment_status, 'UNPAID'), COALESCE(o.payment_method, '-'),
			o.total, o.created_at
		FROM orders o
		JOIN restaurants r ON r.id = o.restaurant_id
		LEFT JOIN tables t ON t.id = o.table_id
		ORDER BY o.created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []PlatformOrder
	for rows.Next() {
		var o PlatformOrder
		err := rows.Scan(
			&o.ID, &o.RestaurantID, &o.RestaurantName, &o.RestaurantSlug, &o.TableName,
			&o.OrderNumber, &o.Status, &o.PaymentStatus, &o.PaymentMethod,
			&o.Total, &o.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *repository) ListOwners(ctx context.Context) ([]PlatformOwner, error) {
	query := `
		SELECT 
			u.id, u.name, u.email, u.status, u.created_at,
			r.id, r.name, r.slug, COALESCE(r.plan, 'PRO'), COALESCE(r.phone, '-')
		FROM users u
		JOIN restaurants r ON r.id = u.restaurant_id
		WHERE u.role = 'OWNER'
		ORDER BY u.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var owners []PlatformOwner
	for rows.Next() {
		var ow PlatformOwner
		err := rows.Scan(
			&ow.ID, &ow.Name, &ow.Email, &ow.Status, &ow.CreatedAt,
			&ow.RestaurantID, &ow.RestaurantName, &ow.RestaurantSlug, &ow.RestaurantPlan, &ow.Phone,
		)
		if err != nil {
			return nil, err
		}
		owners = append(owners, ow)
	}
	return owners, nil
}

func (r *repository) GetSystemHealth(ctx context.Context) (*SystemHealth, error) {
	stat := r.pool.Stat()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	dbStatus := "connected"
	if err := r.pool.Ping(ctx); err != nil {
		dbStatus = "disconnected"
	}

	return &SystemHealth{
		Status:         "operational",
		Database:       dbStatus,
		PoolTotalConns: stat.TotalConns(),
		PoolIdleConns:  stat.IdleConns(),
		Goroutines:     runtime.NumGoroutine(),
		MemoryAllocMB:  float64(mem.Alloc) / 1024 / 1024,
		UptimeSeconds:  int64(time.Since(serverStartTime).Seconds()),
		Timestamp:      time.Now(),
	}, nil
}
