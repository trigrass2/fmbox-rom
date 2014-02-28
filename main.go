
package main

import (
	"log"
	"os"
	"time"
	"bufio"
	"fmt"
	"runtime"
	"flag"
	"sync"

	"github.com/go-av/wpa"
	"github.com/go-av/lush/m"
	"gitcafe.com/nuomi-studio/fmbox-rom.git/audio"
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

var oled *Oled
var fm *DoubanFM
var song m.M

func main() {

	log.SetFlags(log.Lshortfile|log.LstdFlags)
	runtime.GOMAXPROCS(2)

	play := flag.String("play", "", "play mp3 file")
	testLed := flag.Bool("test-led", false, "test pwm led")
	testBtn := flag.Bool("test-btn", false, "test button eint")
	testOled := flag.String("test-oled", "", "test oled display: text1|text2|pattern")
	testGpio := flag.Bool("test-gpio", false, "test gpio pins")
	wpacli := flag.Bool("wpa-cli", false, "wpa-cli mode")
	ctrl := flag.String("ctrl", "ws", "control interface")
	flag.Parse()

	if *play != "" {
		audio.PlayFile(*play)
		return
	}

	if *wpacli {
		wpa.DoCli(flag.Args())
		return
	}

	if modeFmBox {
		gpio.Init()
		BtnInit()
		LedInit()
		oled = NewOled()
	}

	if modeFmBox && *testOled != "" {
		oled.Test(*testOled)
		return
	}

	if modeFmBox && *testLed {
		LedTest()
		return
	}

	if modeFmBox && *testBtn {
		BtnTest()
		return
	}

	if modeFmBox && *testGpio {
		log.Println("PWM Wave in PI4,PI5,PI6,PI7")
		go gpio.Open(8, 4, gpio.Out).Pwm(4, 1, time.Second/5)
		go gpio.Open(8, 5, gpio.Out).Pwm(4, 2, time.Second/5)
		go gpio.Open(8, 6, gpio.Out).Pwm(4, 3, time.Second/5)
		gpio.Open(8, 7, gpio.Out).Pwm(4, 4, time.Second/5)
		return
	}

	if *ctrl == "ws" {
		go CtrlWs()
	} else if *ctrl == "uart" {
		go CtrlUart()
	}

	log.Println("starts fm")

	fm = NewDoubanFM()
	email, pass, channel := fm.LoadConf()
	fm.Channel = channel
	fm.Login(email, pass)

	keyStdin := consoleInputLoop()

	var songList m.A
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func () {
		for {
			for len(songList) <= 1 {
				log.Println("fm: getting songList")
				if list := fm.GetSongList(); len(list) == 0 {
					log.Println("fm: songlist empty. retry")
					time.Sleep(time.Second)
					continue
				} else {
					log.Println("fm: getting songList done.", len(list), "entries loaded")
					songList = append(songList, list...)
				}
			}
			song = songList.M(0)
			wg.Done()
			wg.Add(1)
			songList = songList[1:]
			DispSongLoad(song)
			ctrlSend(m.M{"op": "SongLoad", "song": song})
			audio.CacheQueue(song.S("url"))
			audio.CacheQueue(songList.M(0).S("url"))
			audio.Play(song.S("url"))
			audio.DelCache(song.S("url"))
			fm.EndSong(song)
		}
	}()

	wg.Wait()
	if modeFmBox {
		oled.StartUpdateThread(24)
	}

	onKey := func (i int) {
		log.Println("onKey:", i)
		switch i {
		case BTN_LIKE:
			log.Println("key:", "Like")
			song["like"] = (song.I("like") ^ 1)
			DispShowLike(song.I("like") == 1)
			go fm.LikeSong(song, song.I("like") == 1)
		case BTN_NEXT:
			log.Println("key:", "Next")
			audio.Stop()
			audio.DelCache(song.S("url"))
		case BTN_TRASH:
			log.Println("key:", "Trash")
			audio.Stop()
			audio.DelCache(song.S("url"))
			go fm.TrashSong(song)
		case 10:
			log.Println("")

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

