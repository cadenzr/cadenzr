package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/cadenzr/cadenzr/log"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// jwtCustomClaims are custom claims extending default ones.
type jwtCustomClaims struct {
	Id       int32  // json only supports 32bit?
	Username string `json:"username"`
	jwt.StandardClaims
}

func login(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	shaSum := sha256.Sum256([]byte(password))
	hash := hex.EncodeToString(shaSum[:])
	user := &User{}
	ok, err := find("users", user, map[string]interface{}{"username": username, "password": hash})
	if err != nil {
		log.WithFields(log.Fields{"username": username}).Error("Failed to search user in database.")
		return err
	}

	if ok {
		// Set custom claims
		claims := &jwtCustomClaims{
			Id:       int32(user.Id.Int64),
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
			},
		}

		// Create token with claims
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Generate encoded token and send it as response.
		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{"username": username}).Info("Returning token for user.")
		return c.JSON(http.StatusOK, echo.Map{
			"token": t,
		})
	}

	log.WithFields(log.Fields{"username": username}).Info("Wrong credentials.")
	return c.JSON(http.StatusUnauthorized, echo.Map{
		"message": "Wrong credentials.",
	})
}
