/*
	http mp3 caching and play
*/
package audio

import (
	"io"
	"net/http"
	"sync"
	"log"
	"os/exec"

	"github.com/go-av/fifo"
)

type cacheConn struct {
	uri string
	cancel bool
	buf *fifo.Buffer
}

var cacheLock = &sync.Mutex{}
var cachePool = map[string]*cacheConn{}
var cacheQueue = make(chan *cacheConn, 0)

func doCache(conn *cacheConn) {
	defer conn.buf.Close()
	req, err := http.NewRequest("GET", conn.uri, nil)
	if err != nil {
		log.Println("doCache:", err)
		return
	}
	resp, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		log.Println("doCache:", err2)
	}
	io.Copy(conn.buf, resp.Body)
}

func init() {
	go cacheQueueThread()
	go playThread()
}

func cacheQueueThread() {
	for {
		conn := <-cacheQueue
		if conn.cancel {
			continue
		}
		doCache(conn)
	}
}

func DelCache(uri string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	conn, ok := cachePool[uri]
	if !ok {
		return
	}
	conn.cancel = true
	delete(cachePool, uri)
}

func CacheQueue(uri string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	_, ok := cachePool[uri]
	if ok {
		return
	}
	conn := &cacheConn{uri: uri, buf: fifo.NewBuffer()}
	cachePool[uri] = conn
	cacheQueue <- conn
}

var playQueue = make(chan *cacheConn, 0)
var playEvent = make(chan int, 0)
var PlayEnd = make(chan int, 0)

func playThread() {
	for {
		switch {
		case conn := <-playQueue:
			conn.buf.ResetRead()
			dec := NewMp3Decoder()
			dec.Input = conn.buf
			dec.Output = PcmSink
			go func () {
				dec.Run()
			}()
		case e := <-playEvent:

		}
	}
}

func Play(uri string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	conn, ok := cachePool[uri]
	if !ok {
		return
	}
	playQueue <- conn
}

func Stop() {
}


