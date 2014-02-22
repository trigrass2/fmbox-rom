
package main

import (
	"bufio"
	"net/http"
	"os"
	"log"
	_ "fmt"
	"time"
	"io"

	"github.com/go-av/wpa"
	"github.com/go-av/lush/m"
	"code.google.com/p/go.net/websocket"
)

func ctrlWs(ws *websocket.Conn) {
	ctrlLoop(ws)
}

var ctrlCh chan m.M = make(chan m.M, 1024)

func ctrlSend(r m.M) {
	if len(ctrlCh) < 128 {
		ctrlCh <- r
	} else {
		log.Println("ctrlSend:", "queue full")
	}
}

func ctrlHandle(in m.M) {
	out := m.M{"ts":in.I64("ts")}

	log.Println("ctrl:", "op", in.S("op"))

	switch in.S("op") {

	case "FmStat":
		out["song"] = song
		if fm.Email != "" {
			out["email"] = fm.Email
		}
		ctrlSend(out)

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
			pass := in.S("Password")
			if fm.Login(email, pass) {
				out["r"] = 0
				fm.SaveConf()
			} else {
				out["r"] = 1
				out["err"] = "LoginFailed"
			}
			ctrlSend(out)
		}()

	case "WifiScanResults":
		out["r"] = 0
		out["list"] = wpa.ScanResults()
		ctrlSend(out)

	case "WifiScan":
		go func () {
			out["r"] = 0
			out["list"] = wpa.Scan()
			ctrlSend(out)
		}()

	case "WifiConnect":
		go func () {
			ssid := in.S("Ssid")
			bssid := in.S("Bssid")

			if in.B("SetConfig") {
				wpa.SetConfig(wpa.Config{
					Ssid: ssid,
					Bssid: bssid,
					KeyMgmt: in.S("KeyMgmt"),
					Key: in.S("Key"),
					ScanSsid: in.B("ScanSsid"),
				})
			}

			wpa.Connect(ssid, bssid)
			ok := wpa.WaitCompleted(ssid, bssid, time.Second*10)
			if !ok {
				out["r"] = 1
				out["err"] = "ConnectFailed"
			} else {
				out["r"] = 0
			}

			log.Println("ctrl:", "connect", ssid, bssid, ":", ok)

			if !ok && in.B("SetConfig") {
				wpa.DelConfig(wpa.Config{
					Ssid: ssid,
					Bssid: bssid,
				})
			}

			ctrlSend(out)
		}()
	}

}

func ctrlLoop(rw io.ReadWriter) {
	br := bufio.NewReader(rw)
	log.Println("ctrl:", "starts")

	go func () {
		for {
			r, ok := <-ctrlCh
			if !ok {
				break
			}
			log.Println("ctrl: out", r)
			msg := r.Json()+"\n"

			var err error
			if ws, ok := rw.(*websocket.Conn); ok {
				err = websocket.Message.Send(ws, msg)
			} else {
				_, err = rw.Write([]byte(msg))
			}
			if err != nil {
				log.Println("ctrl: send failed:", err)
				ctrlCh <- r
				break
			}
			log.Println("ctrl: sent", len(msg), "bytes")
		}
		log.Println("ctrl:", "close")
	}()

	for {
		l, err := br.ReadString('\n')
		if err != nil {
			break
		}
		in := m.M{}
		in.FromJson(l)
		log.Println("ctrl: in", in)
		ctrlHandle(in)
	}
}

func CtrlWs() {
	log.Println("ctrl:", "start websocket server")
	http.Handle("/fmbox", websocket.Handler(ctrlWs))
	err := http.ListenAndServe(":8787", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func CtrlUart() {
	log.Println("ctrl:", "start uart")

	uart := "/dev/ttyS4"
	f, err := os.OpenFile(uart, os.O_RDWR, 0744)
	if err != nil {
		log.Println("ctrl:", "open", uart, err)
		return
	}

	ctrlLoop(f)
}

