package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScreenerResult represents cached screener results in the database
type ScreenerResult struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type      string         `gorm:"type:varchar(50);not null;index;uniqueIndex:uniq_screener_result_type_symbol_date" json:"type"` // "inside_day", "high_volume_quarter", "high_volume_year", "high_volume_ever"
	Symbol    string         `gorm:"type:varchar(10);not null;index;uniqueIndex:uniq_screener_result_type_symbol_date" json:"symbol"`
	Date      time.Time      `gorm:"type:date;not null;index;uniqueIndex:uniq_screener_result_type_symbol_date" json:"date"` // Date when the symbol matched the criteria
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (s *ScreenerResult) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the ScreenerResult model
func (ScreenerResult) TableName() string {
	return "screener_results"
}
