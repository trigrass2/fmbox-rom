
package main

import (
	"github.com/go-av/a10/mmap-gpio"
	"gitcafe.com/nuomi-studio/lcdbuf.git"
	"gitcafe.com/nuomi-studio/pcf.git"
	"time"
	"log"
	"sync"
)

type OledOp struct {
	Str string
	X, Y int
	Scroll bool
	buf *lcdbuf.Buf
	tm int
}

type Oled struct {
	buf *lcdbuf.Buf
	font *pcf.File
	l *sync.Mutex
	ops []*OledOp
}

func (o *Oled) Do(ops []*OledOp) {
	o.l.Lock()
	o.ops = ops
	for _, op := range ops {
		op.buf = lcdbuf.PCFText(o.font, op.Str)
		op.tm = 0
		if op.buf.W < o.buf.W {
			op.Scroll = false
		}
	}
	o.l.Unlock()
}

func (o *Oled) getScrollOffset(op *OledOp) int {
	sw := op.buf.W - o.buf.W
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
		if op.Scroll {
			off := o.getScrollOffset(op)
			lcdbuf.DrawOffset(o.buf, op.buf, 0, op.Y*16, off, nil, nil)
		} else {
			lcdbuf.Draw(o.buf, op.buf, 0, op.Y*16, nil, nil)
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
	o.font, _ = pcf.Open("13px.pcf")
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
		o.Do(ops)
		for {
			time.Sleep(time.Second)
		}
	}
}

