
package main

import (
	"time"
	"log"
	"fmt"
	"github.com/go-av/lush/m"
	"github.com/go-av/fmbox-rom/client-mad"
	"github.com/go-av/lcdbuf"
)

var btLogo = &lcdbuf.Buf{
  W: 9,
  H: 16,
  Pix: []byte{
    0x0,0x80,0x0,0x0,0xf0,0x20,0x40,0x80,0x0,
    0x0,0x20,0x11,0xe,0x7f,0x22,0x15,0x8,0x0,
  },
}

var wifiLogo = &lcdbuf.Buf{
  W: 12,
  H: 16,
  Pix: []byte{
    0x0,0x80,0xc0,0x40,0x60,0x20,0x20,0x60,0x40,0xc0,0x80,0x0,
    0x1,0x1,0xc,0x4,0x2,0x62,0x62,0x2,0x4,0xc,0x1,0x1,
  },
}

var btLedOp = &OledOp{Buf:btLogo, Y:3, X:58, Hide:true}
var wifiLedOp = &OledOp{Buf:wifiLogo, Y:3, X:74, Hide:true}

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
			&OledOp{Str:curChan.S("name"), Y:3, X:90, Scroll:true, Inverse:true},
			btLedOp, wifiLedOp,
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

