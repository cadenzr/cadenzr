package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/probers"
	"github.com/cadenzr/cadenzr/scan"

	"github.com/labstack/echo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestArtistControllerIndex(t *testing.T) {
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

		Convey("Test index no-artists.", t, func() {

			req := httptest.NewRequest("get", "/api/artists", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(ArtistController.Index(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusOK)

			response := &struct {
				Data []artistResponse
			}{
				Data: []artistResponse{},
			}
			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
			So(len(response.Data), ShouldEqual, 0)
		})

		// Add a song to the database for testing...
		probers.Initialize()
		go scan.ScanFilesystem("../media/0demo/Curse the Day.mp3")
		<-scan.ScanDone

		Convey("Test index artists.", t, func() {

			req := httptest.NewRequest("get", "/api/artists", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			So(ArtistController.Index(c), ShouldBeNil)
			So(rec.Code, ShouldEqual, http.StatusOK)

			response := &struct {
				Data []artistResponse
			}{
				Data: []artistResponse{},
			}
			err := json.NewDecoder(rec.Result().Body).Decode(response)
			So(err, ShouldEqual, nil)
			So(len(response.Data), ShouldEqual, 1)
			So(response.Data[0].Name, ShouldEqual, "Brain Purist")
			So(len(response.Data[0].Songs), ShouldEqual, 1)
			So(response.Data[0].Songs[0].Name, ShouldEqual, "Curse the Day (Radio Edit)")
		})
	})
}
