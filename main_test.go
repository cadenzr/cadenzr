package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var endpoint string = "http://localhost:8080"
var token string = ""

func Do(method string, postUrl string) (bodyStr string, status int, err error) {
	req, err := http.NewRequest(method, postUrl, nil)
	if err != nil {
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	status = resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	bodyStr = string(body)

	return

}

func TestApi(t *testing.T) {

	configFile = "./main_test_config.json"

	go main()
	time.Sleep(time.Second * 1)

	Convey("Test login", t, func() {

		resp, err := http.PostForm(
			endpoint+"/api/login",
			url.Values{
				"username": {"admin"},
				"password": {"test"},
			},
		)

		So(err, ShouldEqual, nil)
		So(resp.StatusCode, ShouldEqual, 200)

		body, err := ioutil.ReadAll(resp.Body)
		So(err, ShouldEqual, nil)

		var objmap map[string]string

		err = json.Unmarshal(body, &objmap)
		So(err, ShouldEqual, nil)

		//So(objmap, ShouldEqual, "")

		tokenFound, gotToken := objmap["token"]

		So(gotToken, ShouldEqual, true)
		token = tokenFound

	})

	Convey("Test scan", t, func() {

		_, status, err := Do("POST", endpoint+"/api/scan")
		So(err, ShouldEqual, nil)
		So(status, ShouldEqual, 200)

	})

	Convey("Tests for albums", t, func() {

		Convey("Test get albums", func() {

			body, status, err := Do("GET", endpoint+"/api/albums")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

			albums := map[string][]map[string]interface{}{}
			//So(body, ShouldEqual, -1)
			//So(body, ShouldEqual, 1)
			err = json.Unmarshal([]byte(body), &albums)
			So(err, ShouldEqual, nil)
			So(len(albums["data"]), ShouldBeGreaterThanOrEqualTo, 1)
			So(albums["data"][0]["name"], ShouldEqual, "Curse the Day (Single versions)")

		})

		Convey("Test single album", func() {

			body, status, err := Do("GET", endpoint+"/api/albums/1")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

			album := map[string]interface{}{}
			//So(body, ShouldEqual, -1)
			err = json.Unmarshal([]byte(body), &album)
			So(err, ShouldEqual, nil)

			So(album["id"], ShouldEqual, 1)
			So(album["name"], ShouldEqual, "Curse the Day (Single versions)")
			So(album["year"], ShouldEqual, 2016)
			So(album["cover"], ShouldEqual, "/images/8806119b2782b3f3eaa32129c20533bd.jpg")

		})

		Convey("Test albums playlist m3u8", func() {

			_, status, err := Do("GET", endpoint+"/api/albums/1/playlist.m3u8")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

		})

	})

	Convey("Tests for songs", t, func() {

		Convey("Test song stream", func() {

			_, status, err := Do("GET", endpoint+"/api/songs/1/stream")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

		})

		Convey("Test song played", func() {

			_, status, err := Do("POST", endpoint+"/api/songs/1/played")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

		})

	})

	Convey("Tests for playlists", t, func() {

		Convey("Test playlists", func() {

			_, status, err := Do("GET", endpoint+"/api/playlists")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 200)

		})

		Convey("Test single playlist", func() {

			_, status, err := Do("GET", endpoint+"/api/playlist/1")
			So(err, ShouldEqual, nil)
			So(status, ShouldEqual, 404)

			// Need to add one via post, then test again for code 200
			// and test DELETE of playlist

		})

	})

}