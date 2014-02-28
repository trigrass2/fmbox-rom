
package main

import (
	"log"
	"fmt"
	"github.com/go-av/lush/m"
)

func DispSongLoad(s m.M) {
	log.Println("SongLoad:", s.S("title"), "-", s.S("artist"), s.S("url"))
	DispShowLike(s.I("like") == 1)

	if modeFmBox {
		sec := s.I64("length")
		tmstr := fmt.Sprintf("%.2d:%.2d", sec/60, sec%60)
		ops := []*OledOp{
			&OledOp{Str:s.S("artist"), Scroll:true, Y:0},
			&OledOp{Str:"<"+s.S("albumtitle")+">", Scroll:true, Y:1},
			&OledOp{Str:s.S("title"), Scroll:true, Y:2},
			&OledOp{Str:tmstr, Y:3},
		}
		oled.Do(ops)
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

