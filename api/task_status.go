package api

// TaskStatus represents the stage of execution for a task. The associated enumeration is ordered,
// where higher value statuses are closer to complete. It's possible for a task to transition from a
// higher status to a lower one if rescheduled. For example, a "running" experiment can move to
// "starting" if the node it's running on is terminated.
type TaskStatus string

const (
	// TaskStatusCreated means the task is accepted by Beaker.
	// The task will automatically start when eligible.
	TaskStatusCreated TaskStatus = "created"

	// TaskStatusStarting means the task is attempting to start, but is not yet running.
	TaskStatusStarting TaskStatus = "starting"

	// TaskStatusRunning means the task is executing.
	TaskStatusRunning TaskStatus = "running"

	// TaskStatusSuccessful means the task has completed successfully.
	TaskStatusSuccessful TaskStatus = "successful"

	// TaskStatusCanceled means the task was aborted.
	TaskStatusCanceled TaskStatus = "canceled"

	// TaskStatusFailed means the task could not be completed.
	TaskStatusFailed TaskStatus = "failed"
)

// IsEndState is true if the TaskStatus is canceled, failed, or successful
func (ts TaskStatus) IsEndState() bool {
	switch ts {
	case TaskStatusSuccessful, TaskStatusCanceled, TaskStatusFailed:
		return true
	default:
		return false
	}
}
