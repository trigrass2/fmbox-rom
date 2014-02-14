
package audio

import (
	"io"
	_ "log"
	"github.com/go-av/aplay"
)

type PcmSink struct {
	io.Writer
	event chan int
	paused bool
}

const (
	pcmResume = iota
	pcmPause
	pcmRestart
)

var PcmSink0 = NewPcmSink()

func NewPcmSink() *PcmSink {
	return &PcmSink{
		//event: make(chan int, 16),
		Writer: aplay.Input,
	}
}

func (s *PcmSink) Pause() {
	//s.event <- pcmPause
}

func (s *PcmSink) Resume() {
	//s.event <- pcmResume
}

func (s *PcmSink) Restart() {
	//s.event <- pcmRestart
}

/*
func (s *PcmSink) Write(p []byte) (n int, err error) {

	if false {
		log.Println("pcmSink: Write", len(p))
	}

	e := -1
	select {
	case e = <-s.event:
	default:
	}

	for {
		switch e {
		case pcmPause:
			s.paused = true
			log.Println("pcm: Pause")
		case pcmResume:
			s.paused = false
			log.Println("pcm: Resume")
		case pcmRestart:
			err = io.ErrClosedPipe
			log.Println("pcm: Restart")
			return
		}
		if s.paused {
			e = <-s.event
		} else {
			break
		}
	}

	return len(p), nil
}
*/

