package worker

import (
	"time"
)

// Handler function that will be run by the worker and given
// a slice of arguments.
type Handler func(Job) error

// Worker describes how a worker should be implemented.
type Worker interface {
	// Start the worker
	Start() error

	// Stop the worker
	Stop() error

	// Perform a job as soon as possible
	Perform(job Job) (*JobInfo, error)

	// PerformAt performs a job at a particular time
	PerformAt(Job, time.Time) (*JobInfo, error)

	// PerformIn performs a job after waiting for a specified amount of time
	PerformIn(Job, time.Duration) (*JobInfo, error)

	// Register a Handler
	Register(string, Handler) error

	// DeleteJob removes a job from the queue.
	DeleteJob(queue, jobID string) error
}
