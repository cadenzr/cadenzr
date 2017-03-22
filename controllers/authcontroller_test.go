package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthControllerLogin(t *testing.T) {
	e := echo.New()

	withDb(func() {
		Convey("Test creating first user so we can login.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admin",
				"password": "somepassword",
			})

			req := httptest.NewRequest("post", "/api/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			So(UserController.Create(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusCreated)
		})

		Convey("Test login.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admin",
				"password": "somepassword",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusOK)

			response := &struct {
				Token string `json:"token"`
			}{}

			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
			So(response.Token, ShouldNotEqual, "")
		})

		Convey("Test login with wrong username.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admi",
				"password": "somepassword",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Test login with wrong password.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admim",
				"password": "somepasswor",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Test login with wrong username and password.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admi",
				"password": "somepasswor",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Test login with no username.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"password": "somepassword",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Test login with no password.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admin",
			})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Test login with no username and password.", t, func() {
			body, _ := json.Marshal(echo.Map{})

			req := httptest.NewRequest("post", "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AuthController.Login(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusBadRequest)
		})
	})
}
