package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	mydb "github.com/cadenzr/cadenzr/db"

	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/probers"
	"github.com/jmoiron/sqlx"

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
	return mydb.SetupConnection(mydb.SQLITE, config.Database)
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
		log.Fatalf("Failed to create connection to database: %v", err)
	}

	if err := mydb.SetupSchema(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	stopProgram := make(chan struct{})
	handleInterrupt(stopProgram)

	<-stopProgram
	log.Info("Stopping cadenzr...")
	db.Close()
}
