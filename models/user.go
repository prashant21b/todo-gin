package models

import (
	"context"
	"time"

	"todo-app/config"
)

// User represents the user model
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:255;unique;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"` // "-" hides password from JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Todos     []Todo    `json:"todos,omitempty" gorm:"foreignKey:UserID"`
}

// CreateUser inserts a new user into the database with context support
func CreateUser(ctx context.Context, user *User) error {
	return config.GetDBWithContext(ctx).Create(user).Error
}

// GetUserByEmail retrieves a user by their email address with context support
func GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := config.GetDBWithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID with context support
func GetUserByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := config.GetDBWithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
