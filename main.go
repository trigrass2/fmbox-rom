
package main

import (
	"log"
	"os"
	_ "time"
	"bufio"
	"fmt"
	"runtime"
	"flag"
	"sync"

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
				log.Println("consoleKey:", i)
				ch <- i
			}
		}
	}()
	return
}

var modeFmBox = (runtime.GOARCH == "arm")

func main() {

	log.SetFlags(log.Lshortfile|log.LstdFlags)
	runtime.GOMAXPROCS(2)

	play := flag.String("play", "", "play mp3 file")
	flag.Parse()

	if *play != "" {
		audio.PlayFile(*play)
		return
	}

	log.Println("Starts fm")

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
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func () {
		for {
			for len(songList) <= 1 {
				log.Println("fm: getting songList")
				list := fm.GetSongList()
				songList = append(songList, list...)
				log.Println("fm: getting songList done.", len(list), "entries loaded")
			}
			song = songList.M(0)
			wg.Done()
			wg.Add(1)
			songList = songList[1:]
			disp.SongLoad(song)
			audio.CacheQueue(song.S("url"))
			audio.CacheQueue(songList.M(0).S("url"))
			audio.Play(song.S("url"))
			audio.DelCache(song.S("url"))
		}
	}()

	wg.Wait()

	onKey := func (i int) {
		log.Println("onKey:", i)
		switch i {
		case BTN_LIKE:
			log.Println("key:", "Like")
			fm.LikeSong(song)
			disp.ToggleSongLike(song)
		case BTN_NEXT:
			log.Println("key:", "Next")
			audio.Stop()
			audio.DelCache(song.S("url"))
		case BTN_TRASH:
			log.Println("key:", "Trash")
			audio.Stop()
			audio.DelCache(song.S("url"))
			fm.TrashSong(song)
		case BTN_PAUSE:
			audio.Pause()
		}
		log.Println("key:", "done")
	}

	for {
		select {
		case k := <-BtnDown:
			onKey(k)
		case k := <-keyStdin:
			onKey(k)
		}
	}
}

