package table

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTableNotFound = errors.New("table not found")
)

type Table struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Name         string    `json:"name"`
	QRToken      string    `json:"qr_token"`
	Status       string    `json:"status"` // ACTIVE, INACTIVE
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type PublicTableInfo struct {
	Restaurant struct {
		ID             string  `json:"id"`
		Name           string  `json:"name"`
		Slug           string  `json:"slug"`
		LogoURL        *string `json:"logo_url"`
		TaxPercent     float64 `json:"tax_percent"`
		ServicePercent float64 `json:"service_percent"`
	} `json:"restaurant"`
	Table struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		QRToken string `json:"qr_token"`
		Status  string `json:"status"`
	} `json:"table"`
}

type Repository interface {
	Create(ctx context.Context, t *Table) error
	GetByID(ctx context.Context, id, restaurantID string) (*Table, error)
	GetByQRToken(ctx context.Context, qrToken string) (*PublicTableInfo, error)
	ListByRestaurant(ctx context.Context, restaurantID string) ([]Table, error)
	Update(ctx context.Context, t *Table) error
	UpdateQRToken(ctx context.Context, id, restaurantID, newToken string) error
	Delete(ctx context.Context, id, restaurantID string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, t *Table) error {
	query := `
		INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, t.ID, t.RestaurantID, t.Name, t.QRToken, t.Status, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *repository) GetByID(ctx context.Context, id, restaurantID string) (*Table, error) {
	query := `
		SELECT id, restaurant_id, name, qr_token, status, created_at, updated_at
		FROM tables
		WHERE id = $1 AND restaurant_id = $2
	`
	var t Table
	err := r.pool.QueryRow(ctx, query, id, restaurantID).Scan(
		&t.ID, &t.RestaurantID, &t.Name, &t.QRToken, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *repository) GetByQRToken(ctx context.Context, qrToken string) (*PublicTableInfo, error) {
	query := `
		SELECT 
			r.id, r.name, r.slug, r.logo_url, r.tax_percent, r.service_percent,
			t.id, t.name, t.qr_token, t.status
		FROM tables t
		JOIN restaurants r ON r.id = t.restaurant_id
		WHERE t.qr_token = $1 AND t.status = 'ACTIVE' AND r.status = 'ACTIVE'
	`
	var info PublicTableInfo
	err := r.pool.QueryRow(ctx, query, qrToken).Scan(
		&info.Restaurant.ID, &info.Restaurant.Name, &info.Restaurant.Slug, &info.Restaurant.LogoURL,
		&info.Restaurant.TaxPercent, &info.Restaurant.ServicePercent,
		&info.Table.ID, &info.Table.Name, &info.Table.QRToken, &info.Table.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTableNotFound
		}
		return nil, err
	}
	return &info, nil
}

func (r *repository) ListByRestaurant(ctx context.Context, restaurantID string) ([]Table, error) {
	query := `
		SELECT id, restaurant_id, name, qr_token, status, created_at, updated_at
		FROM tables
		WHERE restaurant_id = $1
		ORDER BY name ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var t Table
		if err := rows.Scan(
			&t.ID, &t.RestaurantID, &t.Name, &t.QRToken, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (r *repository) Update(ctx context.Context, t *Table) error {
	t.UpdatedAt = time.Now()
	query := `
		UPDATE tables
		SET name = $1, status = $2, updated_at = $3
		WHERE id = $4 AND restaurant_id = $5
	`
	_, err := r.pool.Exec(ctx, query, t.Name, t.Status, t.UpdatedAt, t.ID, t.RestaurantID)
	return err
}

func (r *repository) UpdateQRToken(ctx context.Context, id, restaurantID, newToken string) error {
	query := `
		UPDATE tables
		SET qr_token = $1, updated_at = $2
		WHERE id = $3 AND restaurant_id = $4
	`
	_, err := r.pool.Exec(ctx, query, newToken, time.Now(), id, restaurantID)
	return err
}

func (r *repository) Delete(ctx context.Context, id, restaurantID string) error {
	query := `DELETE FROM tables WHERE id = $1 AND restaurant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, restaurantID)
	return err
}
