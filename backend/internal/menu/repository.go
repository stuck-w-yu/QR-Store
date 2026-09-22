package menu

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
	ErrMenuNotFound     = errors.New("menu not found")
)

type Repository interface {
	// Categories
	CreateCategory(ctx context.Context, c *Category) error
	GetCategoryByID(ctx context.Context, id, restaurantID string) (*Category, error)
	ListCategories(ctx context.Context, restaurantID string) ([]Category, error)
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, id, restaurantID string) error

	// Menus
	CreateMenu(ctx context.Context, m *Menu) error
	GetMenuByID(ctx context.Context, id, restaurantID string) (*Menu, error)
	ListMenus(ctx context.Context, restaurantID string, categoryID *string) ([]Menu, error)
	UpdateMenu(ctx context.Context, m *Menu) error
	UpdateMenuAvailability(ctx context.Context, id, restaurantID string, available bool) error
	DeleteMenu(ctx context.Context, id, restaurantID string) error

	// Modifiers
	CreateModifier(ctx context.Context, mod *Modifier) error
	ListModifiersByMenu(ctx context.Context, menuID string) ([]Modifier, error)
	DeleteModifier(ctx context.Context, id, restaurantID string) error

	// Public catalog
	GetPublicCatalog(ctx context.Context, restaurantID string) ([]CategoryWithMenus, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// Categories
func (r *repository) CreateCategory(ctx context.Context, c *Category) error {
	query := `
		INSERT INTO categories (id, restaurant_id, name, sort_order, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.RestaurantID, c.Name, c.SortOrder, c.Status, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *repository) GetCategoryByID(ctx context.Context, id, restaurantID string) (*Category, error) {
	query := `
		SELECT id, restaurant_id, name, sort_order, status, created_at, updated_at
		FROM categories
		WHERE id = $1 AND restaurant_id = $2
	`
	var c Category
	err := r.pool.QueryRow(ctx, query, id, restaurantID).Scan(
		&c.ID, &c.RestaurantID, &c.Name, &c.SortOrder, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *repository) ListCategories(ctx context.Context, restaurantID string) ([]Category, error) {
	query := `
		SELECT id, restaurant_id, name, sort_order, status, created_at, updated_at
		FROM categories
		WHERE restaurant_id = $1
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.RestaurantID, &c.Name, &c.SortOrder, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *repository) UpdateCategory(ctx context.Context, c *Category) error {
	c.UpdatedAt = time.Now()
	query := `
		UPDATE categories
		SET name = $1, sort_order = $2, status = $3, updated_at = $4
		WHERE id = $5 AND restaurant_id = $6
	`
	_, err := r.pool.Exec(ctx, query, c.Name, c.SortOrder, c.Status, c.UpdatedAt, c.ID, c.RestaurantID)
	return err
}

func (r *repository) DeleteCategory(ctx context.Context, id, restaurantID string) error {
	query := `DELETE FROM categories WHERE id = $1 AND restaurant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, restaurantID)
	return err
}

// Menus
func (r *repository) CreateMenu(ctx context.Context, m *Menu) error {
	query := `
		INSERT INTO menus (id, restaurant_id, category_id, name, description, price, image_url, available, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		m.ID, m.RestaurantID, m.CategoryID, m.Name, m.Description, m.Price,
		m.ImageURL, m.Available, m.SortOrder, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (r *repository) GetMenuByID(ctx context.Context, id, restaurantID string) (*Menu, error) {
	query := `
		SELECT m.id, m.restaurant_id, m.category_id, c.name, m.name, m.description, m.price, m.image_url, m.available, m.sort_order, m.created_at, m.updated_at
		FROM menus m
		LEFT JOIN categories c ON c.id = m.category_id
		WHERE m.id = $1 AND m.restaurant_id = $2
	`
	var m Menu
	err := r.pool.QueryRow(ctx, query, id, restaurantID).Scan(
		&m.ID, &m.RestaurantID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Description,
		&m.Price, &m.ImageURL, &m.Available, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}

	mods, err := r.ListModifiersByMenu(ctx, m.ID)
	if err == nil {
		m.Modifiers = mods
	}
	return &m, nil
}

func (r *repository) ListMenus(ctx context.Context, restaurantID string, categoryID *string) ([]Menu, error) {
	query := `
		SELECT m.id, m.restaurant_id, m.category_id, c.name, m.name, m.description, m.price, m.image_url, m.available, m.sort_order, m.created_at, m.updated_at
		FROM menus m
		LEFT JOIN categories c ON c.id = m.category_id
		WHERE m.restaurant_id = $1 AND ($2::varchar IS NULL OR m.category_id = $2)
		ORDER BY m.sort_order ASC, m.name ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(
			&m.ID, &m.RestaurantID, &m.CategoryID, &m.CategoryName, &m.Name, &m.Description,
			&m.Price, &m.ImageURL, &m.Available, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (r *repository) UpdateMenu(ctx context.Context, m *Menu) error {
	m.UpdatedAt = time.Now()
	query := `
		UPDATE menus
		SET category_id = $1, name = $2, description = $3, price = $4, image_url = $5, available = $6, sort_order = $7, updated_at = $8
		WHERE id = $9 AND restaurant_id = $10
	`
	_, err := r.pool.Exec(ctx, query,
		m.CategoryID, m.Name, m.Description, m.Price, m.ImageURL, m.Available, m.SortOrder, m.UpdatedAt, m.ID, m.RestaurantID,
	)
	return err
}

func (r *repository) UpdateMenuAvailability(ctx context.Context, id, restaurantID string, available bool) error {
	query := `
		UPDATE menus
		SET available = $1, updated_at = $2
		WHERE id = $3 AND restaurant_id = $4
	`
	_, err := r.pool.Exec(ctx, query, available, time.Now(), id, restaurantID)
	return err
}

func (r *repository) DeleteMenu(ctx context.Context, id, restaurantID string) error {
	query := `DELETE FROM menus WHERE id = $1 AND restaurant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, restaurantID)
	return err
}

// Modifiers
func (r *repository) CreateModifier(ctx context.Context, mod *Modifier) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	modQuery := `
		INSERT INTO menu_modifiers (id, restaurant_id, menu_id, name, type, required, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.Exec(ctx, modQuery, mod.ID, mod.RestaurantID, mod.MenuID, mod.Name, mod.Type, mod.Required, mod.CreatedAt, mod.UpdatedAt); err != nil {
		return err
	}

	optQuery := `
		INSERT INTO menu_modifier_options (id, modifier_id, name, additional_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, opt := range mod.Options {
		if _, err := tx.Exec(ctx, optQuery, opt.ID, mod.ID, opt.Name, opt.AdditionalPrice, opt.CreatedAt, opt.UpdatedAt); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *repository) ListModifiersByMenu(ctx context.Context, menuID string) ([]Modifier, error) {
	query := `
		SELECT id, restaurant_id, menu_id, name, type, required, created_at, updated_at
		FROM menu_modifiers
		WHERE menu_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, menuID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []Modifier
	for rows.Next() {
		var m Modifier
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.MenuID, &m.Name, &m.Type, &m.Required, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}

	// Fetch options for each modifier
	for i := range mods {
		optQuery := `
			SELECT id, modifier_id, name, additional_price, created_at, updated_at
			FROM menu_modifier_options
			WHERE modifier_id = $1
			ORDER BY additional_price ASC, name ASC
		`
		optRows, err := r.pool.Query(ctx, optQuery, mods[i].ID)
		if err == nil {
			var opts []ModifierOption
			for optRows.Next() {
				var opt ModifierOption
				if err := optRows.Scan(&opt.ID, &opt.ModifierID, &opt.Name, &opt.AdditionalPrice, &opt.CreatedAt, &opt.UpdatedAt); err == nil {
					opts = append(opts, opt)
				}
			}
			optRows.Close()
			mods[i].Options = opts
		}
	}

	return mods, nil
}

func (r *repository) DeleteModifier(ctx context.Context, id, restaurantID string) error {
	query := `DELETE FROM menu_modifiers WHERE id = $1 AND restaurant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, restaurantID)
	return err
}

// Public catalog
func (r *repository) GetPublicCatalog(ctx context.Context, restaurantID string) ([]CategoryWithMenus, error) {
	categories, err := r.ListCategories(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	menus, err := r.ListMenus(ctx, restaurantID, nil)
	if err != nil {
		return nil, err
	}

	// Fetch modifiers for each menu
	menuMap := make(map[string][]Menu)
	for _, m := range menus {
		if !m.Available {
			continue
		}
		mods, _ := r.ListModifiersByMenu(ctx, m.ID)
		m.Modifiers = mods

		catID := "uncategorized"
		if m.CategoryID != nil {
			catID = *m.CategoryID
		}
		menuMap[catID] = append(menuMap[catID], m)
	}

	var result []CategoryWithMenus
	for _, cat := range categories {
		if cat.Status != "ACTIVE" {
			continue
		}
		items := menuMap[cat.ID]
		if items == nil {
			items = []Menu{}
		}
		result = append(result, CategoryWithMenus{
			Category: cat,
			Menus:    items,
		})
	}

	// Handle uncategorized if any
	if uncatItems, exists := menuMap["uncategorized"]; exists && len(uncatItems) > 0 {
		result = append(result, CategoryWithMenus{
			Category: Category{
				ID:           "uncategorized",
				RestaurantID: restaurantID,
				Name:         "Lainnya",
				SortOrder:    999,
				Status:       "ACTIVE",
			},
			Menus: uncatItems,
		})
	}

	return result, nil
}
