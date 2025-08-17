package handler

import (
	"fiberapp/internal/service"
	"github.com/gofiber/fiber/v2"
)

type TodoHandler struct {
	todoService *service.TodoService
}

func NewTodoHandler(todoService *service.TodoService) *TodoHandler {
	return &TodoHandler{
		todoService: todoService,
	}
}

func (h *TodoHandler) GetTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID parameter is required",
		})
	}

	todo, err := h.todoService.GetTodo(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	return c.JSON(todo)
}

func (h *TodoHandler) ListTodos(c *fiber.Ctx) error {
	todos, err := h.todoService.ListTodos(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch todos",
		})
	}

	return c.JSON(fiber.Map{
		"data": todos,
		"count": len(todos),
	})
}

func (h *TodoHandler) CreateTodo(c *fiber.Ctx) error {
	var req service.CreateTodoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	todo, err := h.todoService.CreateTodo(c.Context(), req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create todo",
		})
	}

	return c.Status(201).JSON(todo)
}

func (h *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID parameter is required",
		})
	}

	var req service.UpdateTodoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Title is required",
		})
	}

	todo, err := h.todoService.UpdateTodo(c.Context(), id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update todo",
		})
	}

	return c.JSON(todo)
}

func (h *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID parameter is required",
		})
	}

	err := h.todoService.DeleteTodo(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete todo",
		})
	}

	return c.Status(204).Send(nil)
}

func (h *TodoHandler) ToggleComplete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID parameter is required",
		})
	}

	todo, err := h.todoService.ToggleComplete(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to toggle todo completion",
		})
	}

	return c.JSON(todo)
}