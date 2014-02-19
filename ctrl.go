
package main

import (
	"bufio"
	"net/http"
	"os"
	"log"
	"fmt"
	"time"
	"io"

	"github.com/go-av/wpa"
	"github.com/go-av/lush/m"
	"code.google.com/p/go.net/websocket"
)

func ctrlWs(ws *websocket.Conn) {
	ctrlLoop(ws)
}

func ctrlLoop(ws io.ReadWriter) {
	br := bufio.NewReader(ws)
	ch := make(chan m.M, 0)
	log.Println("ctrl:", "starts")
	go func () {
		for {
			r, ok := <-ch
			if !ok {
				break
			}
			log.Println("ctrl: out", r)
			fmt.Fprintln(ws, r.Json())
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
		if !in.Has("ts") {
			continue
		}

		if in.S("op") == "FmLogin" {
			fm.Email = in.S("Email")
			fm.Password = in.S("Password")
			go func (ts int64) {
				out := m.M{"ts": ts}
				ok := fm.Login()
				if ok {
					out["r"] = 0
				} else {
					out["r"] = 1
					out["err"] = "LoginFailed"
				}
				ch <- out
			}(in.I64("ts"))
		}

		if in.S("op") == "WifiScanResults" {
			ch <- m.M{"r": 0, "ts": in.I64("ts"), "list": wpa.ScanResults()}
		}

		if in.S("op") == "WifiScan" {
			go func (ts int64) {
				list := wpa.Scan()
				out := m.M{"r": 0, "ts": ts , "list": list}
				ch <- out
			}(in.I64("ts"))
		}

		if in.S("op") == "WifiConnect" {
			go func (ts int64) {
				ssid := in.S("Ssid")
				bssid := in.S("Bssid")

				if in.B("SetConfig") {
					wpa.SetConfig(wpa.Config{
						Ssid: ssid,
						Bssid: bssid,
						KeyMgmt: in.S("KeyMgmt"),
						Key: in.S("Key"),
					})
				}

				wpa.Connect(ssid, bssid)
				ok := wpa.WaitCompleted(ssid, bssid, time.Second*10)
				out := m.M{"ts": ts}
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

				ch <- out
			}(in.I64("ts"))
		}
	}
	close(ch)
}

// This example demonstrates a trivial echo server.
func CtrlWs() {
	log.Println("ctrl:", "start websocket server")
	http.Handle("/fmbox", websocket.Handler(ctrlWs))
	err := http.ListenAndServe(":8888", nil)
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

