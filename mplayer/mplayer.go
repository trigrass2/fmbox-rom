
package main

import (
	"bytes"
	"runtime"
	"strconv"
	"io"
	"time"
	"fmt"
	"log"
	"strings"
	"os/exec"
	"bufio"
)

type Mplayer struct {
	cmd *exec.Cmd
	cmdOut *bufio.Scanner
	cmdIn io.Writer
	ch chan string
	playStarted, PlayStarted chan int
	PlayEnd chan int
	PlayPos chan time.Duration
}

func (mp *Mplayer) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return len(data), data, nil
	}
	if i := bytes.Index(data, []byte{0x1b, 0x5b, 0x4a, 0xd}); i >= 0 {
		return i+4, data[0:i], nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i+1, data[0:i], nil
	}
	return 0, nil, nil
}

func (mp *Mplayer) parsePos(t string) (ok bool, pos, length time.Duration) {
	if strings.HasPrefix(t, "A:") {
		ok = true
		if a := strings.Fields(t); len(a) >= 4 {
			pos1, _ := strconv.ParseFloat(a[1], 64)
			len1, _ := strconv.ParseFloat(a[4], 64)
			pos = time.Duration(pos1*1000)*time.Second/1000
			length = time.Duration(len1*1000)*time.Second/1000
		}
	}
	return
}

func (mp *Mplayer) Run() {

	if runtime.GOARCH != "arm" {
		mp.cmd = exec.Command("mplayer", "-ao", "null", "-slave", "-idle")
	} else {
		mp.cmd = exec.Command("mplayer", "-slave", "-idle")
	}

	stderr, _ := mp.cmd.StderrPipe()
	stdout, _ := mp.cmd.StdoutPipe()
	mp.cmdIn, _ = mp.cmd.StdinPipe()
	mp.cmdOut = bufio.NewScanner(stdout)
	mp.cmdOut.Split(mp.split)
	mp.cmd.Start()

	mp.PlayStarted = make(chan int, 0)
	mp.playStarted = make(chan int, 0)
	mp.PlayEnd = make(chan int, 0)
	mp.PlayPos = make(chan time.Duration, 0)
	mp.ch = make(chan string, 0)
	read := make(chan string, 0)

	go func () {
		br := bufio.NewReader(stderr)
		for {
			l, err := br.ReadString('\n')
			if err != nil { break }
			l = strings.Trim(l, "\n")
			log.Println("mplayer:", l)
			if strings.Contains(l, "truncated at end") {
				mp.PlayEnd <- 1
			}
		}
	}()

	go func () {
		for mp.cmdOut.Scan() {
			read <- mp.cmdOut.Text()
		}
	}()

	go func () {
		expect := ""
		for {
			select {
			case expect = <-mp.ch:
				log.Println("expect", expect)
			case l := <-read:
				if expect != "" && strings.HasPrefix(l, expect) {
					log.Println("got", l)
					mp.ch <- l[len(expect):]
				}
				if strings.HasPrefix(l, "Starting playback") {
					mp.playStarted <- 1
				}
				if ok, pos, _ := mp.parsePos(l); ok {
					mp.PlayPos <- pos
				}
			}
		}
	}()
}

func (mp *Mplayer) GetProp(name string, to time.Duration) (ans string) {
	mp.ch <- "ANS_"+name+"="
	end := time.Now().Add(to)
	for {
		select {
		case ans = <-mp.ch:
			return
		case <-time.After(time.Second/2):
			fmt.Fprintln(mp.cmdIn, "get_property", name)
		case <-time.After(end.Sub(time.Now())):
			return
		}
	}
	return
}

func (mp *Mplayer) Pos() (pos time.Duration) {
	p := mp.GetProp("time_pos", time.Second)
	var f float64
	fmt.Sscanf(p, "%f", &f)
	return time.Duration(f*1000) * (time.Second / 1000)
}

func (mp *Mplayer) Play(uri string) {
	log.Println("Play", uri)
	uri = "http://localhost:1653" + strings.TrimPrefix(uri, "http:/")
	fmt.Fprintln(mp.cmdIn, "load", uri)
	go func () {
		select {
		case <-time.After(time.Second*10):
			mp.PlayEnd <- 1
		case <-mp.playStarted:
			log.Println("mplayer: PlayStarted")
			mp.PlayStarted <- 1
		}
	}()
}

func (mp *Mplayer) Unpause() {
}

func (mp *Mplayer) Pause() {
}

