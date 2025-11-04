package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Watchlist represents a user's watchlist collection
type Watchlist struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID       `gorm:"type:uuid;not null;index:idx_watchlists_user_id" json:"user_id"`
	Name      string          `gorm:"type:varchar(255);not null" json:"name"`
	Items     []WatchlistItem `gorm:"foreignKey:WatchlistID;constraint:OnDelete:CASCADE" json:"items,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index:idx_watchlists_deleted_at" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (w *Watchlist) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the Watchlist model
func (Watchlist) TableName() string {
	return "watchlists"
}

// WatchlistItem represents a stock in a watchlist
type WatchlistItem struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WatchlistID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_watchlist_items_watchlist_id" json:"watchlist_id"`
	Symbol          string         `gorm:"type:varchar(20);index:idx_watchlist_items_symbol" json:"symbol,omitempty"`
	Name            string         `gorm:"type:varchar(255);not null;index:idx_watchlist_items_name" json:"name"`
	Price           *float64       `gorm:"type:decimal(15,4)" json:"price,omitempty"`
	AfterHoursPrice *float64       `gorm:"type:decimal(15,4)" json:"afterHoursPrice,omitempty"`
	Change          *float64       `gorm:"type:decimal(15,4)" json:"change,omitempty"`
	PercentChange   string         `gorm:"type:varchar(20)" json:"percentChange,omitempty"`
	Logo            string         `gorm:"type:text" json:"logo,omitempty"`
	Starred         bool           `gorm:"type:boolean;default:false;index:idx_watchlist_items_starred" json:"starred"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index:idx_watchlist_items_deleted_at" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (w *WatchlistItem) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the WatchlistItem model
func (WatchlistItem) TableName() string {
	return "watchlist_items"
}
