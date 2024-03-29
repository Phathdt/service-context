package fiberc

import (
	"flag"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	sctx "github.com/phathdt/service-context"
)

const (
	defaultPort = 4000
)

type FiberComponent interface {
	GetPort() int
	SetApp(app *fiber.App)
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

func (e *fiberEngine) SetApp(app *fiber.App) {
	e.app = app
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

	errChan := make(chan error, 1)

	go func() {
		if err := e.app.Listen(fmt.Sprintf(":%d", e.GetPort())); err != nil {
			errChan <- err
		}
	}()

	time.Sleep(1 * time.Second)

	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}

func (e *fiberEngine) Stop() error {
	return e.app.Shutdown()
}

func New(id string) *fiberEngine {
	return &fiberEngine{id: id}
}
