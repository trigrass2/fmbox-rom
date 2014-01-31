
package main

import (
	"time"
	"log"
	"github.com/go-av/a10/mmap-gpio"
)

type Eink struct {
	BUS, SCLK, SDA, BUSY, CS, RST, DC gpio.Pin
}

func NewEink() *Eink {
	/* port: A=0 B=1 C=2 D=3 E=4 F=5 G=6 H=7 I=8 */
	// 8:BUS
	// 9:BUSY
	// 10:RST
	// 11:DC
	// 12:CS
	// 13:SCLK
	// 14:SDA
	e := &Eink{
		BUS: gpio.Open(8, 10, gpio.Out), // PI10 SPI0_CS0
		BUSY: gpio.Open(8, 5, gpio.In), // PI5
		RST: gpio.Open(8, 4, gpio.Out),
		DC: gpio.Open(8, 9, gpio.Out),
		CS: gpio.Open(8, 8, gpio.Out),
		SCLK: gpio.Open(8, 7, gpio.Out),
		SDA: gpio.Open(8, 6, gpio.Out),
	}
	return e
}

func (e *Eink) Test() {
	e.RST.H()
	return
	for {
		log.Println("busy:", e.BUSY.Read())
		time.Sleep(time.Second/1000)
	}
}

func (e *Eink) Ready() {
	i := 0
	for e.BUSY.Read() {
		time.Sleep(time.Second/1000)
		i++
	}
	log.Printf("ready(%d)\n", i)
}

func (e *Eink) Init() {
	e.RST.L()
	e.RST.H()
	e.BUS.L()
	e.CS.H()
	e.SCLK.H()

	log.Println("wait for hw reset")
	e.Ready()

	log.Println("set no deep sleep mode")
	e.Cmd(0x10) //set no deep sleep mode  
	e.Data(0x00)

	log.Println("data enter mode")
	e.Cmd(0x11) //data enter mode  
	e.Data(0x01)

	log.Println("set ram x addr")
	e.Cmd(0x44)//set RAM x address start/end  
	e.Data(0x00)//RAM x address start at 00h  
	e.Data(0x11)//RAM x address end at 11h(17)->72    

	log.Println("set ram y addr")
	e.Cmd(0x45)//set RAM y address start/end  
	e.Data(0xAB)//RAM y address start at 00h  
	e.Data(0x00)//RAM y address start at ABh(171)->172    

	log.Println("set ram x addr count")
	e.Cmd(0x4E)//set RAM x address count to 0  
	e.Data(0x00)

	log.Println("set ram y addr count")
	e.Cmd(0x4F)//set RAM y address count to 0  
	e.Data(0xAB)

	log.Println("bypass RAM data")
	e.Cmd(0x21)//bypass RAM data  
	e.Data(0x03)

	e.Cmd(0xF0)//booster feedback used  
	e.Data(0x1F)

	e.Cmd(0x2C)//vcom voltage  
	e.Data(0xA0)

	e.Cmd(0x3C)//board voltage  
	e.Data(0x63)

	e.Cmd(0x22)//display updata sequence option ,in page 33  
	e.Data(0xC4)//enable sequence: clk -> CP -> LUT -> pattern display  

	log.Println("write lut")
	e.Lut()
	e.Ready()
}

func (e *Eink) Lut() {
	data := []byte {
		0x82,0x00,0x00,0x00,0xAA,0x00,0x00,0x00,
		0xAA,0xAA,0x00,0x00,0xAA,0xAA,0xAA,0x00,
		0x55,0xAA,0xAA,0x00,0x55,0x55,0x55,0x55,
		0xAA,0xAA,0xAA,0xAA,0x55,0x55,0x55,0x55,
		0xAA,0xAA,0xAA,0xAA,0x15,0x15,0x15,0x15,
		0x05,0x05,0x05,0x05,0x01,0x01,0x01,0x01,
		0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		0x00,0x00,0x00,0x00,0x00,0x00,0x00,0x00,
		0x41,0x45,0xF1,0xFF,0x5F,0x55,0x01,0x00,
		0x00,0x00,
	}

	e.Cmd(0x32)
	for _, b := range data {
		e.Data(b)
	}
}

func (e *Eink) Img(pic [3096]byte) {
	log.Println("write img data")
	e.Cmd(0x24)
	for _, b := range pic {
		e.Data(b)
	}
	log.Println("write img data done")

	/*
	log.Println("display update seq")
	e.Cmd(0x22) //display updata sequence option  
	e.Data(0xf7)
	e.Ready()
	*/

	log.Println("master activation")
	e.Cmd(0x20)
	e.Ready()
	return
}

func (e *Eink) clk() {
	e.SCLK.H()
}

func (e *Eink) spi(b byte, dc bool) {
	e.CS.L()

	// When the pin is 
	// pulled HIGH, the data at D1 will be interpreted as data. When the pin is pulled LOW, the data at D1 
	// will be interpreted as command.
	e.DC.Write(dc)

	for i := 0; i < 8; i++ {
		e.SDA.Write(b & 0x80 != 0)

		e.SCLK.L()
	//	time.Sleep(time.Second/1e6)
		e.SCLK.H()
		b = b << 1
	}

	e.CS.H()
}

func (e *Eink) Data(i byte) {
	e.spi(i, true)
}

func (e *Eink) Cmd(i byte) {
	e.spi(i, false)
}

