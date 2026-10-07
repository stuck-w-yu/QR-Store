package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrInvalidCurrentPassword = errors.New("current password does not match")
	ErrEmailAlreadyExists     = errors.New("email is already in use")
	ErrIDAlreadyExists        = errors.New("user ID is already in use")
)

type Claims struct {
	UserID       string `json:"user_id"`
	RestaurantID string `json:"restaurant_id"`
	Role         Role   `json:"role"`
	jwt.RegisteredClaims
}

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	ValidateToken(tokenString string) (*Claims, error)
	HashPassword(password string) (string, error)
	GetUserByID(ctx context.Context, id string) (*UserDTO, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	ResetPassword(ctx context.Context, targetUserID, newPassword string) error
	UpdateAccount(ctx context.Context, userID string, req UpdateAccountRequest) (*AuthResponse, error)
}

type service struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		User: UserDTO{
			ID:           user.ID,
			RestaurantID: user.RestaurantID,
			Name:         user.Name,
			Email:        user.Email,
			Role:         user.Role,
			Status:       user.Status,
		},
		AccessToken: token,
	}, nil
}

func (s *service) generateToken(user *User) (string, error) {
	claims := Claims{
		UserID:       user.ID,
		RestaurantID: user.RestaurantID,
		Role:         user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

func (s *service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *service) GetUserByID(ctx context.Context, id string) (*UserDTO, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &UserDTO{
		ID:           user.ID,
		RestaurantID: user.RestaurantID,
		Name:         user.Name,
		Email:        user.Email,
		Role:         user.Role,
		Status:       user.Status,
	}, nil
}

func (s *service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCurrentPassword
	}

	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, userID, hash)
}

func (s *service) ResetPassword(ctx context.Context, targetUserID, newPassword string) error {
	hash, err := s.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, targetUserID, hash)
}

func (s *service) UpdateAccount(ctx context.Context, userID string, req UpdateAccountRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 1. Verifikasi kata sandi saat ini
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return nil, ErrInvalidCurrentPassword
	}

	currentID := user.ID

	// 2. Jika mengganti ID akun (hanya diperbolehkan jika tidak bentrok)
	if req.NewID != nil && *req.NewID != "" && *req.NewID != currentID {
		newID := *req.NewID
		exists, err := s.repo.CheckIDExists(ctx, newID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrIDAlreadyExists
		}

		if err := s.repo.UpdateUserID(ctx, currentID, newID); err != nil {
			return nil, fmt.Errorf("gagal memperbarui ID pengguna: %w", err)
		}
		currentID = newID
		user.ID = newID
	}

	// 3. Jika mengganti Email login
	if req.Email != nil && *req.Email != "" && *req.Email != user.Email {
		newEmail := *req.Email
		exists, err := s.repo.CheckEmailExists(ctx, newEmail, currentID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrEmailAlreadyExists
		}
		user.Email = newEmail
	}

	// 4. Jika mengganti Nama
	if req.Name != nil && *req.Name != "" {
		user.Name = *req.Name
	}

	// 5. Jika mengganti Kata Sandi baru
	if req.NewPassword != nil && *req.NewPassword != "" {
		if len(*req.NewPassword) < 6 {
			return nil, errors.New("kata sandi baru minimal 6 karakter")
		}
		newHash, err := s.HashPassword(*req.NewPassword)
		if err != nil {
			return nil, fmt.Errorf("gagal memproses hash kata sandi: %w", err)
		}
		user.PasswordHash = newHash
	}

	user.UpdatedAt = time.Now()

	// 6. Simpan detail user di database
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("gagal memperbarui detail akun: %w", err)
	}

	// 7. Generate JWT access token baru yang merefleksikan perubahan
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat token baru: %w", err)
	}

	return &AuthResponse{
		User: UserDTO{
			ID:           user.ID,
			RestaurantID: user.RestaurantID,
			Name:         user.Name,
			Email:        user.Email,
			Role:         user.Role,
			Status:       user.Status,
		},
		AccessToken: token,
	}, nil
}

