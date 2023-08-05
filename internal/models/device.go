package models

import "time"

type Device struct {
	ID        int       `json:"id" gorm:"primaryKey" example:"1"`
	UserID    int       `json:"userId" example:"1"`
	Type      string    `json:"type" example:"web"` // web, android, ios
	Name      string    `json:"name" example:"Chrome 90.0.4430.212 (Linux x86_64)"`
	Token     string    `json:"token" example:"c26BG3n3VJbI0i2H8aXZmG:APA91bGoG3iqJxidumkVRKCniXoA-QsYfSXpc6qhAiWcaIVtAUG9nNKsxoEkL8j4ZVezXVFzJDpIYS6JHtpcJ2af0686djcfKDltLqVLkuWFVHoEnz9NtKV9hgQmof7MURLYQsaokGfM"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
