package rest

import (
	"github.com/gofiber/fiber/v2"
)

// RenderPage Render front-end
func RenderPage(webPage string) func(*fiber.Ctx) error {
	fn := func(ctx *fiber.Ctx) error {
		ctx.Set("Content-Type", "text/html")
		return ctx.SendString(webPage)
	}
	return fn
}
