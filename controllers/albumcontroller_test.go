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

			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
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
				Data []*albumResponse
			}{
				Data: []*albumResponse{},
			}

			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
			So(len(response.Data), ShouldEqual, 0)
		})

		year1 := models.NullInt64{}
		year1.Set(1234)
		year2 := models.NullInt64{}
		year2.Set(1235)
		albums := []*models.Album{
			&models.Album{
				Name: "album1",
				Year: year1,
			},
			&models.Album{
				Name: "album2",
				Year: year2,
			},
		}

		for _, album := range albums {
			db.DB.Create(album)
		}

		Convey("Test returned albums are correct.", t, func() {
			req := httptest.NewRequest("get", "/api/albums", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(AlbumController.Index(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusOK)

			response := &struct {
				Data []*albumResponse
			}{
				Data: []*albumResponse{},
			}

			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
			So(len(response.Data), ShouldEqual, len(albums))

			for i, album := range response.Data {
				So(albums[i].Name, ShouldEqual, album.Name)
				So(albums[i].Year.Int64, ShouldEqual, album.Year.Int64)
			}
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

			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
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
