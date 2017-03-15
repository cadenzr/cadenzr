package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cadenzr/cadenzr/transcoders"

	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/probers"
	"github.com/jmoiron/sqlx"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
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
	Hash     string     `db:"hash"`
	Played   int64      `db:"played"`
}

func NewSong() *Song {
	s := &Song{
		// If we leave these nil, we can get a null pointer dereference when trying to insert in database.
		ArtistId: &NullInt64{},
		AlbumId:  &NullInt64{},
		CoverId:  &NullInt64{},
	}

	return s
}

func (s *Song) SetArtist(artist *Artist) {
	s.Artist = artist
	if artist == nil {
		s.ArtistId = &NullInt64{}
	} else {
		s.ArtistId = &artist.Id
	}
}

func (s *Song) SetAlbum(album *Album) {
	s.Album = album
	if album == nil {
		s.AlbumId = &NullInt64{}
	} else {
		s.AlbumId = &album.Id
	}
}

func (s *Song) SetCover(cover *Image) {
	s.Cover = cover
	if cover == nil {
		s.CoverId = &NullInt64{}
	} else {
		s.CoverId = &cover.Id
	}
}

type Album struct {
	Id      NullInt64  `json:"id" db:"id"`
	Name    string     `json:"name" db:"name"`
	Year    NullInt64  `json:"year" db:"year"`
	CoverId *NullInt64 `db:"cover_id"`
	Cover   *Image     `json:"cover"`

	Songs []*Song
}

func NewAlbum() *Album {
	a := &Album{
		// If we leave these nil, we can get a null pointer dereference when trying to insert in database.
		CoverId: &NullInt64{},
	}

	return a
}

func (a *Album) SetCover(cover *Image) {
	a.Cover = cover
	if cover == nil {
		a.CoverId = &NullInt64{}
	} else {
		a.CoverId = &cover.Id
	}
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

type User struct {
	Id       NullInt64 `db:"id"`
	Username string    `db:"username"`
	Password string    `db:"password"`
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
	//log.Debug(query)

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
	//log.Debug(query)

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
	//log.Debug(query)

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
		log.WithFields(log.Fields{"reason": err.Error(), "table": table}).Error("Could not get model.")
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
	Database string `json:"database"`
	LogLevel string `json:"log_level"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var config = Config{}

func loadConfig() {
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Warn("Could not load config.json.")
	} else {
		json.Unmarshal(raw, &config)
	}

	if config.Port == 0 {
		config.Port = 8080
	}

	if len(config.Database) == 0 {
		config.Database = "file::memory:?mode=memory&cache=shared"
	}

	config.LogLevel = strings.ToLower(config.LogLevel)

	switch config.LogLevel {
	case "debug":
	case "info":
	case "warn":
	case "error":
	default:
		config.LogLevel = "info"
	}

	if len(config.Username) == 0 {
		config.Username = "admin"
		config.Password = ""
	}
}

var db *sqlx.DB

func createSchema() {
	schema, err := ioutil.ReadFile("./schema.sql")
	if err != nil {
		log.Fatalln("Could not load schema file: " + err.Error())
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalln("Failed to create schema: " + err.Error())
	}

	log.WithFields(log.Fields{"database": config.Database}).Info("Database initialized.")
}

func loadDatabase() {
	if strings.HasSuffix(config.Database, ".sqlite") {
		//os.Remove(config.Database)
	}

	var err error
	db, err = sqlx.Open("sqlite3", config.Database)
	if err != nil {
		log.Fatalln("Could not open database: " + err.Error())
	}

	createSchema()

	// Initialize admin user.
	shaSum := sha256.Sum256([]byte(config.Password))
	hash := hex.EncodeToString(shaSum[:])
	user := &User{
		Username: config.Username,
	}

	if ok, _ := find("users", user, map[string]interface{}{"username": user.Username}); ok && user.Password == hash {
		// update password if already exists.
		user.Password = hash
		if err := update("users", user, map[string]interface{}{"username": user.Username}); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to update user password.")
		}
	} else {
		user.Password = hash
		if err := insert("users", user); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to create admin user.")
		}
	}

}

type SongResponse struct {
	Id       int64      `db:"id" json:"id"`
	Name     string     `db:"name" json:"name"`
	Artist   NullString `db:"artist" json:"artist"`
	ArtistId NullInt64  `db:"artist_id" json:"-"`
	AlbumId  NullInt64  `db:"album_id" json:"-"`
	Album    NullString `db:"album" json:"album"`
	Year     NullInt64  `db:"year" json:"year"`
	Genre    NullString `db:"genre" json:"genre"`
	Mime     string     `db:"mime" json:"mime"`
	Cover    NullString `db:"cover" json:"cover"`
	Played   int64      `db:"played" json:"played"`
}

type AlbumResponse struct {
	Id     int64           `db:"id" json:"id"`
	Name   string          `db:"name" json:"name"`
	Year   NullInt64       `db:"year" jsosn:"year"`
	Cover  NullString      `db:"cover" json:"cover"`
	Songs  []*SongResponse `json:"songs"`
	Played int64           `json:"played"`
}

// CalculatePlayed Calculates the number of times this album was played.
// Currently by selecting the min value from the played attribute of the songs.
func (a *AlbumResponse) CalculatePlayed() {
	if len(a.Songs) == 0 {
		return
	}

	a.Played = a.Songs[0].Played

	for _, s := range a.Songs {
		if s.Played < a.Played {
			a.Played = s.Played
		}
	}
}

type UserResponse struct {
	Id       int64  `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
}

func getAlbumSongs(ids ...int64) ([]*SongResponse, error) {
	songs := []*SongResponse{}

	query := `
	SELECT
					"songs"."id" as id,
					"songs"."name" as name,
					"songs"."year" as year,
					"songs"."genre" as genre,
					"songs"."mime" as mime,
					"songs"."played" as played,

					"artists"."name" as artist,
					"artists"."id" as artist_id,

					"albums"."name" as album,
					"albums"."id" as album_id,

					"images"."link" as cover
	FROM "songs"
	LEFT OUTER JOIN "artists" ON "songs"."artist_id" = "artists"."id"
	LEFT OUTER JOIN "albums" ON "songs"."album_id" = "albums"."id"
	LEFT OUTER JOIN "images" ON "songs"."cover_id" = "images"."id"
	WHERE "songs"."album_id" in (?)
	`

	query, args, err := sqlx.In(query, ids)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "ids": ids}).Error("Could not create IN query.")
		return songs, err
	}

	query = db.Rebind(query)
	rows, err := db.Queryx(query, args...)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "ids": ids}).Error("Could not create execute query.")
		return songs, err
	}
	defer rows.Close()

	for rows.Next() {
		result := &SongResponse{}
		if err := rows.StructScan(result); err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "ids": ids}).Error("getAlbumSongs: Failed to scan.")
			continue
		}
		songs = append(songs, result)
	}

	return songs, nil
}

func main() {
	loadConfig()
	logLevel := log.InfoLevel
	switch config.LogLevel {
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	default:
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	probers.Initialize()
	loadDatabase()

	scanCh := make(chan (chan struct{}))
	go scanHandler(scanCh)
	done := make(chan struct{})
	scanCh <- done
	<-done

	e := echo.New()
	e.Use(corsHeader)

	e.Static("/", "app/dist")
	e.Static("/images", "images")

	// Login route
	e.POST("/login", login)

	// Restricted group
	r := e.Group("/api")

	// Configure middleware with the custom claims type
	jwtConf := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: []byte("secret"),
	}
	r.Use(middleware.JWTWithConfig(jwtConf))

	r.POST("/scan", func(c echo.Context) error {
		done := make(chan struct{})
		scanCh <- done

		<-done
		return c.NoContent(http.StatusOK)
	})

	r.GET("/albums", func(c echo.Context) error {
		query := `
			SELECT
				"albums"."id" as id,
				"albums"."name" as name,
				"albums"."year" as year,
				"images"."link" as cover
			FROM "albums"
			LEFT OUTER JOIN "images" ON "albums"."cover_id" = "images"."id"
		`
		rows, err := db.Queryx(query)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not fetch albums.")
			return c.NoContent(http.StatusInternalServerError)
		}
		defer rows.Close()

		results := []*AlbumResponse{}
		albums := map[int64]*AlbumResponse{}
		ids := []int64{}
		for rows.Next() {
			album := &AlbumResponse{}
			if err = rows.StructScan(album); err != nil {
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
				return c.NoContent(http.StatusInternalServerError)
			}

			results = append(results, album)
			albums[album.Id] = album
			ids = append(ids, album.Id)
		}

		songs, err := getAlbumSongs(ids...)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		for _, song := range songs {
			albums[song.AlbumId.Int64].Songs = append(albums[song.AlbumId.Int64].Songs, song)
			albums[song.AlbumId.Int64].CalculatePlayed()
		}

		return c.JSON(http.StatusOK, results)
	})

	r.GET("/albums/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		query := `
			SELECT
				"albums"."id" as id,
				"albums"."name" as name,
				"albums"."year" as year,
				"images"."link" as cover
			FROM "albums"
			LEFT OUTER JOIN "images" ON "albums"."cover_id" = "images"."id"
			WHERE "albums"."id" = ?
		`

		album := &AlbumResponse{}
		err := db.QueryRowx(query, id).StructScan(album)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
			return c.NoContent(http.StatusInternalServerError)
		}

		album.Songs, err = getAlbumSongs(album.Id)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not get album songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		album.CalculatePlayed()

		return c.JSON(http.StatusOK, album)
	})

	e.GET("api/songs/:id/stream", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song := &Song{}
		ok, err := find("songs", song, map[string]interface{}{"id": id})
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		codec := strings.ToLower(c.QueryParam("codec"))

		transcode := false
		var targetCodec transcoders.CodecType
		switch codec {
		case "mp3":
			targetCodec = transcoders.MP3
			transcode = true
		case "vorbis":
			targetCodec = transcoders.VORBIS
			transcode = true
		}

		var streamer Streamer
		if transcode {
			streamer, err = NewTranscodeStreamer(song, targetCodec)
		} else {
			streamer, err = NewFileStreamer(song.Path)
		}

		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not create streamer.")
			return c.NoContent(http.StatusNotFound)
		}

		defer streamer.Close()

		http.ServeContent(c.Response(), c.Request(), song.Name, time.Time{}, streamer)
		return nil
	})

	e.POST("api/songs/:id/played", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)

		query := `
			UPDATE "songs"
			SET "played" = "played" + 1
			WHERE "id" = ?
		`

		r, err := db.Exec(query, id)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not update song played count.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if affected, _ := r.RowsAffected(); affected == 0 {
			return c.NoContent(http.StatusNotFound)
		}

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
}
