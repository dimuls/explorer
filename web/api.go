package main

import (
	"explorer/pg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type APIConfig struct {
	BindAddr string `yaml:"bind_addr"`
	TLSCert  string `yaml:"tls_cert"`
	TLSKey   string `yaml:"tls_key"`
}

type API struct {
	ws *fiber.App
}

func NewAPI(c *APIConfig, s *pg.Explorer) {
	ws := fiber.New()

	ws.Use(recover.New(), logger.New())

	ws.Get("/peers", func(ctx *fiber.Ctx) error {
		offsetStr := ctx.Query("offset", "100")
		lastBlockIDStr := ctx.Query("last_block_id", "")

	})
}
