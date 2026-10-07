package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	ListByRestaurant(ctx context.Context, restaurantID string) ([]User, error)
	Delete(ctx context.Context, id, restaurantID string) error
	UpdatePassword(ctx context.Context, id, passwordHash string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND status = 'ACTIVE'
	`
	var u User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.RestaurantID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE id = $1 AND status = 'ACTIVE'
	`
	var u User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.RestaurantID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *repository) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.RestaurantID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *repository) ListByRestaurant(ctx context.Context, restaurantID string) ([]User, error) {
	query := `
		SELECT id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at
		FROM users
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID,
			&u.RestaurantID,
			&u.Name,
			&u.Email,
			&u.PasswordHash,
			&u.Role,
			&u.Status,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *repository) Delete(ctx context.Context, id, restaurantID string) error {
	query := `DELETE FROM users WHERE id = $1 AND restaurant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, restaurantID)
	return err
}

func (r *repository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, passwordHash, id)
	return err
}
