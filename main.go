package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo"
	_ "github.com/mattn/go-sqlite3"
	id3 "github.com/mikkyang/id3-go"
	id3v2 "github.com/mikkyang/id3-go/v2"
)

type NullInt64 struct {
	sql.NullInt64
}

func (v NullInt64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Int64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullInt64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *int64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullInt64) Set(data int64) {
	v.Int64 = data
	v.Valid = true
}

type NullString struct {
	sql.NullString
}

func (v NullString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullString) Set(data string) {
	v.String = data
	v.Valid = true
}

type Song struct {
	Id       NullInt64  `json:"id" db:"id"`
	Name     string     `json:"name" db:"name"`
	ArtistId *NullInt64 `db:"artist_id"`
	Artist   *Artist    `json:"artist"`
	AlbumId  *NullInt64 `db:"album_id"`
	Album    *Album     `json:"album"`
	Year     NullInt64  `json:"year" db:"year"`
	Genre    NullString `json:"genre" db:"genre"`
	Mime     string     `json:"mime" db:"mime"`
	Path     string     `json:"-" db:"path"`
	CoverId  *NullInt64 `db:"cover_id"`
	Cover    *Image     `json:"cover"`
}

func (s *Song) SetArtist(artist *Artist) {
	s.Artist = artist
	s.ArtistId = &artist.Id
}

func (s *Song) SetAlbum(album *Album) {
	s.Album = album
	s.AlbumId = &album.Id
}

func (s *Song) SetCover(cover *Image) {
	s.Cover = cover
	s.CoverId = &cover.Id
}

type Album struct {
	Id      NullInt64  `json:"id" db:"id"`
	Name    string     `json:"name" db:"name"`
	CoverId *NullInt64 `db:"cover_id"`
	Cover   *Image     `json:"cover"`

	Songs []*Song
}

func (a *Album) SetCover(cover *Image) {
	a.Cover = cover
	a.CoverId = &cover.Id
}

type Artist struct {
	Id   NullInt64 `db:"id"`
	Name string    `db:"name"`
}

type Image struct {
	Id   NullInt64 `db:"id"`
	Path string    `db:"path"`
	Link string    `db:"link"`
	Mime string    `db:"mime"`
	Hash string    `db:"hash"`
}

type Backend struct {
	lock        sync.RWMutex
	nextSongId  uint32
	nextAlbumId uint32

	path   string
	albums map[uint32]*Album
	songs  map[uint32]*Song // Songs that do not belong to any albums.
}

func NewBackend() *Backend {
	return &Backend{
		nextSongId:  1,
		nextAlbumId: 1,
		path:        "./media",
		albums:      map[uint32]*Album{},
		songs:       map[uint32]*Song{},
	}
}

func getSQLColumns(v interface{}) (columns []string) {
	t := reflect.TypeOf(v)
	tv := reflect.ValueOf(v)
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if t.Kind() == reflect.Ptr {
			tv = reflect.Indirect(tv)
			t = t.Elem()
		}

		if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
			t = t.Elem()
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup("db")
		if !ok {
			continue
		}

		columns = append(columns, tag)
	}

	return
}

func getSQLValues(v interface{}) (values []interface{}) {
	t := reflect.TypeOf(v)
	tv := reflect.ValueOf(v)
	for t.Kind() == reflect.Ptr {
		if t.Kind() == reflect.Ptr {
			tv = reflect.Indirect(tv)
			t = t.Elem()
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		_, ok := field.Tag.Lookup("db")
		if !ok {
			continue
		}

		values = append(values, reflect.ValueOf(v).Elem().Field(i).Interface())
	}

	return
}

func update(table string, v interface{}, where map[string]interface{}) error {
	// Remove all indirections so we are left with a struct.
	columns := getSQLColumns(v)
	values := getSQLValues(v)
	for i := range columns {
		columns[i] = columns[i] + "=?"
	}

	wheres := []string{}
	for k, v := range where {
		wheres = append(wheres, k+"=?")
		values = append(values, v)
	}

	query := `
	UPDATE "` + table + `"
	SET ` + strings.Join(columns, ",") + `
	WHERE ` + strings.Join(wheres, " AND ") + `
	`
	log.Info(query)

	_, err := db.Exec(query, values...)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Update " + table + " failed.")
		return err
	}

	return nil
}

// v should be a pointer to a struct.
func insert(table string, v interface{}) error {
	// Remove all indirections so we are left with a struct.
	columns := getSQLColumns(v)
	values := getSQLValues(v)

	query := `
	INSERT INTO "` + table + `" (` + strings.Join(columns, ",") + `)
	VALUES (` + strings.Join(strings.Split(strings.Repeat("?", len(columns)), ""), ",") + `)
	`
	log.Info(query)

	r, err := db.Exec(query, values...)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Insert into " + table + " failed.")
		return err
	}

	id, _ := r.LastInsertId()
	targetId := reflect.ValueOf(v).Elem().FieldByName("Id").Addr().Interface().(*NullInt64)
	targetId.Set(id)
	return nil
}

func find(table string, v interface{}, where map[string]interface{}) (bool, error) {
	underLyingType := reflect.TypeOf(v)
	for underLyingType.Kind() == reflect.Ptr {
		underLyingType = underLyingType.Elem()
	}

	columns := getSQLColumns(v)
	wheres := []string{}
	values := []interface{}{}
	for k, v := range where {
		wheres = append(wheres, k+" = ?")
		values = append(values, v)
	}

	query := `SELECT ` + strings.Join(columns, ",") + ` FROM "` + table + `" WHERE ` + strings.Join(wheres, " AND ")
	log.Info(query)

	var err error
	if underLyingType.Kind() == reflect.Slice || underLyingType.Kind() == reflect.Array {
		err = db.Select(v, query, values...)
	} else {
		err = db.Get(v, query, values...)
	}

	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		log.WithFields(log.Fields{"reason": err.Error(), "table": table}).Info("Could not get model.")
		return false, err
	}

	return true, nil
}

func insertIfNotExists(table string, v interface{}, exists map[string]interface{}) error {
	ok, err := find(table, v, exists)
	if err != nil {
		return err
	}

	if !ok {
		return insert(table, v)
	}

	return err
}

func (b *Backend) scanFilesystem() {

	filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		// Remove our base directory.
		path = path[strings.IndexRune(path, filepath.Separator)+1:]

		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("Failed to handle file/dir.")
			return nil
		}

		if info.IsDir() {
			return nil
		}

		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if !isAudio(mimeType) {
			log.WithFields(log.Fields{"path": path, "mime": mimeType}).Debug("Skipping file. Unknown mime.")
			return nil
		}

		log.WithFields(log.Fields{"path": path, "mime": mimeType}).Info("Found file.")

		_, file := filepath.Split(path)
		s := &Song{
			Name: file,
			Mime: mimeType,
			Path: path,
		}

		mp3File, err := id3.Open("media" + string(filepath.Separator) + path)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Info("Couldn't parse id3 tag")

		} else {
			defer mp3File.Close()
			s.Name = mp3File.Title()
			year, err := strconv.Atoi(mp3File.Year())
			if err == nil {
				s.Year.Set(int64(year))
			}
			s.Genre.Set(mp3File.Genre())
		}

		if len(mp3File.Artist()) > 0 {
			artist := &Artist{
				Name: mp3File.Artist(),
			}

			if insertIfNotExists("artists", artist, map[string]interface{}{"name": artist.Name}) == nil {
				s.SetArtist(artist)
			}
		}

		if len(mp3File.Album()) > 0 {
			album := &Album{
				Name: mp3File.Album(),
			}

			s.SetAlbum(album)
		}

		if apic, ok := mp3File.Frame("APIC").(*id3v2.ImageFrame); ok {
			image := &Image{}

			// Use DetectContentType since it is more reliable than apic.MimeType.
			if coverMime := http.DetectContentType(apic.Data()); isImage(coverMime) {
				image.Mime = coverMime
			} else {
				goto SkipCover
			}

			md5Bytes := md5.Sum(apic.Data())
			image.Hash = hex.EncodeToString(md5Bytes[:])

			if ok, _ := find("images", image, map[string]interface{}{"hash": image.Hash}); !ok {
				extensions, err := mime.ExtensionsByType(image.Mime)
				if err != nil {
					log.WithFields(log.Fields{"reason": err.Error(), "mime": image.Mime}).Info("Failed to guess extension for mime.")
					goto SkipCover
				}

				if extensions == nil || len(extensions) == 0 {
					log.WithFields(log.Fields{"mime": image.Mime}).Info("No extensions found for mime.")
					goto SkipCover
				}

				image.Path = "images/" + image.Hash + extensions[0]
				image.Link = "/images/" + image.Hash + extensions[0]
				if err := ioutil.WriteFile(image.Path, apic.Data(), 0666); err != nil {
					log.WithFields(log.Fields{"reason": err.Error(), "path": image.Path}).Info("Failed to store cover.")
					goto SkipCover
				}

				insert("images", image)
			}

			if image.Id.Valid {
				s.SetCover(image)
				if s.Album != nil && s.Album.Cover == nil {
					s.Album.SetCover(image)
				}
			}
		}
	SkipCover:
		insertIfNotExists("albums", s.Album, map[string]interface{}{"name": s.Album.Name})
		insert("songs", s)
		return nil
	})
}

func (b *Backend) Start() {
	b.scanFilesystem()
}

func isAudio(mime string) bool {
	mime = strings.ToLower(mime)
	return strings.Contains(mime, "audio")
}

func isImage(mime string) bool {
	mime = strings.ToLower(mime)
	return strings.Contains(mime, "image")
}

func parseUint32(str string, fallback uint32) uint32 {
	n, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return uint32(n)
}

func corsHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		return next(c)
	}
}

type Config struct {
	Hostname string `json:"hostname"`
	Port     uint32 `json:"port"`
}

var config = Config{}

func loadConfig() {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Warn("Could not load config.json.")

		config.Port = 8080
		return
	}

	json.Unmarshal(raw, &config)
}

var db *sqlx.DB

func createSchema() {
	schema, err := ioutil.ReadFile("./schema.sql")
	if err != nil {
		panic("Could not load schema file: " + err.Error())
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		panic("Failed to create schema: " + err.Error())
	}

	log.Info("Schema created.")
}

func loadDatabase() {
	os.Remove("./db.sqlite")
	var err error
	db, err = sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		panic("Could not open database: " + err.Error())
	}

	createSchema()
}

type JsonSong struct {
	Id     int64      `db:"id" json:"id"`
	Name   string     `db:"name" json:"name"`
	Artist NullString `db:"artist" json:"artist"`
	Album  NullString `db:"album" json:"album"`
	Year   int        `db:"year" json:"year"`
	Genre  string     `db:"genre" json:"genre"`
	Mime   string     `db:"mime" json:"mime"`
	Cover  NullString `db:"cover" json:"cover"`
}

type JsonAlbum struct {
	Id    int64      `db:"id" json:"id"`
	Name  string     `db:"name" json:"name"`
	Cover NullString `db:"cover" json:"cover"`
}

func getAlbumSongs(id int64) ([]*JsonSong, error) {
	songs := []*JsonSong{}

	query := `
	SELECT
					"songs"."id" as id,
					"songs"."name" as name,
					"songs"."year" as year,
					"songs"."genre" as genre,
					"songs"."mime" as mime,

					"artists"."name" as artist,

					"albums"."name" as album,

					"images"."link" as cover
	FROM "songs"
	JOIN "artists" ON "songs"."artist_id" = "artists"."id"
	JOIN "albums" ON "songs"."album_id" = "albums"."id"
	JOIN "images" ON "songs"."cover_id" = "images"."id"
	WHERE "songs"."album_id" = ?
	`

	rows, err := db.Queryx(query, id)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "album": id}).Error("Could not select songs for album.")
		return songs, err
	}
	defer rows.Close()

	for rows.Next() {
		result := &JsonSong{}
		if err := rows.StructScan(result); err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "album": id}).Error("getAlbumSongs: Failed to scan.")
			continue
		}
		songs = append(songs, result)
	}

	return songs, nil
}

func main() {
	loadConfig()
	loadDatabase()

	backend := NewBackend()
	backend.Start()

	e := echo.New()
	e.Use(corsHeader)

	e.Static("/app", "app/dist")
	e.Static("/images", "images")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/albums", func(c echo.Context) error {
		query := `
			SELECT
				"albums"."id",
				"albums"."name",
				"images"."link"
			FROM "albums"
			JOIN "images" ON "albums"."cover_id" = "images"."id"
		`
		rows, err := db.Query(query)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not fetch albums.")
			return c.NoContent(http.StatusInternalServerError)
		}

		results := []map[string]interface{}{}

		for rows.Next() {
			var id int64
			var name string
			var cover NullString
			if err := rows.Scan(&id, &name, &cover); err != nil {
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
				return c.NoContent(http.StatusInternalServerError)
			}

			songs, err := getAlbumSongs(id)
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			results = append(results, map[string]interface{}{
				"id":    id,
				"name":  name,
				"cover": cover,
				"songs": songs,
			})
		}

		return c.JSON(http.StatusOK, results)
	})

	e.GET("/albums/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		query := `
			SELECT
				"albums"."id",
				"albums"."name",
				"images"."link"
			FROM "albums"
			JOIN "images" ON "albums"."cover_id" = "images"."id"
			WHERE "albums"."id" = ?
		`

		var name string
		var cover NullString
		err := db.QueryRowx(query, id).Scan(&id, &name, &cover)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
			return c.NoContent(http.StatusInternalServerError)
		}

		songs, err := getAlbumSongs(int64(id))
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not get album songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		result := map[string]interface{}{
			"id":    id,
			"name":  name,
			"cover": cover,
			"songs": songs,
		}

		return c.JSON(http.StatusOK, result)
	})

	e.GET("/songs/:id/stream", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song := &Song{}
		ok, err := find("songs", song, map[string]interface{}{"id": id})
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		return c.File(filepath.Join(backend.path, song.Path))
	})

	e.Logger.Fatal(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
}
