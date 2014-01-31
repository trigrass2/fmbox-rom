
package audio

import (
	_ "net/http"
)

var (
	PlayStarted = make(chan int, 0)
	PlayEnd = make(chan int, 0)
)

func Prefetch(uri string, slot int) {
}

func Play(uri string) {
}

