package probers

import (
	"bytes"
	"os"

	"github.com/badgerodon/mp3"
	id3 "github.com/mikkyang/id3-go"
	id3v2 "github.com/mikkyang/id3-go/v2"
)

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
