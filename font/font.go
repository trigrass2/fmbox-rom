
package font

import (
	"log"
	"io/ioutil"
	"github.com/Bitnick2002/freetype-go/freetype"
	"github.com/Bitnick2002/freetype-go/freetype/truetype"
	"image/draw"
	"image"
)

func loadFont(file string) *truetype.Font {
	fontFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	font, err2 := freetype.ParseFont(fontFile)
	if err2 != nil {
		log.Fatal(err2)
	}
	return font
}

func AlbumImg(s1, s2, s3 string) (img *image.Gray) {
	img = image.NewGray(image.Rect(0, 0, 172, 72))
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)

	font1 := loadFont("fangzheng.TTF")
	font2 := loadFont("Carre1.ttf")

	c := freetype.NewContext()
	c.SetDPI(92)
	c.SetClip(img.Bounds())
	c.SetSrc(image.Black)
	c.SetDst(img)

	draw := func (x,y int, s string, font *truetype.Font, fontSize float64) {
		pt := freetype.Pt(x, y+int(c.PointToFix32(fontSize)>>8))
		c.SetFont(font)
		c.SetFontSize(fontSize)
		c.DrawString(s, pt)
	}

	draw(0, 0, s1, font1, 18.5)
	draw(0, 25, s2, font1, 13.75)
	draw(130, 53, s3, font2, 12.0)

	filter2 := func () {
		for i := range img.Pix {
			b := img.Pix[i]
			if b <= 192 {
				b = 0
			} else {
				b = 255
			}
			img.Pix[i] = b
		}
	}

	filter := func () {
		for i := range img.Pix {
			b := img.Pix[i]
			if b >= 192 {
				b = 255
			} else if b >= 128 {
				b = 192
			} else if b >= 64 {
				b = 128
			} else {
				b = 64
			}
			img.Pix[i] = b
		}
	}

	if false { filter() }
	if false { filter2() }

	return
}

func Img2Pic(img *image.Gray) (pic [3096]byte) {
	for y := 0; y < 72; y++ {
		for x := 0; x < 172; x++ {
			b := img.Pix[y*172+x]
			b >>= 6
			i := x*72+y
			j := 3-i%4
			pic[i/4] |= b << byte(j*2)
		}
	}
	return
}

func TestPic() (pic [3096]byte) {
	return Img2Pic(AlbumImg("heh", "haha", "333"))
}

