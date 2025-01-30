package scheduled_tasks

import (
	"encoding/json"
	"time"
)

type Store interface {
	CreateTask(name string, payload json.RawMessage, executeAt *time.Time) (ScheduledTaskResponse, error)
	GetTasks(currentTime *time.Time) ([]ScheduledTaskResponse, error)
}
