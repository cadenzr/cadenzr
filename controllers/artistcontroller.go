package controllers

import "github.com/labstack/echo"

type artistController struct {
}

func (c *artistController) Index(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Show(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Create(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Update(echo.Context) error {
	panic("Not implemented")
}

func (c *artistController) Delete(echo.Context) error {
	panic("Not implemented")
}

// ArtistController Contains the actions for the 'artists' endpoint.
var ArtistController artistController
