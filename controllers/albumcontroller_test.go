package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/models"

	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlbumControllerIndex(t *testing.T) {
	e := echo.New()

	withDb(func() {
		var token string
		Convey("Creating user to access albums.", t, func() {
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

			response := &struct {
				Token string `json:"token"`
			}{}

			json.NewDecoder(rec.Result().Body).Decode(response)
			token = response.Token
		})

		Convey("Test no albums.", t, func() {
			req := httptest.NewRequest("get", "/api/albums", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AlbumController.Index(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusOK)

			response := &struct {
				Data []models.Album
			}{
				Data: []models.Album{},
			}

			json.NewDecoder(rec.Result().Body).Decode(response)
			So(len(response.Data), ShouldEqual, 0)
		})
	})
}

func TestAlbumControllerShow(t *testing.T) {
	e := echo.New()

	withDb(func() {
		var token string
		Convey("Creating user to access albums.", t, func() {
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

			response := &struct {
				Token string `json:"token"`
			}{}

			json.NewDecoder(rec.Result().Body).Decode(response)
			token = response.Token
		})

		Convey("Test non existing album.", t, func() {
			req := httptest.NewRequest("get", "/api/albums/1", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AlbumController.Show(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusNotFound)
		})
	})
}
