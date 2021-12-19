package main

import (
	"authapi/database"
	"authapi/routes"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware/cors"
	"os"
)

func main() {
	database.Connect()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	routes.Setup(app)

	port := os.Getenv("port")

	if port == "" {
		port = ":3000"
	}

	app.Listen(port)
}
