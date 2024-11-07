package webapp

import (
	"embed"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
)

//go:embed html
var staticDir embed.FS

type WebApp struct {
	app *fiber.App
	mu  sync.Mutex
}

func New() *WebApp {
	wa := WebApp{}
	ve := html.New("./views", ".html")
	wa.app = fiber.New(
		fiber.Config{
			Views: ve,
		},
	)
	wa.app.Use("/", filesystem.New())
	wa.app.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Welcome To LGDisplayEmulator")
	})

	return &wa
}

func (wa *WebApp) Start() error {
	wa.mu.Lock()
	defer wa.mu.Unlock()
	wa.app.Listen(":3000")
	return nil
}
