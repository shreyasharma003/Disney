package workers

import (
	"disney/database"
	"disney/jobs"
	"disney/models"
	"log"
	"time"
)

// ViewWorkerPool manages a pool of workers that process view recording jobs
type ViewWorkerPool struct {
	// JobQueue is a buffered channel that receives view jobs
	JobQueue chan jobs.ViewJob
	// NumWorkers specifies the number of concurrent workers
	NumWorkers int
	// done channel to signal graceful shutdown
	done chan struct{}
}

// NewViewWorkerPool creates a new worker pool for processing view jobs
// numWorkers: number of concurrent workers to spawn
// bufferSize: size of the job queue buffer
func NewViewWorkerPool(numWorkers, bufferSize int) *ViewWorkerPool {
	return &ViewWorkerPool{
		JobQueue:   make(chan jobs.ViewJob, bufferSize),
		NumWorkers: numWorkers,
		done:       make(chan struct{}),
	}
}

// Start initializes and starts the worker pool
// Spawns numWorkers goroutines that listen for jobs on the JobQueue
func (vwp *ViewWorkerPool) Start() {
	log.Printf("Starting view worker pool with %d workers\n", vwp.NumWorkers)

	for i := 0; i < vwp.NumWorkers; i++ {
		// Each worker runs independently and continuously processes jobs
		go vwp.worker(i)
	}
}

// worker is a goroutine that continuously processes view jobs
// Each worker listens for jobs on the shared JobQueue and processes them
func (vwp *ViewWorkerPool) worker(workerID int) {
	log.Printf("View worker %d started\n", workerID)

	for {
		select {
		// Received a view job from the queue
		case job := <-vwp.JobQueue:
			vwp.processViewJob(job, workerID)

		// Shutdown signal received
		case <-vwp.done:
			log.Printf("View worker %d shutting down\n", workerID)
			return
		}
	}
}

// processViewJob handles the actual database operation for recording a view
// It creates a View record with user and cartoon IDs
// Uses GORM to safely insert the view record
func (vwp *ViewWorkerPool) processViewJob(job jobs.ViewJob, workerID int) {
	// Create new view record
	newView := models.View{
		CartoonID: job.CartoonID,
		UserID:    &job.UserID,
		ViewedAt:  job.Timestamp,
	}

	// Insert view into database
	// GORM handles this atomically, so multiple workers writing
	// different views won't cause race conditions
	if err := database.DB.Create(&newView).Error; err != nil {
		log.Printf("View worker %d: Error recording view for user %d, cartoon %d: %v\n",
			workerID, job.UserID, job.CartoonID, err)
		return
	}

	log.Printf("View worker %d: Successfully recorded view for user %d, cartoon %d\n",
		workerID, job.UserID, job.CartoonID)
}

// EnqueueViewJob adds a view job to the processing queue
// This is called by HTTP handlers to queue jobs for async processing
// The method returns immediately without waiting for job completion
func (vwp *ViewWorkerPool) EnqueueViewJob(userID, cartoonID uint) {
	job := jobs.ViewJob{
		UserID:    userID,
		CartoonID: cartoonID,
		Timestamp: time.Now(),
	}

	// Send job to queue (non-blocking send, channel is buffered)
	select {
	case vwp.JobQueue <- job:
		// Job enqueued successfully
	case <-time.After(100 * time.Millisecond):
		// Queue is full - log warning but don't block the HTTP request
		log.Printf("View worker pool queue is full, dropping job for user %d, cartoon %d\n",
			userID, cartoonID)
	}
}

// Shutdown gracefully stops the worker pool
// Closes the done channel to signal all workers to stop
func (vwp *ViewWorkerPool) Shutdown() {
	log.Println("Shutting down view worker pool...")
	close(vwp.done)
	// Give workers time to finish current jobs
	time.Sleep(1 * time.Second)
	close(vwp.JobQueue)
}

// GetQueueLength returns the current number of jobs waiting in the queue
// Useful for monitoring worker pool health
func (vwp *ViewWorkerPool) GetQueueLength() int {
	return len(vwp.JobQueue)
}
