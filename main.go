package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
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

type NullFloat64 struct {
	sql.NullFloat64
}

func (v NullFloat64) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Float64)
	} else {
		return json.Marshal(nil)
	}
}

func (v *NullFloat64) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *float64
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Float64 = *x
	} else {
		v.Valid = false
	}
	return nil
}

func (v *NullFloat64) Set(data float64) {
	v.Float64 = data
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
	Id       NullInt64   `json:"id" db:"id"`
	Name     string      `json:"name" db:"name"`
	ArtistId *NullInt64  `db:"artist_id"`
	Artist   *Artist     `json:"artist"`
	AlbumId  *NullInt64  `db:"album_id"`
	Album    *Album      `json:"album"`
	Year     NullInt64   `json:"year" db:"year"`
	Genre    NullString  `json:"genre" db:"genre"`
	Duration NullFloat64 `db:"duration"`
	Mime     string      `json:"mime" db:"mime"`
	Path     string      `json:"-" db:"path"`
	CoverId  *NullInt64  `db:"cover_id"`
	Cover    *Image      `json:"cover"`
	Played   int64       `db:"played"`
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

type Playlist struct {
	Id    NullInt64 `db:"id"`
	Name  string    `db:"name"`
	Songs []*Song
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

func incrementPlayed(songId int64) (bool, error) {
	query := `
		UPDATE "songs"
		SET "played" = "played" + 1
		WHERE "id" = ?
	`

	r, err := db.Exec(query, songId)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "song": songId}).Error("Could not update song played count.")
		return false, err
	}

	if affected, _ := r.RowsAffected(); affected == 0 {
		return false, nil
	}

	return true, nil
}

func parseUint32(str string, fallback uint32) uint32 {
	n, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return uint32(n)
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
		err = json.Unmarshal(raw, &config)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Warn("Could not parse config.json.")
		}
	}

	config.Hostname = strings.TrimSpace(config.Hostname)

	if len(config.Hostname) == 0 {
		config.Hostname = "127.0.0.1"
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

func loadDatabase() error {
	var err error
	db, err = sqlx.Open("sqlite3", config.Database)
	if err != nil {
		log.Fatalln("Could not open database: " + err.Error())
		return err
	}

	createSchema()

	// Initialize admin user.
	shaSum := sha256.Sum256([]byte(config.Password))
	hash := hex.EncodeToString(shaSum[:])
	user := &User{
		Username: config.Username,
	}

	ok, _ := find("users", user, map[string]interface{}{"username": user.Username})
	if ok && user.Password != hash {
		// update password if already exists.
		user.Password = hash
		if err := update("users", user, map[string]interface{}{"username": user.Username}); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to update user password.")
			return err
		}
	} else if !ok {
		user.Password = hash
		if err := insert("users", user); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to create admin user.")
			return err
		}
	}
	return nil
}

type SongResponse struct {
	Id       int64       `db:"id" json:"id"`
	Name     string      `db:"name" json:"name"`
	Artist   NullString  `db:"artist" json:"artist"`
	ArtistId NullInt64   `db:"artist_id" json:"-"`
	AlbumId  NullInt64   `db:"album_id" json:"-"`
	Album    NullString  `db:"album" json:"album"`
	Year     NullInt64   `db:"year" json:"year"`
	Genre    NullString  `db:"genre" json:"genre"`
	Duration NullFloat64 `db:"duration" json:"duration"`
	Mime     string      `db:"mime" json:"mime"`
	Cover    NullString  `db:"cover" json:"cover"`
	Played   int64       `db:"played" json:"played"`
}

type AlbumResponse struct {
	Id     int64           `db:"id" json:"id"`
	Name   string          `db:"name" json:"name"`
	Year   NullInt64       `db:"year" json:"year"`
	Cover  NullString      `db:"cover" json:"cover"`
	Songs  []*SongResponse `json:"songs"`
	Played int64           `json:"played"`
}

type PlaylistResponse struct {
	Id    int64           `db:"id" json:"id"`
	Name  string          `db:"name" json:"name"`
	Songs []*SongResponse `json:"songs"`
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

	if len(ids) == 0 {
		return songs, nil
	}

	query := `
	SELECT
					"songs"."id" as id,
					"songs"."name" as name,
					"songs"."year" as year,
					"songs"."genre" as genre,
					"songs"."mime" as mime,
					"songs"."played" as played,
					"songs"."duration" as duration,

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

func getPlaylistSongs(ids ...int64) (map[int64][]*SongResponse, error) {
	songs := map[int64][]*SongResponse{}

	if len(ids) == 0 {
		return songs, nil
	}

	for _, id := range ids {
		songs[id] = []*SongResponse{}
	}

	query := `
	SELECT
					"songs"."id" as id,
					"songs"."name" as name,
					"songs"."year" as year,
					"songs"."genre" as genre,
					"songs"."mime" as mime,
					"songs"."played" as played,
					"songs"."duration" as duration,

					"artists"."name" as artist,
					"artists"."id" as artist_id,

					"albums"."name" as album,
					"albums"."id" as album_id,

					"images"."link" as cover,

					"playlist_songs"."playlist_id" as playlist_id
	FROM "songs"
	LEFT OUTER JOIN "artists" ON "songs"."artist_id" = "artists"."id"
	LEFT OUTER JOIN "albums" ON "songs"."album_id" = "albums"."id"
	LEFT OUTER JOIN "images" ON "songs"."cover_id" = "images"."id"
	LEFT OUTER JOIN "playlist_songs" ON "songs"."id" = "playlist_songs"."song_id"
	WHERE "playlist_songs"."playlist_id" in (?)
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
		result := &struct {
			SongResponse
			PlaylistId int64 `db:"playlist_id"`
		}{}
		if err := rows.StructScan(result); err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "ids": ids}).Error("getPlaylistSongs: Failed to scan.")
			continue
		}
		songs[result.PlaylistId] = append(songs[result.PlaylistId], &result.SongResponse)
	}

	return songs, nil
}

func handleInterrupt(stopProgram chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stopProgram <- struct{}{}
	}()
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
	if err := loadDatabase(); err != nil {
		return
	}

	stopProgram := make(chan struct{})
	handleInterrupt(stopProgram)

	scanCh := make(chan (chan struct{}))
	go scanHandler(scanCh)
	/*go func() {
		// We listen for ctrl-c interrupt at the end of main. So start this in a new goroutine so that it doesn't block ctrl-c.
		done := make(chan struct{})
		scanCh <- done
		<-done
	}()*/

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

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
			album := &AlbumResponse{
				Songs: []*SongResponse{},
			}
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
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could get album songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		for _, song := range songs {
			albums[song.AlbumId.Int64].Songs = append(albums[song.AlbumId.Int64].Songs, song)
			albums[song.AlbumId.Int64].CalculatePlayed()
		}

		return c.JSON(http.StatusOK, results)
	})

	// TODO SHOULD BE PROTECTED SOMEHOW.
	e.GET("/api/albums/:id/playlist.m3u8", func(c echo.Context) error {
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

		endpoint := "http://" + config.Hostname
		if config.Port != 0 {
			endpoint = endpoint + ":" + strconv.Itoa(int(config.Port))
		}
		endpoint = endpoint + "/api/songs/"

		response := bytes.NewBuffer([]byte{})
		response.WriteString("#EXTM3U\n")
		for _, song := range album.Songs {
			response.WriteString("#EXTINF:" + strconv.Itoa(int(math.Ceil(song.Duration.Float64))) + ", " + song.Artist.String + " - " + song.Name + "\n")
			response.WriteString(endpoint + strconv.Itoa(int(song.Id)) + "/stream?from=m3u8\n")
		}

		//response.WriteString("#EXTINF:419,Alice in Chains - Rotten Apple")
		return c.Stream(http.StatusOK, "text/plain", response)
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

	e.GET("/api/songs/:id/stream", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song := &Song{}
		ok, err := find("songs", song, map[string]interface{}{"id": id})
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Error while searching song stream.")
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

		if c.FormValue("from") == "m3u8" {
			incrementPlayed(int64(id))
		}

		http.ServeContent(c.Response(), c.Request(), song.Name, time.Time{}, streamer)
		return nil
	})

	r.POST("/songs/:id/played", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)

		ok, err := incrementPlayed(int64(id))
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Warn("Could increment played time.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		return c.NoContent(http.StatusOK)
	})

	r.GET("/playlists", func(c echo.Context) error {
		query := `
			SELECT
				"playlists"."id" as id,
				"playlists"."name" as name
			FROM "playlists"
		`
		rows, err := db.Queryx(query)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not fetch playlists.")
			return c.NoContent(http.StatusInternalServerError)
		}
		defer rows.Close()

		results := []*PlaylistResponse{}
		playlists := map[int64]*PlaylistResponse{}
		ids := []int64{}
		for rows.Next() {
			playlist := &PlaylistResponse{
				Songs: []*SongResponse{},
			}
			if err = rows.StructScan(playlist); err != nil {
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan playlist.")
				return c.NoContent(http.StatusInternalServerError)
			}

			results = append(results, playlist)
			playlists[playlist.Id] = playlist
			ids = append(ids, playlist.Id)
		}

		playlistSongs, err := getPlaylistSongs(ids...)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not get playlist songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		for id, songs := range playlistSongs {
			playlists[id].Songs = songs
		}

		return c.JSON(http.StatusOK, results)
	})

	r.POST("/playlists", func(c echo.Context) error {
		name := strings.TrimSpace(c.FormValue("name"))
		if len(name) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		playlist := &Playlist{
			Name: name,
		}

		ok, err := find("playlists", playlist, map[string]interface{}{"name": playlist.Name})
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Error while looking for playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if ok {
			return c.JSON(http.StatusBadRequest, echo.Map{"message": "Playlist with this name already exists."})
		}

		if err := insert("playlists", playlist); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not insert playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		log.WithFields(log.Fields{"playlist": playlist.Id.Int64, "name": playlist.Name}).Info("Created playlist.")

		return c.JSON(http.StatusOK, &PlaylistResponse{
			Id:    playlist.Id.Int64,
			Name:  playlist.Name,
			Songs: []*SongResponse{},
		})
	})

	r.GET("/playlists/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)

		playlist := &PlaylistResponse{
			Id: int64(id),
		}

		ok, err := find("playlists", playlist, map[string]interface{}{"id": playlist.Id})
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Error while looking for playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		playlistSongs, err := getPlaylistSongs(playlist.Id)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not get playlist songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		playlist.Songs = playlistSongs[playlist.Id]
		return c.JSON(http.StatusOK, playlist)
	})

	r.DELETE("/playlists/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)

		query := `DELETE FROM "playlists" WHERE "id" = ?`
		r, err := db.Exec(query, id)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not execute delete playlist query.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if affected, _ := r.RowsAffected(); affected == 0 {
			return c.NoContent(http.StatusNotFound)
		}

		log.WithFields(log.Fields{"playlist": id}).Info("Deleted playlist.")

		return c.NoContent(http.StatusOK)
	})

	r.POST("/upload", upload)

	r.POST("/playlists/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		name := strings.TrimSpace(c.FormValue("name"))
		if len(name) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		playlist := &Playlist{}
		ok, err := find("playlists", playlist, map[string]interface{}{"id": id})
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Error while looking for playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if !ok {
			return c.NoContent(http.StatusNotFound)
		}

		if err := update("playlists", playlist, map[string]interface{}{"id": playlist.Id.Int64}); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Couldn't update playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		log.WithFields(log.Fields{"playlist": playlist.Id.Int64, "name": playlist.Name}).Info("Updated playlist.")

		return c.NoContent(http.StatusOK)
	})

	r.POST("/playlists/:id/songs", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		sids := []int64{}

		formValues, err := c.FormParams()
		if err != nil {
			log.WithFields(log.Fields{"playlist": id, "form": formValues, "reason": err.Error()}).Error("Could not get form params for adding songs to playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		for key, values := range formValues {
			if key == "songs[]" {
				for _, value := range values {
					sids = append(sids, int64(parseUint32(value, 0)))
				}
			}
		}

		query := `INSERT INTO "playlist_songs" ("playlist_id", "song_id") VALUES (?, ?)`
		tx, err := db.Begin()
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not start transaction to insert songs into playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}
		for _, sid := range sids {
			var r sql.Result
			r, err = tx.Exec(query, id, sid)
			if err != nil {
				tx.Rollback()
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to execute query.")
				return c.NoContent(http.StatusInternalServerError)
			}

			if affected, _ := r.RowsAffected(); affected == 0 {
				tx.Rollback()
				return c.NoContent(http.StatusNotFound)
			}
		}

		if err = tx.Commit(); err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Failed to commit songs to playlist_songs.")
			return c.NoContent(http.StatusInternalServerError)
		}

		log.WithFields(log.Fields{"playlist": id, "songs": sids}).Info("Added song to playlist.")

		return c.NoContent(http.StatusOK)
	})

	r.DELETE("/playlists/:pid/songs/:sid", func(c echo.Context) error {
		pid := parseUint32(c.Param("pid"), 0)
		sid := parseUint32(c.Param("sid"), 0)

		query := `DELETE FROM "playlist_songs" WHERE "playlist_id" = ? AND "song_id" = ?`
		r, err := db.Exec(query, pid, sid)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "playlist": pid, "song": sid}).Info("Could not execute delete song from playlist query.")
			return c.NoContent(http.StatusInternalServerError)
		}

		affected, err := r.RowsAffected()
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "playlist": pid, "song": sid}).Info("Could not delete song from playlist.")
			return c.NoContent(http.StatusInternalServerError)
		}

		if affected == 0 {
			return c.NoContent(http.StatusNotFound)
		}

		log.WithFields(log.Fields{"playlist": pid, "song": sid}).Info("Removed song from playlist.")

		return c.NoContent(http.StatusOK)
	})

	go func() {
		e.Logger.Fatal(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
	}()

	<-stopProgram
	log.Info("Stopping cadenzr...")
	db.Close()
}
