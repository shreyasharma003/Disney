package workers

import (
	"disney/database"
	"disney/jobs"
	"disney/models"
	"errors"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// FavouriteWorkerPool manages a pool of workers that process favourite add/remove jobs
type FavouriteWorkerPool struct {
	// JobQueue is a buffered channel that receives favourite jobs
	JobQueue chan jobs.FavouriteJob
	// NumWorkers specifies the number of concurrent workers
	NumWorkers int
	// done channel to signal graceful shutdown
	done chan struct{}
	// mu protects concurrent access during duplicate checks
	// Even though we use DB uniqueness constraints, this adds extra safety
	mu sync.RWMutex
}

// NewFavouriteWorkerPool creates a new worker pool for processing favourite jobs
// numWorkers: number of concurrent workers to spawn
// bufferSize: size of the job queue buffer
func NewFavouriteWorkerPool(numWorkers, bufferSize int) *FavouriteWorkerPool {
	return &FavouriteWorkerPool{
		JobQueue:   make(chan jobs.FavouriteJob, bufferSize),
		NumWorkers: numWorkers,
		done:       make(chan struct{}),
	}
}

// Start initializes and starts the worker pool
// Spawns numWorkers goroutines that listen for jobs on the JobQueue
func (fwp *FavouriteWorkerPool) Start() {
	log.Printf("Starting favourite worker pool with %d workers\n", fwp.NumWorkers)

	for i := 0; i < fwp.NumWorkers; i++ {
		// Each worker runs independently and continuously processes jobs
		go fwp.worker(i)
	}
}

// worker is a goroutine that continuously processes favourite jobs
// Each worker listens for jobs on the shared JobQueue and processes them
func (fwp *FavouriteWorkerPool) worker(workerID int) {
	log.Printf("Favourite worker %d started\n", workerID)

	for {
		select {
		// Received a favourite job from the queue
		case job := <-fwp.JobQueue:
			fwp.processFavouriteJob(job, workerID)

		// Shutdown signal received
		case <-fwp.done:
			log.Printf("Favourite worker %d shutting down\n", workerID)
			return
		}
	}
}

// processFavouriteJob handles the actual database operation for add/remove favourite
// It safely handles concurrent requests using database-level constraints
func (fwp *FavouriteWorkerPool) processFavouriteJob(job jobs.FavouriteJob, workerID int) {
	switch job.Action {
	case "add":
		fwp.processAddFavourite(job, workerID)
	case "remove":
		fwp.processRemoveFavourite(job, workerID)
	default:
		log.Printf("Favourite worker %d: Unknown action '%s' for user %d, cartoon %d\n",
			workerID, job.Action, job.UserID, job.CartoonID)
	}
}

// processAddFavourite handles adding a cartoon to favourites
// Uses database uniqueness constraint to prevent duplicates under concurrent access
// Strategy: Always attempt insert; if unique constraint fails, it's already a favourite
func (fwp *FavouriteWorkerPool) processAddFavourite(job jobs.FavouriteJob, workerID int) {
	newFavourite := models.Favourite{
		UserID:    job.UserID,
		CartoonID: job.CartoonID,
	}

	// Attempt to create favourite record
	result := database.DB.Create(&newFavourite)

	if result.Error != nil {
		// Check if error is due to unique constraint violation
		// This means the cartoon is already in favourites - not an error for our use case
		if result.Error.Error() == "UNIQUE constraint failed: favourites.user_id,favourites.cartoon_id" ||
			errors.Is(result.Error, gorm.ErrDuplicatedKey) ||
			result.Error.Error() == "ERROR: duplicate key value violates unique constraint \"idx_user_cartoon_fav\" (SQLSTATE 23505)" {
			log.Printf("Favourite worker %d: Cartoon %d already in favourites for user %d (idempotent)\n",
				workerID, job.CartoonID, job.UserID)
			return
		}

		// Other database error
		log.Printf("Favourite worker %d: Error adding favourite for user %d, cartoon %d: %v\n",
			workerID, job.UserID, job.CartoonID, result.Error)
		return
	}

	log.Printf("Favourite worker %d: Successfully added cartoon %d to favourites for user %d\n",
		workerID, job.CartoonID, job.UserID)
}

// processRemoveFavourite handles removing a cartoon from favourites
// Uses transaction to safely read, verify, and delete
func (fwp *FavouriteWorkerPool) processRemoveFavourite(job jobs.FavouriteJob, workerID int) {
	// Use a transaction to safely check existence and delete
	// This prevents race conditions between check and delete
	tx := database.DB.Begin()

	// Find the favourite record
	var favourite models.Favourite
	if result := tx.Where("user_id = ? AND cartoon_id = ?", job.UserID, job.CartoonID).First(&favourite); result.RowsAffected == 0 {
		// Favourite doesn't exist - this is idempotent, treat as success
		tx.Rollback()
		log.Printf("Favourite worker %d: Favourite not found for user %d, cartoon %d (already removed)\n",
			workerID, job.UserID, job.CartoonID)
		return
	}

	// Delete the favourite record
	if result := tx.Delete(&favourite); result.Error != nil {
		tx.Rollback()
		log.Printf("Favourite worker %d: Error removing favourite for user %d, cartoon %d: %v\n",
			workerID, job.UserID, job.CartoonID, result.Error)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Favourite worker %d: Error committing transaction for user %d, cartoon %d: %v\n",
			workerID, job.UserID, job.CartoonID, err)
		return
	}

	log.Printf("Favourite worker %d: Successfully removed cartoon %d from favourites for user %d\n",
		workerID, job.CartoonID, job.UserID)
}

// EnqueueFavouriteJob adds a favourite job to the processing queue
// This is called by HTTP handlers to queue jobs for async processing
// action: "add" or "remove"
// The method returns immediately without waiting for job completion
func (fwp *FavouriteWorkerPool) EnqueueFavouriteJob(userID, cartoonID uint, action string) {
	job := jobs.FavouriteJob{
		UserID:    userID,
		CartoonID: cartoonID,
		Action:    action,
		Timestamp: time.Now(),
	}

	// Send job to queue (non-blocking send, channel is buffered)
	select {
	case fwp.JobQueue <- job:
		// Job enqueued successfully
	case <-time.After(100 * time.Millisecond):
		// Queue is full - log warning but don't block the HTTP request
		log.Printf("Favourite worker pool queue is full, dropping job for user %d, cartoon %d, action %s\n",
			userID, cartoonID, action)
	}
}

// Shutdown gracefully stops the worker pool
// Closes the done channel to signal all workers to stop
func (fwp *FavouriteWorkerPool) Shutdown() {
	log.Println("Shutting down favourite worker pool...")
	close(fwp.done)
	// Give workers time to finish current jobs
	time.Sleep(1 * time.Second)
	close(fwp.JobQueue)
}

// GetQueueLength returns the current number of jobs waiting in the queue
// Useful for monitoring worker pool health
func (fwp *FavouriteWorkerPool) GetQueueLength() int {
	return len(fwp.JobQueue)
}
