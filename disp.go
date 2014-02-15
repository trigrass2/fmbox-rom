
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
	d.ShowLike(s.I("like") == 1)
}

func (d Disp) ShowLike(like bool) {
	if modeFmBox {
		if like {
			LedSet(1.0)
		} else {
			LedSet(0.0)
		}
	}
}

