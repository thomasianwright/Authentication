package routes

import (
	"authapi/controllers"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"os"
)

func Setup(app *fiber.App) {
	// Unauthenticated routes
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Post("/api/getuserbytoken", controllers.GetUserByToken)
	app.Post("/api/authenticatetoken", controllers.AuthenticateToken)
	app.Post("/api/activateaccount", controllers.ActivateAccount)

	// Authenticated routes middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("jwtsecret")),
	}))

	// Authenticated Routes
	app.Post("/api/updatepassword", controllers.UpdatePassword)
	app.Post("/api/updateuser", controllers.UpdateUser)

	app.Get("/api/user", controllers.GetUser)

}
