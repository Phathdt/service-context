package jobs

const (
	// Job type constants
	TypeTodoCreated = "todo:created"
	TypeTodoUpdated = "todo:updated"
	TypeTodoDeleted = "todo:deleted"
)

// TodoCreatedPayload represents the payload for todo created job
type TodoCreatedPayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

// TodoUpdatedPayload represents the payload for todo updated job
type TodoUpdatedPayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	UpdatedAt   string `json:"updated_at"`
}

// TodoDeletedPayload represents the payload for todo deleted job
type TodoDeletedPayload struct {
	ID        string `json:"id"`
	DeletedAt string `json:"deleted_at"`
}