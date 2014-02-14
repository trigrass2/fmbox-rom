
package mad

import (
	"testing"
	"io"
	"log"
	"os"
)

func TestMad(t *testing.T) {
	f, _ := os.Open("/var/www/test.mp3")
	f2, _ := os.Create("/tmp/out.raw")
	rate, raw, err := Decode(f)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("rate:", rate)
	_, err = io.Copy(f2, raw)
	f.Close()
	f2.Close()
	if err != nil {
		log.Println(err)
	}
}

