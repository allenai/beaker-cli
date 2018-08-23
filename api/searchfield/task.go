package searchfield

type Task string

const (
	TaskID           Task = "id"
	TaskDescription  Task = "description"
	TaskStartTime    Task = "start_time"
	TaskEndTime      Task = "end_time"
	TaskCreatingUser Task = "user"
)
