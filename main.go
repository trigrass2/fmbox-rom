
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

func consoleInputLoop() (ch chan int) {
	ch = make(chan int, 0)
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
				ch <- i
			}
		}
	}()
	return
}

var modeFmBox = (runtime.GOARCH == "arm")

func main() {
	log.Println("starts")

	fm := NewDoubanFM()
	fm.LoadConf()
	fm.Login()

	disp := &Disp{}

	if modeFmBox {
		gpio.Init()
		EIntBtnInit()
		LedInit()
	}

	keyStdin := consoleInputLoop()

	var song m.M
	var songList m.A

	nextSong := func () {
		for len(songList) <= 1 {
			log.Println("getting songList")
			songList = append(songList, fm.GetSongList()...)
		}
		if song != nil {
			audio.DelCache(song.S("url"))
		}
		song = songList.M(0)
		songList = songList[1:]
		disp.SongLoad(song)
		audio.Play(song.S("url"))
		audio.CacheQueue(songList.M(0).S("url"))
	}

	onKey := func (i int) {
		log.Println("onKey:", i)
		switch i {
		case BTN_LIKE:
			fm.LikeSong(song)
			disp.ToggleSongLike(song)
		case BTN_NEXT:
			nextSong()
		case BTN_TRASH:
			nextSong()
			fm.TrashSong(song)
		case BTN_PAUSE:
			audio.Pause()
		}
	}

	nextSong()

	for {
		select {
		case k := <-BtnDown:
			onKey(k)
		case k := <-keyStdin:
			onKey(k)
		case <-audio.PlayEnd:
			nextSong()
		}
	}
}

