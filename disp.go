
package main

import (
	"time"
	"log"
	"fmt"
	"github.com/go-av/lush/m"
	"gitcafe.com/nuomi-studio/fmbox-rom.git/client-mad"
)

func DispSongLoad(s m.M) {
	curChan := fm.CurChanInfo()
	log.Println("SongLoad:", s.S("title"), "-", s.S("artist"), s.S("url"))
	log.Println("  Chan:", curChan)
	DispShowLike(s.I("like") == 1)

	if modeFmBox {
		tot := s.I64("length")

		tmstr := func (sec int64) string {
			if sec < 0 {
				sec = 0
			}
			return fmt.Sprintf("%.2d:%.2d", sec/60, sec%60)
		}

		strcb := func () string {
			if mad.PlayStart.IsZero() {
				return tmstr(tot)
			} else {
				return tmstr(tot - int64(time.Now().Sub(mad.PlayStart)/time.Second))
			}
		}

		ops := []*OledOp{
			&OledOp{Str:s.S("artist"), Scroll:true, Y:0},
			&OledOp{Str:"<"+s.S("albumtitle")+">", Scroll:true, Y:1},
			&OledOp{Str:s.S("title"), Scroll:true, Y:2},
			&OledOp{StrCb:strcb, Y:3},
			&OledOp{Str:curChan.S("name"), Y:3, X:70, Scroll:true, Inverse:true},
		}
		oled.SetOps(ops)
	}
}

func DispShowLike(like bool) {
	if modeFmBox {
		if like {
			LedSet(ledBright)
		} else {
			LedSet(0.0)
		}
	}
}

