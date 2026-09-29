package database

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"qr-store/backend/pkg/logger"
)

//go:embed schema.sql
var schemaSQL string

func RunMigrationsAndSeed(ctx context.Context, db *DB) error {
	logger.Log.Info("running database migrations...")
	if _, err := db.Pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Ensure new schema columns exist
	_, _ = db.Pool.Exec(ctx, "ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS plan VARCHAR(50) DEFAULT 'PRO'")

	// Ensure Superadmin user exists
	var superadminCount int
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'SUPERADMIN'").Scan(&superadminCount)
	if superadminCount == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO users (id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (email) DO NOTHING
		`, "usr_superadmin", "rst_nusantara", "Super Administrator", "superadmin@qrstore.id", string(hash), "SUPERADMIN", "ACTIVE", time.Now(), time.Now())
		logger.Log.Info("seeded default superadmin user (superadmin@qrstore.id)")
	}

	// Check if seeded
	var count int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM restaurants").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check restaurants table: %w", err)
	}

	if count > 0 {
		logger.Log.Info("database already has data, skipping seed")
		return nil
	}

	logger.Log.Info("seeding initial demonstration data...")
	now := time.Now()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// 1. Seed Restaurant
	restoID := "rst_nusantara"
	restoQuery := `
		INSERT INTO restaurants (id, name, slug, logo_url, address, phone, tax_percent, service_percent, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = db.Pool.Exec(ctx, restoQuery,
		restoID, "Resto Nusantara", "resto-nusantara",
		"https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?w=500&q=80",
		"Jl. Malioboro No. 45, Yogyakarta", "081234567890",
		10.00, 5.00, "ACTIVE", now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to seed restaurant: %w", err)
	}

	// 2. Seed Users
	userQuery := `
		INSERT INTO users (id, restaurant_id, name, email, password_hash, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	users := []struct {
		ID    string
		Name  string
		Email string
		Role  string
	}{
		{"usr_owner", "Budi Owner", "owner@resto.com", "OWNER"},
		{"usr_admin", "Siti Admin", "admin@resto.com", "ADMIN"},
		{"usr_cashier", "Rian Kasir", "cashier@resto.com", "CASHIER"},
		{"usr_kitchen", "Chef Joko", "kitchen@resto.com", "KITCHEN"},
	}
	for _, u := range users {
		_, _ = db.Pool.Exec(ctx, userQuery, u.ID, restoID, u.Name, u.Email, string(hash), u.Role, "ACTIVE", now, now)
	}

	// 3. Seed Tables
	tableQuery := `
		INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	tables := []struct {
		ID      string
		Name    string
		QRToken string
	}{
		{"tbl_01", "Meja 01", "demo-qr-token-table-01"},
		{"tbl_02", "Meja 02", "demo-qr-token-table-02"},
		{"tbl_03", "Meja 03", "demo-qr-token-table-03"},
		{"tbl_04", "Meja 04", "demo-qr-token-table-04"},
		{"tbl_05", "Meja 05", "demo-qr-token-table-05"},
	}
	for _, t := range tables {
		_, _ = db.Pool.Exec(ctx, tableQuery, t.ID, restoID, t.Name, t.QRToken, "ACTIVE", now, now)
	}

	// 4. Seed Categories
	catQuery := `
		INSERT INTO categories (id, restaurant_id, name, sort_order, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	cats := []struct {
		ID    string
		Name  string
		Order int
	}{
		{"cat_food", "Makanan Utama", 1},
		{"cat_drink", "Minuman Segar", 2},
		{"cat_snack", "Camilan & Dessert", 3},
	}
	for _, c := range cats {
		_, _ = db.Pool.Exec(ctx, catQuery, c.ID, restoID, c.Name, c.Order, "ACTIVE", now, now)
	}

	// 5. Seed Menus
	menuQuery := `
		INSERT INTO menus (id, restaurant_id, category_id, name, description, price, image_url, available, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	menus := []struct {
		ID       string
		CatID    string
		Name     string
		Desc     string
		Price    int64
		ImageURL string
	}{
		{
			ID: "menu_nasgor", CatID: "cat_food", Name: "Nasi Goreng Spesial Nusantara",
			Desc: "Nasi goreng bumbu rempah dengan suwiran ayam, telur, acar segar, dan kerupuk udang.",
			Price: 28000,
			ImageURL: "https://images.unsplash.com/photo-1603133872878-684f208fb84b?w=600&q=80",
		},
		{
			ID: "menu_ayam_bakar", CatID: "cat_food", Name: "Ayam Bakar Madu Pedas",
			Desc: "Paha ayam bakar bumbu madu karamel dengan sambal terasi khas dan lalapan segar.",
			Price: 34000,
			ImageURL: "https://images.unsplash.com/photo-1598515214211-89d3c73ae83b?w=600&q=80",
		},
		{
			ID: "menu_miegor", CatID: "cat_food", Name: "Mie Goreng Seafood Jawa",
			Desc: "Mie pipih goreng dengan udang, cumi, bakso ikan, dan sayuran segar.",
			Price: 30000,
			ImageURL: "https://images.unsplash.com/photo-1585032226651-759b368d7246?w=600&q=80",
		},
		{
			ID: "menu_esteh", CatID: "cat_drink", Name: "Es Teh Manis Melati",
			Desc: "Teh melati seduh harum khas Jawa dengan gula tebu asli.",
			Price: 6000,
			ImageURL: "https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=600&q=80",
		},
		{
			ID: "menu_alpukat", CatID: "cat_drink", Name: "Jus Alpukat Kocok Cokelat",
			Desc: "Alpukat mentega legit dikocok lembut dengan siraman kental manis cokelat premium.",
			Price: 18000,
			ImageURL: "https://images.unsplash.com/photo-1600271886742-f049cd451bba?w=600&q=80",
		},
		{
			ID: "menu_pisang", CatID: "cat_snack", Name: "Pisang Goreng Keju Karamel",
			Desc: "Pisang raja goreng krispi dengan taburan keju cheddar melimpah dan lelehan saus karamel.",
			Price: 20000,
			ImageURL: "https://images.unsplash.com/photo-1528735602780-2552fd46c7af?w=600&q=80",
		},
	}

	for _, m := range menus {
		_, _ = db.Pool.Exec(ctx, menuQuery, m.ID, restoID, m.CatID, m.Name, m.Desc, m.Price, m.ImageURL, true, 0, now, now)
	}

	// 6. Seed Modifiers for Nasi Goreng
	modID := "mod_pedas"
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO menu_modifiers (id, restaurant_id, menu_id, name, type, required, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, modID, restoID, "menu_nasgor", "Tingkat Kepedasan", "SINGLE", true, now, now)

	modOpts := []struct {
		ID    string
		Name  string
		Price int64
	}{
		{"opt_pedas_0", "Level 0 (Tidak Pedas)", 0},
		{"opt_pedas_1", "Level 1 (Sedang)", 0},
		{"opt_pedas_2", "Level 2 (Pedas Banget)", 2000},
	}
	for _, opt := range modOpts {
		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO menu_modifier_options (id, modifier_id, name, additional_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, opt.ID, modID, opt.Name, opt.Price, now, now)
	}

	// Extra Topping for Nasi Goreng
	modID2 := "mod_topping"
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO menu_modifiers (id, restaurant_id, menu_id, name, type, required, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, modID2, restoID, "menu_nasgor", "Extra Topping", "MULTIPLE", false, now, now)

	toppingOpts := []struct {
		ID    string
		Name  string
		Price int64
	}{
		{"opt_top_egg", "Telur Mata Sapi Setengah Matang", 5000},
		{"opt_top_cheese", "Keju Mozzarella Parut", 6000},
		{"opt_top_sosis", "Sosis Bakar Iris", 4000},
	}
	for _, opt := range toppingOpts {
		_, _ = db.Pool.Exec(ctx, `
			INSERT INTO menu_modifier_options (id, modifier_id, name, additional_price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, opt.ID, modID2, opt.Name, opt.Price, now, now)
	}

	// 7. Seed Registers
	registerQuery := `
		INSERT INTO registers (id, restaurant_id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	registers := []struct {
		ID   string
		Name string
	}{
		{"reg_pos_01", "POS 01 - Kasir Utama"},
		{"reg_pos_02", "POS 02 - Kasir Bar"},
	}
	for _, reg := range registers {
		_, _ = db.Pool.Exec(ctx, registerQuery, reg.ID, restoID, reg.Name, "ACTIVE", now, now)
	}

	logger.Log.Info("seed data completed successfully!")
	return nil
}
