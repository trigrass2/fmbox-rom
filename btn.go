
package main

import (
	"log"
	"github.com/go-av/a10/mmap-gpio"
)

var BtnDown = make(chan int, 0)

func PollEIntBtn() {
	for {
		l := gpio.PollEInt()
		for i := uint32(0); i < 32; i++ {
			if l & (1<<i) != 0 {
				log.Println("Eint", i)
				if i >= 22 && i <= 25 {
					BtnDown <- int(i-22)
				}
			}
		}
	}
}

func EIntBtnInit() {
	/* port: A=0 B=1 C=2 D=3 E=4 F=5 G=6 H=7 I=8 */
	gpio.SetIntMode(22, 1) // eint22=-edge keydown
	gpio.SetIntMode(23, 1) // eint23=-edge
	gpio.SetIntMode(24, 1) // eint24=-edge
	gpio.SetPinMode(8, 10, 6) // PI10=eint22
	gpio.SetPinMode(8, 11, 6) // PI11=eint23
	gpio.SetPinMode(8, 12, 6) // PI12=eint24
	gpio.SetPinPull(8, 10, 1) // PI10=pull up
	gpio.SetPinPull(8, 11, 1) // PI11=pull up
	gpio.SetPinPull(8, 12, 1) // PI12=pull up
	gpio.SetIntEnable(22, 1) // enable eint22
	gpio.SetIntEnable(23, 1) // enable eint23
	gpio.SetIntEnable(24, 1) // enable eint24
}

