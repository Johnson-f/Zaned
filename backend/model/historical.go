package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Historical represents historical stock price data in the system
type Historical struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Symbol    string         `gorm:"type:varchar(20);not null;index" json:"symbol"`
	Epoch     int64          `gorm:"type:bigint;not null;index" json:"epoch"`
	Range     string         `gorm:"type:varchar(10);not null;column:range" json:"range"`
	Interval  string         `gorm:"type:varchar(10);not null;column:interval" json:"interval"`
	Open      float64        `gorm:"type:decimal(15,4);not null" json:"open"`
	High      float64        `gorm:"type:decimal(15,4);not null" json:"high"`
	Low       float64        `gorm:"type:decimal(15,4);not null" json:"low"`
	Close     float64        `gorm:"type:decimal(15,4);not null" json:"close"`
	AdjClose  *float64       `gorm:"type:decimal(15,4)" json:"adjClose,omitempty"`
	Volume    int64          `gorm:"type:bigint;not null" json:"volume"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate hook to generate UUID if not set
func (h *Historical) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name for the Historical model
func (Historical) TableName() string {
	return "historical"
}
