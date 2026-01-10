package jobs

import "time"

// ViewJob represents a job to record a cartoon view
// Used by worker pool to safely process view recordings under high concurrency
type ViewJob struct {
	UserID    uint
	CartoonID uint
	Timestamp time.Time
}

// FavouriteJob represents a job to add or remove a favourite
// Used by worker pool to safely process favourite operations under high concurrency
type FavouriteJob struct {
	UserID    uint
	CartoonID uint
	Action    string // "add" or "remove"
	Timestamp time.Time
}

// ViewJobResponse represents the result of a view job (for response tracking)
type ViewJobResponse struct {
	Success   bool
	Error     string
	ViewCount int64
}

// FavouriteJobResponse represents the result of a favourite job
type FavouriteJobResponse struct {
	Success bool
	Error   string
	Message string
}
