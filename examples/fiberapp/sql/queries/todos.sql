-- name: GetTodo :one
SELECT * FROM todos WHERE id = $1 LIMIT 1;

-- name: ListTodos :many
SELECT * FROM todos
ORDER BY created_at DESC;

-- name: ListTodosByStatus :many
SELECT * FROM todos
WHERE completed = $1
ORDER BY created_at DESC;

-- name: CreateTodo :one
INSERT INTO todos (
  title, description
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateTodo :one
UPDATE todos
SET title = $2,
    description = $3,
    completed = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1;

-- name: ToggleTodoComplete :one
UPDATE todos
SET completed = NOT completed,
    updated_at = NOW()
WHERE id = $1
RETURNING *;