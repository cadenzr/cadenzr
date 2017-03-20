package probers

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/cadenzr/cadenzr/log"
)

type AudioMeta struct {
	Title       string
	Artist      string
	Album       string
	Track       int
	TotalTracks int
	Year        int
	AlbumArtist string
	Genre       string
	Duration    float64

	CoverBufer []byte
}

func (a *AudioMeta) IsComplete() bool {
	complete := true

	complete = complete && len(a.Title) > 0
	complete = complete && len(a.Artist) > 0
	complete = complete && len(a.Album) > 0
	complete = complete && a.Year > 0
	complete = complete && len(a.Genre) > 0
	complete = complete && a.Duration > 0
	complete = complete && a.CoverBufer != nil

	return complete
}

// Merge Set a's attributes to the attributes that b has more.
func (a *AudioMeta) Merge(b *AudioMeta) {
	if len(a.Title) == 0 {
		a.Title = b.Title
	}

	if len(a.Artist) == 0 {
		a.Artist = b.Artist
	}

	if len(a.Album) == 0 {
		a.Album = b.Album
	}

	if a.Track != b.Track && a.Track == 0 {
		a.Track = b.Track
	}

	if a.TotalTracks != b.TotalTracks && a.TotalTracks == 0 {
		a.TotalTracks = b.TotalTracks
	}

	if a.Year != b.Year && a.Year == 0 {
		a.Year = b.Year
	}

	if len(a.AlbumArtist) == 0 {
		a.AlbumArtist = b.AlbumArtist
	}

	if len(a.Genre) == 0 {
		a.Genre = b.Genre
	}

	if a.Duration != b.Duration && a.Duration == 0 {
		a.Duration = b.Duration
	}

	if a.CoverBufer == nil {
		a.CoverBufer = b.CoverBufer
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

type AudioProber interface {
	// Probe Only returns nil if there was an error.
	ProbeAudio(file string) (*AudioMeta, error)

	fmt.Stringer
}

type prober struct {
	Mime    *regexp.Regexp
	Probers []AudioProber
}

var probers = []*prober{}

func Initialize() {
	id3Prober := &id3AudioProber{}
	_ = id3Prober
	genericTagProber := &genericTagAudioProber{}
	ffProber := &ffprobeAudioProber{}

	mp3Probers := []AudioProber{}
	flacProbers := []AudioProber{}

	mp3Probers = append(mp3Probers, genericTagProber)

	if ffProber.hasFFprobe() {
		mp3Probers = append(mp3Probers, ffProber)
		flacProbers = append(flacProbers, ffProber)
	}

	probers = append(probers, &prober{
		Mime:    regexp.MustCompile("audio/(mpeg|mp3)"),
		Probers: mp3Probers,
	})

	probers = append(probers, &prober{
		Mime:    regexp.MustCompile("audio/flac"),
		Probers: flacProbers,
	})

	for _, prober := range probers {
		log.Infof("Registered probes for '%s': %s", prober.Mime, prober.Probers)
	}
}

func HasProber(mime string) bool {
	for _, prober := range probers {
		if prober.Mime.MatchString(mime) {
			return true
		}
	}

	return false
}

// ProbeAudioFile Probes a file and tries to fill in AudioMeta.
// meta is not nil if there was no error.
func ProbeAudioFile(file string) (meta *AudioMeta, err error) {
	log.Debugln(file)
	ext := filepath.Ext(file)
	mime := mime.TypeByExtension(ext)
	if len(mime) == 0 {
		var fh *os.File
		fh, err = os.Open(file)
		if err != nil {
			return
		}
		defer fh.Close()

		b := make([]byte, 512)
		_, err = fh.Read(b)
		if err != nil {
			return
		}

		// Always returns a valid MIME.
		mime = http.DetectContentType(b)
	}

	meta = &AudioMeta{}
	done := false
	foundProber := false
	for _, prober := range probers {
		if prober.Mime.MatchString(mime) {
			for _, ap := range prober.Probers {
				log.WithFields(log.Fields{"mime": mime, "file": file}).Debugf("Using prober '%s'.", ap)
				foundProber = true
				tmpMeta, err := ap.ProbeAudio(file)
				if err != nil {
					continue
				}

				meta.Merge(tmpMeta)
				if meta.IsComplete() {
					done = true
					break
				}
			}

			if done {
				break
			}
		}
	}

	if !foundProber {
		return nil, errors.New("No prober found for '" + mime + "'")
	}

	return
}
