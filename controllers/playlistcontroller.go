package controllers

import (
	"net/http"
	"strings"

	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/log"
	"github.com/cadenzr/cadenzr/models"
	"github.com/labstack/echo"
)

type playlistResponse struct {
	ID    uint            `json:"id"`
	Name  string          `json:"name"`
	Songs []*songResponse `json:"songs"`
}

func TransformPlaylist(playlist *models.Playlist) *playlistResponse {
	r := &playlistResponse{}
	r.ID = playlist.ID
	r.Name = playlist.Name

	if playlist.Songs != nil {
		r.Songs = TransformSongs(playlist.Songs...)
	}

	return r
}

func TransformPlaylists(playlists ...*models.Playlist) []*playlistResponse {
	r := []*playlistResponse{}

	for _, playlist := range playlists {
		r = append(r, TransformPlaylist(playlist))
	}

	return r
}

type playlistController struct {
}

func (c *playlistController) Index(ctx echo.Context) error {
	playlists := []*models.Playlist{}
	if gormDB := db.DB.Preload("Songs").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").Find(&playlists); gormDB.Error != nil {
		log.Errorf("PlaylistController::Index Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, echo.Map{
		"data": TransformPlaylists(playlists...),
	})
}

func (c *playlistController) Show(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))

	playlist := &models.Playlist{}
	gormDB := db.DB.Preload("Songs").Preload("Songs.Album").Preload("Songs.Artist").Preload("Songs.Cover").Where("id = ?", id).Find(&playlist)

	if gormDB.RecordNotFound() {
		log.Debugf("PlaylistController::Show Playlist '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	} else if gormDB.Error != nil {
		log.Errorf("PlaylistController::Show Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, TransformPlaylist(playlist))
}

func (c *playlistController) Create(ctx echo.Context) error {
	params := &struct {
		Name string `json:"name" form:"name"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Debugf("PlaylistController::Create Binding params failed: %v", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	params.Name = strings.TrimSpace(params.Name)
	if len(params.Name) == 0 {
		log.Errorf("PlaylistController::Create Playlist name too short: '%s'", params.Name)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var count uint64
	if gormDB := db.DB.Table("playlists").Where("name = ?", params.Name).Count(&count); gormDB.Error != nil {
		log.Errorf("PlaylistController::Create Checking if playlist already exists failed. %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if count != 0 {
		log.Debugf("PlaylistController::Create Playlist '%s' already exists.", params.Name)
		return ctx.NoContent(http.StatusUnauthorized)
	}

	playlist := &models.Playlist{
		Name: params.Name,
	}

	gormDB := db.DB.Create(playlist)
	if gormDB.Error != nil {
		log.Errorf("PlaylistController::Create Creating new playlist failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.WithFields(log.Fields{"id": playlist.ID, "name": playlist.Name}).Info("New playlist created.")
	return ctx.JSON(http.StatusCreated, TransformPlaylist(playlist))
}

func (c *playlistController) Delete(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))
	gormDB := db.DB.Unscoped().Delete(models.Playlist{}, "id = ?", id)
	if gormDB.Error != nil {
		log.Errorf("PlaylistController::Delete Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if gormDB.RowsAffected == 0 {
		log.Debugf("PlaylistController::Delete Playlist '%d' not found.", id)
		return ctx.NoContent(http.StatusNotFound)
	}

	log.WithFields(log.Fields{"id": id}).Info("Deleted playlist.")
	return ctx.NoContent(http.StatusOK)
}

func (c *playlistController) AddSongs(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))
	params := struct {
		Sids []uint `json:"songs" form:"songs[]"`
	}{
		Sids: []uint{},
	}

	if err := ctx.Bind(&params); err != nil {
		log.Debugf("PlaylistController::AddSongs Binding params failed: %v", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		log.Errorf("PlaylistController::AddSongs Could not start transaction: %v", tx.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	for _, sid := range params.Sids {
		if gormDB := tx.Exec("INSERT INTO playlist_songs (playlist_id, song_id) VALUES(?,?)", id, sid); gormDB.Error != nil {
			tx.Rollback()
			log.Errorf("PlaylistController::AddSongs Could not insert song into playlist: %v", gormDB.Error)
			return ctx.NoContent(http.StatusInternalServerError)
		}
	}

	if gormDB := tx.Commit(); gormDB.Error != nil {
		tx.Rollback()
		log.Errorf("PlaylistController::AddSongs Could not commits song into playlist: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.WithFields(log.Fields{"playlist": id, "songs": params.Sids}).Info("Added songs to playlist.")
	return ctx.NoContent(http.StatusOK)
}

func (c *playlistController) DeleteSong(ctx echo.Context) error {
	id := StrToUint(ctx.Param("id"))
	sid := StrToUint(ctx.Param("sid"))

	gormDB := db.DB.Exec("DELETE FROM playlist_songs WHERE playlist_id = ? AND song_id = ?", id, sid)
	if gormDB.Error != nil {
		log.Errorf("PlaylistController::DeleteSong Database failed: %v", gormDB.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	} else if gormDB.RowsAffected == 0 {
		log.Debugf("PlaylistController::DeleteSong Song '%d' not found in playlist %d.", sid, id)
		return ctx.NoContent(http.StatusNotFound)
	}

	log.WithFields(log.Fields{"playlist": id, "song": sid}).Info("Deleted song from playlist.")
	return ctx.NoContent(http.StatusOK)
}

// PlaylistController Contains the actions for the 'playlist' endpoint.
var PlaylistController playlistController
