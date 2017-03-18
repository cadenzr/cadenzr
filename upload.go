package main

import (
	"io"
	"os"

	"net/http"

	"github.com/cadenzr/cadenzr/log"
	"github.com/labstack/echo"
)

func upload(c echo.Context) error {

	//-----------
	// Read file
	//-----------

	// Source
	file, err := c.FormFile("file")
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Couldn't find upload file.")
		return c.NoContent(http.StatusInternalServerError)
	}
	src, err := file.Open()
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Couldn't open upload file.")
		return c.NoContent(http.StatusInternalServerError)
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("media/uploads/" + file.Filename)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Couldn't upload file.")
		return c.NoContent(http.StatusInternalServerError)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.WithFields(log.Fields{"reason": err.Error()}).Error("Couldn't copy uploaded file.")
		return c.NoContent(http.StatusInternalServerError)
	}

	scanFilesystem("media/uploads/" + file.Filename)

	return c.NoContent(http.StatusOK)
}
