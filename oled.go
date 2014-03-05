
package main

import (
	"github.com/go-av/a10/mmap-gpio"
	"github.com/go-av/lcdbuf"
	"github.com/go-av/pcf"
	"time"
	"log"
	"sync"
)

type OledOp struct {
	StrCb func () string
	Str string
	Hide bool
	Inverse bool
	X, Y, W int
	Scroll bool
	Buf *lcdbuf.Buf
	tm int
}

type Oled struct {
	buf *lcdbuf.Buf
	font *pcf.File
	l *sync.Mutex
	ops []*OledOp
}

func (o *Oled) initOp(op *OledOp) {
	if op.StrCb != nil {
		op.Str = op.StrCb()
	}
	if op.Buf == nil {
		op.Buf = lcdbuf.PCFText(o.font, op.Str)
	}
	if op.Inverse {
		op.Buf.Inverse()
	}
	op.tm = 0
	if op.Scroll && op.X+op.Buf.W < o.buf.W {
		op.Scroll = false
	}
}

func (o *Oled) SetOp(i int, op *OledOp) {
	o.l.Lock()
	o.ops[i] = op
	o.initOp(o.ops[i])
	o.l.Unlock()
}

func (o *Oled) SetOps(ops []*OledOp) {
	o.l.Lock()
	o.ops = ops
	for _, op := range ops {
		o.initOp(op)
	}
	o.l.Unlock()
}

func (o *Oled) getScrollOffset(op *OledOp) int {
	sw := op.X+op.Buf.W - o.buf.W
	tm := op.tm
	pause := 20

	// paused at left
	if tm < pause {
		return 0
	}
	tm -= pause

	// move right
	if tm < sw {
		return tm
	}
	tm -= sw

	// paused at right
	if tm < pause {
		return sw
	}
	tm -= pause

	// move left
	if tm < sw {
		return sw - tm
	}

	op.tm = 0
	return 0
}

func (o *Oled) Update() {
	o.l.Lock()
	o.buf.Clear()
	for _, op := range o.ops {
		if op.StrCb != nil {
			str := op.StrCb()
			if op.Str != str {
				op.Str = str
				op.Buf = lcdbuf.PCFText(o.font, op.Str)
			}
		}
		if !op.Hide {
			if op.Scroll {
				off := o.getScrollOffset(op)
				lcdbuf.DrawOffset(o.buf, op.Buf, op.X, op.Y*16, off)
			} else {
				lcdbuf.Draw(o.buf, op.Buf, op.X, op.Y*16)
			}
		}
		op.tm++
	}
	gpio.Call5(o.buf.Pix)
	o.l.Unlock()
}

func (o *Oled) StartUpdateThread(fps int) {
	go func () {
		for {
			o.Update()
			time.Sleep(time.Second/time.Duration(fps))
		}
	}()
}

func NewOled() *Oled {
	o := &Oled{
		buf: lcdbuf.New(128, 64),
		l: &sync.Mutex{},
	}
	o.font, _ = pcf.Open("/etc/13px.pcf")
	return o
}

func (o *Oled) Test(s string) {
	switch s {
	case "scr":
		b := make([]byte, 128*8)
		b[33] = 0x23
		b[44] = 0x11
		gpio.Call5(b)

	case "maxfps":
		i := 0
		go func () {
			for {
				b := make([]byte, 128*8)
				b[33] = 0x23
				b[44] = 0x11
				gpio.Call5(b)
				i++
			}
		}()
		for {
			time.Sleep(time.Second)
			log.Println("fps", i)
			i = 0
		}

	case "scroll":
		log.Println("start scrolling")
		o.StartUpdateThread(24)
		ops := []*OledOp{
			&OledOp{Str:"Another Brick in the Wall, Pt. 2", Scroll:true, Y:0},
			&OledOp{Str:"<The Wall> 1234566789900--87555", Scroll:true, Y:1},
			&OledOp{Str:"00:12 12231923812903812903810928390218", Scroll:true, Y:2},
		}
		o.SetOps(ops)
		for {
			time.Sleep(time.Second)
		}
	}
}

