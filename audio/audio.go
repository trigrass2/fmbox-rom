
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

	"github.com/go-av/wget"
	"gitcafe.com/nuomi-studio/fifo.git"
	"gitcafe.com/nuomi-studio/fmbox-rom.git/client-mad"
)

type cacheConn struct {
	req *wget.Request
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
	c.buf.Close()
	c.req.Close()
}

func doCache(conn *cacheConn) (err error) {
	log.Println("doCache: starts", conn.uri)
	log.Println("doCache: pool has", len(cachePool), "entries")

	var out io.Reader
	if out, err = conn.req.GetReader(); err != nil {
		log.Println("doCache: wget failed:", err)
		conn.Close()
		return
	}

	var n int64
	log.Println("doCache: startIo", conn.uri)
	n, _ = io.Copy(conn.buf, out)

	log.Println("doCache: done", conn.uri, err, n/1024, "KiB")
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

func DelAllCache() {
	log.Println("cache:", "del all")
	cacheLock.Lock()
	for _, conn := range cachePool {
		conn.Close()
	}
	cachePool = map[string]*cacheConn{}
	cacheLock.Unlock()
}

func DelCache(uri string) {
	cacheLock.Lock()
	if conn, ok := cachePool[uri]; !ok {
		log.Println("cache:", "del", uri)
	} else {
		conn.Close()
	}
	delete(cachePool, uri)
	cacheLock.Unlock()
}

func CacheQueue(uri string) (conn *cacheConn) {
	var ok bool

	cacheLock.Lock()
	conn, ok = cachePool[uri]
	cacheLock.Unlock()
	if ok {
		return
	}

	conn = &cacheConn{
		uri: uri,
		buf: fifo.NewBuffer(),
		req: wget.NewRequest(uri, 10),
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

func Stop() {
	mad.StopPlay()
}

