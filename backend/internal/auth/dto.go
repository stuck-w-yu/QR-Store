package auth

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserDTO struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         Role   `json:"role"`
	Status       string `json:"status"`
}

type AuthResponse struct {
	User        UserDTO `json:"user"`
	AccessToken string  `json:"access_token"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type UpdateAccountRequest struct {
	NewID           *string `json:"new_id"`
	Name            *string `json:"name"`
	Email           *string `json:"email"`
	CurrentPassword string  `json:"current_password" binding:"required"`
	NewPassword     *string `json:"new_password"`
}
