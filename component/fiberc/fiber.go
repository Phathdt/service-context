package fiberc

import (
	"flag"

	sctx "github.com/phathdt/service-context"
	"github.com/phathdt/service-context/component/fiberc/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flogger "github.com/gofiber/fiber/v2/middleware/logger"
)

const (
	defaultPort = 4000
)

type FiberComponent interface {
	GetPort() int
	GetApp() *fiber.App
}

type fiberEngine struct {
	id     string
	port   int
	logger sctx.Logger
	app    *fiber.App
}

func (e *fiberEngine) GetPort() int {
	return e.port
}

func (e *fiberEngine) GetApp() *fiber.App {
	return e.app
}

func (e *fiberEngine) ID() string {
	return e.id
}

func (e *fiberEngine) InitFlags() {
	flag.IntVar(&e.port, "fiber-port", defaultPort, "fiber server port. Default 4000")
}

func (e *fiberEngine) Activate(sv sctx.ServiceContext) error {
	e.logger = sv.Logger(e.id)

	e.logger.Info("init engine...")
	app := fiber.New(fiber.Config{BodyLimit: 100 * 1024 * 1024})

	app.Use(flogger.New(flogger.Config{
		Format: `{"ip":${ip}, "timestamp":"${time}", "status":${status}, "latency":"${latency}", "method":"${method}", "path":"${path}"}` + "\n",
	}))
	app.Use(compress.New())
	app.Use(cors.New())
	app.Use(middleware.Recover(sv))

	app.Get("/", ping())
	e.app = app

	return nil
}

func (e *fiberEngine) Stop() error {
	return e.app.Shutdown()
}

func New(id string) *fiberEngine {
	return &fiberEngine{id: id}
}

func ping() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(&fiber.Map{
			"msg": "pong",
		})
	}
}
