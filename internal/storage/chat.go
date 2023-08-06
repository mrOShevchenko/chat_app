package storage

import (
	"chat_app/internal/models"
	"gorm.io/gorm"
)

// ChatRepo encapsulates the logic for accessing chats from the data source.
type ChatRepo struct {
	db *gorm.DB
}

// NewChatRepo creates a new instance of ChatRepo.
func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// FindByID retrieves the chat with the provided ID from the database.
func (r *ChatRepo) FindByID(id int) (*models.Chat, error) {
	chat := new(models.Chat)
	err := r.db.Preload("Users.BlockedUsers").Find(chat, id).Error
	if err != nil {
		return nil, err
	}
	return chat, nil
}

// FindByUserID retrieves all chats of the user with the provided ID.
func (r *ChatRepo) FindByUserID(id int) ([]*models.Chat, error) {
	var chats []*models.Chat
	err := r.db.Preload("Users").
		Joins("JOIN chat_users ON chat_users.chat_id = chats.id").
		Where("chat_users.user_id = ?", id).
		Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

// FindPrivateChatByUsersArray finds the private chat for the provided users.
func (r *ChatRepo) FindPrivateChatByUsersArray(users []*models.User) (*models.Chat, error) {
	chat := new(models.Chat)
	err := r.db.Preload("Users").
		Where("type = ?", "private").
		Joins("JOIN chat_users ON chat_users.chat_id = chats.id").
		Where("chat_users.user_id IN ?", extractUserIDs(users)).
		Group("chats.id").
		Having("COUNT(DISTINCT chat_users.user_id) = ?", len(users)).
		First(chat).Error

	if err != nil {
		return nil, err
	}

	return chat, nil
}

// Create adds a new chat to the database.
func (r *ChatRepo) Create(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

// Update modifies an existing chat in the database.
func (r *ChatRepo) Update(chat *models.Chat) error {
	return r.db.Save(chat).Error
}

// Delete removes a chat from the database.
func (r *ChatRepo) Delete(chat *models.Chat) error {
	if err := r.db.Table("chat_users").Where("chat_id = ?", chat.ID).Delete(&models.User{}).Error; err != nil {
		return err
	}
	return r.db.Delete(chat).Error
}

// extractUserIDs extracts the IDs from the provided users.
func extractUserIDs(users []*models.User) []int {
	ids := make([]int, len(users))
	for i, user := range users {
		ids[i] = user.ID
	}
	return ids
}
