package api

// TaskStatus represents the stage of execution for a task. The associated enumeration is ordered,
// where higher value statuses are closer to complete. It's possible for a task to transition from a
// higher status to a lower one if rescheduled. For example, a "running" experiment can move to
// "initializing" if the node it's running on is terminated.
type TaskStatus string

// TODO: Remove DEPRECATED values
const (
	// TaskStatusCreated means the task is accepted by Beaker.
	// DEPRECATED.
	// The task will automatically start when eligible.
	TaskStatusCreated TaskStatus = "created"

	// TaskStatusStarting means the task is attempting to start, but is not yet running.
	// DEPRECATED.
	TaskStatusStarting TaskStatus = "starting"

	// TaskStatusSuccessful means the task has completed successfully.
	// DEPRECATED.
	TaskStatusSuccessful TaskStatus = "successful"

	// TaskStatusCanceled means the task was aborted.
	// DEPRECATED.
	TaskStatusCanceled TaskStatus = "canceled"

	// TaskStatusSubmitted means a task is accepted by Beaker.
	// The task will automatically start when eligible.
	TaskStatusSubmitted TaskStatus = "submitted"

	// TaskStatusProvisioning means a task has been submitted for execution and
	// Beaker is waiting for compute resources to become available.
	TaskStatusProvisioning TaskStatus = "provisioning"

	// TaskStatusInitializing means a task is attempting to start, but is not yet running.
	TaskStatusInitializing TaskStatus = "initializing"

	// TaskStatusRunning means a task is executing.
	TaskStatusRunning TaskStatus = "running"

	// TaskStatusSucceeded means a task has completed successfully.
	TaskStatusSucceeded TaskStatus = "succeeded"

	// TaskStatusSkipped means a task will never run due to failed or invalid prerequisites.
	TaskStatusSkipped TaskStatus = "skipped"

	// TaskStatusStopped means a task was interrupted by a user.
	TaskStatusStopped TaskStatus = "stopped"

	// TaskStatusFailed means a task could not be completed.
	TaskStatusFailed TaskStatus = "failed"
)

// IsEndState is true if the TaskStatus is canceled, failed, or successful
func (ts TaskStatus) IsEndState() bool {
	switch ts {
	// TODO: Delete this when the deprecated states go away.
	case TaskStatusSuccessful, TaskStatusCanceled:
		fallthrough
	// New end states
	case TaskStatusSucceeded, TaskStatusSkipped, TaskStatusStopped, TaskStatusFailed:
		return true
	default:
		return false
	}
}
