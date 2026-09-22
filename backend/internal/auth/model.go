package auth

import "time"

type Role string

const (
	RoleOwner   Role = "OWNER"
	RoleAdmin   Role = "ADMIN"
	RoleCashier Role = "CASHIER"
	RoleKitchen Role = "KITCHEN"
)

type User struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
