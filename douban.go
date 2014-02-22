
package main

import (
	"log"
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
	ApiParam m.M
	Email, Password string
	ConfSec string
	Channel string

	l *sync.Mutex
}

func NewDoubanFM() *DoubanFM {
	fm := &DoubanFM{
		ApiParam: m.M{
			"app_name": "radio_desktop_win",
			"version": "100",
		},
		ConfSec: "douban",
		Channel: "0",
		l: &sync.Mutex{},
	}

	return fm
}

func (f *DoubanFM) LoadConf() (email, pass, channel string) {
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		email, _ = c.GetValue(f.ConfSec, "email")
		pass, _ = c.GetValue(f.ConfSec, "password")
		channel, _ = c.GetValue(f.ConfSec, "channel")
	}
	return
}

func (f *DoubanFM) SaveConf() {
	os.Create(f.confFile())
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		c.SetValue(f.ConfSec, "email", f.Email)
		c.SetValue(f.ConfSec, "password", f.Password)
		c.SetValue(f.ConfSec, "channel", f.Channel)
		goconfig.SaveConfigFile(c, f.confFile())
	}
}

func (f *DoubanFM) confFile() string {
	return "fm.cfg"
}

func (f *DoubanFM) Api(method, path string, p m.M) (j m.M) {
	u, _ := url.ParseRequestURI("http://www.douban.com"+path)

	f.l.Lock()
	q := u.Query()
	p.Add(f.ApiParam).Each(func (k, v string) {
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

func (f *DoubanFM) EndSong(s m.M) {
	f.Api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":"e", "channel":f.Channel})
}

func (f *DoubanFM) TrashSong(s m.M) {
	f.Api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":"b", "channel":f.Channel})
}

func (f *DoubanFM) LikeSong(s m.M, like bool) {
	t := ""
	if like {
		t = "r"
	} else {
		t = "u"
	}
	f.Api("GET", "/j/app/radio/people", m.M{"sid":s.S("sid"), "type":t, "channel":f.Channel})
}

func (f *DoubanFM) GetSongList() m.A {
	r := f.Api("GET", "/j/app/radio/people", m.M{"type":"n", "channel":f.Channel})
	return r.A("song")
}

func (f *DoubanFM) Logout() {
	log.Println("douban: logout")
	f.l.Lock()
	f.Email = ""
	f.Password = ""
	delete(f.ApiParam, "token")
	delete(f.ApiParam, "user_id")
	delete(f.ApiParam, "expire")
	f.l.Unlock()
}

func (f *DoubanFM) Login(email, pass string) bool {
	log.Println("douban: login")
	if email == "" && pass == "" {
		log.Println("douban: login missing username or password")
		return false
	}
	r := f.Api("POST", "/j/app/login", m.M{
		"email": email, "password": pass,
	})
	if r.S("token") != "" {
		f.l.Lock()
		f.ApiParam["token"] = r.S("token")
		f.ApiParam["user_id"] = r.S("user_id")
		f.ApiParam["expire"] = r.S("expire")
		f.l.Unlock()
		log.Println("douban: login ok")
		return true
	} else {
		log.Println("douban: login failed", r)
		return false
	}
}


