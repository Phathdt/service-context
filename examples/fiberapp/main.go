package main

import (
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	sctx "github.com/phathdt/service-context"
	"github.com/phathdt/service-context/component/fiberc"

	flogger "github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	sc := sctx.NewServiceContext(
		sctx.WithName("demo"),
		sctx.WithComponent(fiberc.New("fiber")),
	)

	logger := sctx.GlobalLogger().GetLogger("service")

	time.Sleep(time.Second * 1)

	NewRouter(sc)

	if err := sc.Load(); err != nil {
		logger.Fatal(err)
	}

	// gracefully shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	_ = sc.Stop()
	logger.Info("Server exited")
}

func ping() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(&fiber.Map{
			"msg": "pong",
		})
	}
}

func NewRouter(sc sctx.ServiceContext) {
	app := fiber.New(fiber.Config{BodyLimit: 100 * 1024 * 1024})
	app.Use(flogger.New(flogger.Config{
		Format: `{"ip":${ip}, "timestamp":"${time}", "status":${status}, "latency":"${latency}", "method":"${method}", "path":"${path}"}` + "\n",
	}))
	app.Use(compress.New())
	app.Use(cors.New())

	app.Get("/", ping())

	fiberComp := sc.MustGet("fiber").(fiberc.FiberComponent)
	fiberComp.SetApp(app)
}
