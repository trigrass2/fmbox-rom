
package main

import (
	"runtime"
	"time"
	"log"
	"github.com/go-av/lush/m"
)

type LcdDisp struct {
}

func (ld LcdDisp) SongLoad(s m.M) {
	dur := time.Duration(s.I64("length"))*time.Second
	log.Println("SongLoad", s.S("title"), "-", s.S("artist"), dur)
	if runtime.GOARCH == "arm" {
	}
}

func (ld LcdDisp) SongStart() {
	log.Println("SongStart")
}

func (ld LcdDisp) SongPos(pos time.Duration) {
	//log.Println("SongPos", pos)
}

func (ld LcdDisp) SongEnd() {
	log.Println("SongEnd")
}

func (ld LcdDisp) ShowSongLike(s m.M) {
	if runtime.GOARCH == "arm" {
		if s.I("like") > 0 {
			LedSet(1.0)
		} else {
			LedSet(0)
		}
	}
}

func (ld LcdDisp) ToggleSongLike(s m.M) {
}

