package api

import (
	"github.com/fako1024/btscale/pkg/scale"
	"github.com/gofiber/fiber/v2"
)

// API denotes a REST API for a scale
type API struct {
	scale  scale.Scale
	router *fiber.App
}

// New instantiates a new API
func New(s scale.Scale, endpoint string) *API {

	api := API{
		scale:  s,
		router: fiber.New(),
	}

	// Setup routes
	api.router.Post("/toggle_buzzer", api.handleToggleBuzzer())

	// Start to listen in goroutine
	go func() {
		if err := api.router.Listen(endpoint); err != nil {
			panic(err)
		}
	}()

	return &api
}

func (api *API) handleToggleBuzzer() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return api.scale.ToggleBuzzingOnTouch()
	}
}
