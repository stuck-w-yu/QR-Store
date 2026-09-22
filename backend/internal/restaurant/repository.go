package restaurant

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
)

type Restaurant struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	LogoURL        *string   `json:"logo_url"`
	Address        *string   `json:"address"`
	Phone          *string   `json:"phone"`
	TaxPercent     float64   `json:"tax_percent"`
	ServicePercent float64   `json:"service_percent"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UpdateRestaurantRequest struct {
	Name           *string  `json:"name"`
	LogoURL        *string  `json:"logo_url"`
	Address        *string  `json:"address"`
	Phone          *string  `json:"phone"`
	TaxPercent     *float64 `json:"tax_percent"`
	ServicePercent *float64 `json:"service_percent"`
	Status         *string  `json:"status"`
}

type Repository interface {
	GetByID(ctx context.Context, id string) (*Restaurant, error)
	GetBySlug(ctx context.Context, slug string) (*Restaurant, error)
	Update(ctx context.Context, id string, req UpdateRestaurantRequest) (*Restaurant, error)
	Create(ctx context.Context, r *Restaurant) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetByID(ctx context.Context, id string) (*Restaurant, error) {
	query := `
		SELECT id, name, slug, logo_url, address, phone, tax_percent, service_percent, status, created_at, updated_at
		FROM restaurants
		WHERE id = $1
	`
	var res Restaurant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.Address, &res.Phone,
		&res.TaxPercent, &res.ServicePercent, &res.Status, &res.CreatedAt, &res.UpdatedAt,
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
		SELECT id, name, slug, logo_url, address, phone, tax_percent, service_percent, status, created_at, updated_at
		FROM restaurants
		WHERE slug = $1
	`
	var res Restaurant
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&res.ID, &res.Name, &res.Slug, &res.LogoURL, &res.Address, &res.Phone,
		&res.TaxPercent, &res.ServicePercent, &res.Status, &res.CreatedAt, &res.UpdatedAt,
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
	query := `
		INSERT INTO restaurants (id, name, slug, logo_url, address, phone, tax_percent, service_percent, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		res.ID, res.Name, res.Slug, res.LogoURL, res.Address, res.Phone,
		res.TaxPercent, res.ServicePercent, res.Status, res.CreatedAt, res.UpdatedAt,
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
	current.UpdatedAt = time.Now()

	query := `
		UPDATE restaurants
		SET name = $1, logo_url = $2, address = $3, phone = $4, tax_percent = $5, service_percent = $6, status = $7, updated_at = $8
		WHERE id = $9
	`
	_, err = r.pool.Exec(ctx, query,
		current.Name, current.LogoURL, current.Address, current.Phone,
		current.TaxPercent, current.ServicePercent, current.Status, current.UpdatedAt, id,
	)
	if err != nil {
		return nil, err
	}
	return current, nil
}
