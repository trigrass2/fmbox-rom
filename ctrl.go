
package main

import (
	"bufio"
	"net/http"
	"os"
	"log"
	"strings"
	"io"
	"net"
	"time"
	"sync"

	"github.com/go-av/wpa"
	"github.com/go-av/lush/m"
	"code.google.com/p/go.net/websocket"
)

var ctrlLock = &sync.Mutex{}
var ctrlChans = m.A{}

func ctrlSend(in m.M) {
	ctrlLock.Lock()
	ctrlChans.Each(func (ch chan m.M) {
		ch <- in
	})
	ctrlLock.Unlock()
}

func ctrlHandle(ch chan m.M, in m.M) {
	out := m.M{"ts":in.I64("ts")}

	log.Println("ctrl:", "op", in.S("op"))

	switch in.S("op") {

	case "FmStat":
		out["song"] = fm.Song(0)
		out["email"] = fm.CurUser()
		ch <- out

	case "FmSetChan":
		if fm.SetChan(in.S("channel")) {
			fm.SaveConf()
			BtnDown <- BTN_NEXT
		}

	case "FmChanList":
		go func () {
			out["list"] = fm.GetChanList()
			ch <- out
		}()

	case "FmNext":
		BtnDown <- BTN_NEXT

	case "FmTrash":
		BtnDown <- BTN_TRASH

	case "FmLike":
		BtnDown <- BTN_LIKE

	case "FmLogout":
		fm.Logout()
		fm.SaveConf()

	case "FmLogin":
		go func () {
			email := in.S("Email")
			password := in.S("Password")
			if ok, cookie := fm.Login(email, password); ok {
				out["r"] = 0
				fm.SetCookie(email, password, cookie)
				fm.SaveConf()
			} else {
				out["r"] = 1
				out["err"] = "LoginFailed"
			}
			ch <- out
		}()

	case "WifiScanResults":
		out["r"] = 0
		out["list"] = wpa.ScanResults()
		ch <- out

	case "WifiScan":
		go func () {
			out["r"] = 0
			out["list"] = wpa.Scan()
			ch <- out
		}()

	case "WifiConnect":
		go func () {
			ssid := in.S("Ssid")
			bssid := in.S("Bssid")

			ok := wpa.Connect(ssid, bssid, wpa.Config{
				KeyMgmt: in.S("KeyMgmt"),
				Key: in.S("Key"),
				ScanSsid: in.B("ScanSsid"),
			})

			if !ok {
				out["r"] = 1
				out["err"] = "ConnectFailed"
			} else {
				out["r"] = 0
			}
			log.Println("ctrl:", "connect", ssid, bssid, "result:", ok)
			ch <- out
		}()
	}

}

func escapeAT(in string) (out string) {
	return strings.Replace(in, "AT", `\x41\x54`, -1)
}

func ctrlLoop(rw io.ReadWriter) {
	br := bufio.NewReader(rw)
	log.Println("ctrl:", "loop starts")

	ch := make(chan m.M, 1024)
	ctrlLock.Lock()
	ctrlChans = append(ctrlChans, ch)
	ctrlLock.Unlock()

	go func () {
		for {
			r, ok := <-ch
			if !ok {
				break
			}
			log.Println("ctrl: out", r)
			msg := r.Json()+"\n"
			msg = escapeAT(msg)

			log.Println("ctrl: out", r)

			var err error
			if ws, ok := rw.(*websocket.Conn); ok {
				err = websocket.Message.Send(ws, msg)
			} else {
				_, err = rw.Write([]byte(msg))
			}
			if err != nil {
				log.Println("ctrl: send failed:", err)
				ch <- r
				break
			}
			log.Println("ctrl: sent", len(msg), "bytes")
		}
		log.Println("ctrl:", "conn close")
	}()

	for {
		l, err := br.ReadString('\n')
		if err != nil {
			break
		}
		in := m.M{}
		in.FromJson(l)
		log.Println("ctrl: in", in)
		ctrlHandle(ch, in)
	}

	ctrlLock.Lock()
	ctrlChans = ctrlChans.Del(ch)
	ctrlLock.Unlock()
}

func ctrlWs() {
	log.Println("ctrl:", "start websocket server")
	http.Handle("/fmbox", websocket.Handler(func (ws *websocket.Conn) {
		log.Println("ctrl:", "websocket accept")
		ctrlLoop(ws)
		log.Println("ctrl:", "websocket close")
	}))
	err := http.ListenAndServe(":8787", nil)
	if err != nil {
		log.Println("ctrl: websocket listen failed:", err)
	}
}

func ctrlUart() {
	log.Println("ctrl:", "start uart")

	uart := "/dev/ttyS2"
	f, err := os.OpenFile(uart, os.O_RDWR, 0744)
	if err != nil {
		log.Println("ctrl:", "open", uart, err)
		return
	}

	ctrlLoop(f)
}

func ctrlReverseTcp() {
	log.Println("ctrl:", "start reverse tcp")

	addr := "125.39.155.32:1984"
	for {
		conn, err := net.DialTimeout("tcp4", addr, time.Second*20)
		if err != nil {
			log.Println("ctrl:", "dial", err)
			time.Sleep(time.Second)
			continue
		}
		log.Println("ctrl:", "reverse tcp connected")
		ctrlLoop(conn)
		log.Println("ctrl:", "reverse tcp close")
	}
}

