package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Screener represents stock market data in the system
type Screener struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Symbol    string         `gorm:"type:varchar(20);not null;index" json:"symbol"`
	Open      float64        `gorm:"type:decimal(15,4);not null" json:"open"`
	High      float64        `gorm:"type:decimal(15,4);not null" json:"high"`
	Low       float64        `gorm:"type:decimal(15,4);not null" json:"low"`
	Close     float64        `gorm:"type:decimal(15,4);not null" json:"close"`
	Volume    int64          `gorm:"type:bigint;not null" json:"volume"`
	Logo      string         `gorm:"type:text" json:"logo,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (s *Screener) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the Screener model
func (Screener) TableName() string {
	return "screener"
}
