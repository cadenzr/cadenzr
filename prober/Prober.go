package prober

import (
	"encoding/json"
	"os/exec"
	"strconv"
)

type AudioMeta struct {
	Title       string
	Artist      string
	Album       string
	Track       int
	TotalTracks int
	Year        int
	AlbumArtist string
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

type AudioProber interface {
	// Probe Only returns nil if there was an error.
	ProbeAudio(file string) (*AudioMeta, error)
}

type ffAudioProber struct{}

func (f *ffAudioProber) ProbeAudio(file string) (meta *AudioMeta, err error) {
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

	return
}

// ProbeAudioFile Probes a file and tries to fill in AudioMeta.
// meta is not nil if there was no error.
func ProbeAudioFile(file string) (meta *AudioMeta, err error) {
	ffprober := &ffAudioProber{}
	return ffprober.ProbeAudio(file)
}
