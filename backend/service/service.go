package service

import (
	"errors"
	"screener/backend/database"
	"screener/backend/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExampleService contains business logic for example operations
type ExampleService struct {
	db *gorm.DB
}

// NewExampleService creates a new instance of ExampleService
func NewExampleService() *ExampleService {
	return &ExampleService{
		db: database.GetDB(),
	}
}

// GetUserData fetches user-specific data based on user ID
func (s *ExampleService) GetUserData(userID string) (*model.Example, error) {
	// Parse userID string to UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var example model.Example
	result := s.db.Where("user_id = ?", uid).First(&example)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &example, nil
}

// CreateUserData creates a new example record for a user
func (s *ExampleService) CreateUserData(userID string, content string) (*model.Example, error) {
	// Parse userID string to UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	example := &model.Example{
		UserID:  uid,
		Content: content,
	}

	result := s.db.Create(example)
	if result.Error != nil {
		return nil, result.Error
	}

	return example, nil
}

// ProcessUserRequest processes a user request with their ID
func (s *ExampleService) ProcessUserRequest(userID string, data string) error {
	// Parse userID string to UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	// Example: Update user's example data
	example := &model.Example{}
	result := s.db.Where("user_id = ?", uid).First(example)
	if result.Error != nil {
		return result.Error
	}

	example.Content = data
	result = s.db.Save(example)
	return result.Error
}

// GetUserExamples fetches all examples for a user
func (s *ExampleService) GetUserExamples(userID string) ([]model.Example, error) {
	// Parse userID string to UUID
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	var examples []model.Example
	result := s.db.Where("user_id = ?", uid).Find(&examples)
	if result.Error != nil {
		return nil, result.Error
	}

	return examples, nil
}

