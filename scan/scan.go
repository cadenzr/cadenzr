package scan

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/cadenzr/cadenzr/probers"
)

var ScanDone = make(chan struct{})

func isImage(mime string) bool {
	mime = strings.ToLower(mime)
	return strings.Contains(mime, "image")
}

func ScanHandler(scanCh chan (chan struct{})) {
	requests := []chan struct{}{}
	scanning := false
	for {
		select {
		case done := <-scanCh:
			requests = append(requests, done)
			if scanning == false {
				scanning = true
				go ScanFilesystem("media")
			}
		case <-ScanDone:
			for _, done := range requests {

				done <- struct{}{}
			}
			requests = []chan struct{}{}
			scanning = false
		}
	}
}

func ScanFilesystem(mediaDir string) {
	defer func() {
		ScanDone <- struct{}{}
	}()

	start := time.Now()
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
		var exists uint
		gormDB := db.DB.Table("songs").Where("path = ?", path).Count(&exists)
		if gormDB.Error != nil {
			log.Errorf("scanFilesystem Could not check if path exists. Database failed %v.", gormDB.Error)
			return nil
		}

		if exists != 0 {
			log.Debugf("Skipping file. Path already in database: %s.", path)
			return nil
		}

		meta, err := probers.ProbeAudioFile(path)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "file": path}).Error("Probing file failed.")
			return nil
		}

		var cover *models.Image
		if meta.CoverBufer != nil {
			mimeCover := http.DetectContentType(meta.CoverBufer)
			if isImage(mimeCover) {
				md5Sum := md5.Sum(meta.CoverBufer)
				hash := hex.EncodeToString(md5Sum[:])

				extensions, _ := mime.ExtensionsByType(mimeCover)
				if extensions != nil && len(extensions) > 0 {
					destination := "images/" + hash + extensions[0]
					cover = &models.Image{
						Path: destination,
						Link: "/" + destination,
						Mime: mimeCover,
						Hash: hash,
					}

					if err = ioutil.WriteFile(destination, meta.CoverBufer, 0666); err == nil {
						cover = &models.Image{
							Path: destination,
							Link: "/" + destination,
							Mime: mimeCover,
							Hash: hash,
						}

						gormDB := db.DB.Table("images").Where("hash = ?", cover.Hash).First(cover)
						if gormDB.RecordNotFound() {
							gormDB = db.DB.Create(cover)
						}

						if gormDB.Error != nil {
							log.WithFields(log.Fields{"reason": err.Error(), "file": cover.Path}).Error("Could not get or insert image.")
							cover = nil
						}

					} else {
						log.WithFields(log.Fields{"reason": err.Error(), "destination": destination}).Error("Failed to write cover to disk.")
					}
				}
			}
		}

		song := &models.Song{}
		song.Name = meta.Title
		if len(song.Name) == 0 {
			log.WithFields(log.Fields{"file": path}).Error("Could not get name of song.")
			return nil
		}
		song.Mime = mimeType
		song.Path = path
		if cover != nil {
			song.Cover = cover
		}
		if len(meta.Genre) > 0 {
			song.Genre.Set(meta.Genre)
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No genre found.")
		}
		if meta.Year != 0 {
			song.Year.Set(int64(meta.Year))
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No year found.")
		}
		if len(meta.Album) > 0 {
			album := &models.Album{
				Name: meta.Album,
				Year: song.Year,
			}
			if cover != nil {
				album.Cover = cover
			}
			if gormDB := db.DB.FirstOrCreate(album, "name = ?", album.Name); gormDB.Error != nil {
				log.Errorf("Could not create/get album '%s': %v", album.Name, gormDB.Error)
				return nil
			}

			song.Album = album
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No album found.")
		}
		if len(meta.Artist) > 0 {
			artist := &models.Artist{
				Name: meta.Artist,
			}

			if gormDB := db.DB.FirstOrCreate(artist, "name = ?", artist.Name); gormDB.Error != nil {
				log.Errorf("Could not create/get artist '%s': %v", artist.Name, gormDB.Error)
				return nil
			}

			song.Artist = artist
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No Artist found.")
		}

		if meta.Duration != 0 {
			song.Duration.Set(meta.Duration)
		} else {
			log.WithFields(log.Fields{"file": path}).Debug("No duration found.")
		}

		if gormDB := db.DB.Create(song); gormDB.Error != nil {
			log.Errorf("Could not create song '%s': %v", song.Name, gormDB.Error)
			return nil
		}

		newFiles++
		return nil
	})

	dur := time.Since(start)
	log.Infof("Added %d new songs in %.2f seconds.", newFiles, dur.Seconds())
}
