package storage

import (
	"chat_app/internal/models"
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepo represents a repository for accessing and manipulating User data.
type UserRepo struct {
	db    *gorm.DB
	cache *redis.Client
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(db *gorm.DB, cache *redis.Client) *UserRepo {
	return &UserRepo{
		db:    db,
		cache: cache,
	}
}

// FindAll retrieves all users from the database.
func (r *UserRepo) FindAll() ([]*models.User, error) {
	var users []*models.User
	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindByID retrieves a user by ID from the database.
func (r *UserRepo) FindByID(id int) (*models.User, error) {
	user := new(models.User)
	err := r.db.Preload("FollowedUsers").Preload("BlockedUsers").Preload("Chats.Users").Preload("Devices").Find(user, id).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByUsername retrieves a user by username from the database.
func (r *UserRepo) FindByUsername(username string) (*models.User, error) {
	user := new(models.User)
	err := r.db.Where("username = ?", username).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindArrayByPartUsername retrieves users whose username starts with a given substring.
func (r *UserRepo) FindArrayByPartUsername(username string, order string, limit int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.Where("username LIKE ?", username+"%").Order("username " + order).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// Create adds a new user to the database.
func (r *UserRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user in the database.
func (r *UserRepo) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdateWithAssociations updates a user along with its associations in the database.
func (r *UserRepo) UpdateWithAssociations(user *models.User) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(user).Error
}

// ReplaceFollowedUsers replaces a user's followed users in the database.
func (r *UserRepo) ReplaceFollowedUsers(user *models.User, followedUsers []*models.User) error {
	return r.db.Model(user).Association("FollowedUsers").Replace(followedUsers)
}

// ReplaceBlockedUsers replaces a user's blocked users in the database.
func (r *UserRepo) ReplaceBlockedUsers(user *models.User, blockedUsers []*models.User) error {
	return r.db.Model(user).Association("BlockedUsers").Replace(blockedUsers)
}

// Delete deletes a user from the database.
func (r *UserRepo) Delete(user *models.User) error {
	return r.db.Delete(user).Error
}

// ResetPassword changes a user's password.
func (r *UserRepo) ResetPassword(user *models.User, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	return r.db.Save(user).Error
}

// PasswordMatches checks if a plain text password matches the hashed password of a user.
func (r *UserRepo) PasswordMatches(user *models.User, plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
