package probers

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"
	"strconv"
)

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
