package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	sctx "github.com/phathdt/service-context"
)

type JobHandlers struct {
	logger sctx.Logger
}

func NewJobHandlers() *JobHandlers {
	return &JobHandlers{
		logger: sctx.GlobalLogger().GetLogger("jobs"),
	}
}

// HandleTodoCreated handles todo created jobs
func (h *JobHandlers) HandleTodoCreated(ctx context.Context, t *asynq.Task) error {
	var payload TodoCreatedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error("Failed to unmarshal TodoCreatedPayload", err.Error())
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	h.logger.Info(fmt.Sprintf("Processing TODO CREATED job - ID: %s, Title: %s, Description: %s, CreatedAt: %s",
		payload.ID, payload.Title, payload.Description, payload.CreatedAt))

	// Simulate some work
	// In a real application, you might:
	// - Send notifications
	// - Update analytics
	// - Sync with external services
	// - Generate reports
	
	h.logger.Info("TODO CREATED job processed successfully", "id", payload.ID)
	return nil
}

// HandleTodoUpdated handles todo updated jobs
func (h *JobHandlers) HandleTodoUpdated(ctx context.Context, t *asynq.Task) error {
	var payload TodoUpdatedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error("Failed to unmarshal TodoUpdatedPayload", err.Error())
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	h.logger.Info(fmt.Sprintf("Processing TODO UPDATED job - ID: %s, Title: %s, Completed: %t, UpdatedAt: %s",
		payload.ID, payload.Title, payload.Completed, payload.UpdatedAt))

	h.logger.Info("TODO UPDATED job processed successfully", "id", payload.ID)
	return nil
}

// HandleTodoDeleted handles todo deleted jobs
func (h *JobHandlers) HandleTodoDeleted(ctx context.Context, t *asynq.Task) error {
	var payload TodoDeletedPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error("Failed to unmarshal TodoDeletedPayload", err.Error())
		return fmt.Errorf("json.Unmarshal failed: %v", err)
	}

	h.logger.Info(fmt.Sprintf("Processing TODO DELETED job - ID: %s, DeletedAt: %s",
		payload.ID, payload.DeletedAt))

	h.logger.Info("TODO DELETED job processed successfully", "id", payload.ID)
	return nil
}