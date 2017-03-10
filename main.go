package main

import (
	"bytes"
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
	id3 "github.com/mikkyang/id3-go"
	id3v2 "github.com/mikkyang/id3-go/v2"
)

type Song struct {
	lock sync.RWMutex

	Id             uint32 `json:"id"`
	Name           string `json:"name"`
	Artist         string `json:"artist"`
	Album          string `json:"album"`
	Year           string `json:"year"`
	Genre          string `json:"genre"`
	Mime           string `json:"mime"`
	Path           string `json:"-"`
	StreamLocation string `json:"stream_location"`
	Cover          string `json:"cover"`
}

type Album struct {
	lock sync.RWMutex

	Id   uint32 `json:"id"`
	Name string `json:"name"`
	Year string `json:"year"`
	Path string `json:"-"`

	Songs []*Song `json:"songs"`
}

func (a *Album) GetSongs() []*Song {
	a.lock.RLock()
	defer a.lock.RUnlock()

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
			Id:             b.nextSongId,
			Name:           file,
			Mime:           mimeType,
			Path:           path,
			StreamLocation: "/stream/songs/" + strconv.Itoa(int(b.nextSongId)),
			Cover:          "/songs/" + strconv.Itoa(int(b.nextSongId)) + "/cover",
		}

		mp3File, err := id3.Open("media" + string(filepath.Separator) + path)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Info("Couldn't parse id3 tag")
		} else {
			defer mp3File.Close()
			s.Name = mp3File.Title()
			s.Artist = mp3File.Artist()
			s.Album = mp3File.Album()
			s.Year = mp3File.Year()
			s.Genre = mp3File.Genre()
		}

		b.nextSongId++
		b.songs[s.Id] = s

		//albumName := dir[strings.LastIndexFunc(dir[:len(dir)-1], func(r rune) bool { return r == filepath.Separator })+1 : len(dir)-1]
		//albumName := s.Album
		album := b.albumByName(s.Album)
		if album == nil {
			album = &Album{
				Id:   b.nextAlbumId,
				Name: s.Album,
				Year: s.Year,
				Path: path,
			}
			b.nextAlbumId++

			b.albums[album.Id] = album
			log.WithFields(log.Fields{"name": album.Name}).Info("Created new album.")
		}

		album.Songs = append(album.Songs, s)
		log.WithFields(log.Fields{"name": s.Name, "album": album.Name}).Info("Added song.")

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

type APIC struct {
	Mime        string
	Description string
	Data        []byte
}

func parseAPIC(b []byte) *APIC {
	/**
	This frame contains a picture directly related to the audio file.
	Image format is the MIME type and subtype [MIME] for the image. In
	the event that the MIME media type name is omitted, "image/" will be
	implied. The "image/png" [PNG] or "image/jpeg" [JFIF] picture format
	should be used when interoperability is wanted. Description is a
	short description of the picture, represented as a terminated
	text string. There may be several pictures attached to one file, each
	in their individual "APIC" frame, but only one with the same content
	descriptor. There may only be one picture with the picture type
	declared as picture type $01 and $02 respectively. There is the
	possibility to put only a link to the image file by using the 'MIME
	type' "-->" and having a complete URL [URL] instead of picture data.
	The use of linked files should however be used sparingly since there
	is the risk of separation of files.

		<Header for 'Attached picture', ID: "APIC">
		Text encoding      $xx
		MIME type          <text string> $00
		Picture type       $xx
		Description        <text string according to encoding> $00 (00)
		Picture data       <binary data>


	Picture type:  $00  Other
								 $01  32x32 pixels 'file icon' (PNG only)
								 $02  Other file icon
								 $03  Cover (front)
								 $04  Cover (back)
								 $05  Leaflet page
								 $06  Media (e.g. label side of CD)
								 $07  Lead artist/lead performer/soloist
								 $08  Artist/performer
								 $09  Conductor
								 $0A  Band/Orchestra
								 $0B  Composer
								 $0C  Lyricist/text writer
								 $0D  Recording Location
								 $0E  During recording
								 $0F  During performance
								 $10  Movie/video screen capture
								 $11  A bright coloured fish
								 $12  Illustration
								 $13  Band/artist logotype
								 $14  Publisher/Studio logotype
	**/
	apic := &APIC{}

	// Skip encoding
	b = b[1:]

	mimeEnd := bytes.IndexByte(b, 0x03)
	apic.Mime = string(b[:mimeEnd])
	b = b[mimeEnd+1:]

	apic.Data = b

	return apic
}

func main() {
	loadConfig()

	backend := NewBackend()
	backend.Start()

	e := echo.New()
	e.Use(corsHeader)

	e.Static("/app", "app/dist")

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/albums", func(c echo.Context) error {
		albums := backend.Albums()
		return c.JSON(http.StatusOK, albums)
	})

	e.GET("/albums/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		album := backend.AlbumById(id)
		if album == nil {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusOK, album)
	})

	e.GET("/stream/songs/:id", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		song := backend.SongById(id)
		if song == nil {
			return c.NoContent(http.StatusNotFound)
		}

		/*contents, err := os.Open(filepath.Join(backend.path, song.Path))
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "song": song.Id}).Error("Could not open song for streaming.")
			return c.NoContent(http.StatusInternalServerError)
		}*/

		//c.Response().Header().Add("Accept-Ranges", "bytes") // Allow seeking.
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
