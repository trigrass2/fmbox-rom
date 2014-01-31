
package main

import (
	"log"
	"os"
	_ "time"
	"bufio"
	"fmt"
	"runtime"

	"github.com/go-av/lush/m"
	"github.com/go-av/douban.fm/audio"
	"github.com/go-av/a10/mmap-gpio"
)

func main() {
	runtime.GOMAXPROCS(4)

	log.Println("starts")

	fm := NewDoubanFM()
	fm.LoadConf()
	fm.Login()

	log.Println("login ok")

	disp := &LcdDisp{}
	if runtime.GOARCH == "arm" {
		gpio.Init()
		EIntBtnInit()
		LedInit()
	}

	keyStdin := make(chan int, 0)

	go func () {
		br := bufio.NewReader(os.Stdin)
		for {
			l, err := br.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			var i int
			n, _ := fmt.Sscanf(l, "%d", &i)
			if n == 1 {
				keyStdin <- i
			}
		}
	}()

	var song m.M
	var songList m.A

	nextSong := func () {
		for len(songList) == 0 {
			songList = fm.GetSongList()
		}
		song = songList.M(0)
		songList = songList[1:]
		disp.SongLoad(song)
		audio.Play(song.S("url"))
	}

	onKey := func (i int) {
		switch i {
		case 0: // like
			fm.LikeSong(song)
			disp.ToggleSongLike(song)
		case 1: // next
			nextSong()
		case 2: // trash
			fm.TrashSong(song)
			nextSong()
		}
	}

	nextSong()

	for {
		select {
		case k := <-BtnDown:
			onKey(k)
		case k := <-keyStdin:
			onKey(k)
		case <-audio.PlayStarted:
			disp.SongStart()
		case <-audio.PlayEnd:
			disp.SongEnd()
			nextSong()
		}
	}
}

