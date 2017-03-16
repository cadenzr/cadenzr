package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cadenzr/cadenzr/transcoders"
)

type Streamer interface {
	io.Closer
	io.ReadSeeker
}

type FileStreamer struct {
	f *os.File
}

func (s *FileStreamer) Read(p []byte) (n int, err error) {
	n, err = s.f.Read(p)
	return
}

func (s *FileStreamer) Seek(offset int64, whence int) (n int64, err error) {
	n, err = s.f.Seek(offset, whence)

	return
}

func (s *FileStreamer) Close() error {
	return s.f.Close()
}

func NewFileStreamer(path string) (streamer Streamer, err error) {
	fh, err := os.Open(path)
	if err != nil {
		return
	}

	streamer = &FileStreamer{
		f: fh,
	}

	return
}

func NewTranscodeStreamer(song *Song, codec transcoders.CodecType) (streamer Streamer, err error) {
	cachePath := "cache" + string(filepath.Separator) + "transcodings" + string(filepath.Separator) + song.Hash + codec.Extension()
	err = os.MkdirAll(filepath.Dir(cachePath), 0755)
	if err != nil {
		return
	}

	if _, err = os.Stat(cachePath); !os.IsNotExist(err) {
		return NewFileStreamer(cachePath)
	}

	originalFile, err := os.Open(song.Path)
	if err != nil {
		return
	}
	defer originalFile.Close()

	cacheFile, err := os.Create(cachePath)
	if err != nil {
		return
	}
	defer cacheFile.Close()

	transcoder, err := transcoders.NewTranscoder(originalFile, codec)
	if err != nil {
		return
	}

	_, err = io.Copy(cacheFile, transcoder)
	if err != nil {
		// So that it first gets closed?
		defer os.Remove(cachePath)
		return
	}

	streamer, err = NewFileStreamer(cachePath)
	return
}
