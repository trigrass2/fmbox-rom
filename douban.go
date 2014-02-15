
package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"bytes"
	"io"

	"github.com/Unknwon/goconfig"
	"github.com/go-av/lush/m"
)

type DoubanFM struct {
	ApiParam m.M
	Email, Password, Channel string
	Song m.A
}

func NewDoubanFM() *DoubanFM {
	fm := &DoubanFM{
		ApiParam: m.M{
			"app_name": "radio_desktop_win",
			"version": "100",
		},
		Channel: "0",
	}

	return fm
}

func (f *DoubanFM) confFile() string {
	return "fm.cfg"
}

func (f *DoubanFM) LoadConf() {
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		f.Email, _ = c.GetValue("user", "email")
		f.Password , _ = c.GetValue("user", "password")
		f.Channel, _ = c.GetValue("user", "channel")
	}
}

func (f *DoubanFM) SaveConf() {
	os.Create(f.confFile())
	if c, err := goconfig.LoadConfigFile(f.confFile()); err == nil {
		c.SetValue("user", "email", f.Email)
		c.SetValue("user", "password", f.Password)
		c.SetValue("user", "channel", f.Channel)
		goconfig.SaveConfigFile(c, f.confFile())
	}
}

func (f *DoubanFM) Api(method, path string, p m.M) (j m.M) {
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

func (f *DoubanFM) Login() {
	r := f.Api("POST", "/j/app/login", m.M{
		"email": f.Email, "password": f.Password,
	})
	if r.S("token") != "" {
		f.ApiParam["token"] = r.S("token")
		f.ApiParam["user_id"] = r.S("user_id")
		f.ApiParam["expire"] = r.S("expire")
		log.Println("douban: login ok")
	} else {
		log.Println("douban: login failed", r)
	}
}


