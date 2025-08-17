package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"fiberapp/internal/db"
	"fiberapp/internal/jobs"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type TodoService struct {
	queries     *db.Queries
	redis       *redis.Client
	asynqClient *asynq.Client
}

func NewTodoService(queries *db.Queries, redis *redis.Client, asynqClient *asynq.Client) *TodoService {
	return &TodoService{
		queries:     queries,
		redis:       redis,
		asynqClient: asynqClient,
	}
}

type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type TodoResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *TodoService) GetTodo(ctx context.Context, id string) (*TodoResponse, error) {
	cacheKey := fmt.Sprintf("todo:%s", id)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var todo TodoResponse
		if err := json.Unmarshal([]byte(cached), &todo); err == nil {
			return &todo, nil
		}
	}

	todoID := pgtype.UUID{}
	if err := todoID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	todo, err := s.queries.GetTodo(ctx, todoID)
	if err != nil {
		return nil, err
	}

	response := s.mapToResponse(todo)

	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return response, nil
}

func (s *TodoService) ListTodos(ctx context.Context) ([]*TodoResponse, error) {
	cacheKey := "todos:all"

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var todos []*TodoResponse
		if err := json.Unmarshal([]byte(cached), &todos); err == nil {
			return todos, nil
		}
	}

	todos, err := s.queries.ListTodos(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = s.mapToResponse(todo)
	}

	if data, err := json.Marshal(responses); err == nil {
		s.redis.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return responses, nil
}

func (s *TodoService) CreateTodo(ctx context.Context, req CreateTodoRequest) (*TodoResponse, error) {
	description := pgtype.Text{}
	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	todo, err := s.queries.CreateTodo(ctx, db.CreateTodoParams{
		Title:       req.Title,
		Description: description,
	})
	if err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)

	response := s.mapToResponse(todo)

	cacheKey := fmt.Sprintf("todo:%s", response.ID)
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	// Enqueue background job for todo created
	jobPayload := jobs.TodoCreatedPayload{
		ID:          response.ID,
		Title:       response.Title,
		Description: response.Description,
		CreatedAt:   response.CreatedAt.Format(time.RFC3339),
	}
	if err := s.enqueueJob(jobs.TypeTodoCreated, jobPayload); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to enqueue todo created job: %v\n", err)
	}

	return response, nil
}

func (s *TodoService) UpdateTodo(ctx context.Context, id string, req UpdateTodoRequest) (*TodoResponse, error) {
	todoID := pgtype.UUID{}
	if err := todoID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	description := pgtype.Text{}
	if req.Description != "" {
		description.String = req.Description
		description.Valid = true
	}

	todo, err := s.queries.UpdateTodo(ctx, db.UpdateTodoParams{
		ID:          todoID,
		Title:       req.Title,
		Description: description,
		Completed:   req.Completed,
	})
	if err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)

	response := s.mapToResponse(todo)

	cacheKey := fmt.Sprintf("todo:%s", id)
	s.redis.Del(ctx, cacheKey)
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	// Enqueue background job for todo updated
	jobPayload := jobs.TodoUpdatedPayload{
		ID:          response.ID,
		Title:       response.Title,
		Description: response.Description,
		Completed:   response.Completed,
		UpdatedAt:   response.UpdatedAt.Format(time.RFC3339),
	}
	if err := s.enqueueJob(jobs.TypeTodoUpdated, jobPayload); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to enqueue todo updated job: %v\n", err)
	}

	return response, nil
}

func (s *TodoService) DeleteTodo(ctx context.Context, id string) error {
	todoID := pgtype.UUID{}
	if err := todoID.Scan(id); err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}

	err := s.queries.DeleteTodo(ctx, todoID)
	if err != nil {
		return err
	}

	s.invalidateCache(ctx)
	s.redis.Del(ctx, fmt.Sprintf("todo:%s", id))

	// Enqueue background job for todo deleted
	jobPayload := jobs.TodoDeletedPayload{
		ID:        id,
		DeletedAt: time.Now().Format(time.RFC3339),
	}
	if err := s.enqueueJob(jobs.TypeTodoDeleted, jobPayload); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to enqueue todo deleted job: %v\n", err)
	}

	return nil
}

func (s *TodoService) ToggleComplete(ctx context.Context, id string) (*TodoResponse, error) {
	todoID := pgtype.UUID{}
	if err := todoID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	todo, err := s.queries.ToggleTodoComplete(ctx, todoID)
	if err != nil {
		return nil, err
	}

	s.invalidateCache(ctx)

	response := s.mapToResponse(todo)

	cacheKey := fmt.Sprintf("todo:%s", id)
	s.redis.Del(ctx, cacheKey)
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return response, nil
}

func (s *TodoService) mapToResponse(todo db.Todo) *TodoResponse {
	var description string
	if todo.Description.Valid {
		description = todo.Description.String
	}

	id := uuid.UUID(todo.ID.Bytes)

	return &TodoResponse{
		ID:          id.String(),
		Title:       todo.Title,
		Description: description,
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt.Time,
		UpdatedAt:   todo.UpdatedAt.Time,
	}
}

func (s *TodoService) invalidateCache(ctx context.Context) {
	s.redis.Del(ctx, "todos:all")
}

// enqueueJob helper method to enqueue background jobs
func (s *TodoService) enqueueJob(jobType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}

	task := asynq.NewTask(jobType, payloadBytes)
	_, err = s.asynqClient.Enqueue(task)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}
