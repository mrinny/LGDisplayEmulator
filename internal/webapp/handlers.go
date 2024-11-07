package webapp

import "github.com/gofiber/fiber/v2"

type Handlers struct{}

func NewHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) HandleGetIndex(ctx *fiber.Ctx) error {
	context := fiber.Map{}
	return ctx.Render("index", context)
}
