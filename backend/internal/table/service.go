package table

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type CreateTableRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateTableRequest struct {
	Name   string `json:"name" binding:"required"`
	Status string `json:"status" binding:"required"`
}

type Service interface {
	CreateTable(ctx context.Context, restaurantID string, req CreateTableRequest) (*Table, error)
	GetTable(ctx context.Context, id, restaurantID string) (*Table, error)
	GetPublicTableInfo(ctx context.Context, qrToken string) (*PublicTableInfo, error)
	ListTables(ctx context.Context, restaurantID string) ([]Table, error)
	UpdateTable(ctx context.Context, id, restaurantID string, req UpdateTableRequest) (*Table, error)
	RegenerateQRToken(ctx context.Context, id, restaurantID string) (string, error)
	GenerateQRCodePNG(ctx context.Context, id, restaurantID, frontendBaseURL string) ([]byte, error)
	DeleteTable(ctx context.Context, id, restaurantID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTable(ctx context.Context, restaurantID string, req CreateTableRequest) (*Table, error) {
	now := time.Now()
	t := &Table{
		ID:           "tbl_" + uuid.New().String()[:8],
		RestaurantID: restaurantID,
		Name:         req.Name,
		QRToken:      "qr_" + uuid.New().String(),
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) GetTable(ctx context.Context, id, restaurantID string) (*Table, error) {
	return s.repo.GetByID(ctx, id, restaurantID)
}

func (s *service) GetPublicTableInfo(ctx context.Context, qrToken string) (*PublicTableInfo, error) {
	return s.repo.GetByQRToken(ctx, qrToken)
}

func (s *service) ListTables(ctx context.Context, restaurantID string) ([]Table, error) {
	return s.repo.ListByRestaurant(ctx, restaurantID)
}

func (s *service) UpdateTable(ctx context.Context, id, restaurantID string, req UpdateTableRequest) (*Table, error) {
	t, err := s.repo.GetByID(ctx, id, restaurantID)
	if err != nil {
		return nil, err
	}

	t.Name = req.Name
	t.Status = req.Status
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) RegenerateQRToken(ctx context.Context, id, restaurantID string) (string, error) {
	newToken := "qr_" + uuid.New().String()
	if err := s.repo.UpdateQRToken(ctx, id, restaurantID, newToken); err != nil {
		return "", err
	}
	return newToken, nil
}

func (s *service) GenerateQRCodePNG(ctx context.Context, id, restaurantID, frontendBaseURL string) ([]byte, error) {
	t, err := s.repo.GetByID(ctx, id, restaurantID)
	if err != nil {
		return nil, err
	}

	qrURL := fmt.Sprintf("%s/order?token=%s", frontendBaseURL, t.QRToken)
	return qrcode.Encode(qrURL, qrcode.Medium, 256)
}

func (s *service) DeleteTable(ctx context.Context, id, restaurantID string) error {
	return s.repo.Delete(ctx, id, restaurantID)
}
