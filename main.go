package main

import (
	"database/sql"
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
	ArtistId NullInt64  `db:"artist_id"`
	Artist   *Artist    `json:"artist"`
	AlbumId  NullInt64  `db:"album_id"`
	Album    *Album     `json:"album"`
	Year     NullInt64  `json:"year" db:"year"`
	Genre    NullString `json:"genre" db:"genre"`
	Mime     string     `json:"mime" db:"mime"`
	Path     string     `json:"-" db:"path"`
	CoverId  NullInt64  `db:"cover_id"`
	Cover    *Image     `json:"cover"`
}

func (s *Song) SetArtist(artist *Artist) {
	s.Artist = artist
	s.ArtistId = artist.Id
}

func (s *Song) SetAlbum(album *Album) {
	s.Album = album
	s.AlbumId = album.Id
}

func (s *Song) SetCover(cover *Image) {
	s.Cover = cover
	s.CoverId = cover.Id
}

type Album struct {
	Id      NullInt64 `json:"id" db:"id"`
	Name    string    `json:"name" db:"name"`
	CoverId NullInt64 `db:"cover_id"`
	Cover   *Image    `json:"cover"`

	Songs []*Song
}

func (a *Album) SetCover(cover *Image) {
	a.Cover = cover
	a.CoverId = cover.Id
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

func getSQLValues(v interface{}) (columns []string, values []interface{}) {
	// Remove all indirections so we are left with a struct.
	for reflect.ValueOf(v).Kind() == reflect.Ptr {
		v = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}

	t := reflect.TypeOf(v)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup("db")
		if !ok {
			continue
		}

		columns = append(columns, tag)
		values = append(values, reflect.ValueOf(v).Field(i).Interface())
	}

	return
}

// v should be a pointer to a struct.
func insert(table string, v interface{}) error {
	// Remove all indirections so we are left with a struct.
	columns, values := getSQLValues(v)

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

func insertIfNotExists(table string, v interface{}, exists map[string]interface{}) error {
	columns, _ := getSQLValues(v)
	wheres := []string{}
	values := []interface{}{}
	for k, v := range exists {
		wheres = append(wheres, k+" = ?")
		values = append(values, v)
	}

	query := `SELECT ` + strings.Join(columns, ",") + ` FROM "` + table + `" WHERE ` + strings.Join(wheres, " AND ")
	log.Info(query)

	err := db.Get(v, query, values...)
	switch {
	case err == sql.ErrNoRows:
		err = insert(table, v)
	case err != nil:
		log.WithFields(log.Fields{"reason": err.Error(), "table": table}).Info("Could not get model.")
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

		if insert("songs", s) != nil {
			return nil
		}

		if len(mp3File.Artist()) > 0 {
			artist := &Artist{
				Name: mp3File.Artist(),
			}

			err := insertIfNotExists("artists", artist, map[string]interface{}{
				"name": mp3File.Artist(),
			})

			if err == nil {
				s.SetArtist(artist)
			}
		}

		log.Println(s)

		return nil
	})
}

func (b *Backend) albumByName(name string) *Album {
	for _, album := range b.albums {
		if album.Name == name {
			return album
		}
	}

	return nil
}

func (b *Backend) Albums() []*Album {
	b.lock.RLock()
	defer b.lock.RUnlock()

	albums := []*Album{}
	for _, album := range b.albums {
		albums = append(albums, album)
	}

	return albums
}

func (b *Backend) AlbumById(id uint32) *Album {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.albums[id]
}

func (b *Backend) SongById(id uint32) *Song {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.songs[id]
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
	db, err = sqlx.Open("sqlite3", "./db.sqlite")
	if err != nil {
		panic("Could not open database: " + err.Error())
	}

	createSchema()
}

func getSong(id uint32) (*Song, error) {
	song := &Song{}

	query := "SELECT `songs`.`id`, `songs`.`name`, `artists`.`name`, `albums`.`name`, `songs`.`year`, `songs`.`genre`, `songs`.`mime`, `songs`.`path`, `images`.`link` FROM `songs` JOIN `artists` ON `songs`.`artist_id` = `artists`.`id` JOIN `albums` ON `songs`.`album_id` = `albums`.`id` JOIN `images` ON `songs`.`cover_id` = `images`.`id` WHERE `songs`.`id` = ?"
	err := db.QueryRow(query, id).Scan(&song.Id, &song.Name, &song.Artist, &song.Album, &song.Year, &song.Genre, &song.Mime, &song.Path, &song.Cover)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not select songs for album.")
		return nil, err
	}

	return song, nil
}

func getAlbumSongs(id int64) ([]*Song, error) {
	songs := []*Song{}

	query := "SELECT `songs`.`id`, `songs`.`name`, `artists`.`name`, `albums`.`name`, `songs`.`year`, `songs`.`genre`, `songs`.`mime`, `songs`.`path`, `images`.`link` FROM `songs` JOIN `artists` ON `songs`.`artist_id` = `artists`.`id` JOIN `albums` ON `songs`.`album_id` = `albums`.`id` JOIN `images` ON `songs`.`cover_id` = `images`.`id` WHERE `songs`.`album_id` = ?"
	rows, err := db.Query(query, id)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not select songs for album.")
		return songs, err
	}

	for rows.Next() {
		song := &Song{}

		err := rows.Scan(&song.Id, &song.Name, &song.Artist, &song.Album, &song.Year, &song.Genre, &song.Mime, &song.Path, &song.Cover)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan song.")
			return songs, err
		}

		songs = append(songs, song)
	}

	return songs, nil
}

func main() {
	log.Println(getSQLValues(&Song{}))

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
		rows, err := db.Query("SELECT `albums`.`id`, `albums`.`name`, `images`.`link` FROM `albums` JOIN `images` ON `albums`.`cover_id` = `images`.`id`")
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not fetch albums.")
			return c.NoContent(http.StatusInternalServerError)
		}

		albums := []Album{}

		for rows.Next() {
			var album Album
			if err := rows.Scan(&album.Id, &album.Name, &album.Cover); err != nil {
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
				return c.NoContent(http.StatusInternalServerError)
			}

			album.Songs, err = getAlbumSongs(album.Id.Int64)
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			albums = append(albums, album)
		}

		return c.JSON(http.StatusOK, albums)
	})

	e.GET("/albums/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)

		var album Album
		err := db.QueryRow("SELECT `id`, `name` FROM `albums` WHERE `id` = ?", id).Scan(&album.Id, &album.Name)
		switch {
		case err == sql.ErrNoRows:
			log.WithFields(log.Fields{"album": id}).Error("Album not found.")
			return c.NoContent(http.StatusNotFound)
		case err != nil:
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
			return c.NoContent(http.StatusInternalServerError)
		}

		album.Songs, err = getAlbumSongs(album.Id.Int64)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, album)
	})

	e.GET("/songs/:id/stream", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song, err := getSong(id)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		if song == nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.File(filepath.Join(backend.path, song.Path))
	})

	e.Logger.Fatal(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
}
