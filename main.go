package main

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo"
)

type Song struct {
	lock sync.RWMutex

	Id   uint32 `json:"id"`
	Name string `json:"name"`
	Mime string `json:"mime"`
	Path string `json:"-"`
}

type Album struct {
	lock sync.RWMutex

	Id   uint32 `json:"id"`
	Name string `json:"name"`
	Path string `json:"-"`

	Songs []*Song `json:"-"`
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

		log.WithFields(log.Fields{"path": path, "mime": mimeType}).Debug("Found file.")

		i := strings.LastIndexFunc(path, func(r rune) bool { return r == filepath.Separator })
		if i == -1 {
			// No seperator -> Song in root directory -> No album.
			s := &Song{
				Id:   b.nextSongId,
				Name: path,
				Mime: mimeType,
				Path: "",
			}
			b.nextSongId++
			b.songs[s.Id] = s

			log.WithFields(log.Fields{"name": s.Name}).Info("Added song.")
		} else {
			// Seperator so guess that directory is album name.
			dir, file := filepath.Split(path)
			albumName := dir[strings.LastIndexFunc(dir[:len(dir)-1], func(r rune) bool { return r == filepath.Separator })+1 : len(dir)-1]
			album := b.albumByName(albumName)
			if album == nil {
				album = &Album{
					Id:   b.nextAlbumId,
					Name: albumName,
					Path: path,
				}
				b.nextAlbumId++

				b.albums[album.Id] = album
				log.WithFields(log.Fields{"name": album.Name}).Info("Created new album.")
			}

			s := &Song{
				Id:   b.nextSongId,
				Name: file,
				Mime: mimeType,
				Path: path,
			}
			b.nextSongId++
			b.songs[s.Id] = s

			album.Songs = append(album.Songs, s)
			log.WithFields(log.Fields{"name": s.Name, "album": album.Name}).Info("Added song.")
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

func main() {
	backend := NewBackend()
	backend.Start()

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/albums", func(c echo.Context) error {
		albums := backend.Albums()
		return c.JSON(http.StatusOK, albums)
	})

	e.GET("/albums/:id/songs", func(c echo.Context) error {
		id := parseUint32(c.Param("id"), 0)
		album := backend.AlbumById(id)
		if album == nil {
			return c.NoContent(http.StatusNotFound)
		}

		songs := album.GetSongs()

		return c.JSON(http.StatusOK, songs)
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

	e.Logger.Fatal(e.Start(":8080"))
}
