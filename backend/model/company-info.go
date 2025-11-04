package model

import (
	"time"

	"gorm.io/gorm"
)

// CompanyInfo represents company information and statistics
type CompanyInfo struct {
	Symbol           string         `gorm:"type:varchar(20);primaryKey;not null" json:"symbol"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Price            string         `gorm:"type:varchar(50)" json:"price,omitempty"`
	AfterHoursPrice  string         `gorm:"type:varchar(50)" json:"afterHoursPrice,omitempty"`
	Change           string         `gorm:"type:varchar(50)" json:"change,omitempty"`
	PercentChange    string         `gorm:"type:varchar(50)" json:"percentChange,omitempty"`
	Open             string         `gorm:"type:varchar(50)" json:"open,omitempty"`
	High             string         `gorm:"type:varchar(50)" json:"high,omitempty"`
	Low              string         `gorm:"type:varchar(50)" json:"low,omitempty"`
	YearHigh         string         `gorm:"type:varchar(50)" json:"yearHigh,omitempty"`
	YearLow          string         `gorm:"type:varchar(50)" json:"yearLow,omitempty"`
	Volume           int64          `gorm:"type:bigint" json:"volume,omitempty"`
	AvgVolume        int64          `gorm:"type:bigint" json:"avgVolume,omitempty"`
	MarketCap        string         `gorm:"type:varchar(50)" json:"marketCap,omitempty"`
	Beta             string         `gorm:"type:varchar(50)" json:"beta,omitempty"`
	PE               string         `gorm:"type:varchar(50)" json:"pe,omitempty"`
	EarningsDate     string         `gorm:"type:varchar(100)" json:"earningsDate,omitempty"`
	Sector           string         `gorm:"type:varchar(255)" json:"sector,omitempty"`
	Industry         string         `gorm:"type:varchar(255)" json:"industry,omitempty"`
	About            string         `gorm:"type:text" json:"about,omitempty"`
	Employees        string         `gorm:"type:varchar(50)" json:"employees,omitempty"`
	FiveDaysReturn   string         `gorm:"type:varchar(50)" json:"fiveDaysReturn,omitempty"`
	OneMonthReturn   string         `gorm:"type:varchar(50)" json:"oneMonthReturn,omitempty"`
	ThreeMonthReturn string         `gorm:"type:varchar(50)" json:"threeMonthReturn,omitempty"`
	SixMonthReturn   string         `gorm:"type:varchar(50)" json:"sixMonthReturn,omitempty"`
	YtdReturn        string         `gorm:"type:varchar(50)" json:"ytdReturn,omitempty"`
	YearReturn       string         `gorm:"type:varchar(50)" json:"yearReturn,omitempty"`
	ThreeYearReturn  string         `gorm:"type:varchar(50)" json:"threeYearReturn,omitempty"`
	FiveYearReturn   string         `gorm:"type:varchar(50)" json:"fiveYearReturn,omitempty"`
	TenYearReturn    string         `gorm:"type:varchar(50)" json:"tenYearReturn,omitempty"`
	MaxReturn        string         `gorm:"type:varchar(50)" json:"maxReturn,omitempty"`
	Logo             string         `gorm:"type:text" json:"logo,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for the CompanyInfo model
func (CompanyInfo) TableName() string {
	return "company_info"
}
