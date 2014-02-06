
package main

import (
	"time"
	"log"
	"github.com/go-av/lush/m"
)

type Disp struct {
}

func (d Disp) SongLoad(s m.M) {
	dur := time.Duration(s.I64("length"))*time.Second
	log.Println("SongLoad:", s.S("title"), "-", s.S("artist"), dur, s.S("url"))
	d.ShowSongLike(s)
}

func (d Disp) ShowSongLike(s m.M) {
	if modeFmBox {
		if s.I("like") > 0 {
			LedSet(1.0)
		} else {
			LedSet(0)
		}
	}
}

func (d Disp) ToggleSongLike(s m.M) {
	if s.I("like") > 0 {
		delete(s, "like")
	} else {
		s["like"] = 1
	}
	d.ShowSongLike(s)
}

