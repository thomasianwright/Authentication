package routes

import (
	"authapi/controllers"
	"github.com/gofiber/fiber"
)

func Setup(app *fiber.App) {
	app.Post("/api/register", controllers.Register)
	app.Post("/api/login", controllers.Login)
	app.Get("/api/user", controllers.User)
	app.Post("/api/getuserbytoken", controllers.GetUserByToken)
	app.Post("/api/authenticatetoken", controllers.AuthenticateToken)
	app.Post("/api/activateaccount", controllers.ActivateAccount)
}
