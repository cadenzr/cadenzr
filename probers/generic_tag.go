package probers

import (
	"os"

	"github.com/badgerodon/mp3"
	"github.com/dhowden/tag"
)

type genericTagAudioProber struct{}

func (p *genericTagAudioProber) ProbeAudio(file string) (meta *AudioMeta, err error) {

	f, err := os.Open(file) // For read access.
	if err != nil {
		return
	}

	m, err := tag.ReadFrom(f)
	if err != nil {
		return
	}

	meta = &AudioMeta{}

	meta.Title = m.Title()
	meta.Artist = m.Artist()
	meta.Album = m.Album()
	meta.Year = m.Year()
	meta.Genre = m.Genre()
	meta.Track, meta.TotalTracks = m.Track()

	meta.CoverBufer = m.Picture().Data

	d, err := mp3.Length(f)
	if err == nil {
		defer f.Close()
		meta.Duration = d.Seconds()
	}

	return
}

func (p *genericTagAudioProber) String() string {
	return "genericTagAudioProber"
}
