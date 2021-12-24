package main

import (
	"authapi/database"
	"authapi/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"os"
)

func main() {
	database.Connect()

	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format: "[${ip}:${port}] ${method} ${path} ${status} ${latency} ${res_length}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	routes.Setup(app)

	port := os.Getenv("port")

	if port == "" {
		port = ":3000"
	}

	err := app.Listen(port)
	if err != nil {
		panic("Port in use")
	}
}
