package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cadenzr/cadenzr/db"

	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

func withDb(cb func()) {
	if err := db.SetupConnection(db.SQLITE, "file:memdb1?mode=memory&cache=shared"); err != nil {
		panic(err)
	}

	if err := db.SetupSchema(); err != nil {
		panic(err)
	}

	cb()

	if err := db.Shutdown(); err != nil {
		panic(err)
	}
}

func TestUserControllerCreateUser(t *testing.T) {
	e := echo.New()

	withDb(func() {
		Convey("Test first user creation.", t, func() {
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
			var count uint64
			db.DB.Table("users").Count(&count)
			So(count, ShouldEqual, 1)
		})

		Convey("Test only existing users can create new users.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "admin2",
				"password": "somepassword2",
			})

			req := httptest.NewRequest("post", "/api/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(UserController.Create(c), ShouldBeNil)
			So(rec.Code, ShouldBeIn, 401, 403)
			var count uint64
			db.DB.Table("users").Count(&count)
			So(count, ShouldEqual, 1)
		})

		Convey("Test if we can create user after logging in.", t, func() {
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

			json.NewDecoder(rec.Result().Body).Decode(response)

			body, _ = json.Marshal(echo.Map{
				"username": "admin2",
				"password": "somepassword2",
			})

			req = httptest.NewRequest("post", "/api/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+response.Token)
			rec = httptest.NewRecorder()
			c = e.NewContext(req, rec)

			So(UserController.Create(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusCreated)
			var count uint64
			db.DB.Table("users").Count(&count)
			So(count, ShouldEqual, 2)
		})
	})
}

func TestUserControllerCreateInvalidUser(t *testing.T) {
	e := echo.New()

	withDb(func() {
		Convey("Test empty username.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"username": "",
				"password": "somepassword",
			})

			req := httptest.NewRequest("post", "/api/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(UserController.Create(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, 400)
			var count uint64
			db.DB.Table("users").Count(&count)
			So(count, ShouldEqual, 0)
		})

		Convey("Test no username field.", t, func() {
			body, _ := json.Marshal(echo.Map{
				"password": "somepassword2",
			})

			req := httptest.NewRequest("post", "/api/users", bytes.NewBuffer(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(UserController.Create(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, 400)
			var count uint64
			db.DB.Table("users").Count(&count)
			So(count, ShouldEqual, 0)
		})
	})
}
