package caching

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Persister manages the background persistence worker
type Persister struct {
	persistenceService *PersistenceService
	ticker             *time.Ticker
	stopChan           chan bool
	running            bool
}

// NewPersister creates a new persister instance
func NewPersister() *Persister {
	return &Persister{
		persistenceService: NewPersistenceService(),
		stopChan:           make(chan bool),
		running:            false,
	}
}

// Start starts the background persistence worker
// It runs on a schedule (hourly or daily, configurable)
func (p *Persister) Start() error {
	if p.running {
		log.Println("[PERSISTER] Persister is already running")
		return nil
	}

	// Get persistence schedule from environment
	schedule := p.getPersistenceSchedule()
	log.Printf("[PERSISTER] Starting persistence worker with schedule: %v", schedule)

	p.ticker = time.NewTicker(schedule)
	p.running = true

	// Run persistence immediately on start (optional)
	// Uncomment if you want immediate persistence on startup
	// go p.runPersistence()

	// Start the worker goroutine
	go p.worker()

	log.Println("[PERSISTER] Persistence worker started")
	return nil
}

// Stop stops the background persistence worker
func (p *Persister) Stop() {
	if !p.running {
		return
	}

	log.Println("[PERSISTER] Stopping persistence worker...")
	p.running = false

	if p.ticker != nil {
		p.ticker.Stop()
	}

	p.stopChan <- true
	log.Println("[PERSISTER] Persistence worker stopped")
}

// worker runs the persistence loop
func (p *Persister) worker() {
	for {
		select {
		case <-p.ticker.C:
			p.runPersistence()
		case <-p.stopChan:
			return
		}
	}
}

// runPersistence executes the persistence operation with error handling and retry logic
func (p *Persister) runPersistence() {
	log.Println("[PERSISTER] Running scheduled persistence...")
	startTime := time.Now()

	// Run persistence with retry logic
	maxRetries := 3
	retryDelay := 5 * time.Second

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = p.persistenceService.PersistAll()
		if err == nil {
			duration := time.Since(startTime)
			log.Printf("[PERSISTER] Persistence completed successfully in %v", duration)
			return
		}

		if attempt < maxRetries {
			log.Printf("[PERSISTER] Persistence attempt %d failed, retrying in %v: %v", attempt, retryDelay, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	duration := time.Since(startTime)
	log.Printf("[PERSISTER] Persistence failed after %d attempts in %v: %v", maxRetries, duration, err)
}

// getPersistenceSchedule gets the persistence schedule from environment variable
// Defaults to 1 hour if not set
func (p *Persister) getPersistenceSchedule() time.Duration {
	scheduleStr := os.Getenv("CACHE_PERSISTENCE_SCHEDULE")
	if scheduleStr == "" {
		scheduleStr = "1h" // Default: hourly
	}

	// Parse duration
	duration, err := time.ParseDuration(scheduleStr)
	if err != nil {
		log.Printf("[PERSISTER] Warning: Invalid CACHE_PERSISTENCE_SCHEDULE '%s', using default 1h", scheduleStr)
		return 1 * time.Hour
	}

	// Validate minimum interval (at least 1 minute)
	if duration < 1*time.Minute {
		log.Printf("[PERSISTER] Warning: CACHE_PERSISTENCE_SCHEDULE too short (%v), using minimum 1m", duration)
		return 1 * time.Minute
	}

	return duration
}

// GetNextRunTime returns when the next persistence will run
func (p *Persister) GetNextRunTime() time.Time {
	if p.ticker == nil {
		return time.Time{}
	}

	// Calculate next run time based on schedule
	schedule := p.getPersistenceSchedule()
	return time.Now().Add(schedule)
}

// IsRunning returns whether the persister is currently running
func (p *Persister) IsRunning() bool {
	return p.running
}

// TriggerManualPersistence manually triggers persistence (for admin endpoints)
func (p *Persister) TriggerManualPersistence() error {
	log.Println("[PERSISTER] Manual persistence triggered")
	return p.persistenceService.PersistAll()
}

// Helper function to parse duration from string with support for "daily"
func parsePersistenceSchedule(scheduleStr string) (time.Duration, error) {
	if scheduleStr == "daily" || scheduleStr == "24h" {
		return 24 * time.Hour, nil
	}

	// Try parsing as duration string
	duration, err := time.ParseDuration(scheduleStr)
	if err != nil {
		// Try parsing as hours (e.g., "1" = 1 hour)
		if hours, parseErr := strconv.Atoi(scheduleStr); parseErr == nil {
			return time.Duration(hours) * time.Hour, nil
		}
		return 0, err
	}

	return duration, nil
}

