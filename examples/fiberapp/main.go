package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	sctx "github.com/phathdt/service-context"
	"github.com/phathdt/service-context/component/fiberc"
)

func main() {
	sc := sctx.NewServiceContext(
		sctx.WithName("demo"),
		sctx.WithComponent(fiberc.New("fiber")),
	)

	serviceLogger := sctx.GlobalLogger().GetLogger("service")

	time.Sleep(time.Second * 1)

	if err := sc.Load(); err != nil {
		serviceLogger.Fatal(err)
	}

	fiberComp := sc.MustGet("fiber").(fiberc.FiberComponent)

	app := fiber.New()

	app.Get("/", ping())

	if err := app.Listen(fmt.Sprintf(":%d", fiberComp.GetPort())); err != nil {
		serviceLogger.Fatal(err)
	}
}

func ping() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(200).JSON(&fiber.Map{
			"msg": "pong",
		})
	}
}
