package controllers

import (
	"authapi/database"
	"authapi/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	uuid2 "github.com/nu7hatch/gouuid"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
	"time"
)

const (
	LoginSuccess      = 1
	LoginNotActivated = 2
	LoginNotFound     = 3
)

const (
	RegisterSuccess = 1
	RegisterExists  = 2
)

const (
	ActivateSuccess  = 1
	ActivateNotFound = 2
)

const (
	AuthorizeSuccess = 1
	AuthorizeFailed  = 2
)

func Register(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)

	user := models.User{
		Firstname: data["firstname"],
		Lastname:  data["lastname"],
		Username:  data["username"],
		Email:     data["email"],
		Password:  password,
		Activated: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := database.DB.Create(&user)
	if err != nil {
		return c.JSON(fiber.Map{
			"status":  RegisterExists,
			"message": "Username or email already exists",
		})
	}

	uuid, _ := uuid2.NewV4()

	activate := models.Activation{
		UserId: user.Id,
		Guid:   uuid.String(),
	}

	errs := database.DB.Create(&activate)

	if errs != nil {
		uuid, _ := uuid2.NewV4()

		database.DB.Create(&models.Activation{
			UserId: user.Id,
			Guid:   uuid.String(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  RegisterSuccess,
		"message": "Successfully registered",
	})
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	var user models.User

	database.DB.Where("email = ?", data["email"]).Or("username = ?", data["email"]).First(&user)

	if user.Id == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message":     "User not found",
			"loginStatus": LoginNotFound,
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message":     "incorrect password",
			"loginStatus": LoginNotFound,
		})
	}

	if !user.Activated {
		c.Status(fiber.StatusNotAcceptable)
		return c.JSON(fiber.Map{
			"message":     "Please check your email to activate your account!",
			"loginStatus": LoginNotActivated,
		})
	}

	user.LastLogin = time.Now()

	database.DB.Save(&user)

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Issuer:    strconv.Itoa(int(user.Id)),
		ExpiresAt: time.Now().Add(time.Hour * 3).Unix(),
		IssuedAt:  time.Now().Unix(),
	})

	token, tokenErr := claims.SignedString([]byte(os.Getenv("jwtsecret")))
	if tokenErr != nil {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message": "Server Error",
		})
	}

	tokenCookie := fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 3),
		HTTPOnly: true,
	}

	c.Cookie(&tokenCookie)

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message":     "success",
		"token":       token,
		"loginStatus": LoginSuccess,
	})
}

func User(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")

	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("jwtsecret")), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	database.DB.Where("id = ?", claims.Issuer).First(&user)

	return c.JSON(user)
}

func AuthenticateToken(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err := jwt.ParseWithClaims(data["token"], &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("jwtsecret")), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		return c.JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message": "authorized",
	})

}

func GetUserByToken(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err.Error())
		return err
	}

	token, err := jwt.ParseWithClaims(data["token"], &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("jwtsecret")), nil
	})

	if err != nil {
		c.Status(fiber.StatusUnauthorized)
		fmt.Println(err.Error())
		fmt.Println(data["token"])

		return c.JSON(fiber.Map{
			"message": "unauthorized",
		})
	}

	claims := token.Claims.(*jwt.StandardClaims)

	var user models.User

	database.DB.Where("id = ?", claims.Issuer).First(&user)

	return c.JSON(user)
}

func ActivateAccount(c *fiber.Ctx) error {
	var data map[string]string
	c.BodyParser(&data)
	var activation models.Activation

	database.DB.Where("guid = ?", data["guid"]).First(&activation)

	if activation.Id == 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(fiber.Map{
			"message": "Token not found",
		})
	}
	var user models.User

	database.DB.Where("id = ?", activation.UserId).First(&user)

	user.Activated = true

	database.DB.Save(&user)
	database.DB.Delete(&activation)

	c.Status(fiber.StatusAccepted)
	return c.JSON(fiber.Map{
		"message": "success",
	})
}
