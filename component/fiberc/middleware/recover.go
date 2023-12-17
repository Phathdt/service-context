package middleware

import (
	"net/http"

	sctx "github.com/phathdt/service-context"

	"github.com/gofiber/fiber/v2"
)

type CanGetStatusCode interface {
	StatusCode() int
}

func Recover(serviceCtx sctx.ServiceContext) fiber.Handler {
	// Return new handler
	return func(ctx *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				if appErr, ok := err.(CanGetStatusCode); ok {
					ctx.Status(appErr.StatusCode()).JSON(appErr)

				} else {
					ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{
						"code":    http.StatusInternalServerError,
						"status":  "internal server error",
						"message": "something went wrong, please try again or contact supporters",
					})
				}

				serviceCtx.Logger("service").Errorf("%+v \n", err)
			}
		}()

		// Return err if existed, else move to next handler
		return ctx.Next()
	}
}
