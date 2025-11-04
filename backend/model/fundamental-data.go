package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FundamentalData represents financial statement data
type FundamentalData struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Symbol        string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_fundamental_symbol_type_freq" json:"symbol"`
	StatementType string         `gorm:"type:varchar(50);not null;uniqueIndex:idx_fundamental_symbol_type_freq" json:"statement_type"`
	Frequency     string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_fundamental_symbol_type_freq" json:"frequency"`
	Statement     string         `gorm:"type:jsonb;not null" json:"statement"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (f *FundamentalData) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the FundamentalData model
func (FundamentalData) TableName() string {
	return "fundamental_data"
}
