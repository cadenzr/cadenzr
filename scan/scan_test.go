package scan

import (
	"github.com/cadenzr/cadenzr/db"
	"github.com/cadenzr/cadenzr/models"
	"github.com/cadenzr/cadenzr/probers"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestScan(t *testing.T) {

	Convey("Scan dir", t, func() {
		//Demo audio file: '../media/demo/Curse the Day.mp3' downloaded from Jamendo

		if err := db.SetupConnection(db.SQLITE, "file::memory:?mode=memory&cache=shared"); err != nil {
			So(err, ShouldBeNil)
		}

		if err := db.SetupSchema(); err != nil {
			So(err, ShouldBeNil)
		}

		probers.Initialize()
		go ScanFilesystem("../media/0demo/")
		<-ScanDone

		song := &models.Song{}
		gormDB := db.DB.First(&song, "id = ?", 1)
		So(gormDB.RecordNotFound(), ShouldBeFalse)
		So(gormDB.Error, ShouldBeNil)
		So(song.Name, ShouldEqual, "Curse the Day (Radio Edit)")
		So(song.Year.Int64, ShouldEqual, 2016)
		So(song.Track.Int64, ShouldEqual, 1)
		So(song.TotalTracks.Int64, ShouldEqual, 42)
	})

}
