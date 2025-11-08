package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MarketStatistics represents daily aggregated market statistics
type MarketStatistics struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Date            time.Time      `gorm:"type:date;not null;uniqueIndex" json:"date"`
	StocksUp        int            `gorm:"type:integer;not null;default:0" json:"stocksUp"`        // +0.01% or more
	StocksDown      int            `gorm:"type:integer;not null;default:0" json:"stocksDown"`        // Below -0.01%
	StocksUnchanged int            `gorm:"type:integer;not null;default:0" json:"stocksUnchanged"`   // Between -0.01% and +0.01%
	TotalStocks     int            `gorm:"type:integer;not null;default:0" json:"totalStocks"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (m *MarketStatistics) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the MarketStatistics model
func (MarketStatistics) TableName() string {
	return "market_statistics"
}

