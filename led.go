
package main

import (
	"time"
	"log"
	"github.com/go-av/a10/mmap-gpio"
)

func LedInit() {
	gpio.SetPinMode(8, 3, 2) // PI3=PWM1

	ch1 := uint32(0) // prescalar
	ch1 |= 1<<4 // enable
	ch1 |= 1<<5 // act state: 0 lowlevel 1 highlevel
	ch1 |= 1<<6 // gating 0=mask 1=pulse
	ch1 |= 0<<7 // mode 0=cycle mode 1=pulse mode.
	gpio.Writel(1, 0, ch1<<15) // Set pwm1 ctrl

	log.Println("led init")
}

func LedSet(i float64) {
	cyc := uint32(64)<<16 // entire cycles
	cyc |= uint32(64*i) // active cycles
	gpio.Writel(1, 8, cyc) // Set pwm1 period
}

func LedTest() {
	log.Println("test led")
	for {
		for i := 0.0; i < 1.0; i += 0.1 {
			LedSet(i)
			time.Sleep(time.Second/5)
		}
		for i := 1.0; i > 0.0; i -= 0.1 {
			LedSet(i)
			time.Sleep(time.Second/5)
		}
	}
}

