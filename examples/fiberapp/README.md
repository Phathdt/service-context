# Fiber App with FX and Service Context Integration

This example demonstrates the integration between `service-context` and `fx` dependency injection framework, featuring a complete todo list API with PostgreSQL database and Redis caching.

## Features

- **fx Dependency Injection**: Clean dependency management with uber-go/fx
- **CLI Interface**: Command-line interface using urfave/cli/v2
- **Database Integration**: PostgreSQL with SQLC for type-safe queries
- **Redis Caching**: Automatic caching layer for improved performance
- **Service Context**: Integration with phathdt/service-context framework
- **RESTful API**: Complete CRUD operations for todo management

## Prerequisites

- Go 1.24+
- Docker and Docker Compose (for PostgreSQL and Redis)
- SQLC installed globally

## Quick Start

1. **Start dependencies**:
```bash
docker-compose up -d
```

2. **Build the application**:
```bash
go build -o fiberapp
```

3. **Run the application**:
```bash
./fiberapp --postgres-db-dsn "postgres://postgres:postgres@localhost:5432/todoapp?sslmode=disable" --redis-uri "redis://localhost:6379"
```

## API Endpoints

### Health Check
- `GET /` - Ping endpoint
- `GET /health` - Health check

### Todo Management
- `GET /api/v1/todos` - List all todos
- `POST /api/v1/todos` - Create a new todo
- `GET /api/v1/todos/:id` - Get a specific todo
- `PUT /api/v1/todos/:id` - Update a todo
- `DELETE /api/v1/todos/:id` - Delete a todo
- `PATCH /api/v1/todos/:id/toggle` - Toggle todo completion status

### Example Requests

**Create a todo**:
```bash
curl -X POST http://localhost:4000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "description": "Master Go programming language"}'
```

**List todos**:
```bash
curl http://localhost:4000/api/v1/todos
```

**Update a todo**:
```bash
curl -X PUT http://localhost:4000/api/v1/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go Advanced", "description": "Master advanced Go concepts", "completed": true}'
```

## Architecture

### Dependency Injection with FX

The application uses fx to manage dependencies and lifecycle:

```go
fx.New(
    fx.Provide(
        NewServiceContext,
        NewFiberApp,
        NewPostgresConnection,
        NewRedisConnection,
        NewQueries,
        service.NewTodoService,
        handler.NewTodoHandler,
    ),
    fx.Invoke(NewRouter),
)
```

### Service Context Integration

Service context manages:
- Fiber web server component
- PostgreSQL database connection
- Redis cache connection

### Caching Strategy

Redis is used for:
- Individual todo caching (5 minutes TTL)
- Todo list caching (2 minutes TTL)
- Automatic cache invalidation on mutations

## CLI Options

```
--name value, -n value     application name (default: "demo")
--port value, -p value     server port (default: "8080")  
--db-host value           database host (default: "localhost")
--db-port value           database port (default: "5432")
--db-user value           database user (default: "postgres")
--db-password value       database password (default: "postgres")
--db-name value           database name (default: "todoapp")
--redis-host value        redis host (default: "localhost")
--redis-port value        redis port (default: "6379")
```

## Database Schema

The application uses a simple todos table:

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## Key Integration Points

1. **Service Context → FX**: Service context components are provided to fx
2. **FX → Services**: Database and Redis clients are injected into services  
3. **Services → Handlers**: Business logic is injected into HTTP handlers
4. **Lifecycle Management**: FX manages startup/shutdown coordination

This example showcases how service-context and fx can work together to create a clean, maintainable application architecture.