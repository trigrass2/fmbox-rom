
/*
	http mp3 caching and play (thread unsafe)
*/
package audio

import (
	"io"
	"log"
	"sync"
	"net"
	"net/http"

	"github.com/go-av/fifo"
	"github.com/go-av/douban.fm/client-mad"
)

type cacheConn struct {
	l *sync.Mutex
	conns []net.Conn
	req *http.Request
	uri string
	cancel bool
	buf *fifo.Buffer
}

var cachePool = map[string]*cacheConn{}
var cacheQueue = make(chan *cacheConn, 16)

func (c *cacheConn) Cancel() {
	c.cancel = true
	c.buf.Close()
	c.l.Lock()
	for _, nc := range c.conns {
		if nc != nil {
			nc.Close()
		}
	}
	c.l.Unlock()
}

func doCache(conn *cacheConn) {
	defer conn.buf.Close()
	log.Println("cacheStart:", conn.uri)
	log.Println("cachePool:", len(cachePool), "entries")

	trans := &http.Transport{
		Dial: func (netw, addr string) (net.Conn, error) {
			c, err := net.Dial(netw, addr)
			conn.l.Lock()
			conn.conns = append(conn.conns, c)
			conn.l.Unlock()
			return c, err
		},
	}

	cli := &http.Client{
		Transport: trans,
	}

	resp, err2 := cli.Do(conn.req)
	if err2 != nil {
		log.Println("doCache:", err2)
		return
	}
	if resp.Body == nil {
		log.Println("doCache:", "http response no body")
		return
	}
	log.Println("cacheStartData:", conn.uri)
	_, err := io.Copy(conn.buf, resp.Body)
	resp.Body.Close()
	log.Println("cacheDone:", conn.uri, err)
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
	conn.Cancel()
	delete(cachePool, uri)
	log.Println("delCache", uri)
}

func CacheQueue(uri string) {
	_, ok := cachePool[uri]
	if ok {
		return
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println("cacheQueue: uri invalid", uri)
		return
	}
	conn := &cacheConn{
		uri: uri,
		buf: fifo.NewBuffer(),
		req: req,
		l: &sync.Mutex{},
	}
	cachePool[uri] = conn
	cacheQueue <- conn
	log.Println("cacheQueue:", uri)
}

var playQueue = make(chan *cacheConn, 0)
var PlayEnd = make(chan int, 0)

func playQueueThread() {
	for {
		conn := <-playQueue
		conn.buf.ResetRead()
		log.Println("play:", conn.uri)
		dec := mad.NewDecoder()
		dec.R = conn.buf
		dec.W = PcmSink0
		err := dec.Run()
		if !conn.cancel {
			PlayEnd <- 1
			log.Println("playEnd:", conn.uri)
		} else {
			log.Println("playCanceled:", conn.uri, err)
		}
	}
}

func Play(uri string) {
	CacheQueue(uri)
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

