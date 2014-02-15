
package main

import (
	"time"
	"log"
	"github.com/go-av/a10/mmap-gpio"
)

var BtnDown = make(chan int, 0)

const (
	BTN_LIKE = iota
	BTN_NEXT
	BTN_TRASH
	BTN_PAUSE
)

func BtnTest() {
	log.Println("press a button")
	for {
		i := <-BtnDown
		log.Println("BtnDown", i)
	}
}

var lastHit [3]time.Time

func btnHit(i int) {
	if lastHit[i].IsZero() {
		lastHit[i] = time.Now()
	}
	last := lastHit[i]
	now := time.Now()
	if now.Sub(last) < time.Second/5 {
		return
	}

	lastHit[i] = now
	BtnDown <- i
}

func BtnPoll() {
	for {
		l := gpio.PollEInt()
		for i := uint32(0); i < 32; i++ {
			if l & (1<<i) != 0 {
				log.Println("BtnEint:", i)
				if i >= 23 && i <= 25 {
					btnHit(int(i) - 23)
				}
			}
		}
	}
}

func BtnInit() {
	/* port: A=0 B=1 C=2 D=3 E=4 F=5 G=6 H=7 I=8 */
	gpio.SetIntMode(23, 1) // eint23=-edge keydown
	gpio.SetIntMode(24, 1) // eint24=-edge
	gpio.SetIntMode(25, 1) // eint25=-edge
	gpio.SetPinMode(8, 11, 6) // PI11=eint22
	gpio.SetPinMode(8, 12, 6) // PI12=eint23
	gpio.SetPinMode(8, 13, 6) // PI13=eint24
	gpio.SetPinPull(8, 11, 1) // PI11=pull up
	gpio.SetPinPull(8, 12, 1) // PI12=pull up
	gpio.SetPinPull(8, 13, 1) // PI13=pull up
	gpio.SetIntEnable(23, 1) // enable eint23
	gpio.SetIntEnable(24, 1) // enable eint24
	gpio.SetIntEnable(25, 1) // enable eint25
	go BtnPoll()

	log.Println("btn: init")
}

