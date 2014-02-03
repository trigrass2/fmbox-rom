
package audio

import (
	"io"
)

type PcmSink struct {
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
		event: make(chan int, 16),
	}
}

func (s *PcmSink) Pause() {
	s.event <- pcmPause
}

func (s *PcmSink) Resume() {
	s.event <- pcmResume
}

func (s *PcmSink) Restart() {
	s.event <- pcmRestart
}

func (s *PcmSink) Write(p []byte) (n int, err error) {

	e := -1
	select {
	case e = <-s.event:
	default:
	}

	for {
		switch e {
		case pcmPause:
			s.paused = true
		case pcmResume:
			s.paused = false
		case pcmRestart:
			err = io.ErrClosedPipe
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

