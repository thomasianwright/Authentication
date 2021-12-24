package controllers

import (
	"authapi/database"
	"authapi/models"
	"encoding/json"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	uuid2 "github.com/nu7hatch/gouuid"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
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
	UpdateSuccess    = 1
	UpdateNotSuccess = 2
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

	error := database.DB.Create(&user)

	if error.Error != nil {
		fmt.Println(error.Error)
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

	if errs.Error != nil {
		uuid, _ := uuid2.NewV4()

		database.DB.Create(&models.Activation{
			UserId: user.Id,
			Guid:   uuid.String(),
		})
	}

	jsonReq, _ := json.Marshal(map[string]string{
		"firstname": user.Firstname,
		"lastname":  user.Lastname,
		"email":     user.Email,
		"token":     uuid.String(),
	})

	resp, _ := httpclient.WithHeader("Content-Type", "application/json").PostJson("https://localhost:7033/Email", string(jsonReq))

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	type Response struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}

	var responseData Response

	json.Unmarshal(bodyBytes, &responseData)
	m, _ := json.Marshal(responseData)
	fmt.Println(string(m))

	if responseData.Status != 1 {
		return c.JSON(fiber.Map{
			"status":  2,
			"message": "Register failed contact support to activate your account.",
		})
	}

	fmt.Println("User: " + user.Email + " Activation Code: " + uuid.String())
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
		c.Status(fiber.StatusOK)
		return c.JSON(fiber.Map{
			"message":     "User not found",
			"loginStatus": LoginNotFound,
		})
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		c.Status(fiber.StatusOK)
		return c.JSON(fiber.Map{
			"message":     "incorrect password",
			"loginStatus": LoginNotFound,
		})
	}

	if !user.Activated {
		c.Status(fiber.StatusOK)
		return c.JSON(fiber.Map{
			"message":     "Please check your email to activate your account!",
			"loginStatus": LoginNotActivated,
		})
	}

	user.LastLogin = time.Now()

	database.DB.Save(&user)

	// Create the Claims
	claims := jwt.MapClaims{
		"issuer":    "authapi",
		"id":        user.Id,
		"firstname": user.Firstname,
		"lastname":  user.Lastname,
		"username":  user.Username,
		"email":     user.Email,
		"exp":       time.Now().Add(time.Hour * 3).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("jwtsecret")))
	if err != nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(fiber.Map{
			"message": "Something went wrong",
			"status":  500,
		})
	}

	c.Status(fiber.StatusOK)
	return c.JSON(fiber.Map{
		"message":     "success",
		"token":       t,
		"loginStatus": LoginSuccess,
		"user":        user,
	})
}

func AuthenticateToken(c *fiber.Ctx) error {
	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err.Error())
		return err
	}

	_, err := jwt.ParseWithClaims(data["token"], jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	token, err := jwt.ParseWithClaims(data["token"], jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	claims := token.Claims.(jwt.MapClaims)

	var user models.User

	database.DB.Where("id = ?", claims["id"]).First(&user)

	return c.JSON(user)
}

func ActivateAccount(c *fiber.Ctx) error {
	var data map[string]string
	c.BodyParser(&data)
	var activation models.Activation

	database.DB.Where("guid = ?", data["guid"]).First(&activation)

	if activation.Id == 0 {
		return c.JSON(fiber.Map{
			"status":  ActivateNotFound,
			"message": "Token not found",
		})
	}
	var user models.User

	database.DB.Where("id = ?", activation.UserId).First(&user)

	user.Activated = true

	database.DB.Save(&user)
	database.DB.Delete(&activation)

	return c.JSON(fiber.Map{
		"status":  ActivateSuccess,
		"message": "success",
		"user":    user,
	})
}

func GetUser(c *fiber.Ctx) error {
	jwtUser := c.Locals("user").(*jwt.Token)
	claims := jwtUser.Claims.(jwt.MapClaims)
	id := uint(claims["id"].(float64))

	var user models.User
	database.DB.Where("id = ?", id).First(&user)

	return c.JSON(user)
}

func UpdateUser(c *fiber.Ctx) error {
	jwtUser := c.Locals("user").(*jwt.Token)
	claims := jwtUser.Claims.(jwt.MapClaims)
	id := uint(claims["id"].(float64))

	var user models.User
	database.DB.Where("id = ?", id).First(&user)

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err.Error())
		return c.JSON(fiber.Map{
			"status":  0,
			"message": "Something went wrong",
		})
	}

	user.Firstname = data["firstname"]
	user.Lastname = data["lastname"]
	user.Username = data["username"]
	user.Email = data["email"]

	database.DB.Save(&user)

	c.Status(fiber.StatusOK)

	return c.JSON(fiber.Map{
		"status":  UpdateSuccess,
		"message": "success",
		"user":    user,
	})
}

func UpdatePassword(c *fiber.Ctx) error {
	jwtUser := c.Locals("user").(*jwt.Token)
	claims := jwtUser.Claims.(jwt.MapClaims)
	fmt.Println(json.Marshal(claims))
	id := uint(claims["id"].(float64))

	var user models.User
	database.DB.Where("id = ?", id).First(&user)

	var data map[string]string

	if err := c.BodyParser(&data); err != nil {
		fmt.Println(err.Error())
		return c.JSON(fiber.Map{
			"status":  0,
			"message": "Something went wrong",
		})
	}
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["oldPassword"])); err != nil {
		fmt.Println(err.Error())
		c.Status(fiber.StatusBadRequest)
		return c.JSON(fiber.Map{
			"message":     "incorrect password",
			"loginStatus": LoginNotFound,
		})
	}

	newPassword, _ := bcrypt.GenerateFromPassword([]byte(data["newPassword"]), 14)
	user.Password = newPassword
	user.UpdatedAt = time.Now()
	database.DB.Save(&user)

	c.Status(fiber.StatusOK)

	return c.JSON(fiber.Map{
		"status":  UpdateSuccess,
		"message": "success",
		"user":    user,
	})
}
