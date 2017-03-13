package probers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	log "github.com/cadenzr/cadenzr/log"
	id3 "github.com/mikkyang/id3-go"
	id3v2 "github.com/mikkyang/id3-go/v2"
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

	CoverBufer []byte
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

type ffprobeAudioProber struct{}

func (f *ffprobeAudioProber) getCover(file string, meta *AudioMeta) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return
	}

	cmd := exec.Command(ffmpeg,
		"-i", file,
		"-f", "image2",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	buf, err := ioutil.ReadAll(stdout)
	if err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		return
	}

	meta.CoverBufer = buf
}

func (p *ffprobeAudioProber) ProbeAudio(file string) (meta *AudioMeta, err error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return
	}

	cmd := exec.Command(ffprobe,
		"-print_format", "json",
		"-show_entries", "format_tags",
		file,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	response := struct {
		Format struct {
			Tags struct {
				Title       string `json:"title"`
				Artist      string `json:"artist"`
				Album       string `json:"album"`
				Track       string `json:"track"`
				TotalTracks string `json:"totaltracks"`
				Date        string `json:"date"`
				AlbumArtist string `json:"album_artist"`
			}
		}
	}{}

	if err = json.NewDecoder(stdout).Decode(&response); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		return
	}

	meta = &AudioMeta{}
	meta.Album = response.Format.Tags.Album
	meta.Title = response.Format.Tags.Title
	meta.Artist = response.Format.Tags.Artist
	meta.Album = response.Format.Tags.Album
	meta.Track = parseInt(response.Format.Tags.Track)
	meta.TotalTracks = parseInt(response.Format.Tags.TotalTracks)
	meta.Year = parseInt(response.Format.Tags.Date)
	meta.AlbumArtist = response.Format.Tags.AlbumArtist

	p.getCover(file, meta)

	return
}

func (p *ffprobeAudioProber) hasFFprobe() bool {
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return false
	}

	return true
}

func (p *ffprobeAudioProber) String() string {
	return "ffprobeAudioProber"
}

type id3AudioProber struct{}

func (p *id3AudioProber) cleanString(s string) string {
	// Something with macos?
	return string(bytes.Trim([]byte(s), "\x00"))
}

func (p *id3AudioProber) ProbeAudio(file string) (meta *AudioMeta, err error) {
	id3f, err := id3.Open(file)
	if err != nil {
		return
	}
	defer id3f.Close()

	meta = &AudioMeta{}
	meta.Title = p.cleanString(id3f.Title())
	meta.Year = parseInt(p.cleanString(id3f.Year()))
	meta.Genre = p.cleanString(id3f.Genre())
	meta.Artist = p.cleanString(id3f.Artist())
	meta.Album = p.cleanString(id3f.Album())

	if apic := id3f.Frame("APIC"); apic != nil {
		meta.CoverBufer = apic.(*id3v2.ImageFrame).Data()
	}

	return
}

func (p *id3AudioProber) String() string {
	return "id3Prober"
}

type prober struct {
	Mime    *regexp.Regexp
	Probers []AudioProber
}

var probers = []*prober{}

func Initialize() {
	id3Prober := &id3AudioProber{}
	ffProber := &ffprobeAudioProber{}

	probers = append(probers, &prober{
		Mime: regexp.MustCompile("audio/(mpeg|mp3)"),
		Probers: []AudioProber{
			id3Prober,
		},
	})

	if ffProber.hasFFprobe() {
		probers = append(probers, &prober{
			Mime: regexp.MustCompile("audio/.*"),
			Probers: []AudioProber{
				ffProber,
			},
		})
	}

	for _, prober := range probers {
		log.Infof("Registered probes for '%s': %s", prober.Mime, prober.Probers)
	}

}

// ProbeAudioFile Probes a file and tries to fill in AudioMeta.
// meta is not nil if there was no error.
func ProbeAudioFile(file string) (meta *AudioMeta, err error) {
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

	hasProber := false
	for _, prober := range probers {
		if prober.Mime.MatchString(mime) {
			for _, ap := range prober.Probers {
				meta, err = ap.ProbeAudio(file)
				// Just use first prober that works.
				// TODO: Maybe we can use next prober if this one failed.
				log.WithFields(log.Fields{"mime": mime, "file": file}).Debugf("Using prober '%s'.", ap)
				hasProber = true
				break
			}

			if hasProber {
				break
			}
		}
	}

	if !hasProber {
		log.WithFields(log.Fields{"file": file, "mime": mime}).Info("No prober found for file.")
		return meta, errors.New("No prober found for mime `" + mime + "`.")
	}

	return
}
