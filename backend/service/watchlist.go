package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WatchlistService contains business logic for watchlist operations
type WatchlistService struct {
	db *gorm.DB
}

// NewWatchlistService creates a new instance of WatchlistService
func NewWatchlistService() *WatchlistService {
	return &WatchlistService{
		db: database.GetDB(),
	}
}

// Watchlist CRUD Operations

// CreateWatchlist creates a new watchlist
func (s *WatchlistService) CreateWatchlist(watchlist *model.Watchlist) error {
	if watchlist == nil {
		return errors.New("watchlist cannot be nil")
	}
	if watchlist.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if watchlist.Name == "" {
		return errors.New("name is required")
	}

	result := s.db.Create(watchlist)
	if result.Error != nil {
		return fmt.Errorf("failed to create watchlist: %w", result.Error)
	}

	return nil
}

// GetWatchlistByID fetches a watchlist by ID (with items)
func (s *WatchlistService) GetWatchlistByID(id string) (*model.Watchlist, error) {
	var watchlist model.Watchlist
	result := s.db.Preload("Items").Where("id = ?", id).First(&watchlist)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &watchlist, nil
}

// GetWatchlistsByUserID fetches all watchlists for a user
func (s *WatchlistService) GetWatchlistsByUserID(userID uuid.UUID) ([]model.Watchlist, error) {
	var watchlists []model.Watchlist
	result := s.db.Preload("Items").Where("user_id = ?", userID).Order("created_at DESC").Find(&watchlists)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch watchlists: %w", result.Error)
	}

	return watchlists, nil
}

// UpdateWatchlist updates an existing watchlist
func (s *WatchlistService) UpdateWatchlist(id string, watchlist *model.Watchlist) error {
	if watchlist == nil {
		return errors.New("watchlist cannot be nil")
	}

	var existing model.Watchlist
	result := s.db.Where("id = ?", id).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("record not found")
		}
		return result.Error
	}

	// Update only allowed fields
	existing.Name = watchlist.Name

	updateResult := s.db.Save(&existing)
	if updateResult.Error != nil {
		return fmt.Errorf("failed to update watchlist: %w", updateResult.Error)
	}

	*watchlist = existing
	return nil
}

// DeleteWatchlist deletes a watchlist (soft delete)
func (s *WatchlistService) DeleteWatchlist(id string) error {
	result := s.db.Where("id = ?", id).Delete(&model.Watchlist{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete watchlist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

// WatchlistItem CRUD Operations

// AddItemToWatchlist adds a new item to a watchlist
func (s *WatchlistService) AddItemToWatchlist(watchlistID uuid.UUID, item *model.WatchlistItem) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}
	if item.Name == "" {
		return errors.New("name is required")
	}

	// Check if watchlist exists
	var watchlist model.Watchlist
	if err := s.db.Where("id = ?", watchlistID).First(&watchlist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("watchlist not found")
		}
		return err
	}

	// Check if item already exists in watchlist
	var existingItem model.WatchlistItem
	result := s.db.Where("watchlist_id = ? AND name = ?", watchlistID, item.Name).First(&existingItem)
	if result.Error == nil {
		return errors.New("item already exists in watchlist")
	}
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	// Set watchlist ID and create item
	item.WatchlistID = watchlistID
	createResult := s.db.Create(item)
	if createResult.Error != nil {
		return fmt.Errorf("failed to add item to watchlist: %w", createResult.Error)
	}

	return nil
}

// GetWatchlistItems fetches all items for a watchlist
func (s *WatchlistService) GetWatchlistItems(watchlistID uuid.UUID) ([]model.WatchlistItem, error) {
	var items []model.WatchlistItem
	result := s.db.Where("watchlist_id = ?", watchlistID).Order("created_at ASC").Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch watchlist items: %w", result.Error)
	}

	return items, nil
}

// GetWatchlistItemByID fetches a watchlist item by ID
func (s *WatchlistService) GetWatchlistItemByID(id string) (*model.WatchlistItem, error) {
	var item model.WatchlistItem
	result := s.db.Where("id = ?", id).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &item, nil
}

// UpdateWatchlistItem updates an existing watchlist item
func (s *WatchlistService) UpdateWatchlistItem(id string, item *model.WatchlistItem) error {
	if item == nil {
		return errors.New("item cannot be nil")
	}

	var existing model.WatchlistItem
	result := s.db.Where("id = ?", id).First(&existing)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("record not found")
		}
		return result.Error
	}

	// Update allowed fields
	if item.Name != "" {
		existing.Name = item.Name
	}
	if item.Price != nil {
		existing.Price = item.Price
	}
	if item.AfterHoursPrice != nil {
		existing.AfterHoursPrice = item.AfterHoursPrice
	}
	if item.Change != nil {
		existing.Change = item.Change
	}
	if item.PercentChange != "" {
		existing.PercentChange = item.PercentChange
	}
	if item.Logo != "" {
		existing.Logo = item.Logo
	}
	existing.Starred = item.Starred

	updateResult := s.db.Save(&existing)
	if updateResult.Error != nil {
		return fmt.Errorf("failed to update watchlist item: %w", updateResult.Error)
	}

	*item = existing
	return nil
}

// DeleteWatchlistItem removes an item from a watchlist
func (s *WatchlistService) DeleteWatchlistItem(id string) error {
	result := s.db.Where("id = ?", id).Delete(&model.WatchlistItem{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete watchlist item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

// ToggleItemStarred toggles the starred status of a watchlist item
func (s *WatchlistService) ToggleItemStarred(id string) (*model.WatchlistItem, error) {
	var item model.WatchlistItem
	result := s.db.Where("id = ?", id).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	item.Starred = !item.Starred
	updateResult := s.db.Save(&item)
	if updateResult.Error != nil {
		return nil, fmt.Errorf("failed to toggle starred status: %w", updateResult.Error)
	}

	return &item, nil
}

// GetStarredItems fetches all starred items for a user's watchlists
func (s *WatchlistService) GetStarredItems(userID uuid.UUID) ([]model.WatchlistItem, error) {
	var items []model.WatchlistItem
	result := s.db.Joins("JOIN watchlists ON watchlist_items.watchlist_id = watchlists.id").
		Where("watchlists.user_id = ? AND watchlist_items.starred = ?", userID, true).
		Order("watchlist_items.created_at DESC").
		Find(&items)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch starred items: %w", result.Error)
	}

	return items, nil
}

// BatchUpdateItems updates multiple items in a watchlist (useful for price updates)
func (s *WatchlistService) BatchUpdateItems(items []model.WatchlistItem) error {
	if len(items) == 0 {
		return errors.New("items cannot be empty")
	}

	for _, item := range items {
		if item.ID == uuid.Nil {
			return errors.New("item ID is required for batch update")
		}

		updateResult := s.db.Model(&model.WatchlistItem{}).
			Where("id = ?", item.ID).
			Updates(map[string]interface{}{
				"price":             item.Price,
				"after_hours_price": item.AfterHoursPrice,
				"change":            item.Change,
				"percent_change":    item.PercentChange,
				"logo":              item.Logo,
			})
		if updateResult.Error != nil {
			return fmt.Errorf("failed to update item %s: %w", item.ID, updateResult.Error)
		}
	}

	return nil
}
