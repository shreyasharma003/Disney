package models

import (
	"time"
)

//User Table
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // excluded from JSON
	Role         string    `gorm:"type:varchar(50);default:'user';not null" json:"role"` // user/admin
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

//Table naming manually
func (User) TableName() string {
	return "users"
}

//Genre Table
type Genre struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
}

// Table naming manually
func (Genre) TableName() string {
	return "genres"
}
