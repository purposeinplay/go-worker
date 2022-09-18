package worker

import "encoding/json"

// Args are the arguments passed into a job.
type Args map[string]any

func (a Args) String() string {
	b, _ := json.Marshal(a) // nolint: errchkjson
	return string(b)
}

// Job to be processed by a Worker.
type Job struct {
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
