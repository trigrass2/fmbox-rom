
package audio

import (
	"io"
)

type pcmSink struct {
}

var PcmSink = pcmSink{}
var PcmEvent = make(chan int, 8)
var PcmPaused = false

const (
	PCM_RESUME = iota
	PCM_PAUSE
	PCM_RESTART
)

func init() {
}

func (s pcmSink) Write(p []byte) (n int, err error) {

	e := -1
	select {
	case e = <-PcmEvent:
	default:
	}

	for {
		switch e {
		case PCM_PAUSE:
			PcmPaused = true
		case PCM_RESUME:
			PcmPaused = false
		case PCM_RESTART:
			err = io.ErrClosedPipe
			return
		}
		if PcmPaused {
			e = <-PcmEvent
		} else {
			break
		}
	}

	return len(p), nil
}

