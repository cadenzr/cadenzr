package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/probers"
)

var scanDone = make(chan struct{})

func isImage(mime string) bool {
	mime = strings.ToLower(mime)
	return strings.Contains(mime, "image")
}

func scanHandler(scanCh chan (chan struct{})) {
	requests := []chan struct{}{}
	scanning := false
	for {
		select {
		case done := <-scanCh:
			requests = append(requests, done)
			if scanning == false {
				scanning = true
				go scanFilesystem("media")
			}
		case <-scanDone:
			for _, done := range requests {

				done <- struct{}{}
			}
			requests = []chan struct{}{}
			scanning = false
		}
	}
}

func scanFilesystem(mediaDir string) {
	defer func() {
		scanDone <- struct{}{}
	}()

	newFiles := 0

	filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("Failed to handle file/dir.")
			return nil
		}

		if info.IsDir() {
			return nil
		}

		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if !probers.HasProber(mimeType) {
			log.WithFields(log.Fields{"path": path, "mime": mimeType}).Debug("Skipping file. Unknown mime.")
			return nil
		}

		log.WithFields(log.Fields{"path": path, "mime": mimeType}).Debug("Found file.")

		// First check if file with same path exists.
		query := `SELECT "id" FROM "songs" WHERE "path" = ?`
		var exists int64
		err = db.QueryRow(query, path).Scan(&exists)
		switch {
		case err == sql.ErrNoRows:
		case err != nil:
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("Failed to check if path exists in database.")
			return nil
		default:
			log.WithFields(log.Fields{"path": path}).Info("Skipping file. Already in database.")
			return nil
		}

		meta, err := probers.ProbeAudioFile(path)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "file": path}).Error("Probing file failed.")
			return nil
		}

		var cover *Image
		if meta.CoverBufer != nil {
			mimeCover := http.DetectContentType(meta.CoverBufer)
			if isImage(mimeCover) {
				md5Sum := md5.Sum(meta.CoverBufer)
				hash := hex.EncodeToString(md5Sum[:])
				extensions, _ := mime.ExtensionsByType(mimeCover)
				if extensions != nil && len(extensions) > 0 {
					destination := "images/" + hash + extensions[0]
					if err = ioutil.WriteFile(destination, meta.CoverBufer, 0666); err == nil {
						cover = &Image{
							Path: destination,
							Link: "/" + destination,
							Mime: mimeCover,
							Hash: hash,
						}
					} else {
						log.WithFields(log.Fields{"reason": err.Error(), "destination": destination}).Error("Failed to write cover to disk.")
					}
				}
			}
		}

		// TODO mime type is also calculated in ProbeAudioFile. Use that one?
		s := NewSong()
		s.Name = meta.Title
		if len(s.Name) == 0 {
			log.WithFields(log.Fields{"file": path}).Error("Could not get name of song.")
			return nil
		}
		s.Mime = mimeType
		s.Path = path
		if cover != nil {
			s.SetCover(cover)
			insertIfNotExists("images", s.Cover, map[string]interface{}{"hash": s.Cover.Hash})
		}
		if len(meta.Genre) > 0 {
			s.Genre.Set(meta.Genre)
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No genre found.")
		}
		if meta.Year != 0 {
			s.Year.Set(int64(meta.Year))
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No year found.")
		}
		if len(meta.Album) > 0 {
			s.SetAlbum(&Album{
				Name: meta.Album,
				Year: s.Year,
			})

			s.Album.SetCover(s.Cover)

			if err = insertIfNotExists("albums", s.Album, map[string]interface{}{"name": s.Album.Name}); err != nil {
				log.WithFields(log.Fields{"reason": err.Error(), "album": s.Album.Name}).Error("Failed to insert album.")
			}
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No album found.")
		}
		if len(meta.Artist) > 0 {
			s.SetArtist(&Artist{
				Name: meta.Artist,
			})

			if err = insertIfNotExists("artists", s.Artist, map[string]interface{}{"name": s.Artist.Name}); err != nil {
				log.WithFields(log.Fields{"reason": err.Error(), "artist": s.Artist.Name}).Error("Failed to insert artist.")
			}
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No Artist found.")
		}

		if meta.Duration != 0 {
			s.Duration.Set(meta.Duration)
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No duration found.")
		}

		ok, err := find("songs", &Song{}, map[string]interface{}{"name": s.Name, "album_id": s.AlbumId})
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "song": s.Name}).Error("Failed to check if song already exists.")
		} else if !ok {
			if err := insert("songs", s); err != nil {
				log.WithFields(log.Fields{"reason": err.Error(), "song": s.Name}).Error("Failed to insert song.")
			} else {
				newFiles++
			}
		} else {
			albumName := ""
			if s.Album != nil {
				albumName = s.Album.Name
			}
			log.WithFields(log.Fields{"song": s.Name, "album": albumName}).Info("Song already in database.")
		}

		return nil
	})

	log.Infof("Added %d new songs.", newFiles)
}
