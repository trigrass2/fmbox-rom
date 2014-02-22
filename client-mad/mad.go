
package mad

import (
	"log"
	"io"
	"fmt"
	"encoding/binary"
	"os/exec"
	"sync"
)

func Decode(r io.Reader) (rate uint32, raw io.ReadCloser, err error) {
	cmd := exec.Command("mad")
	cmd.Stdin = r
	raw, _ = cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		log.Println("mad:", "start mad failed:", err)
		return
	}

	lock.Lock()
	mad = cmd
	lock.Unlock()

	err = binary.Read(raw, binary.LittleEndian, &rate)
	if err != nil {
		log.Println("mad: read rate failed:", err)
	}
	return
}

var lock = &sync.Mutex{}
var aplay *exec.Cmd
var mad *exec.Cmd

func StopPlay() {
	lock.Lock()
	if aplay != nil && aplay.Process != nil {
		aplay.Process.Kill()
	}
	if mad != nil && mad.Process != nil {
		mad.Process.Kill()
	}
	lock.Unlock()
}

func Play(r io.Reader) (err error) {
	StopPlay()

	var rate uint32
	var raw io.ReadCloser
	rate, raw, err = Decode(r)
	if err != nil {
		return
	}

	cmd := exec.Command("aplay", "-c", "2", "-f", "S16_LE", "-r", fmt.Sprint(rate))
	cmd.Stdin = raw
	err = cmd.Start()
	if err != nil {
		log.Println("mad: start aplay failed:", err)
		return
	}

	lock.Lock()
	aplay = cmd
	lock.Unlock()

	log.Println("mad: aplay starts. samplerate", rate)

	err = cmd.Wait()
	log.Println("mad: aplay end:", err)

	return
}

