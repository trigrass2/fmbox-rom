
package main

import (
	"log"
	"fmt"
	"net/url"
	"time"
	"os"
	"sync"

	"github.com/go-av/wget"
	"github.com/Unknwon/goconfig"
	"github.com/go-av/lush/m"
)

type DoubanFM struct {
	apiParam m.M
	cookie m.M
	email, password string

	confFile string
	confSec string

	channel string
	channels m.A

	songs m.A

	l *sync.Mutex
}

func NewDoubanFM() *DoubanFM {
	fm := &DoubanFM{
		apiParam: m.M{
			"app_name": "radio_desktop_win",
			"version": "100",
		},
		cookie: m.M{},
		confSec: "douban",
		confFile: "/etc/fm.cfg",
		l: &sync.Mutex{},
	}

	return fm
}

func (f *DoubanFM) LoadConf() (email, pass, channel string) {
	log.Println("douban:", "load config")
	if c, err := goconfig.LoadConfigFile(f.confFile); err == nil {
		email, _ = c.GetValue(f.confSec, "email")
		pass, _ = c.GetValue(f.confSec, "password")
		channel, _ = c.GetValue(f.confSec, "channel")
	}
	return
}

func (f *DoubanFM) SaveConf() {
	f.l.Lock()
	defer f.l.Unlock()
	log.Println("douban:", "save config")
	os.Create(f.confFile)
	if c, err := goconfig.LoadConfigFile(f.confFile); err == nil {
		c.SetValue(f.confSec, "email", f.email)
		c.SetValue(f.confSec, "password", f.password)
		c.SetValue(f.confSec, "channel", f.channel)
		goconfig.SaveConfigFile(c, f.confFile)
	}
}

func (f *DoubanFM) api(method, path string, p m.M) (j m.M) {
	u, _ := url.ParseRequestURI("http://www.douban.com"+path)

	f.l.Lock()
	q := u.Query()
	p.Add(f.apiParam).Add(f.cookie).Each(func (k, v string) {
		q.Set(k, v)
	})
	f.l.Unlock()

	log.Println("douban:", "api:", path, p)

	if method == "GET" {
		u.RawQuery = q.Encode()
		j = wget.NewRequest(u.String(), 0).GetJson()
	} else {
		r := wget.NewRequest(u.String(), 0)
		r.PostData = q.Encode()
		j = r.GetJson()
	}

	return
}

func (f *DoubanFM) SetChan(channel string) (changed bool) {
	f.l.Lock()
	if f.channel != channel {
		f.channel = channel
		f.songs = m.A{}
		changed = true
	}
	f.l.Unlock()
	return
}

func (f *DoubanFM) CurChan() string {
	f.l.Lock()
	defer f.l.Unlock()
	return f.channel
}

func (f *DoubanFM) CurChanInfo() (rc m.M) {
	f.l.Lock()
	defer f.l.Unlock()
	rc = m.M{}
	f.channels.Each(func (c m.M) {
		if fmt.Sprint(c.I("channel_id")) == f.channel {
			rc = c
		}
	})
	return
}

func (f *DoubanFM) GetChanList() m.A {
	var a m.A
	for {
		log.Println("douban:", "getting channels list")
		r := f.api("GET", "/j/app/radio/channels", m.M{})
		a = r.A("channels")
		if len(a) > 0 {
			break
		}
		log.Println("douban:", "  get channel empty, retry")
		time.Sleep(time.Second)
	}
	f.l.Lock()
	f.channels = a
	f.channels.Each(func (c m.M) {
		log.Println("  ", c)
	})
	f.l.Unlock()
	return f.channels
}

func (f *DoubanFM) EndSong(s m.M) {
	f.api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":"e", "channel":f.CurChan()})
}

func (f *DoubanFM) TrashSong(s m.M) {
	f.api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":"b", "channel":f.CurChan()})
}

func (f *DoubanFM) LikeSong(s m.M, like bool) {
	t := ""
	if like {
		t = "r"
	} else {
		t = "u"
	}
	f.api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":t, "channel":f.CurChan()})
}

func (f *DoubanFM) Next() {
	f.l.Lock()
	defer f.l.Unlock()
	if len(f.songs) > 0 {
		f.songs = f.songs[1:]
	}
}

func (f *DoubanFM) SetSong(i int, s m.M) {
	f.l.Lock()
	defer f.l.Unlock()
	if i < len(f.songs) {
		f.songs[i] = s
	}
}

func (f *DoubanFM) Song(i int) m.M {
	f.l.Lock()
	defer f.l.Unlock()
	if i < len(f.songs) {
		return f.songs.M(i)
	}
	return m.M{}
}

func (f *DoubanFM) UpdateSongList() {
	for {
		f.l.Lock()
		if len(f.songs) >= 2 {
			f.l.Unlock()
			break
		}
		f.l.Unlock()

		log.Println("douban:", "getting song list")
		r := f.api("GET", "/j/app/radio/people", m.M{"type":"n", "channel":f.CurChan()})
		a := r.A("song")

		f.l.Lock()
		//if len(a) == 0 {
		//	f.channel = "1"
		//	log.Println("douban:   getsonglist failed: fall back to channel 1")
		//}
		f.songs = append(f.songs, a...)
		f.l.Unlock()
	}
}

func (f *DoubanFM) Logout() {
	log.Println("douban: logout")
	f.l.Lock()
	f.email = ""
	f.password = ""
	f.cookie = m.M{}
	f.l.Unlock()
}

func (f *DoubanFM) CurUser() string {
	f.l.Lock()
	defer f.l.Unlock()
	return f.email
}

func (f *DoubanFM) SetCookie(email, password string, cookie m.M) {
	f.l.Lock()
	f.email = email
	f.password = password
	f.cookie = cookie
	f.l.Unlock()
}

func (f *DoubanFM) Login(email, password string) (ok bool, cookie m.M) {
	log.Println("douban: login")
	if email == "" && password == "" {
		log.Println("douban: login missing username or password")
		return
	}
	r := f.api("POST", "/j/app/login", m.M{
		"email": email, "password": password,
	})
	if r.S("token") != "" {
		cookie = m.M{
			"token": r.S("token"),
			"user_id": r.S("user_id"),
			"expire": r.S("expire"),
		}
		log.Println("douban: login ok")
		ok = true
		return
	}

	log.Println("douban: login failed", r)
	return
}


