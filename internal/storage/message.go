package storage

import (
	"chat_app/internal/models"
	"gorm.io/gorm"
)

// MessageRepo encapsulates the logic for accessing messages from the data source.
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo creates a new instance of MessageRepo.
func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// FindAll retrieves all messages from the database.
func (r *MessageRepo) FindAll() ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// FindByID retrieves the message with the provided ID from the database.
func (r *MessageRepo) FindByID(id int) (*models.Message, error) {
	message := new(models.Message)
	err := r.db.Find(message, id).Error
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Create adds a new message to the database.
func (r *MessageRepo) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

// Update modifies an existing message in the database.
func (r *MessageRepo) Update(message *models.Message) error {
	return r.db.Save(message).Error
}

// GetMessages retrieves a list of messages associated with the chatID.
// The 'from' parameter determines the starting ID from which messages are retrieved, and 'limit' specifies the number of messages.
func (r *MessageRepo) GetMessages(chatID, from, limit int) ([]*models.Message, error) {
	var messages []*models.Message
	db := r.db.Preload("Sender").
		Where("chat_id = ?", chatID).
		Limit(limit).
		Order("id desc")

	if from != 0 {
		db = db.Where("id < ?", from)
	}

	err := db.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	return messages, nil
}
