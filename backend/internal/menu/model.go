package menu

import "time"

type Category struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	Name         string    `json:"name"`
	SortOrder    int       `json:"sort_order"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Menu struct {
	ID           string           `json:"id"`
	RestaurantID string           `json:"restaurant_id"`
	CategoryID   *string          `json:"category_id"`
	CategoryName *string          `json:"category_name,omitempty"`
	Name         string           `json:"name"`
	Description  *string          `json:"description"`
	Price        int64            `json:"price"` // in Rupiah (e.g. 25000)
	ImageURL     *string          `json:"image_url"`
	Available    bool             `json:"available"`
	SortOrder    int              `json:"sort_order"`
	Modifiers    []Modifier       `json:"modifiers,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type Modifier struct {
	ID           string           `json:"id"`
	RestaurantID string           `json:"restaurant_id"`
	MenuID       *string          `json:"menu_id"`
	Name         string           `json:"name"`
	Type         string           `json:"type"` // SINGLE, MULTIPLE
	Required     bool             `json:"required"`
	Options      []ModifierOption `json:"options"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type ModifierOption struct {
	ID              string    `json:"id"`
	ModifierID      string    `json:"modifier_id"`
	Name            string    `json:"name"`
	AdditionalPrice int64     `json:"additional_price"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CategoryWithMenus struct {
	Category
	Menus []Menu `json:"menus"`
}
