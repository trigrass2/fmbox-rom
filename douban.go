
package main

import (
	"log"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"bytes"
	"sync"
	"io"

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
	p.Add(f.apiParam).Each(func (k, v string) {
		q.Set(k, v)
	})
	f.l.Unlock()

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
		log.Println("douban: api", err)
		return
	}

	b := new(bytes.Buffer)
	io.Copy(b, resp.Body)

	err = j.FromJson(b.String())
	if err != nil {
		log.Println("douban: api", err, q)
		return
	}

	return
}

func (f *DoubanFM) SetChan(channel string) {
	f.l.Lock()
	f.channel = channel
	f.l.Unlock()
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
	r := f.api("GET", "/j/app/radio/channels", m.M{})
	f.l.Lock()
	f.channels = r.A("channels")
	log.Println("douban: getChanList", f.channels)
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

func (f *DoubanFM) GetSongList() m.A {
	r := f.api("GET", "/j/app/radio/people", m.M{"type":"n", "channel":f.CurChan()})
	a := r.A("song")

	f.l.Lock()
	if len(a) == 0 {
		log.Println("douban:", "cannot get songs in channel", f.channel, ". change to 1")
		f.channel = "1"
	}
	f.l.Unlock()

	return a
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


