
package main

import (
	"log"
	"os"
	"time"
	"bufio"
	"fmt"
	"runtime"
	"flag"

	"github.com/go-av/file"
	"github.com/go-av/wpa"
	"github.com/go-av/lush/m"
	"github.com/go-av/fmbox-rom/audio"
	"github.com/go-av/a10/mmap-gpio"
)

func consoleInputLoop() (ch chan int) {
	ch = make(chan int, 0)
	go func () {
		br := bufio.NewReader(os.Stdin)
		for {
			l, err := br.ReadString('\n')
			if err != nil {
				return
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

var ledBright float64
var debugApi bool

func main() {

	log.SetFlags(log.Lshortfile|log.LstdFlags)
	runtime.GOMAXPROCS(2)

	play := flag.String("play", "", "play mp3 file")
	flag.Float64Var(&ledBright, "led-bright", 0.05, "like led brightness")
	testLed := flag.Bool("test-led", false, "test pwm led")
	testBtn := flag.Bool("test-btn", false, "test button eint")
	testOled := flag.String("test-oled", "", "test oled display: text1|text2|pattern")
	testGpio := flag.Bool("test-gpio", false, "test gpio pins")
	wpacli := flag.Bool("wpa-cli", false, "wpa-cli mode")
	ctrl := flag.String("ctrl", "uart", "control interface")
	logto := flag.String("log", "", "rotate log to file")
	flag.BoolVar(&debugApi, "debug-api", false, "debug douban api")
	flag.Parse()

	if *logto != "" {
		log.SetOutput(file.AppendTo(*logto).LimitSize(1024*1024))
	}

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
		go ctrlWs()
	} else if *ctrl == "uart" {
		go ctrlUart()
	}

	if modeFmBox {
		oled.StartUpdateThread(24)
	}

	log.Println("starts fm")

	oled.SetOps([]*OledOp{
		&OledOp{Str:"连接中..", Y:1, X:40},
		wifiLedOp, btLedOp,
	})

	fm = NewDoubanFM()
	email, password, channel := fm.LoadConf()
	fm.SetChan(channel)
	if ok, cookie := fm.Login(email, password); ok {
		fm.SetCookie(cookie)
	}
	fm.GetChanList()
	fm.UpdateSongList()

	wifiLedOp.Hide = false

	go func () {
		for {
			btLedOp.Hide = uartLastAlive.IsZero() || time.Now().Sub(uartLastAlive) > time.Second*5
			wifiLedOp.Hide = !( wpa.Status() == "COMPLETED" )
			time.Sleep(time.Second*5)
		}
	}()

	keyStdin := consoleInputLoop()

	go func () {
		for {
			fm.UpdateSongList()
			DispSongLoad(fm.Song(0))
			ctrlSend(m.M{"op": "SongLoad", "song": fm.Song(0), "channel": fm.CurChan()})
			audio.CacheQueue(fm.Song(0).S("url"))
			audio.CacheQueue(fm.Song(1).S("url"))
			audio.Play(fm.Song(0).S("url"))
			audio.DelCache(fm.Song(0).S("url"))
			go fm.EndSong(fm.Song(0))
			fm.Next()
		}
	}()

	onKey := func (i int) {
		log.Println("onKey:", i)
		song := fm.Song(0)
		switch i {
		case BTN_LIKE:
			log.Println("key:", "Like")
			song["like"] = (song.I("like") ^ 1)
			fm.SetSong(0, song)
			DispShowLike(song.I("like") == 1)
			ctrlSend(m.M{"op": "SongLike", "like": song.I("like") == 1})
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

