package models

type Chat struct {
	ID    int     `json:"id" gorm:"primaryKey" example:"1"`
	Type  string  `json:"type" example:"private"`
	Users []*User `json:"users,omitempty" gorm:"many2many:chat_users"`
}

type ChatRepository interface {
	FindByID(id int) (*Chat, error)
	FindByUserID(id int) (*[]Chat, error)
	FindPrivateChatByUsersArray(users []*User) (*Chat, error)
	Create(chat *Chat) error
	Update(chat *Chat) error
	Delete(chat *Chat) error
}
