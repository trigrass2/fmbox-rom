
package main

import (
	"bytes"
	"net/url"
	"net/http"
	"log"
	"io"
	"os"
	_ "time"
	"bufio"
	"fmt"
	"runtime"

	"github.com/go-av/douban.fm/mmap-gpio"
	"github.com/go-av/lush/m"
	"github.com/Unknwon/goconfig"
)

type FM struct {
	ApiParam m.M
	Email, Password, Channel string
	Song m.A
}

func NewFM() *FM {
	fm := &FM{
		ApiParam: m.M{
			"app_name": "radio_desktop_win",
			"version": "100",
		},
		Channel: "0",
	}

	return fm
}

func (f *FM) confFile() string {
	return "fm.cfg"
}

func (f *FM) LoadConf() {
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		f.Email, _ = c.GetValue("user", "email")
		f.Password , _ = c.GetValue("user", "password")
		f.Channel, _ = c.GetValue("user", "channel")
	}
}

func (f *FM) SaveConf() {
	os.Create(f.confFile())
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		c.SetValue("user", "email", f.Email)
		c.SetValue("user", "password", f.Password)
		c.SetValue("user", "channel", f.Channel)
		goconfig.SaveConfigFile(c, f.confFile())
	}
}

func (f *FM) Api(method, path string, p m.M) (j m.M) {
	u, _ := url.ParseRequestURI("http://www.douban.com"+path)

	q := u.Query()
	p.Add(f.ApiParam).Each(func (k, v string) {
		q.Set(k, v)
	})

	var resp *http.Response
	var err error

	if method == "GET" {
		u.RawQuery = q.Encode()
		r, _ := http.NewRequest(method, u.String(), nil)
		resp, err = http.DefaultClient.Do(r)
	} else {
		resp, err = http.PostForm(u.String(), q)
	}

	j = m.M{}
	if err != nil {
		log.Println(err)
		return
	}

	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)

	err = j.FromJson(b.String())
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func (f *FM) TrashSong(s m.M) {
}

func (f *FM) LikeSong(s m.M) {
}

func (f *FM) NextSong() m.M {
	if len(f.Song) == 0 {
		f.Song = f.GetSongList()
	}
	if len(f.Song) > 0 {
		r := f.Song.M(0)
		f.Song = f.Song[1:]
		return r
	}
	return nil
}

func (f *FM) GetSongList() m.A {
	r := f.Api("GET", "/j/app/radio/people", m.M{"type":"n", "channel":f.Channel})
	return r.A("song")
}

func (f *FM) Login() {
	r := f.Api("POST", "/j/app/login", m.M{
		"email": f.Email, "password": f.Password,
	})
	if len(r) > 0 {
		f.ApiParam["token"] = r.S("token")
		f.ApiParam["user_id"] = r.S("user_id")
		f.ApiParam["expire"] = r.S("expire")
	}
}

func main() {
	runtime.GOMAXPROCS(4)

	log.Println("starts")

	fm := NewFM()
	fm.LoadConf()
	fm.Login()

	mp := &Mplayer{}
	mp.Run()

	disp := &LcdDisp{}
	var song m.M

	if runtime.GOARCH == "arm" {
		gpio.Init()
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

	nextSong := func () {
		song = fm.NextSong()
		if song != nil {
			disp.SongLoad(song)
			mp.Play(song.S("url"))
		}
	}

	likeSong := func () {
		if song != nil {
			fm.LikeSong(song)
			disp.ToggleSongLike(song)
		}
	}

	onKey := func (i int) {
		switch i {
		case 0: // like
			likeSong()
		case 1: // next
			nextSong()
		case 2: // trash
			nextSong()
		}
	}

	nextSong()

	for {
		select {
		case k := <-gpio.BtnDown:
			onKey(k)
		case k := <-keyStdin:
			onKey(k)
		case <-mp.PlayStarted:
			disp.SongStart()
		case <-mp.PlayEnd:
			disp.SongEnd()
			nextSong()
		}
	}
}

