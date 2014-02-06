
package mad

import (
	"testing"
	"os"
)

func TestMad(t *testing.T) {
	dec := NewDecoder()
	f, _ := os.Open("/var/www/test.mp3")
	f2, _ := os.Create("/tmp/out.raw")
	dec.R = f
	dec.W = f2
	dec.Run()
	f.Close()
	f2.Close()
}

