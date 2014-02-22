
/*
	http mp3 caching and play (thread unsafe)
*/
package audio

import (
	"os"
	"io"
	"log"
	"time"
	"sync"
	"net"
	"net/http"

	"gitcafe.com/nuomi-studio/fifo.git"
	"gitcafe.com/nuomi-studio/fmbox-rom.git/client-mad"
)

type cacheConn struct {
	conns []net.Conn
	req *http.Request
	uri string
	buf *fifo.Buffer
	tm time.Time
}

var cacheLock = &sync.Mutex{}
var cachePool = map[string]*cacheConn{}
var cacheWait = make(chan int, 1024)
var curPlayConn *cacheConn

func init() {
	go cacheQueueThread()
}

func (c *cacheConn) Close() {
	cacheLock.Lock()
	c.buf.Close()
	for _, nc := range c.conns {
		if nc != nil {
			nc.Close()
		}
	}
	cacheLock.Unlock()
}

func doCache(conn *cacheConn) (err error) {
	log.Println("cacheStart:", conn.uri)
	log.Println("cachePool:", len(cachePool), "entries")

	trans := &http.Transport{
		Dial: func (netw, addr string) (net.Conn, error) {
			log.Println("cache:", "  Dial:", addr)
			c, err := net.Dial(netw, addr)
			log.Println("cache:", "  DialDone:", addr)
			cacheLock.Lock()
			conn.conns = append(conn.conns, c)
			cacheLock.Unlock()
			return c, err
		},
	}

	cli := &http.Client{
		Transport: trans,
	}

	resp, err2 := cli.Do(conn.req)
	if err2 != nil {
		log.Println("cacheErr:", err2)
		conn.Close()
		return err2
	}
	if resp.Body == nil {
		log.Println("cacheErr:", "http response no body")
		conn.Close()
		return err2
	}
	log.Println("cacheStartIo:", conn.uri)
	n, err := io.Copy(conn.buf, resp.Body)
	resp.Body.Close()

	log.Println("cacheDone:", conn.uri, err, n/1024, "KiB")
	conn.buf.CloseWrite()
	return nil
}

func latestConn() (latest *cacheConn) {
	tm := time.Time{}
	cacheLock.Lock()
	for _, c := range cachePool {
		if !c.tm.IsZero() && (tm.IsZero() || c.tm.Before(tm)) {
			tm = c.tm
			latest = c
		}
	}
	cacheLock.Unlock()
	return
}

func cacheQueueThread() {
	for {
		<-cacheWait
		c := latestConn()
		if c == nil {
			continue
		}

		doCache(c)

		cacheLock.Lock()
		c.tm = time.Time{}
		cacheLock.Unlock()
	}
}

func DelCache(uri string) {

	cacheLock.Lock()
	conn, ok := cachePool[uri]
	cacheLock.Unlock()
	if !ok {
		log.Println("delCache: NotExists", uri)
		return
	}

	conn.Close()

	cacheLock.Lock()
	delete(cachePool, uri)
	cacheLock.Unlock()
	log.Println("delCache", uri)
}

func CacheQueue(uri string) (conn *cacheConn) {
	var ok bool

	cacheLock.Lock()
	conn, ok = cachePool[uri]
	cacheLock.Unlock()
	if ok {
		return
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println("cacheQueue: uri invalid", uri)
		return
	}
	conn = &cacheConn{
		uri: uri,
		buf: fifo.NewBuffer(),
		req: req,
		tm: time.Now(),
	}

	cacheLock.Lock()
	cachePool[uri] = conn
	cacheLock.Unlock()

	log.Println("cacheQueue:", uri)
	cacheWait <- 1
	return
}

func PlayFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Println("Open:", err)
		return
	}
	err = mad.Play(f)
	if err != nil {
		log.Println("play:", err)
	}
}

func Play(uri string) {
	log.Println("play:", uri)
	conn := CacheQueue(uri)
	conn.buf.ResetRead()

	cacheLock.Lock()
	curPlayConn = conn
	cacheLock.Unlock()

	err := mad.Play(conn.buf)
	if err != nil {
		log.Println("playErr:", err)
	}

	log.Println("playEnd:", conn.uri)
}

func Resume() {
}

func Stop() {
	mad.StopPlay()
}

func Pause() {
}

