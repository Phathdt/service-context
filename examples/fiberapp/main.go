package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"fiberapp/internal/components/asynqc"
	"fiberapp/internal/components/asynqw"
	"fiberapp/internal/config"
	"fiberapp/internal/db"
	"fiberapp/internal/handler"
	"fiberapp/internal/jobs"
	"fiberapp/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	sctx "github.com/phathdt/service-context"
	"github.com/phathdt/service-context/component/pgxc"
	"github.com/phathdt/service-context/component/redisc"
	"github.com/redis/go-redis/v9"
	slogfiber "github.com/samber/slog-fiber"
	"github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

// Remove AppConfig wrapper - we only need the config itself

func main() {
	app := &cli.App{
		Name:  "fiberapp",
		Usage: "A Fiber web application with fx dependency injection and config management",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yml",
				Usage:   "path to configuration file",
			},
		},
		Action: func(cCtx *cli.Context) error {
			configPath := cCtx.String("config")

			// Load configuration from YAML + env overrides
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				return err
			}

			fxApp := fx.New(
				fx.Provide(
					func() *config.Config { return cfg },
					NewServiceContextAndLoad,
					NewFiberApp,
					NewPostgresConnection,
					NewRedisConnection,
					NewAsynqClient,
					NewAsynqWorker,
					NewJobHandlers,
					NewQueries,
					service.NewTodoService,
					handler.NewTodoHandler,
				),
				fx.Invoke(
					NewRouter,
					RegisterJobHandlers,
					func(lc fx.Lifecycle, sc sctx.ServiceContext) {
						lc.Append(fx.Hook{
							OnStop: func(ctx context.Context) error {
								logger := sctx.GlobalLogger().GetLogger("service")
								_ = sc.Stop()
								logger.Info("Server exited")
								return nil
							},
						})
					}),
			)

			fxApp.Run()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func NewServiceContextAndLoad(cfg *config.Config) sctx.ServiceContext {

	// Create components with configuration passed directly in constructors
	pgxComp := pgxc.New("postgres", "postgres", cfg.Database.GetDSN())
	redisComp := redisc.New("redis", cfg.Redis.GetURI())
	asynqClientComp := asynqc.New("asynq-client", cfg.Redis.GetURI())
	asynqWorkerComp := asynqw.New("asynq-worker", cfg.Redis.GetURI())

	// Create service context WITHOUT fiber component first
	sc := sctx.NewServiceContext(
		sctx.WithName("fiberapp"),
		sctx.WithComponent(pgxComp),
		sctx.WithComponent(redisComp),
		sctx.WithComponent(asynqClientComp),
		sctx.WithComponent(asynqWorkerComp),
	)

	// Load all components (postgres, redis, asynq client & worker)
	if err := sc.Load(); err != nil {
		panic(err)
	}

	return sc
}


func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 100 * 1024 * 1024})
	logger := sctx.GlobalLogger().GetLogger("fiber").GetSLogger()

	app.Use(slogfiber.New(logger))
	app.Use(compress.New())
	app.Use(cors.New())

	return app
}

func NewPostgresConnection(sc sctx.ServiceContext) *pgxpool.Pool {
	// Components should already be activated by sc.Load()
	return sc.MustGet("postgres").(pgxc.PgxComp).GetConn()
}

func NewRedisConnection(sc sctx.ServiceContext) *redis.Client {
	// Components should already be activated by sc.Load()
	return sc.MustGet("redis").(redisc.RedisComponent).GetClient()
}

func NewQueries(pool *pgxpool.Pool) *db.Queries {
	return db.New(pool)
}

func NewAsynqClient(sc sctx.ServiceContext) *asynq.Client {
	// Components already activated by sc.Load()
	return sc.MustGet("asynq-client").(asynqc.AsynqClientComponent).GetClient()
}

func NewAsynqWorker(sc sctx.ServiceContext) asynqw.AsynqWorkerComponent {
	// Components already activated by sc.Load()
	return sc.MustGet("asynq-worker").(asynqw.AsynqWorkerComponent)
}

func NewJobHandlers() *jobs.JobHandlers {
	return jobs.NewJobHandlers()
}

func RegisterJobHandlers(worker asynqw.AsynqWorkerComponent, handlers *jobs.JobHandlers) {
	worker.RegisterHandler(jobs.TypeTodoCreated, handlers.HandleTodoCreated)
	worker.RegisterHandler(jobs.TypeTodoUpdated, handlers.HandleTodoUpdated)
	worker.RegisterHandler(jobs.TypeTodoDeleted, handlers.HandleTodoDeleted)
}

func NewRouter(app *fiber.App, sc sctx.ServiceContext, cfg *config.Config, todoHandler *handler.TodoHandler) {
	app.Get("/", ping())
	app.Get("/health", health())

	api := app.Group("/api/v1")

	todos := api.Group("/todos")
	todos.Get("/", todoHandler.ListTodos)
	todos.Post("/", todoHandler.CreateTodo)
	todos.Get("/:id", todoHandler.GetTodo)
	todos.Put("/:id", todoHandler.UpdateTodo)
	todos.Delete("/:id", todoHandler.DeleteTodo)
	todos.Patch("/:id/toggle", todoHandler.ToggleComplete)

	// Start the Fiber server directly
	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
			panic(err)
		}
	}()
}

func ping() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(&fiber.Map{
			"msg": "pong",
		})
	}
}

func health() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(&fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	}
}
