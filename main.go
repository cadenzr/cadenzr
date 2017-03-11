package main

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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
	Id     uint32     `json:"id"`
	Name   string     `json:"name"`
	Artist NullString `json:"artist"`
	Album  NullString `json:"album"`
	Year   NullInt64  `json:"year"`
	Genre  NullString `json:"genre"`
	Mime   string     `json:"mime"`
	Path   string     `json:"-"`
	Cover  NullInt64  `json:"cover"`
}

type Album struct {
	Id   uint32 `json:"id"`
	Name string `json:"name"`
	Year string `json:"year"`
	Path string `json:"-"`

	Songs []*Song `json:"songs"`
}

func (a *Album) GetSongs() []*Song {

	songs := []*Song{}
	for _, song := range a.Songs {
		songs = append(songs, song)
	}

	return songs
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
			Id:   b.nextSongId,
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
			s.Artist.Set(mp3File.Artist())
			s.Album.Set(mp3File.Album())
			year, err := strconv.Atoi(mp3File.Year())
			if err == nil {
				s.Year.Set(int64(year))
			}
			s.Genre.Set(mp3File.Genre())
		}

		insertArtist := "INSERT OR FAIL INTO `artists` (`name`) VALUES (?)"
		artistId := int64(0)
		if len(mp3File.Artist()) > 0 {
			err := db.QueryRow("SELECT `id` FROM `artists` WHERE `name` = ?", mp3File.Artist()).Scan(&artistId)
			if err != nil {
				r, err := db.Exec(insertArtist, mp3File.Artist())
				if err != nil {
					log.WithFields(log.Fields{"reason": err.Error(), "artist": mp3File.Artist()}).Error("Failed to insert artist.")
				} else {
					artistId, _ = r.LastInsertId()
				}
			}
		}

		insertAlbum := "INSERT OR FAIL INTO `albums` (`name`) VALUES (?)"
		albumId := int64(0)
		if len(mp3File.Album()) > 0 {
			err := db.QueryRow("SELECT `id` FROM `albums` WHERE `name` = ?", mp3File.Album()).Scan(&albumId)
			if err != nil {
				r, err := db.Exec(insertAlbum, mp3File.Album())
				if err != nil {
					log.WithFields(log.Fields{"reason": err.Error(), "album": mp3File.Album()}).Error("Failed to insert album.")
				} else {
					albumId, _ = r.LastInsertId()
				}
			}
		}

		insertSong := "INSERT OR IGNORE INTO `songs` (`name`, `artist_id`, `album_id`, `year`, `genre`, `mime`, `path`, `cover_id`) VALUES (?,?,?,?,?,?,?,?)"
		songValues := [9]interface{}{}
		songValues[0] = s.Name
		if artistId != 0 {
			songValues[1] = artistId
		}
		if albumId != 0 {
			songValues[2] = albumId
		}
		songValues[3] = s.Year
		songValues[4] = s.Genre
		songValues[5] = s.Mime
		songValues[6] = s.Path

		_, err = db.Exec(insertSong, songValues[:]...)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "song": s.Name, "artist": artistId, "album": albumId}).Error("Failed to insert song.")
		}

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

var db *sql.DB

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
	db, err = sql.Open("sqlite3", "./db.sqlite")
	if err != nil {
		panic("Could not open database: " + err.Error())
	}

	createSchema()
}

func getSong(id uint32) (*Song, error) {
	song := &Song{}

	query := "SELECT `songs`.`id`, `songs`.`name`, `artists`.`name`, `albums`.`name`, `songs`.`year`, `songs`.`genre`, `songs`.`mime`, `songs`.`path`, `songs`.`cover_id` FROM `songs` JOIN `artists` ON `songs`.`artist_id` = `artists`.`id` JOIN `albums` ON `songs`.`album_id` = `albums`.`id` WHERE `songs`.`id` = ?"
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

func getAlbumSongs(id uint32) ([]*Song, error) {
	songs := []*Song{}

	query := "SELECT `songs`.`id`, `songs`.`name`, `artists`.`name`, `albums`.`name`, `songs`.`year`, `songs`.`genre`, `songs`.`mime`, `songs`.`path`, `songs`.`cover_id` FROM `songs` JOIN `artists` ON `songs`.`artist_id` = `artists`.`id` JOIN `albums` ON `songs`.`album_id` = `albums`.`id` WHERE `songs`.`album_id` = ?"
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
	loadConfig()
	loadDatabase()

	backend := NewBackend()
	backend.Start()

	e := echo.New()
	e.Use(corsHeader)

	e.Static("/app", "app/dist")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/albums", func(c echo.Context) error {
		rows, err := db.Query("SELECT `id`, `name` FROM `albums`")
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not fetch albums.")
			return c.NoContent(http.StatusInternalServerError)
		}

		albums := []Album{}

		for rows.Next() {
			var album Album
			if err := rows.Scan(&album.Id, &album.Name); err != nil {
				log.WithFields(log.Fields{"reason": err.Error()}).Error("Could not scan album.")
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

		album.Songs, err = getAlbumSongs(album.Id)
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

	e.GET("/songs/:id/cover", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song := backend.SongById(id)
		if song == nil {
			return c.NoContent(http.StatusNotFound)
		}

		mp3File, err := id3.Open("./media/" + song.Path)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "song": song.Id}).Info("Couldn't parse id3 tags to get cover art.")
			return c.NoContent(http.StatusNoContent)
		}

		apicFrames := mp3File.Frames("APIC")
		if len(apicFrames) == 0 {
			log.WithFields(log.Fields{"song": song.Id}).Info("Couldn't find APIC frame to get cover art.")
			return c.NoContent(http.StatusNoContent)
		}
		apicFrame := apicFrames[0].(*id3v2.ImageFrame)
		log.WithFields(log.Fields{"song": song.Id, "mime": apicFrame.MIMEType()}).Info("Found cover art.")

		// TODO: mime shouldn't be hardcoded.
		return c.Blob(http.StatusOK, apicFrame.MIMEType(), apicFrame.Data())
	})

	e.Logger.Fatal(e.Start(config.Hostname + ":" + strconv.Itoa(int(config.Port))))
}
