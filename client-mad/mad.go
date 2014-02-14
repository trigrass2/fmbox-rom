
package mad

import (
	"log"
	"io"
	"net"
	"fmt"
	"encoding/binary"
	"os/exec"
	"sync/atomic"
	"unsafe"
)

func Decode(r io.Reader) (rate uint32, raw io.ReadCloser, err error) {
	var c net.Conn
	c, err = net.Dial("tcp", "localhost:91")
	if err != nil {
		log.Println("mad: Dial", err)
		return
	}

	go func () {
		io.Copy(c, r)
		c.(*net.TCPConn).CloseWrite()
	}()

	err = binary.Read(c, binary.LittleEndian, &rate)
	if err != nil {
		log.Println("mad: getRate", err)
	}
	raw = c
	return
}

var curCmd unsafe.Pointer

func StopPlay() {
	oldCmd := (*exec.Cmd)(atomic.LoadPointer(&curCmd))
	if oldCmd != nil {
		oldCmd.Process.Kill()
	}
}

func Play(r io.Reader) (err error) {
	var rate uint32
	var raw io.ReadCloser
	rate, raw, err = Decode(r)
	if err != nil {
		return
	}

	log.Println("mad: rate", rate)

	StopPlay()

	cmd := exec.Command("aplay", "-c", "2", "-f", "S16_LE", "-r", fmt.Sprint(rate))
	w, _ := cmd.StdinPipe()
	err = cmd.Start()
	if err != nil {
		log.Println("mad: start aplay", err)
		return
	}

	atomic.StorePointer(&curCmd, unsafe.Pointer(cmd))
	_, err = io.Copy(w, raw)
	if err != nil {
		log.Println("mad: aplay interrupted", err)
	}
	raw.Close()
	cmd.Wait()
	log.Println("mad: aplay end")
	atomic.StorePointer(&curCmd, unsafe.Pointer(nil))

	return
}

