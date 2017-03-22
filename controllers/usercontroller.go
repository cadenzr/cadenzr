package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// TODO: Refactor this away. E.g getUser() function as in laravel.
func isAuthenticated(ctx echo.Context) bool {
	tokenStr := ctx.Request().Header.Get("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// TODO: Is this signing checking correct? Also it should probably be done somewhere else. Because auth has to use the same signing method.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// Also used in authcontroller. Should be refactored somewhere else.
		return []byte("secret"), nil
	})

	if err != nil {
		return false
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true
	}

	return false
}

type userController struct {
}

func (c *userController) Index(echo.Context) error {
	panic("Not implemented")
}

func (c *userController) Show(echo.Context) error {
	panic("Not implemented")
}

func (c *userController) Create(ctx echo.Context) error {
	params := &struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Debugf("UserController::Create Binding params failed: %v", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	params.Username = strings.TrimSpace(params.Username)
	if len(params.Username) == 0 {
		log.Errorf("UserController::Create Username is too short: '%s'", params.Username)
		return ctx.NoContent(http.StatusBadRequest)
	}
	hashSum := sha256.Sum256([]byte(params.Password))
	params.Password = hex.EncodeToString(hashSum[:])
	user := &models.User{
		Username: params.Username,
		Password: params.Password,
	}

	var count uint64
	if gormDB := db.DB.Table("users").Count(&count); gormDB.Error != nil {
		log.Errorf("UserController::Create Checking if this is first user failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if count != 0 && !isAuthenticated(ctx) {
		log.Info("UserController::Create Only existing users can create new users.")
		return ctx.NoContent(http.StatusUnauthorized)
	}

	gormDB := db.DB.Create(user)
	if gormDB.Error != nil {
		log.Errorf("UserController::Create Creating new user failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.WithFields(log.Fields{"id": user.ID, "username": user.Username}).Info("New user registered.")
	return ctx.JSON(http.StatusCreated, user)
}

func (c *userController) Update(echo.Context) error {
	panic("Not implemented")
}

func (c *userController) Delete(echo.Context) error {
	panic("Not implemented")
}

// UserController Contains the actions for the 'users' endpoint.
var UserController userController
