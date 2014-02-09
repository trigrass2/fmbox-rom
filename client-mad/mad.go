
package mad

import (
	"log"
	"io"
	"net"
)

type Decoder struct {
	R io.Reader
	W io.Writer
}

func (d *Decoder) Run() error {
	c, err := net.Dial("tcp", "localhost:91")
	if err != nil {
		log.Println("mad: dial", err)
		return err
	}
	end := make(chan int, 0)
	go func () {
		b := make([]byte, 16)
		c.Read(b)
		d.W.Write(b)
		log.Println("mad: data comes")
		io.Copy(d.W, c)
		end <- 1
	}()
	_, err = io.Copy(c, d.R)
	c.(*net.TCPConn).CloseWrite()
	<-end
	c.Close()
	return err
}

func NewDecoder() *Decoder {
	return &Decoder{}
}
