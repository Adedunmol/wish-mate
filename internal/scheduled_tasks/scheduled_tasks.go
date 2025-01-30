package scheduled_tasks

import (
	"encoding/json"
	"time"
)

type ScheduledTaskResponse struct {
	ID        int             `json:"id"`
	TaskName  string          `json:"task_name"`
	Payload   json.RawMessage `json:"payload"`
	Status    string          `json:"status"`
	ExecuteAt time.Time       `json:"execute_at"`
}
