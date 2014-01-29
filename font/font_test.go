
package font

import (
	"testing"
	"os"
	"image/png"
)

func TestFont(t *testing.T) {
	img := AlbumImg("哈哈", "The XX", "33:44")
	f, _ := os.Create("/var/www/pic.png")
	png.Encode(f, img)
	f.Close()
}

