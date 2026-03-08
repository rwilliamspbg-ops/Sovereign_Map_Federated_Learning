package scheduler

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// JobPriority controls dequeue order (higher first).
type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
)

// Job describes a training task scheduled for execution.
type Job struct {
	ID        string
	NodeID    string
	RoundID   int
	Payload   map[string]interface{}
	Priority  JobPriority
	CreatedAt time.Time
}

// JobQueue stores pending training tasks with priority ordering.
type JobQueue struct {
	mu      sync.Mutex
	maxSize int
	jobs    []Job
	seen    map[string]struct{}
}

// NewJobQueue creates a queue with optional max size (0 means unlimited).
func NewJobQueue(maxSize int) *JobQueue {
	return &JobQueue{maxSize: maxSize, seen: make(map[string]struct{})}
}

// Enqueue adds a new job to the queue.
func (q *JobQueue) Enqueue(job Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if job.ID == "" {
		return fmt.Errorf("job id is required")
	}
	if _, exists := q.seen[job.ID]; exists {
		return fmt.Errorf("job %s already queued", job.ID)
	}
	if q.maxSize > 0 && len(q.jobs) >= q.maxSize {
		return fmt.Errorf("job queue is full")
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}

	q.jobs = append(q.jobs, job)
	q.seen[job.ID] = struct{}{}

	sort.SliceStable(q.jobs, func(i, j int) bool {
		if q.jobs[i].Priority != q.jobs[j].Priority {
			return q.jobs[i].Priority > q.jobs[j].Priority
		}
		return q.jobs[i].CreatedAt.Before(q.jobs[j].CreatedAt)
	})

	return nil
}

// Dequeue removes and returns the highest-priority job.
func (q *JobQueue) Dequeue() (Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return Job{}, false
	}
	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	delete(q.seen, job.ID)
	return job, true
}

// Len returns number of pending jobs.
func (q *JobQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}

// PendingIDs returns a snapshot of queued job IDs in dequeue order.
func (q *JobQueue) PendingIDs() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]string, 0, len(q.jobs))
	for _, j := range q.jobs {
		out = append(out, j.ID)
	}
	return out
}
