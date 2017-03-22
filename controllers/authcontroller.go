package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// UserLoginClaim is the claim that will be used for jwt.
type UserLoginClaim struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// Secret used for signing tokens.
var Secret = []byte("secret")

type authController struct {
}

func (c *authController) Login(ctx echo.Context) error {
	params := &struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Debugf("AuthController::Login Binding params failed: %v", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	hashSum := sha256.Sum256([]byte(params.Password))
	params.Password = hex.EncodeToString(hashSum[:])
	user := &models.User{}

	gormDB := db.DB.Find(user, "username = ?", params.Username)
	authenticated := false
	if gormDB.RecordNotFound() {
		log.WithFields(log.Fields{"username": params.Username}).Info("AuthController::Login Username not found.")
	} else if gormDB.Error != nil {
		log.Errorf("AuthController::Login Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if params.Password != user.Password {
		fmt.Println(params.Password)
		fmt.Println(user.Password)
		log.WithFields(log.Fields{"username": params.Username}).Info("AuthController::Login Wrong password.")
	} else {
		authenticated = true
	}

	if !authenticated {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid username or password.",
		})
	}

	claims := &UserLoginClaim{
		ID:       user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			// TODO: Get expiration time from config.
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}

	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// TODO: Get key from config.
	// And this is also used in usercontroller. Should be places somewhere else...
	signedToken, err := unsignedToken.SignedString(Secret)
	if err != nil {
		log.Errorf("AuthController::Login Could not sign jwt token: %v", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.WithFields(log.Fields{"id": user.ID, "username": user.Username, "token": signedToken}).Info("Generated user login token.")
	return ctx.JSON(http.StatusOK, echo.Map{"token": signedToken})
}

// AuthController Contains the actions for authentication.
var AuthController authController
