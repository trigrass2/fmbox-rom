
package main

import (
	"github.com/go-av/a10/mmap-gpio"
	"time"
	"log"
)

type Oled struct {
	sclk, sdin, rst, dc gpio.Pin
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
	o.cmd(0xa1)//--Set SEG/Column Mapping     0xa0左右反置 0xa1正常
	o.cmd(0xc8)//Set COM/Row Scan Direction   0xc0上下反置 0xc8正常
	o.cmd(0xa6)//--set normal display
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
	o.cmd(0xa6)// Disable Inverse Display On (0xa6/a7) 
	o.cmd(0xaf)//--turn on oled panel

	o.fill(0x00)  //初始清屏
	o.pos(0,0)

	o.fill(0xab)

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

func NewOled() *Oled {
	return &Oled{
		sclk: gpio.Open(8, 4, gpio.Out), //PI4
		sdin: gpio.Open(8, 5, gpio.Out), //PI5
		rst: gpio.Open(8, 6, gpio.Out), //PI6
		dc: gpio.Open(8, 7, gpio.Out), //PI7
	}
}

func OledTest() {
	o := NewOled()
	o.Init()
	b := byte(44)
	for {
		b += 31
		b |= 0x81
		o.fill(b)
		time.Sleep(time.Second)
	}
}

