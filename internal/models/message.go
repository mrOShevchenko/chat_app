package models

type Message struct {
	ID        int    `json:"id" gorm:"primaryKey" example:"1"`
	Type      string `json:"type" example:"text"` // text, error, system
	ChatID    int    `json:"chatID,omitempty" gorm:"foreignKey:ChatID" example:"1"`
	Content   string `json:"content" example:"twit-twit"`
	SenderID  int    `json:"senderId,omitempty" gorm:"foreignKey:SenderID" example:"2"`
	Sender    User   `json:"sender,omitempty"`
	Status    string `json:"status,omitempty" example:"sent"` // sent, delivered, read
	CreatedAt int64  `json:"createdAt,omitempty" example:"1230000000"`
}

type MessageRepository interface {
	FindByID(id int) (*Message, error)
	FindAll() (*[]Message, error)
	Create(meassage *Message) error
	Update(message *Message) error
	GetMessages(chatID, from, limit int) (*[]Message, error)
}
