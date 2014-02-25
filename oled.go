
package main

import (
	"github.com/go-av/a10/mmap-gpio"
	"gitcafe.com/nuomi-studio/lcdbuf.git"
	"gitcafe.com/nuomi-studio/pcf.git"
	"time"
	"fmt"
	"log"
	"sync"
)

type Oled struct {
	sclk, sdin, rst, dc gpio.Pin
	buf *lcdbuf.Buf
	font *pcf.File
	l *sync.Mutex
}

func (o *Oled) Init() {
	o.sclk.H()
	o.rst.L()
	time.Sleep(time.Millisecond*50)
	o.rst.H()

	o.cmd(0xae)//--turn off oled panel
	o.cmd(0x00)//---set low column address
	o.cmd(0x10)//---set high column address
	o.cmd(0x40)//--set start line address  Set Mapping RAM Display Start Line (0x00~0x3F)
	o.cmd(0x81)//--set contrast control register
	o.cmd(0xcf) // Set SEG Output Current Brightness
	o.cmd(0xa0)//--Set SEG/Column Mapping     0xa0左右反置 0xa1正常
	o.cmd(0xc0)//Set COM/Row Scan Direction   0xc0上下反置 0xc8正常
	o.cmd(0xa6)//--set display: a6 normal a7 inverse
	o.cmd(0xa8)//--set multiplex ratio(1 to 64)
	o.cmd(0x3f)//--1/64 duty
	o.cmd(0xd3)//-set display offset	Shift Mapping RAM Counter (0x00~0x3F)
	o.cmd(0x00)//-not offset
	o.cmd(0xd5)//--set display clock divide ratio/oscillator frequency
	o.cmd(0x80)//--set divide ratio, Set Clock as 100 Frames/Sec
	o.cmd(0xd9)//--set pre-charge period
	o.cmd(0xf1)//Set Pre-Charge as 15 Clocks & Discharge as 1 Clock
	o.cmd(0xda)//--set com pins hardware configuration
	o.cmd(0x12)
	o.cmd(0xdb)//--set vcomh
	o.cmd(0x40)//Set VCOM Deselect Level
	o.cmd(0x20)//-Set Page Addressing Mode (0x00/0x01/0x02)
	o.cmd(0x02)//
	o.cmd(0x8d)//--set Charge Pump enable/disable
	o.cmd(0x14)//--set(0x10) disable
	o.cmd(0xa4)// Disable Entire Display On (0xa4/0xa5)
	o.cmd(0xaf)//--turn on oled panel

	o.fill(0x00)
	o.pos(0,0)

	log.Println("oled:", "init")
}

func (o *Oled) wr(b byte) {
	o.sclk.L()
	for i := 0; i < 8; i++ {
		if b & 0x80 != 0 {
			o.sdin.H()
		} else {
			o.sdin.L()
		}
		o.sclk.H()
		o.sclk.L()
		b <<= 1
	}
}

func (o *Oled) dat(b byte) {
	o.dc.H()
	o.wr(b)
}

func (o *Oled) cmd(b byte) {
	o.dc.L()
	o.wr(b)
}

func (o *Oled) pos(x, y byte) {
	o.cmd(0xb0+y)
	o.cmd((( x & 0xf0 ) >> 4) | 0x10)
	o.cmd(( x & 0x0f ) | 0x01)
}

func (o *Oled) fill(b byte) {
	var x, y byte
	for y = 0; y < 8; y++ {
		o.cmd(0xb0+y)
		o.cmd(0x01)
		o.cmd(0x10)
		for x = 0; x < 128; x++ {
			o.dat(b)
		}
	}
}

func (o *Oled) Text(x, y int, str string) {
	o.l.Lock()
	lcdbuf.Draw(o.buf, lcdbuf.PCFText(o.font, str), x, y*16, o.pos, o.dat)
	o.l.Unlock()
}

func (o *Oled) StopScroll(y int) {
}

type scroller struct {
	ch chan bool
}

func (s scroller) Stop() {
	s.ch <- false
}

func (o *Oled) ScrollText(y int, str string) *scroller {
	s := &scroller{ch: make(chan bool)}
	tbuf := lcdbuf.PCFText(o.font, str)
	go func () {
		wait := func (t int) bool {
			select {
			case <-s.ch:
				return true
			case <-time.After(time.Millisecond*time.Duration(t)):
			}
			return false
		}
		draw := func (i int) {
			o.l.Lock()
			lcdbuf.DrawOffset(o.buf, tbuf, 0, y*16, i, o.pos, o.dat)
			o.l.Unlock()
		}
		w := tbuf.W - o.buf.W
		log.Println(w)
		for {
			for i := 0; i < w; i++ {
				draw(i)
				if wait(1) {
					return
				}
			}
			if wait(15) {
				return
			}
			for i := w-1; i >= 0; i-- {
				draw(i)
				if wait(1) {
					return
				}
			}
			if wait(15) {
				return
			}
		}
	}()
	return s
}

func NewOled() *Oled {
	o := &Oled{
		sclk: gpio.Open(8, 4, gpio.Out), //PI4
		sdin: gpio.Open(8, 5, gpio.Out), //PI5
		rst: gpio.Open(8, 6, gpio.Out), //PI6
		dc: gpio.Open(8, 7, gpio.Out), //PI7
		buf: lcdbuf.New(128, 64),
		l: &sync.Mutex{},
	}
	o.font, _ = pcf.Open("13px.pcf")
	return o
}

func (o *Oled) Test(s string) {
	switch s {
	case "pattern":
		o.pos(3, 0)
		for i := 0; i < 128-3; i++ {
			o.dat(1<<byte(i%8))
		}
	case "text1":
		o.Text(0, 0, "测试哈哈")
		log.Println("draw text end")
	case "text2":
		o.Text(0, 0, "PinkFloyd")
		o.Text(0, 1, "Another Brick in the Wall, Pt. 2")
		o.Text(0, 2, "<The Wall>")
		o.Text(0, 3, "00:12")
		log.Println("draw text end")
	case "scroll":
		log.Println("start scrolling")
		a := o.ScrollText(0, "Another Brick in the Wall, Pt. 2")
		b := o.ScrollText(1, "<The Wall> is an great album hehehe")
		c := o.ScrollText(2, "00:12 12231923812903812903810928390218")
		for {
			if false {
				n := time.Now()
				tmstr := fmt.Sprintf("%.2d:%.2d:%.2d", n.Hour(), n.Minute(), n.Second())
				o.Text(0, 3, tmstr)
			}
			time.Sleep(time.Second)
		}
		a.Stop()
		b.Stop()
		c.Stop()

	case "speed":
		for n := 0; ; n++ {
			for j := 0; j < 4; j++ {
				lcdbuf.Draw(o.buf, lcdbuf.PCFText(o.font, fmt.Sprintf("%.6d", n+j)), 6, j*16, nil, nil)
			}
			for i := 0; i < 8; i++ {
				o.pos(0, byte(i))
				for j := 0; j < 128; j++ {
					o.dat(o.buf.Pix[i*128+j])
				}
			}
		}
	}
}

