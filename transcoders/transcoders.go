package transcoders

import (
	"io"
	"os/exec"

	"github.com/Cadenzr/Cadenzr/log"
)

type CodecType uint8

func (c *CodecType) String() string {
	switch *c {
	case MP3:
		return "mp3"
	case VORBIS:
		return "vorbis"
	default:
		panic("Unknown codec type")
	}
}

func (c *CodecType) Extension() string {
	switch *c {
	case MP3:
		return ".mp3"
	case VORBIS:
		return ".ogg"
	default:
		panic("Unknown codec type")
	}
}

const (
	MP3 CodecType = iota
	VORBIS
)

type Transcoder struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (t *Transcoder) Read(p []byte) (n int, err error) {
	n, err = t.stdout.Read(p)
	return
}

func NewTranscoder(input io.Reader, codec CodecType) (io.Reader, error) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, err
	}

	codecString := ""
	formatString := ""
	switch codec {
	case MP3:
		codecString = "mp3"
		formatString = "mp3"
	case VORBIS:
		codecString = "libvorbis"
		formatString = "ogg"
	default:
		panic("Unknown codec type")
	}

	cmd := exec.Command(ffmpeg,
		"-i", "-",
		"-codec:a", codecString,
		"-f", formatString,
		"pipe:1",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	transcoder := &Transcoder{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}

	go func() {
		io.Copy(transcoder.stdin, input)
		transcoder.stdin.Close()

		if err := cmd.Wait(); err != nil {
			log.WithFields(log.Fields{"reason": err}).Error("Wait failed.")
		}
	}()

	return transcoder, nil
}
