package probers

import(
    "testing"
    . "github.com/smartystreets/goconvey/convey"
)

func TestProbers(t *testing.T) {

    Convey("Demo audio file", t, func() {
        //Demo audio file: '../media/demo/Curse the Day.mp3' downloaded from Jamendo
        
        Initialize()
        
        meta, err := ProbeAudioFile("../media/demo/Curse the Day.mp3")
        So(err, ShouldEqual, nil)
        
        So(meta.Title, ShouldEqual, "Curse the Day (Radio Edit)")
        So(meta.Artist, ShouldEqual, "Brain Purist")
        So(meta.Album, ShouldEqual, "Curse the Day (Single versions)")
        So(meta.Year, ShouldEqual, 2016)
        //So(meta.Track, ShouldEqual, 1)
        So(meta.Genre, ShouldEqual, "Pop")
        
        
    })

}