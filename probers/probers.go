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

	"github.com/badgerodon/mp3"
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
		"-show_entries", "format=duration:format_tags",
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
			Duration string `json:"duration"`
			Tags     struct {
				Title       string `json:"title"`
				Artist      string `json:"artist"`
				Album       string `json:"album"`
				Genre       string `json:"genre"`
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
	meta.Genre = response.Format.Tags.Genre
	meta.Album = response.Format.Tags.Album
	meta.Track = parseInt(response.Format.Tags.Track)
	meta.TotalTracks = parseInt(response.Format.Tags.TotalTracks)
	meta.Year = parseInt(response.Format.Tags.Date)
	meta.AlbumArtist = response.Format.Tags.AlbumArtist
	meta.Duration, _ = strconv.ParseFloat(response.Format.Duration, 64)

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

	f, err := os.Open(file)
	if err != nil {
		return
	}
	d, err := mp3.Length(f)
	if err == nil {
		defer f.Close()
		meta.Duration = d.Seconds()
	}

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

	mp3Probers := []AudioProber{}
	flacProbers := []AudioProber{}

	mp3Probers = append(mp3Probers, id3Prober)

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
