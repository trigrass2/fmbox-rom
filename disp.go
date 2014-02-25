
package main

import (
	"time"
	"log"
	"fmt"
	"github.com/go-av/lush/m"
)

func DispSongLoad(s m.M) {
	dur := time.Duration(s.I64("length"))*time.Second
	log.Println("SongLoad:", s.S("title"), "-", s.S("artist"), dur, s.S("url"))
	DispShowLike(s.I("like") == 1)

	if modeFmBox {
		oled.Text(0, 0, s.S("artist"))
		oled.Text(0, 1, s.S("title"))
		oled.Text(0, 2, s.S("album"))
		sec := s.I("len")
		oled.Text(0, 3, fmt.Sprintf("%.2d:%.2d", sec/60, sec%60))
	}
}

func DispShowLike(like bool) {
	if modeFmBox {
		if like {
			LedSet(1.0)
		} else {
			LedSet(0.0)
		}
	}
}

