package worker

import (
	"encoding/json"
	"time"
)

// Args are the arguments passed into a job.
type Args map[string]any

func (a Args) String() string {
	b, _ := json.Marshal(a) // nolint: errchkjson
	return string(b)
}

// Job describes the job to be processed by a Worker.
type Job struct {
	// ID is the identifier of the task.
	ID string

	// Handler that will be run by the worker
	Handler string
	// Queue the job should be placed into
	Queue string
	// Args that will be passed to the Handler when run
	Args Args
}

func (j Job) String() string {
	b, _ := json.Marshal(j) // nolint: errchkjson
	return string(b)
}

// JobInfo describes the task and it's metadata
type JobInfo struct {
	// ID is the identifier of the task.
	ID string

	// Retries is the number of times the task has retried so far.
	Retries int

	// LastFailedAt is the time of the last failure if any.
	// If the task has no failures, LastFailedAt is zero time (i.e. time.Time{}).
	LastFailedAt time.Time
}
