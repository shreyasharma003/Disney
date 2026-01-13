package models

import (
	"time"
)

// User Table
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`                  // excluded from JSON
	Role         string    `gorm:"type:varchar(50);default:'user';not null" json:"role"` // user/admin
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Age          int       `gorm:"type:int;not null" json:"age"`
}

// Table naming manually
func (User) TableName() string {
	return "users"
}

// Genre Table
type Genre struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
}

// Table naming manually
func (Genre) TableName() string {
	return "genres"
}

// AgeGroup Table
type AgeGroup struct {
	ID    uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Label string `gorm:"type:varchar(50);not null;uniqueIndex" json:"label"`
}

// Table naming manually
func (AgeGroup) TableName() string {
	return "age_groups"
}

// Cartoon Table
type Cartoon struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"type:varchar(255);not null;index" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	PosterURL   string    `gorm:"type:varchar(500)" json:"poster_url"`
	ReleaseYear int       `gorm:"type:int" json:"release_year"`
	GenreID     uint      `gorm:"not null;index" json:"genre_id"`
	AgeGroupID  uint      `gorm:"not null;index" json:"age_group_id"`
	IsFeatured  bool      `gorm:"default:false;index" json:"is_featured"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Foreign key relationships
	Genre      Genre       `gorm:"foreignKey:GenreID;constraint:OnDelete:RESTRICT" json:"genre,omitempty"`
	AgeGroup   AgeGroup    `gorm:"foreignKey:AgeGroupID;constraint:OnDelete:RESTRICT" json:"age_group,omitempty"`
	Characters []Character `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"characters,omitempty"`
}

// Table naming manually
func (Cartoon) TableName() string {
	return "cartoons"
}

// Character Table
type Character struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"type:varchar(255);not null;index" json:"name"`
	ImageURL  string `gorm:"type:varchar(500)" json:"image_url"`
	CartoonID uint   `gorm:"not null;index" json:"cartoon_id"`

	// Foreign key relationship
	Cartoon Cartoon `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"cartoon,omitempty"`
}

// Table naming manually
func (Character) TableName() string {
	return "characters"
}

// Rating Table
type Rating struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_user_cartoon" json:"user_id"`
	CartoonID uint      `gorm:"not null;index:idx_user_cartoon" json:"cartoon_id"`
	Rating    int       `gorm:"type:int;not null;check:rating >= 1 AND rating <= 10" json:"rating"`
	CreatedAt time.Time `json:"created_at"`

	// Foreign key relationships
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Cartoon Cartoon `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"cartoon,omitempty"`
}

// Table naming manually
func (Rating) TableName() string {
	return "ratings"
}

// Favourite Table
type Favourite struct {
	ID        uint `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint `gorm:"not null;uniqueIndex:idx_user_cartoon_fav" json:"user_id"`
	CartoonID uint `gorm:"not null;uniqueIndex:idx_user_cartoon_fav" json:"cartoon_id"`

	// Foreign key relationships
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Cartoon Cartoon `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"cartoon,omitempty"`
}

// Table naming manually
func (Favourite) TableName() string {
	return "favourites"
}

// View Table (analytics)
type View struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CartoonID uint      `gorm:"not null;index" json:"cartoon_id"`
	UserID    *uint     `gorm:"index" json:"user_id,omitempty"` // nullable
	ViewedAt  time.Time `gorm:"not null;index" json:"viewed_at"`

	// Foreign key relationships
	Cartoon Cartoon `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"cartoon,omitempty"`
	User    *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

// Table naming manually
func (View) TableName() string {
	return "views"
}

// AdminLog Table
type AdminLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID   uint      `gorm:"not null;index" json:"admin_id"`
	Action    string    `gorm:"type:varchar(255);not null" json:"action"`
	Entity    string    `gorm:"type:varchar(255);not null" json:"entity"`
	CreatedAt time.Time `json:"created_at"`

	// Foreign key relationship
	Admin User `gorm:"foreignKey:AdminID;constraint:OnDelete:CASCADE" json:"admin,omitempty"`
}

// Table naming manually
func (AdminLog) TableName() string {
	return "admin_logs"
}

// RequestLog Table
type RequestLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       *uint     `gorm:"index" json:"user_id,omitempty"` // nullable for anonymous requests
	Endpoint     string    `gorm:"type:varchar(500);not null;index" json:"endpoint"`
	Method       string    `gorm:"type:varchar(10);not null" json:"method"`
	ResponseTime int       `gorm:"type:int" json:"response_time"` // in milliseconds
	StatusCode   int       `gorm:"type:int;index" json:"status_code"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`

	// Foreign key relationship
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

// Table naming manually
func (RequestLog) TableName() string {
	return "request_logs"
}

// TimeTable Table (Show Schedule)
type TimeTable struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CartoonID uint      `gorm:"not null;index" json:"cartoon_id"`
	ShowTime  time.Time `gorm:"not null;index" json:"show_time"`
	DayOfWeek string    `gorm:"type:varchar(20);not null" json:"day_of_week"` // Monday, Tuesday, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Foreign key relationship
	Cartoon Cartoon `gorm:"foreignKey:CartoonID;constraint:OnDelete:CASCADE" json:"cartoon,omitempty"`
}

// Table naming manually
func (TimeTable) TableName() string {
	return "time_tables"
}
