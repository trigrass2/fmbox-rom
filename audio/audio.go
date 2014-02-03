
/*
	http mp3 caching and play (thread unsafe)
*/
package audio

import (
	"io"
	"log"
	"net/http"
	"os/exec"

	"github.com/go-av/fifo"
)

type cacheConn struct {
	uri string
	cancel bool
	buf *fifo.Buffer
}

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
	resp.Body.Close()
}

func init() {
	go cacheQueueThread()
	go playQueueThread()
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
	conn, ok := cachePool[uri]
	if !ok {
		return
	}
	conn.cancel = true
	conn.buf.Close()
	delete(cachePool, uri)
}

func CacheQueue(uri string) {
	_, ok := cachePool[uri]
	if ok {
		return
	}
	conn := &cacheConn{uri: uri, buf: fifo.NewBuffer()}
	cachePool[uri] = conn
	cacheQueue <- conn
}

var playQueue = make(chan *cacheConn, 0)
var PlayEnd = make(chan int, 0)

func playQueueThread() {
	for {
		conn := <-playQueue
		conn.buf.ResetRead()
		dec := exec.Command("mad")
		dec.Stdin = conn.buf
		dec.Stdout = PcmSink0
		if dec.Run() == nil {
			PlayEnd <- 1
		}
	}
}

func Play(uri string) {
	CacheQueue(uri)
	Stop()
	playQueue <- cachePool[uri]
}

func Resume() {
	PcmSink0.Resume()
}

func Pause() {
	PcmSink0.Pause()
}

func Stop() {
	PcmSink0.Restart()
}

